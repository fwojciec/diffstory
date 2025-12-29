# diffstory

![diffstory walkthrough](examples/story-walkthrough.gif)

A diff viewer designed for reviewing AI-generated code changes. Instead of line-by-line diffs, diffstory uses an LLM to classify changes into a structured narrative with semantic sections.

## Installation

```bash
go install github.com/fwojciec/diffview/cmd/diffstory@latest
```

## Quick Start

```bash
# Set your Gemini API key
export GEMINI_API_KEY="your-api-key"

# Run in any git repository with changes
diffstory
```

## Features

- **Git-native analysis** - Auto-detects base branch from `origin/HEAD` and analyzes your current branch
- **LLM-powered classification** - Uses Gemini to classify changes by type (bugfix, feature, refactor) and narrative pattern
- **Semantic sections** - Groups related hunks by role (problem, fix, test, core, supporting)
- **Interactive TUI** - Syntax-highlighted diff viewer with keyboard navigation
- **Eval case management** - Save and replay analyzed diffs for evaluation

## Usage

### Analyze Current Branch

```bash
diffstory
```

Analyzes the diff between your current branch and its base branch, classifies it with Gemini, and opens an interactive TUI.

### Replay Saved Cases

```bash
diffstory replay <file.jsonl> [index]
```

Re-opens a previously saved eval case. The index is zero-based and defaults to 0.

## How It Works

1. Detects your base branch from `origin/HEAD`
2. Gets the diff (`base...HEAD`)
3. Sends the diff to Gemini for classification
4. Displays results in an interactive TUI with:
   - Change type and narrative pattern
   - Summary of changes
   - Sections grouping related hunks by semantic role

## Requirements

- Git repository with a configured remote
- `GEMINI_API_KEY` environment variable

## License

[MIT](LICENSE)
