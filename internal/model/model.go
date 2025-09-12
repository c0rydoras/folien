package model

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/c0rydoras/folien/internal/navigation"
	"github.com/c0rydoras/folien/internal/preprocessor"
	"github.com/c0rydoras/folien/pkg/parser"
	"github.com/c0rydoras/folien/pkg/util"

	"github.com/c0rydoras/folien/internal/code"
	"github.com/c0rydoras/folien/internal/meta"
	"github.com/c0rydoras/folien/styles"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

var (
	tabSpaces = strings.Repeat(" ", 4)
)

const (
	delimiter = "\n---\n"
)

type HideInternalError int

const (
	// Hide All Internal Errors
	All HideInternalError = iota
	// Hide all Internal Errors except for the last one (on the current slide)
	AllButLast
	// Don't hide any Internal Errors
	None
)

// Model represents the model of this presentation, which contains all the
// state related to the current folien.
type Model struct {
	Slides   []string
	Page     int
	Author   string
	Date     string
	Theme    glamour.TermRendererOption
	Paging   string
	FileName string
	viewport viewport.Model
	buffer   string
	// VirtualText is used for additional information that is not part of the
	// original folien, it will be displayed on a slide and reset on page change
	VirtualText  string
	Search       navigation.Search
	Preprocessor *preprocessor.Config
	// TODO: move into some proper config struct
	HideInternalErrors HideInternalError
	AllowExecution     bool
	ready              bool
}

type fileWatchMsg struct{}

var fileInfo os.FileInfo

// Init initializes the model and begins watching the folien file for changes
// if it exists.
func (m Model) Init() tea.Cmd {
	if m.FileName == "" {
		return nil
	}
	fileInfo, _ = os.Stat(m.FileName)
	return fileWatchCmd()
}

func fileWatchCmd() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return fileWatchMsg{}
	})
}

// Load loads all of the content and metadata for the presentation.
func (m *Model) Load() error {
	var content string
	var err error

	if m.FileName != "" && m.FileName != "-" {
		content, err = util.ReadFile(m.FileName)
	} else {
		content, err = readStdin()
	}

	if err != nil {
		return err
	}

	content = strings.ReplaceAll(content, "\r", "")
	metaData, exists := meta.New().Parse(content)

	if exists {
		content = parser.RemoveFrontMatter(content)
	}
	folien := strings.Split(content, delimiter)

	m.Slides = folien

	if m.Preprocessor != nil {
		m.Slides = m.Preprocessor.Process(folien)
	}

	m.Author = metaData.Author
	m.Date = metaData.Date
	m.Paging = metaData.Paging
	if m.Theme == nil {
		m.Theme = styles.SelectTheme(metaData.Theme)
	}

	m.updateViewportContent()

	return nil
}

// Update updates the presentation model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		footerHeight := 3
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-footerHeight)
			m.viewport.YPosition = 0
			m.ready = true
			m.updateViewportContent()
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - footerHeight
		}
		return m, nil

	case tea.KeyMsg:
		keyPress := msg.String()

		if m.Search.Active {
			switch msg.Type {
			case tea.KeyEnter:
				// execute current buffer
				if m.Search.Query() != "" {
					m.Search.Execute(&m)
				} else {
					m.Search.Done()
				}
				// cancel search
				return m, nil
			case tea.KeyCtrlC, tea.KeyEscape:
				// quit command mode
				m.Search.SetQuery("")
				m.Search.Done()
				return m, nil
			}

			var cmd tea.Cmd
			m.Search.SearchTextInput, cmd = m.Search.SearchTextInput.Update(msg)
			return m, cmd
		}

		switch keyPress {
		case "/":
			// Begin search
			m.Search.Begin()
			m.Search.SearchTextInput.Focus()
			return m, nil
		case "ctrl+n":
			// Go to next occurrence
			m.Search.Execute(&m)
		case "ctrl+e":
			// Run code blocks
			blocks, err := code.Parse(m.Slides[m.Page])
			if err != nil {
				// We couldn't parse the code block on the screen
				m.VirtualText = "\n" + err.Error()
				m.updateViewportContent()
				return m, nil
			}
			if !m.AllowExecution {
				m.VirtualText = "\n Execution is disabled"
				m.updateViewportContent()
				return m, nil
			}
			var outs []string
			for i, block := range blocks {
				res := code.Execute(block)
				if res.ExitCode == code.ExitCodeInternalError {
					if m.HideInternalErrors == All {
						continue
					}
					if m.HideInternalErrors == AllButLast && i != len(blocks)-1 {
						continue
					}
				}
				outs = append(outs, res.Out)
			}
			m.VirtualText = strings.Join(outs, "\n")
			m.updateViewportContent()
		case "y":
			blocks, err := code.Parse(m.Slides[m.Page])
			if err != nil {
				return m, nil
			}
			for _, b := range blocks {
				_ = clipboard.WriteAll(b.Code)
			}
			return m, nil
		case "ctrl+c", "q":
			return m, tea.Quit
		default:
			if m.shouldHandleViewportNavigation(keyPress) {
				if keyPress == "j" || keyPress == "k" {
					repeat := 1
					if m.bufferIsNumeric() {
						if r, err := strconv.Atoi(m.buffer); err == nil && r > 0 {
							repeat = r
						}
						m.buffer = ""
					}

					for i := 0; i < repeat; i++ {
						m.viewport, cmd = m.viewport.Update(msg)
						cmds = append(cmds, cmd)
					}
					return m, tea.Batch(cmds...)
				} else {
					m.viewport, cmd = m.viewport.Update(msg)
					cmds = append(cmds, cmd)
					return m, tea.Batch(cmds...)
				}
			}

			newState := navigation.Navigate(navigation.State{
				Buffer:      m.buffer,
				Page:        m.Page,
				TotalSlides: len(m.Slides),
			}, keyPress)
			m.buffer = newState.Buffer
			if newState.Page != m.Page {
				m.SetPage(newState.Page)
				m.updateViewportContent()
				m.viewport.GotoTop()
			}
		}

	case fileWatchMsg:
		newFileInfo, err := os.Stat(m.FileName)
		if err == nil && newFileInfo.ModTime() != fileInfo.ModTime() {
			fileInfo = newFileInfo
			_ = m.Load()
			if m.Page >= len(m.Slides) {
				m.Page = len(m.Slides) - 1
			}
			m.updateViewportContent()
		}
		return m, fileWatchCmd()
	}

	if !m.Search.Active {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the current slide in the presentation and the status bar which
// contains the author, date, and pagination information.
func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	slide := styles.Slide.Render(m.viewport.View())

	var left string
	if m.Search.Active {
		// render search bar
		left = m.Search.SearchTextInput.View()
	} else {
		// render author and date
		left = styles.Author.Render(m.Author) + styles.Date.Render(m.Date)
	}

	right := styles.Page.Render(m.paging())
	status := styles.Status.Render(styles.JoinHorizontal(left, right, m.viewport.Width))

	return fmt.Sprintf("%s\n%s", slide, status)
}

func (m *Model) shouldHandleViewportNavigation(keyPress string) bool {
	scrollKeys := map[string]bool{
		"up":     true,
		"down":   true,
		"pgup":   true,
		"pgdown": true,
		"home":   true,
		"end":    true,
		"ctrl+u": true,
		"ctrl+d": true,
		"ctrl+b": true,
		"ctrl+f": true,
	}

	if keyPress == "j" || keyPress == "k" {
		return true
	}

	return scrollKeys[keyPress]
}

func (m *Model) updateViewportContent() {
	if !m.ready || len(m.Slides) == 0 {
		return
	}

	r, _ := glamour.NewTermRenderer(m.Theme, glamour.WithWordWrap(m.viewport.Width))
	slide := m.Slides[m.Page]
	slide = code.HideComments(slide)
	slide, err := r.Render(slide)
	slide = strings.ReplaceAll(slide, "\t", tabSpaces)
	slide += m.VirtualText
	if err != nil {
		slide = fmt.Sprintf("Error: Could not render markdown! (%v)", err)
	}

	slide = "\n\n" + slide

	m.viewport.SetContent(slide)
}

func (m *Model) bufferIsNumeric() bool {
	if m.buffer == "" {
		return false
	}
	for _, r := range m.buffer {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func (m *Model) paging() string {
	switch strings.Count(m.Paging, "%d") {
	case 2:
		return fmt.Sprintf(m.Paging, m.Page+1, len(m.Slides))
	case 1:
		return fmt.Sprintf(m.Paging, m.Page+1)
	default:
		return m.Paging
	}
}

func readStdin() (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	if stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0 {
		return "", errors.New("no input provided")
	}

	reader := bufio.NewReader(os.Stdin)
	var b strings.Builder

	for {
		r, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}
		_, err = b.WriteRune(r)
		if err != nil {
			return "", err
		}
	}

	return b.String(), nil
}

// CurrentPage returns the current page the presentation is on.
func (m *Model) CurrentPage() int {
	return m.Page
}

// SetPage sets which page the presentation should render.
func (m *Model) SetPage(page int) {
	if m.Page == page {
		return
	}

	m.VirtualText = ""
	m.Page = page
	m.updateViewportContent()
}

// Pages returns all the folien in the presentation.
func (m *Model) Pages() []string {
	return m.Slides
}
