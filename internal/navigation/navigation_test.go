package navigation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNavigation(t *testing.T) {
	tests := []struct {
		keys   string
		target int
	}{
		{target: 0},
		{keys: "l", target: 1},
		{keys: "jjjjjjjjjj", target: 0},
		{keys: "jjjjjjjjjjjjj", target: 0},
		{keys: "G", target: 10},
		{keys: "llgg", target: 0},
		{keys: "2jll", target: 2},
		{keys: "0jl", target: 1},
		{keys: "-11G", target: 10},
		{keys: "0G", target: 0},
		{keys: "3G", target: 2},
		{keys: "11G", target: 10},
		{keys: "101G", target: 10},
		{keys: "nnN", target: 1},
	}

	for _, tt := range tests {
		t.Run(tt.keys, func(t *testing.T) {
			currentState := State{
				Buffer:      "",
				Page:        0,
				TotalSlides: 11,
			}

			for _, key := range strings.Split(tt.keys, "") {
				currentState = Navigate(currentState, key)
			}

			targetState := State{Page: tt.target, TotalSlides: 11}
			assert.Equal(t, targetState, currentState)
		})
	}
}
