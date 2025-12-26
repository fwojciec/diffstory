# Token-Based Word Diff Design

## Problem

Current worddiff implementation uses diff-match-patch at character level with DiffCleanupSemantic. This produces partial identifier highlighting (`myVariable` vs `myValue` shows `myVa` as common), which is confusing for code review.

## Solution

Replace character-level diffing with token-based array diffing. Tokenize code into identifiers, operators, etc., then diff the token arrays. This eliminates partial identifier highlighting entirely.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Library | `pmezard/go-difflib` | Works on slices natively, has `SequenceMatcher.Ratio()` |
| Package name | `difflib/` | Ben Johnson pattern: name after dependency |
| Similarity threshold | 0.4, internal | Caller doesn't need to know; low similarity = full replacement |
| Whitespace handling | Separate tokens | More accurate for code; `mergeSegments` combines for output |
| Tokenizer | Simple regex | YAGNI; handles 90% of cases |
| Tests | Rewrite entirely | Behavior changes fundamentally |

## Architecture

### Package Structure

```
difflib/
├── difflib.go      # Differ type, Diff method, tokenize, all logic
└── difflib_test.go # Test suite + benchmarks
```

### Differ Type

```go
type Differ struct {
    tokenPattern *regexp.Regexp
}

func NewDiffer() *Differ {
    return &Differ{
        tokenPattern: regexp.MustCompile(
            `([a-zA-Z_][a-zA-Z0-9_]*)|` + // identifiers
            `([0-9]+\.?[0-9]*)|` +         // numbers
            `("[^"]*"|'[^']*')|` +         // string literals
            `([+\-*/=<>!&|^%]+)|` +        // operators
            `([(){}\[\];,.])|` +           // punctuation
            `(\s+)`,                        // whitespace
        ),
    }
}
```

### Algorithm Flow

```
1. old == new? → return single unchanged segment (fast path)
2. Tokenize both strings
3. Create SequenceMatcher, get matching blocks (once)
4. Compute ratio from blocks: 2.0 * matchedCount / totalTokens
5. If ratio < 0.4 → return everything as changed
6. Build segments from matching blocks
7. Merge adjacent segments with same Changed status
```

### Segment Building

```go
func buildSegments(oldTokens, newTokens []string, blocks []difflib.Match) (oldSegs, newSegs []Segment) {
    oldIdx, newIdx := 0, 0

    for _, block := range blocks {
        // Gap before match = changed
        if oldIdx < block.A {
            oldSegs = append(oldSegs, Segment{
                Text:    strings.Join(oldTokens[oldIdx:block.A], ""),
                Changed: true,
            })
        }
        if newIdx < block.B {
            newSegs = append(newSegs, Segment{
                Text:    strings.Join(newTokens[newIdx:block.B], ""),
                Changed: true,
            })
        }

        // Match = unchanged
        if block.Size > 0 {
            text := strings.Join(oldTokens[block.A:block.A+block.Size], "")
            oldSegs = append(oldSegs, Segment{Text: text, Changed: false})
            newSegs = append(newSegs, Segment{Text: text, Changed: false})
        }

        oldIdx = block.A + block.Size
        newIdx = block.B + block.Size
    }

    return mergeSegments(oldSegs), mergeSegments(newSegs)
}
```

## Test Cases

### Core Improvement (no partial identifiers)

```go
{"myVariable", "myValue"}     // entire tokens differ
{"getUserName", "getUserEmail"} // entire tokens differ
```

### Similarity Threshold

```go
{"return x + 1", "return x + 2"}        // high similarity → word diff
{"func foo() {}", "type Bar struct{}"}  // low similarity → full replacement
```

### Edge Cases

```go
{"", ""}           // both empty
{"", "text"}       // old empty
{"text", ""}       // new empty
{"same", "same"}   // identical (fast path)
{"hello 👋", "hello 🌍"}  // unicode
```

### Token Boundaries

```go
{"x + y", "x - y"}   // operators as tokens
{"a  b", "a b"}      // whitespace tokens differ
```

## Benchmarks

```go
BenchmarkDiffer_Diff/short_similar
BenchmarkDiffer_Diff/short_different
BenchmarkDiffer_Diff/long_line
BenchmarkDiffer_Diff/identical
```

Run with `go test -bench=. -benchmem ./difflib/`

## Migration

1. Create `difflib/` package with new implementation
2. Update imports: `worddiff` → `difflib` in:
   - `cmd/diffview/main.go`
   - `bubbletea/viewer.go`
   - `bubbletea/viewer_test.go`
3. Delete `worddiff/` package
4. Remove `sergi/go-diff` dependency
5. Add `pmezard/go-difflib` dependency

## Validation

- [ ] No partial identifier highlighting (test: myVariable → myValue)
- [ ] Similarity threshold skips noisy diffs
- [ ] Performance acceptable (<100ms per line pair)
- [ ] `make validate` passes
