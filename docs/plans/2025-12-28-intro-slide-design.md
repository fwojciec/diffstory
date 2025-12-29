# Intro Slide Design

## Problem

When reviewing code with diffstory, users jump straight into code hunks without context. An intro slide would orient reviewers before diving in.

## Solution

Add a non-code "section 0" that displays the classification summary and a roadmap of upcoming sections.

## User Flow

1. `diffstory` launches at intro slide (section 0)
2. Press `s` to advance to section 1 (first code section)
3. Press `S` from section 1 to return to intro
4. `]`/`[` navigate files within sections as before

## Intro Slide Layout

```
┌─────────────────────────────────────────────────┐
│                                                 │
│  Add expiry check to token validation           │  ← Summary
│                                                 │
│  Sections:                                      │
│    1. Problem: Auth token validation fails      │  ← Section list
│    2. Fix: Add expiry check                     │
│    3. Test: Verify token rejection              │
│                                                 │
│                                                 │
│                         [s] next section        │  ← Hint
└─────────────────────────────────────────────────┘
│ overview                             100%       │  ← Status bar
└─────────────────────────────────────────────────┘
```

## Implementation

Changes localized to `bubbletea/story.go`:

1. **Section indexing**: `activeSection` starts at -1 (intro). Code sections are 0-indexed internally, display as 1-N.

2. **Intro rendering**: When `activeSection == -1`, `renderContent()` returns the intro layout instead of `renderDiff()`.

3. **Status bar**: Shows "overview" when on intro, existing section title otherwise.

4. **No new fields**: Uses existing `StoryClassification.Summary` and `Section.Title`.

## Not In Scope

- Graphical diagrams (future enhancement)
- Extracted slide component (extract if we add more slides)
- Changes to classification or domain types
