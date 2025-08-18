package model

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
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

	return nil
}

// Update updates the presentation model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
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
				return m, nil
			}
			var outs []string
			for _, block := range blocks {
				res := code.Execute(block)
				outs = append(outs, res.Out)
			}
			m.VirtualText = strings.Join(outs, "\n")
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
			newState := navigation.Navigate(navigation.State{
				Buffer:      m.buffer,
				Page:        m.Page,
				TotalSlides: len(m.Slides),
			}, keyPress)
			m.buffer = newState.Buffer
			m.SetPage(newState.Page)
		}

	case fileWatchMsg:
		newFileInfo, err := os.Stat(m.FileName)
		if err == nil && newFileInfo.ModTime() != fileInfo.ModTime() {
			fileInfo = newFileInfo
			_ = m.Load()
			if m.Page >= len(m.Slides) {
				m.Page = len(m.Slides) - 1
			}
		}
		return m, fileWatchCmd()
	}
	return m, nil
}

// View renders the current slide in the presentation and the status bar which
// contains the author, date, and pagination information.
func (m Model) View() string {
	r, _ := glamour.NewTermRenderer(m.Theme, glamour.WithWordWrap(m.viewport.Width))
	slide := m.Slides[m.Page]
	slide = code.HideComments(slide)
	slide, err := r.Render(slide)
	slide = strings.ReplaceAll(slide, "\t", tabSpaces)
	slide += m.VirtualText
	if err != nil {
		slide = fmt.Sprintf("Error: Could not render markdown! (%v)", err)
	}
	slide = styles.Slide.Render(slide)

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
	return styles.JoinVertical(slide, status, m.viewport.Height)
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
}

// Pages returns all the folien in the presentation.
func (m *Model) Pages() []string {
	return m.Slides
}
