package chroma_test

import (
	"testing"

	chromalib "github.com/alecthomas/chroma/v2"
	"github.com/fwojciec/diffstory"
	"github.com/fwojciec/diffstory/chroma"
	"github.com/fwojciec/diffstory/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testStyleFunc returns a style function using the test palette.
func testStyleFunc() func(chromalib.TokenType) diffview.Style {
	return chroma.StyleFromPalette(lipgloss.TestTheme().Palette())
}

func TestTokenizer_Tokenize(t *testing.T) {
	t.Parallel()

	t.Run("tokenizes Go code", func(t *testing.T) {
		t.Parallel()

		tokenizer, err := chroma.NewTokenizer(testStyleFunc())
		require.NoError(t, err)
		tokens := tokenizer.Tokenize("go", `package main`)

		require.NotEmpty(t, tokens, "expected tokens for valid Go code")

		// Reconstruct the source from tokens
		var reconstructed string
		for _, tok := range tokens {
			reconstructed += tok.Text
		}
		assert.Equal(t, "package main", reconstructed)

		// Check that keyword "package" gets a style
		var foundPackageKeyword bool
		for _, tok := range tokens {
			if tok.Text == "package" {
				foundPackageKeyword = true
				assert.NotEmpty(t, tok.Style.Foreground, "keyword should have foreground color")
			}
		}
		assert.True(t, foundPackageKeyword, "should find 'package' keyword token")
	})

	t.Run("returns nil for unsupported language", func(t *testing.T) {
		t.Parallel()

		tokenizer, err := chroma.NewTokenizer(testStyleFunc())
		require.NoError(t, err)
		tokens := tokenizer.Tokenize("nonexistent-language-xyz", "some code")

		assert.Nil(t, tokens)
	})

	t.Run("handles empty source", func(t *testing.T) {
		t.Parallel()

		tokenizer, err := chroma.NewTokenizer(testStyleFunc())
		require.NoError(t, err)
		tokens := tokenizer.Tokenize("go", "")

		assert.Empty(t, tokens)
	})

	t.Run("styles function names", func(t *testing.T) {
		t.Parallel()

		tokenizer, err := chroma.NewTokenizer(testStyleFunc())
		require.NoError(t, err)
		// Code with a function definition
		tokens := tokenizer.Tokenize("go", `func foo() {}`)

		require.NotEmpty(t, tokens)

		var fooStyle diffview.Style
		for _, tok := range tokens {
			if tok.Text == "foo" {
				fooStyle = tok.Style
				break
			}
		}

		assert.NotEmpty(t, fooStyle.Foreground, "function name should have color")
	})

	t.Run("uses colors from provided palette", func(t *testing.T) {
		t.Parallel()

		// Use test palette which has known colors
		palette := lipgloss.TestTheme().Palette()
		tokenizer, err := chroma.NewTokenizer(chroma.StyleFromPalette(palette))
		require.NoError(t, err)
		tokens := tokenizer.Tokenize("go", `package main`)

		require.NotEmpty(t, tokens)

		// Find the "package" keyword and verify it uses the palette's keyword color
		for _, tok := range tokens {
			if tok.Text == "package" {
				assert.Equal(t, string(palette.Keyword), tok.Style.Foreground,
					"keyword should use palette's keyword color")
				assert.True(t, tok.Style.Bold, "keyword should be bold")
				return
			}
		}
		t.Fatal("did not find 'package' keyword in tokens")
	})

	t.Run("returns error for nil styleFunc", func(t *testing.T) {
		t.Parallel()

		_, err := chroma.NewTokenizer(nil)
		assert.Error(t, err)
	})
}

func TestTokenizer_TokenizeLines(t *testing.T) {
	t.Parallel()

	t.Run("tokenizes multi-line comments correctly", func(t *testing.T) {
		t.Parallel()

		tokenizer, err := chroma.NewTokenizer(testStyleFunc())
		require.NoError(t, err)

		// Multi-line JSDoc comment - each line should be recognized as comment
		source := "/**\n * Config options\n */"
		lineTokens := tokenizer.TokenizeLines("javascript", source)

		require.Len(t, lineTokens, 3, "should have tokens for 3 lines")

		// All three lines should contain comment tokens with comment styling
		palette := lipgloss.TestTheme().Palette()
		expectedCommentColor := string(palette.Comment)

		for lineNum, tokens := range lineTokens {
			require.NotEmpty(t, tokens, "line %d should have tokens", lineNum)
			// At least one token on each line should have comment color
			var hasCommentStyle bool
			for _, tok := range tokens {
				if tok.Style.Foreground == expectedCommentColor {
					hasCommentStyle = true
					break
				}
			}
			assert.True(t, hasCommentStyle,
				"line %d should have comment styling, got tokens: %v", lineNum, tokens)
		}
	})

	t.Run("handles single line correctly", func(t *testing.T) {
		t.Parallel()

		tokenizer, err := chroma.NewTokenizer(testStyleFunc())
		require.NoError(t, err)

		source := "const x = 1"
		lineTokens := tokenizer.TokenizeLines("javascript", source)

		require.Len(t, lineTokens, 1)
		require.NotEmpty(t, lineTokens[0])

		// Reconstruct and verify
		var reconstructed string
		for _, tok := range lineTokens[0] {
			reconstructed += tok.Text
		}
		assert.Equal(t, "const x = 1", reconstructed)
	})

	t.Run("handles empty source", func(t *testing.T) {
		t.Parallel()

		tokenizer, err := chroma.NewTokenizer(testStyleFunc())
		require.NoError(t, err)

		lineTokens := tokenizer.TokenizeLines("go", "")
		assert.Empty(t, lineTokens)
	})

	t.Run("returns nil for unsupported language", func(t *testing.T) {
		t.Parallel()

		tokenizer, err := chroma.NewTokenizer(testStyleFunc())
		require.NoError(t, err)

		lineTokens := tokenizer.TokenizeLines("nonexistent-language-xyz", "some code")
		assert.Nil(t, lineTokens)
	})

	t.Run("single-line comment still works", func(t *testing.T) {
		t.Parallel()

		tokenizer, err := chroma.NewTokenizer(testStyleFunc())
		require.NoError(t, err)

		source := "// single line comment"
		lineTokens := tokenizer.TokenizeLines("javascript", source)

		require.Len(t, lineTokens, 1)
		require.NotEmpty(t, lineTokens[0])

		palette := lipgloss.TestTheme().Palette()
		expectedCommentColor := string(palette.Comment)

		var hasCommentStyle bool
		for _, tok := range lineTokens[0] {
			if tok.Style.Foreground == expectedCommentColor {
				hasCommentStyle = true
				break
			}
		}
		assert.True(t, hasCommentStyle, "single-line comment should have comment styling")
	})

	t.Run("handles multi-line comment with empty lines", func(t *testing.T) {
		t.Parallel()

		tokenizer, err := chroma.NewTokenizer(testStyleFunc())
		require.NoError(t, err)

		// JSDoc-style comment with empty line (common pattern)
		source := "/**\n * Description\n *\n * @param foo\n */"
		lineTokens := tokenizer.TokenizeLines("javascript", source)

		// Should have 5 lines
		require.Len(t, lineTokens, 5, "should have tokens for 5 lines")

		palette := lipgloss.TestTheme().Palette()
		expectedCommentColor := string(palette.Comment)

		// All lines (including the empty-content line 3) should have comment styling
		for lineNum, tokens := range lineTokens {
			// Line 3 (" *") may have minimal tokens but should still be comment-styled
			if len(tokens) == 0 {
				continue // Empty token slice is acceptable for whitespace-only lines
			}
			var hasCommentStyle bool
			for _, tok := range tokens {
				if tok.Style.Foreground == expectedCommentColor {
					hasCommentStyle = true
					break
				}
			}
			assert.True(t, hasCommentStyle,
				"line %d should have comment styling, got tokens: %v", lineNum, tokens)
		}
	})
}
