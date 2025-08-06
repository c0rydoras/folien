package preprocessor

import (
	"reflect"
	"testing"
)

func TestCollectHeadings(t *testing.T) {
	tests := []struct {
		name     string
		slide    string
		expected map[int]string
	}{
		{
			name:     "empty slide",
			slide:    "",
			expected: map[int]string{},
		},
		{
			name:     "no headings",
			slide:    "no headings",
			expected: map[int]string{},
		},
		{
			name:  "single h1",
			slide: `# h1`,
			expected: map[int]string{
				1: "h1",
			},
		},
		{
			name: "multiple heading levels",
			slide: `# h1
## h2
### h3
`,
			expected: map[int]string{
				1: "h1",
				2: "h2",
				3: "h3",
			},
		},
		{
			name: "multiple headings at same level",
			slide: `
# nono
## h2
# h1`,
			expected: map[int]string{
				1: "h1",
				2: "h2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectHeadings([]byte(tt.slide))
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("collectHeadings() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAddHeadings(t *testing.T) {
	folien := []string{
		"# Slide 1\nContent of slide 1",
		"## Slide 2\nContent of slide 2",
		"Content of slide 3",
		"# Slide 4\ncontent of slide 4",
		"## Slide 5\ncontent of slide 5",
	}

	expected := []string{
		"# Slide 1\nContent of slide 1",
		"# Slide 1\n## Slide 2\nContent of slide 2",
		"# Slide 1\n## Slide 2\nContent of slide 3",
		"# Slide 4\ncontent of slide 4",
		"# Slide 4\n## Slide 5\ncontent of slide 5",
	}

	result := AddHeadings(folien, 2)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("AddHeadings() = %v, want %v", result, expected)
	}
}
