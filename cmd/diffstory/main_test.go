package main_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fwojciec/diffview"
	main "github.com/fwojciec/diffview/cmd/diffstory"
	"github.com/fwojciec/diffview/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp_Run_ReturnsDiffAndClassification(t *testing.T) {
	t.Parallel()

	diffInput := `diff --git a/hello.go b/hello.go
new file mode 100644
index 0000000..e69de29
--- /dev/null
+++ b/hello.go
@@ -0,0 +1,3 @@
+package main
+
+func hello() {}
`

	expectedClassification := &diffview.StoryClassification{
		ChangeType: "feature",
		Narrative:  "core-periphery",
		Summary:    "Add hello function",
		Sections: []diffview.Section{
			{
				Role:  "core",
				Title: "Add function",
				Hunks: []diffview.HunkRef{
					{File: "hello.go", HunkIndex: 0, Category: "core"},
				},
			},
		},
	}

	app := &main.App{
		Input: strings.NewReader(diffInput),
		Classifier: &mock.StoryClassifier{
			ClassifyFn: func(_ context.Context, input diffview.ClassificationInput) (*diffview.StoryClassification, error) {
				// Verify the input has the parsed diff
				require.Len(t, input.Diff.Files, 1)
				require.Equal(t, "hello.go", input.Diff.Files[0].NewPath)
				return expectedClassification, nil
			},
		},
	}

	diff, classification, err := app.Run(context.Background())
	require.NoError(t, err)

	// Verify diff was parsed correctly
	require.NotNil(t, diff)
	require.Len(t, diff.Files, 1)
	assert.Equal(t, "hello.go", diff.Files[0].NewPath)

	// Verify classification was returned
	require.NotNil(t, classification)
	assert.Equal(t, "feature", classification.ChangeType)
	assert.Equal(t, "Add hello function", classification.Summary)
}

func TestApp_Run_ReadsFromFilePath(t *testing.T) {
	t.Parallel()

	diffContent := `diff --git a/hello.go b/hello.go
new file mode 100644
index 0000000..e69de29
--- /dev/null
+++ b/hello.go
@@ -0,0 +1,3 @@
+package main
+
+func hello() {}
`
	// Create a temp file with the diff
	tmpDir := t.TempDir()
	diffPath := filepath.Join(tmpDir, "test.patch")
	err := os.WriteFile(diffPath, []byte(diffContent), 0o644)
	require.NoError(t, err)

	app := &main.App{
		FilePath: diffPath,
		Classifier: &mock.StoryClassifier{
			ClassifyFn: func(_ context.Context, _ diffview.ClassificationInput) (*diffview.StoryClassification, error) {
				return &diffview.StoryClassification{ChangeType: "feature"}, nil
			},
		},
	}

	diff, classification, err := app.Run(context.Background())
	require.NoError(t, err)
	require.NotNil(t, diff)
	require.NotNil(t, classification)
	assert.Len(t, diff.Files, 1)
}

func TestApp_Run_PassesDiffToClassifier(t *testing.T) {
	t.Parallel()

	diffInput := `diff --git a/src/auth.go b/src/auth.go
index 0000000..e69de29
--- a/src/auth.go
+++ b/src/auth.go
@@ -1,3 +1,4 @@
 package auth

+func login() {}
 func logout() {}
`

	var capturedInput diffview.ClassificationInput
	app := &main.App{
		Input: strings.NewReader(diffInput),
		Classifier: &mock.StoryClassifier{
			ClassifyFn: func(_ context.Context, input diffview.ClassificationInput) (*diffview.StoryClassification, error) {
				capturedInput = input
				return &diffview.StoryClassification{ChangeType: "feature"}, nil
			},
		},
	}

	_, _, err := app.Run(context.Background())
	require.NoError(t, err)

	// Verify the diff was passed to the classifier
	require.Len(t, capturedInput.Diff.Files, 1)
	assert.Equal(t, "src/auth.go", capturedInput.Diff.Files[0].NewPath)
	require.Len(t, capturedInput.Diff.Files[0].Hunks, 1)
}

func TestApp_Run_ClassifierError(t *testing.T) {
	t.Parallel()

	diffInput := `diff --git a/hello.go b/hello.go
new file mode 100644
--- /dev/null
+++ b/hello.go
@@ -0,0 +1 @@
+package main
`

	app := &main.App{
		Input: strings.NewReader(diffInput),
		Classifier: &mock.StoryClassifier{
			ClassifyFn: func(_ context.Context, _ diffview.ClassificationInput) (*diffview.StoryClassification, error) {
				return nil, errors.New("API error")
			},
		},
	}

	_, _, err := app.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}

func TestApp_Run_FileNotFound(t *testing.T) {
	t.Parallel()

	app := &main.App{
		FilePath: "/nonexistent/path/to/diff.patch",
		Classifier: &mock.StoryClassifier{
			ClassifyFn: func(_ context.Context, _ diffview.ClassificationInput) (*diffview.StoryClassification, error) {
				return &diffview.StoryClassification{ChangeType: "feature"}, nil
			},
		},
	}

	_, _, err := app.Run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file")
}

func TestApp_Run_EmptyDiff(t *testing.T) {
	t.Parallel()

	// Empty input - no diff content at all
	diffInput := ""

	app := &main.App{
		Input: strings.NewReader(diffInput),
		Classifier: &mock.StoryClassifier{
			ClassifyFn: func(_ context.Context, _ diffview.ClassificationInput) (*diffview.StoryClassification, error) {
				t.Error("Classifier should not be called for empty diff")
				return &diffview.StoryClassification{ChangeType: "feature"}, nil
			},
		},
	}

	_, _, err := app.Run(context.Background())
	require.Error(t, err)
	assert.Equal(t, main.ErrNoChanges, err)
}
