package bubbletea_test

import (
	"bytes"
	"io"

	diffview "github.com/fwojciec/diffstory"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// trueColorRenderer creates a lipgloss renderer that outputs true colors.
// This is useful for testing color output without affecting global state.
func trueColorRenderer() *lipgloss.Renderer {
	r := lipgloss.NewRenderer(io.Discard)
	r.SetColorProfile(termenv.TrueColor)
	return r
}

// extractLastLine returns the last non-empty line from the output.
func extractLastLine(s string) string {
	lines := bytes.Split([]byte(s), []byte("\n"))
	for i := len(lines) - 1; i >= 0; i-- {
		line := bytes.TrimSpace(lines[i])
		if len(line) > 0 {
			return string(lines[i])
		}
	}
	return ""
}

// mockTokenizer implements diffview.Tokenizer for testing.
type mockTokenizer struct {
	TokenizeFn func(language, source string) []diffview.Token
}

func (m *mockTokenizer) Tokenize(language, source string) []diffview.Token {
	return m.TokenizeFn(language, source)
}

// mockLanguageDetector implements diffview.LanguageDetector for testing.
type mockLanguageDetector struct {
	DetectFromPathFn func(path string) string
}

func (m *mockLanguageDetector) DetectFromPath(path string) string {
	return m.DetectFromPathFn(path)
}

// mockWordDiffer implements diffview.WordDiffer for testing.
type mockWordDiffer struct {
	DiffFn func(old, new string) (oldSegs, newSegs []diffview.Segment)
}

func (m *mockWordDiffer) Diff(old, new string) (oldSegs, newSegs []diffview.Segment) {
	return m.DiffFn(old, new)
}
