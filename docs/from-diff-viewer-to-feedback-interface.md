# From Diff Viewer to Feedback Interface

A conceptual exploration of what "reviewing AI-generated code" actually means, and how the tool should evolve beyond traditional diff presentation.

## The Medium Shift

Current PR review is "newspapers on the web" - we took `diff -u` from 1974, wrapped it in web UIs, added comments. The fundamental unit remains "here are the lines that changed, approve or reject."

That made sense when humans wrote each line deliberately. With AI-generated code, it's like reviewing the newspaper's typesetting when you should be asking whether the story is true.

## What's Different With AI-Generated Code

1. **The prompt is the source code.** The diff is just a rendering. But we review the rendering, not the source.

2. **The AI can regenerate instantly.** Feedback isn't "fix line 42" - it's "try a different approach." The loop is fundamentally different.

3. **Verification beats inspection.** Reading every line is automation bias. You need to *test* whether it does what you asked, not *approve* what it wrote.

4. **The AI has receipts.** It knows what it looked at, what it considered, what it rejected. That metadata is invisible in current tools.

## The Reframe: Feedback Loop Interface

The diff viewer might not be about viewing diffs at all - or rather, it's about evaluating diffs in terms of the feedback mechanisms that produce them.

| Old Model | New Model |
|-----------|-----------|
| Review *output* | Participate in *process* |
| After-the-fact | During-creation |
| Approve/reject | Steer/refine |
| One-shot | Iterative |
| "Is this correct?" | "Is this going the right direction?" |

### The Right Altitude

The engineer needs to be:
- Close enough to see what's actually being written
- Far enough to see patterns across files/decisions
- At the right altitude to intervene meaningfully

Not line-by-line (old code review), not pure delegation (just accept/reject), but somewhere in between where they can see patterns and make course corrections.

### Intervention That Compounds

The key insight: intervention isn't just about fixing the current output. When you intervene at step 3 of a 10-step process:
1. You fix step 3
2. You shape steps 4-10
3. You potentially improve future sessions

The tool should enable **course correction that compounds** - tightening feedback loops that refine outputs over time.

## The Vision

Three components:

1. **A stream view** of the agent's work as it happens (what's it looking at, what's it deciding, what's it writing)

2. **Intervention points** where the engineer can redirect ("stop, try X instead", "that's the wrong file", "you're overcomplicating this")

3. **A record** of interventions that becomes context for future runs

The diff at the end becomes almost a *receipt* - proof that the feedback loop converged to something. The real value is the tightening that happened along the way.

This is less "code review tool" and more "agent collaboration interface with code as the medium."

## Practical Constraints: Available Inputs

We have to be modest about what inputs are actually available. The extension path:

### Layer 1: Now
**Git diff + smart presentation**

What we have:
- Git diff (old text → new text)
- Commit messages (if the agent writes good ones)
- Branch names (may encode task info)
- File structure (can infer some semantics)

What we don't have:
- The original prompt/task
- What files were examined but not changed
- Reasoning traces
- Rejected alternatives
- Sequence/timing of changes
- Confidence signals

Even with just git diff, semantic chunking, dependency ordering, and progressive disclosure reduce cognitive load on the output.

### Layer 1.5: Beads Integration
**Git diff + task context**

Beads issues provide:
- Task specification (the "why")
- Acceptance criteria (verification checklist)
- Dependencies (what this builds on)
- Related issues closed together

This gives us the "original prompt" equivalent - correlating intent to implementation:

```
beads-abc: "Add JWT authentication"
├── Description: what was requested
├── Dependencies: what this builds on
├── Commits: 3 commits on branch feature/beads-abc
│   └── Diff: semantic view of changes
└── Validation: criteria from issue
```

### Layer 2: Claude Hooks Traces
**Git diff + task context + process visibility**

Hooks fire on tool calls, prompt submissions, and other events. A trace-generating hook could capture:

```jsonl
{"event": "tool_call", "tool": "Read", "file": "auth.go", "timestamp": ...}
{"event": "tool_call", "tool": "Edit", "file": "auth.go", "success": true}
```

This unlocks:
- What files were examined (not just changed)
- Sequence of operations (the "story" of the change)
- Time spent in different phases
- Failed attempts (tool calls that errored)

Architecture sketch:
1. Hook writes events to `.claude/traces/<session-id>.jsonl`
2. Commit message or branch metadata references session ID
3. Diff viewer reads trace file alongside git diff
4. Correlates: "this hunk in auth.go came after examining user.go and middleware.go"

### Layer 3: Future
**Real-time integration**

Direct integration where you watch the agent work in real-time, not just review after. The full feedback loop interface.

## Design Principle

Each layer unlocks more of the vision, but layer 1 is useful on its own. Building layer 1 well teaches you what questions reviewers actually have, which tells you what inputs you need to answer them.

The tool teaches you what it needs to become.

## References

- [Intelligent Diff Presentation](./intellingent-diff-presentation.md) - Cognitive science foundations
- Ryo Lu (Cursor head of design) on interface as bottleneck: "The models aren't the bottleneck anymore - it's the interface... our thoughts aren't linear, they're spatial, visual, emotional"
- Marshall McLuhan's "medium is the message" - new mediums initially replicate old forms before finding native expression
