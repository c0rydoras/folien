package meta_test

import (
	"fmt"
	"os/user"
	"testing"
	"time"

	"github.com/c0rydoras/folien/internal/meta"
	"github.com/stretchr/testify/assert"
)

func TestMeta_ParseHeader(t *testing.T) {
	user, _ := user.Current()
	date := time.Now().Format("2006-01-02")

	tests := []struct {
		name      string
		folienhow string
		want      *meta.Meta
	}{
		{
			name:      "Parse theme from header",
			folienhow: fmt.Sprintf("---\ntheme: %q\n", "dark"),
			want: &meta.Meta{
				Theme:  "dark",
				Author: user.Name,
				Date:   date,
				Paging: "Slide %d / %d",
			},
		},
		{
			name:      "Fallback to default if no theme provided",
			folienhow: "\n# Header Slide\n > Subtitle\n",
			want: &meta.Meta{
				Theme:  "default",
				Author: user.Name,
				Date:   date,
				Paging: "Slide %d / %d",
			},
		},
		{
			name:      "Parse author from header",
			folienhow: fmt.Sprintf("---\nauthor: %q\n", "gopher"),
			want: &meta.Meta{
				Theme:  "default",
				Author: "gopher",
				Date:   date,
				Paging: "Slide %d / %d",
			},
		},
		{
			name:      "Fallback to default if no author provided",
			folienhow: "\n# Header Slide\n > Subtitle\n",
			want: &meta.Meta{
				Theme:  "default",
				Author: user.Name,
				Date:   date,
				Paging: "Slide %d / %d",
			},
		},
		{
			name:      "Parse static date from header",
			folienhow: fmt.Sprintf("---\ndate: %q\n", "31/01/1970"),
			want: &meta.Meta{
				Theme:  "default",
				Author: user.Name,
				Date:   "31/01/1970",
				Paging: "Slide %d / %d",
			},
		},
		{
			name:      "Parse go-styled date from header",
			folienhow: fmt.Sprintf("---\ndate: %q\n", "MMM dd, YYYY"),
			want: &meta.Meta{
				Theme:  "default",
				Author: user.Name,
				Date:   time.Now().Format("Jan 2, 2006"),
				Paging: "Slide %d / %d",
			},
		},
		{
			name:      "Parse YYYY-MM-DD date from header",
			folienhow: fmt.Sprintf("---\ndate: %q\n", "YYYY-MM-DD"),
			want: &meta.Meta{
				Theme:  "default",
				Author: user.Name,
				Date:   time.Now().Format("2006-01-02"),
				Paging: "Slide %d / %d",
			},
		},
		{
			name:      "Parse dd/mm/YY date from header",
			folienhow: fmt.Sprintf("---\ndate: %q\n", "dd/mm/YY"),
			want: &meta.Meta{
				Theme:  "default",
				Author: user.Name,
				Date:   time.Now().Format("2/1/06"),
				Paging: "Slide %d / %d",
			},
		},
		{
			name:      "Parse MMM dd, YYYY date from header",
			folienhow: fmt.Sprintf("---\ndate: %q\n", "MMM dd, YYYY"),
			want: &meta.Meta{
				Theme:  "default",
				Author: user.Name,
				Date:   time.Now().Format("Jan 2, 2006"),
				Paging: "Slide %d / %d",
			},
		},
		{
			name:      "Parse MMMM DD, YYYY date from header",
			folienhow: fmt.Sprintf("---\ndate: %q\n", "MMMM DD, YYYY"),
			want: &meta.Meta{
				Theme:  "default",
				Author: user.Name,
				Date:   time.Now().Format("January 02, 2006"),
				Paging: "Slide %d / %d",
			},
		},
		{
			name:      "Fallback to default if no date provided",
			folienhow: "\n# Header Slide\n > Subtitle\n",
			want: &meta.Meta{
				Theme:  "default",
				Author: user.Name,
				Date:   date,
				Paging: "Slide %d / %d",
			},
		},
		{
			name:      "Parse paging from header",
			folienhow: fmt.Sprintf("---\npaging: %q\n", "%d of %d"),
			want: &meta.Meta{
				Theme:  "default",
				Author: user.Name,
				Date:   date,
				Paging: "%d of %d",
			},
		},
		{
			name:      "Fallback to default if no numebring provided",
			folienhow: "\n# Header Slide\n > Subtitle\n",
			want: &meta.Meta{
				Theme:  "default",
				Author: user.Name,
				Date:   date,
				Paging: "Slide %d / %d",
			},
		},
		{
			name:      "Fallback if first slide is valid yaml",
			folienhow: "---\n# Header Slide---\nContent\n",
			want: &meta.Meta{
				Theme:  "default",
				Author: user.Name,
				Date:   date,
				Paging: "Slide %d / %d",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &meta.Meta{}
			got, hasMeta := m.Parse(tt.folienhow)
			if !hasMeta {
				assert.NotNil(t, got)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want *meta.Meta
	}{
		{name: "Create meta struct", want: &meta.Meta{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, meta.New(), tt.want)
		})
	}
}

func ExampleMeta_Parse() {
	header := `
---
theme: "dark"
author: "Gopher"
date: "Apr. 4, 2021"
paging: "%d"
---
`
	// Parse the header from the markdown
	// file
	m, _ := meta.New().Parse(header)

	// Print the return theme
	// meta
	fmt.Println(m.Theme)
	fmt.Println(m.Author)
	fmt.Println(m.Date)
	fmt.Println(m.Paging)
}
