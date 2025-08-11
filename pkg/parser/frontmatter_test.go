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
