package parser_test

import (
	"testing"

	"github.com/c0rydoras/folien/pkg/parser"
	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	Title string `yaml:"title"`
	Theme string `yaml:"theme"`
	Count int    `yaml:"count"`
}

func TestUnmarshalFrontMatter_ValidYAML(t *testing.T) {
	source := []byte(`---
title: "Test Presentation"
theme: "dark"
count: 42
---

# Content here`)

	result, err := parser.UnmarshalFrontMatter[TestConfig](source)

	assert.NoError(t, err)
	assert.Equal(t, "Test Presentation", result.Title)
	assert.Equal(t, "dark", result.Theme)
	assert.Equal(t, 42, result.Count)
}

func TestUnmarshalFrontMatter_NoFrontMatter(t *testing.T) {
	source := []byte(`# Just content without frontmatter`)

	result, err := parser.UnmarshalFrontMatter[TestConfig](source)

	assert.NoError(t, err)
	assert.Equal(t, TestConfig{}, result)
}

func TestUnmarshalFrontMatter_ValidTOML(t *testing.T) {
	source := []byte(`+++
title = "Test Presentation"
theme = "light"
count = 100
+++`)

	result, err := parser.UnmarshalFrontMatter[TestConfig](source)

	assert.NoError(t, err)
	assert.Equal(t, "Test Presentation", result.Title)
	assert.Equal(t, "light", result.Theme)
	assert.Equal(t, 100, result.Count)
}

func TestRemoveFrontMatter_WithFrontMatter(t *testing.T) {
	source := `---
title: "Test Presentation"
theme: "dark"
---

# Main Content

This is the actual content of the document.`

	result := parser.RemoveFrontMatter(source)

	assert.Contains(t, result, "# Main Content")
	assert.Contains(t, result, "This is the actual content")
	assert.NotContains(t, result, "title:")
	assert.NotContains(t, result, "theme:")
}

func TestRemoveFrontMatter_WithFrontMatterTOML(t *testing.T) {
	source := `+++
title = 320
+++

# Main Content

This is the actual content of the document.`

	result := parser.RemoveFrontMatter(source)

	assert.Contains(t, result, "# Main Content")
	assert.Contains(t, result, "This is the actual content")
	assert.NotContains(t, result, "title:")
	assert.NotContains(t, result, "theme:")
}

func TestRemoveFrontMatter_NoFrontMatter(t *testing.T) {
	source := `# Just Content

This document has no frontmatter, just regular markdown content.`

	result := parser.RemoveFrontMatter(source)

	assert.Contains(t, result, "# Just Content")
	assert.Contains(t, result, "This document has no frontmatter")
}
