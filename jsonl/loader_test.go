package jsonl_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fwojciec/diffstory/jsonl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_Load(t *testing.T) {
	t.Parallel()

	t.Run("loads valid JSONL file", func(t *testing.T) {
		t.Parallel()

		// Create temp file with valid JSONL
		dir := t.TempDir()
		path := filepath.Join(dir, "cases.jsonl")
		content := `{"input":{"repo":"","commits":[{"hash":"abc123","message":""}],"diff":{"files":[]}},"story":{"change_type":"refactor","narrative":"","summary":"Refactored foo","sections":[]}}
{"input":{"repo":"","commits":[{"hash":"def456","message":""}],"diff":{"files":[]}},"story":{"change_type":"feature","narrative":"","summary":"Added bar","sections":[]}}`
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

		loader := jsonl.NewLoader()
		cases, err := loader.Load(path)

		require.NoError(t, err)
		assert.Len(t, cases, 2)
		assert.Equal(t, "abc123", cases[0].Input.FirstCommitHash())
		assert.Equal(t, "refactor", cases[0].Story.ChangeType)
		assert.Equal(t, "def456", cases[1].Input.FirstCommitHash())
		assert.Equal(t, "feature", cases[1].Story.ChangeType)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		t.Parallel()

		loader := jsonl.NewLoader()
		_, err := loader.Load("/nonexistent/path.jsonl")

		assert.Error(t, err)
	})

	t.Run("returns error for malformed JSON line", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "bad.jsonl")
		content := `{"input":{"commits":[{"hash":"abc123"}]},"story":{}}
not valid json
{"input":{"commits":[{"hash":"def456"}]},"story":{}}`
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

		loader := jsonl.NewLoader()
		_, err := loader.Load(path)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "line 2")
	})

	t.Run("handles empty file", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "empty.jsonl")
		require.NoError(t, os.WriteFile(path, []byte(""), 0o644))

		loader := jsonl.NewLoader()
		cases, err := loader.Load(path)

		require.NoError(t, err)
		assert.Empty(t, cases)
	})

	t.Run("skips empty lines", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "with-blanks.jsonl")
		content := `{"input":{"commits":[{"hash":"abc123"}]},"story":{"change_type":"refactor","summary":"x"}}

{"input":{"commits":[{"hash":"def456"}]},"story":{"change_type":"feature","summary":"y"}}`
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

		loader := jsonl.NewLoader()
		cases, err := loader.Load(path)

		require.NoError(t, err)
		assert.Len(t, cases, 2)
	})

	t.Run("handles large lines exceeding default buffer", func(t *testing.T) {
		t.Parallel()

		// Create a line larger than default scanner buffer (64KB)
		// Generate a message with 100KB of padding
		largeMessage := strings.Repeat("x", 100*1024)
		dir := t.TempDir()
		path := filepath.Join(dir, "large.jsonl")
		content := `{"input":{"commits":[{"hash":"abc123","message":"` + largeMessage + `"}]},"story":{"change_type":"refactor","summary":"x"}}`
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))

		loader := jsonl.NewLoader()
		cases, err := loader.Load(path)

		require.NoError(t, err)
		require.Len(t, cases, 1)
		assert.Equal(t, "abc123", cases[0].Input.FirstCommitHash())
	})
}
