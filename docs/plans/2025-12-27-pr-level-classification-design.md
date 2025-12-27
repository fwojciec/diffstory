# PR-Level Classification Design

## Problem

The current eval dataset contains individual commits, but work is structured into PRs. This loses context:

- "Address PR review feedback" commits are meaningless alone
- "Start work" / "Close" commits are workflow noise
- The PR tells the complete story, commits are chapters

The classifier itself is agnostic to commit boundaries - it organizes hunks into narrative sections regardless of source. The limitation is in data collection, not classification.

## Design Decisions

### Classification granularity: PR-level

PRs are the natural unit of review. A developer's mental model is "review this PR" not "review these 5 commits individually." The classifier should receive the full PR context.

### Commit messages: Include them

Historical PRs have rich commit message data. Including them gives the classifier more context. For the runtime case (uncommitted changes), this section is simply empty.

### Collection approach: Replace, don't extend

The commit-level collector is being replaced entirely. No `--mode` flag, no backward compatibility with old dataset format.

### Git context at runtime: Verify before enriching

When a diff is piped in, we can query git for context. But we should verify the piped diff matches what git thinks the branch contains. If they match (good hygiene), enrich with commit messages. If not (uncommitted changes), warn and classify without context.

## Data Structure Changes

### Before

```go
type ClassificationInput struct {
    Commit CommitInfo `json:"Commit"`
    Diff   Diff       `json:"Diff"`
}

type CommitInfo struct {
    Hash    string `json:"Hash"`
    Repo    string `json:"Repo"`
    Message string `json:"Message"`
}
```

### After

```go
type ClassificationInput struct {
    Repo    string        `json:"repo"`
    Branch  string        `json:"branch"`
    Commits []CommitBrief `json:"commits"`
    Diff    Diff          `json:"diff"`
}

type CommitBrief struct {
    Hash    string `json:"hash"`
    Message string `json:"message"`
}
```

Usage scenarios:
- **PR from history**: `Branch="diffview-zv1"`, `Commits` has all PR commits
- **Local uncommitted**: `Branch="feature-x"`, `Commits` empty or partial
- **Piped foreign diff**: `Branch=""`, `Commits` empty

## GitRunner Interface Changes

### Before

```go
type GitRunner interface {
    Log(ctx context.Context, repoPath string, limit int) ([]string, error)
    Show(ctx context.Context, repoPath, hash string) (string, error)
    Message(ctx context.Context, repoPath, hash string) (string, error)
}
```

### After

```go
type GitRunner interface {
    // For historical collection (PR boundaries)
    MergeCommits(ctx context.Context, repoPath string, limit int) ([]string, error)

    // Range operations (shared by collection and runtime)
    CommitsInRange(ctx context.Context, repoPath, base, head string) ([]CommitBrief, error)
    DiffRange(ctx context.Context, repoPath, base, head string) (string, error)

    // For runtime context enrichment
    CurrentBranch(ctx context.Context, repoPath string) (string, error)
    MergeBase(ctx context.Context, repoPath, ref1, ref2 string) (string, error)
}
```

### Git commands

| Method | Git command |
|--------|-------------|
| `MergeCommits` | `git log --merges --format=%H -n <limit>` |
| `CommitsInRange` | `git log --format=%H%x00%s <base>..<head>` |
| `DiffRange` | `git diff <base>...<head>` |
| `CurrentBranch` | `git rev-parse --abbrev-ref HEAD` |
| `MergeBase` | `git merge-base <ref1> <ref2>` |

### PR extraction from merge commit

For a merge commit hash, the PR range is:
- Base: `merge^1` (the main branch before merge)
- Head: `merge^2` (the feature branch tip)

The merge commit message format (GitHub default): `Merge pull request #N from user/branch`

## Collector Changes

### Before

```
for each commit hash:
    git show <hash> → single commit diff
    git message <hash> → single message
    emit EvalCase
```

### After

```
for each merge commit:
    parse merge message → PR number, branch name
    commits = git log merge^1..merge^2
    diff = git diff merge^1...merge^2
    apply min/max line filters
    emit EvalCase with {repo, branch, commits, diff}
```

### Filtering

Line filters apply to the combined PR diff. Default `--max-lines` should increase (PRs are larger than individual commits). Suggest 1000-1500 as new default.

## Formatter Changes

### Before

```
<commit_message>
Fix bug in parser
</commit_message>

<diff>
=== FILE: foo.go ===
...
</diff>
```

### After

```
<context>
Repository: diffview
Branch: diffview-zv1

Commits:
- af44c89: Address PR feedback: skip redundant viewport updates
- 51fad8d: Fix extra blank lines in diff rendering
- 1680dfb: Close diffview-zv1
</context>

<diff>
=== FILE: foo.go ===
...
</diff>
```

When `Commits` is empty (no git context), omit the "Commits:" section entirely.

## Runtime Behavior

### Mode 1: Piped diff (no verification)

```bash
git diff origin/main | diffview
```

diffview receives raw diff on stdin. If in a git repo:
1. Query current branch
2. Query merge-base with origin/main
3. Generate expected diff: `git diff merge-base..HEAD`
4. Compare to piped diff:
   - **Match**: Full enrichment (branch + commits)
   - **Superset**: Uncommitted changes detected, warn user
   - **No match**: Foreign diff, classify without context

### Mode 2: Git-aware (no pipe)

```bash
diffview --base=origin/main
```

diffview queries git directly, no comparison needed. Always gets full context.

## Implementation Order

1. **Data structures** - Update `ClassificationInput`, add `CommitBrief`
2. **GitRunner** - Add new methods to interface and `git.Runner` implementation
3. **Collector** - Rewrite to iterate merge commits
4. **Formatter** - Update to produce new `<context>` format
5. **Dataset** - Generate from diffview + locdoc repos
6. **Classifier prompt** - No changes needed (already generic)

## Future Work (tracked in beads)

| Issue | Description |
|-------|-------------|
| `diffview-1ib` | Runtime git context enrichment for piped diffs |
| `diffview-a2m` | Section-aware diff rendering (navigation, collapse, category styling) |

## Validation

- [ ] Collector produces PR-level EvalCases from merge commits
- [ ] Each case has branch name and all commit messages
- [ ] Combined diff matches `git diff merge^1...merge^2`
- [ ] Line filtering works on combined PR diff
- [ ] Formatter produces valid `<context>` section
- [ ] Classifier successfully classifies PR-level diffs
- [ ] Dataset generated from both repos
