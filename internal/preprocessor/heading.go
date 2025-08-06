package preprocessor

import (
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func collectHeadings(slideContent []byte) map[int]string {
	parser := goldmark.DefaultParser()
	reader := text.NewReader(slideContent)
	doc := parser.Parse(reader)

	latestHeadings := make(map[int]string)

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if h, ok := n.(*ast.Heading); ok {
			headingText := string(h.Lines().Value(slideContent))
			latestHeadings[h.Level] = headingText
		}
		return ast.WalkContinue, nil
	})

	if err != nil {
		return make(map[int]string)
	}
	return latestHeadings

}

func AddHeadings(folien []string, maxLevel int) []string {
	if len(folien) == 0 {
		return folien
	}

	var newSlides []string

	acc := make(map[int]string)

	for _, slideContent := range folien {
		currentHeadings := collectHeadings([]byte(slideContent))

		headingsToAdd := make(map[int]string)
		for level := 1; level <= maxLevel; level++ {
			if newHeading, ok := currentHeadings[level]; ok {
				acc[level] = newHeading
				for j := level + 1; j <= maxLevel; j++ {
					delete(acc, j)
				}
			} else {
				if inheritedHeading, ok := acc[level]; ok {
					headingsToAdd[level] = inheritedHeading
				}
			}
		}

		var prefix strings.Builder
		for level := 1; level <= maxLevel; level++ {
			if headingText, ok := headingsToAdd[level]; ok {
				line := fmt.Sprintf("%s %s\n", strings.Repeat("#", level), headingText)
				prefix.WriteString(line)
			}
		}

		newSlide := prefix.String() + slideContent
		newSlides = append(newSlides, newSlide)
	}

	return newSlides
}
