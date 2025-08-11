// Package meta implements markdown frontmatter parsing for simple
// folien configuration
package meta

import (
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/c0rydoras/folien/pkg/parser"
)

// Meta contains all of the data to be parsed
// out of a markdown file's header section
type Meta struct {
	Theme  string `yaml:"theme"`
	Author string `yaml:"author"`
	Date   string `yaml:"date"`
	Paging string `yaml:"paging"`
}

// New creates a new instance of the
// slideshow meta header object
func New() *Meta {
	return &Meta{}
}

// Parse parses metadata from a slideshows header slide
// including theme information
//
// If no front matter is provided, it will fallback to the default theme and
// return false to acknowledge that there is no front matter in this slide
func (m *Meta) Parse(presentation string) (*Meta, bool) {
	fallback := &Meta{
		Theme:  defaultTheme(),
		Author: defaultAuthor(),
		Date:   defaultDate(),
		Paging: defaultPaging(),
	}

	tmp, err := parser.UnmarshalFrontMatter[Meta]([]byte(presentation))
	if err != nil {
		return fallback, false
	}

	// If all fields are empty, assume no frontmatter was found
	if tmp.Theme == "" && tmp.Author == "" && tmp.Date == "" && tmp.Paging == "" {
		return fallback, false
	}

	if tmp.Theme != "" {
		m.Theme = tmp.Theme
	} else {
		m.Theme = fallback.Theme
	}

	if tmp.Author != "" {
		m.Author = tmp.Author
	} else {
		m.Author = fallback.Author
	}

	if tmp.Date != "" {
		parsedDate := parseDate(tmp.Date)
		if parsedDate == tmp.Date {
			m.Date = tmp.Date
		} else {
			m.Date = time.Now().Format(parsedDate)
		}
	} else {
		m.Date = fallback.Date
	}

	if tmp.Paging != "" {
		m.Paging = tmp.Paging
	} else {
		m.Paging = fallback.Paging
	}

	return m, true
}

func defaultTheme() string {
	theme := os.Getenv("GLAMOUR_STYLE")
	if theme == "" {
		return "default"
	}
	return theme
}

func defaultAuthor() string {
	user, err := user.Current()
	if err != nil {
		return ""
	}

	return user.Name
}

func defaultDate() string {
	return time.Now().Format(parseDate("YYYY-MM-DD"))
}

func defaultPaging() string {
	return "Slide %d / %d"
}

func parseDate(value string) string {
	pairs := [][]string{
		{"YYYY", "2006"},
		{"YY", "06"},
		{"MMMM", "January"},
		{"MMM", "Jan"},
		{"MM", "01"},
		{"mm", "1"},
		{"DD", "02"},
		{"dd", "2"},
	}

	for _, p := range pairs {
		value = strings.ReplaceAll(value, p[0], p[1])
	}
	return value
}
