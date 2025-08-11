package parser

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func CollectCodeBlocks(source []byte) []*ast.FencedCodeBlock {
	parser := goldmark.DefaultParser()
	reader := text.NewReader(source)
	doc := parser.Parse(reader)

	codeBlocks := []*ast.FencedCodeBlock{}

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if c, ok := n.(*ast.FencedCodeBlock); ok {
			codeBlocks = append(codeBlocks, c)
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return []*ast.FencedCodeBlock{}
	}
	return codeBlocks
}
