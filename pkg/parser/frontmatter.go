package parser

import (
	"regexp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
)

func UnmarshalFrontMatter[T any](source []byte) (T, error) {
	var zero T

	md := goldmark.New(
		goldmark.WithExtensions(&frontmatter.Extender{}),
	)

	ctx := parser.NewContext()
	_ = md.Parser().Parse(text.NewReader(source), parser.WithContext(ctx))

	d := frontmatter.Get(ctx)
	if d == nil {
		return zero, nil
	}

	var data T
	if err := d.Decode(&data); err != nil {
		return zero, err
	}

	return data, nil
}

var frontMatterRegex = regexp.MustCompile(`(?s)^([-+]{3})\n(.*?\n)([-+]{3})\n`)

func RemoveFrontMatter(source string) string {
	matches := frontMatterRegex.FindStringSubmatch(source)
	if len(matches) == 4 && matches[1] == matches[3] {
		return source[len(matches[0]):]
	}
	return source
}
