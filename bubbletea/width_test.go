package bubbletea_test

import (
	"testing"

	"github.com/fwojciec/diffstory/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestDisplayWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "simple text",
			input:    "hello",
			expected: 5,
		},
		{
			name:     "single tab at start",
			input:    "\t",
			expected: 8, // tab expands to column 8
		},
		{
			name:     "tab after one char",
			input:    "a\t",
			expected: 8, // 'a' at col 0, tab expands to col 8
		},
		{
			name:     "tab after seven chars",
			input:    "1234567\t",
			expected: 8, // 7 chars, tab expands to col 8
		},
		{
			name:     "tab after eight chars",
			input:    "12345678\t",
			expected: 16, // 8 chars, tab expands to col 16
		},
		{
			name:     "multiple tabs",
			input:    "\t\t",
			expected: 16, // first tab to 8, second tab to 16
		},
		{
			name:     "mixed content with tabs",
			input:    "abc\tdef",
			expected: 11, // 'abc' (3), tab to 8, 'def' (3) = 11
		},
		{
			name:     "typescript style indentation",
			input:    "\t\tconst x = 1;",
			expected: 28, // two tabs (16) + 12 chars
		},
		{
			name:     "unicode with tabs",
			input:    "日本\t語",
			expected: 10, // 2-width chars (4) + tab to 8 + 2-width char (2) = 10
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := bubbletea.DisplayWidth(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
