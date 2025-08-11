package parser

import (
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
