# Intelligent Diff Presentation for AI-Agent Code Review: A Design Framework

When developers review AI-generated code, they face a fundamentally different challenge than traditional code review. The goal shifts from line-by-line approval to understanding *what the AI did* and *why*, then providing actionable feedback to guide the next iteration. Research reveals that **cognitive load is the primary bottleneck**: reviewers can only effectively process 200-400 lines per session, context-switching costs 20% of cognitive capacity per switch, and attention degrades sharply after 60-90 minutes. A well-designed diff viewer must work with these constraints, not against them.

This framework synthesizes findings from cognitive science, program comprehension research, existing tooling patterns, and AI agent workflows to define how Claude Code's output should be structured for human review.

## Cognitive science reveals why large diffs overwhelm reviewers

The brain's working memory holds only **7±2 chunks** of information simultaneously, with just 2-3 elements actively processable at once. The landmark Cisco/SmartBear study established concrete limits: beyond **400 lines of code**, defect-finding ability diminishes dramatically. At review rates faster than 500 LOC/hour, quality collapses—reviewers shift from reading to skimming.

Context switching compounds the problem. Each file transition forces a **mental model rebuild**, consuming cognitive resources that should go toward understanding the code itself. Carnegie Mellon research found developers juggling multiple contexts spend only 20% of cognitive energy on actual work—the rest is lost to mental overhead. After interruptions, regaining full focus takes an average of **23 minutes**.

These constraints have direct implications for diff presentation:

- **Chunk changes into 200-400 line segments** with clear progress indicators
- **Minimize file transitions** by grouping semantically related changes together
- **Preserve review state** so interrupted sessions can resume without cognitive penalty
- **Surface fatigue indicators** after 60-90 minutes of continuous review

Google's internal research confirms this at scale: across 9 million code reviews, the median change size was just **24 lines**, with over 35% of changes modifying only a single file. Small, focused changes reviewed quickly (median under 4 hours) produce better outcomes than large PRs that languish.

## Mental models determine how developers comprehend changes

Experienced reviewers construct mental models at three layers, according to Letovsky's code comprehension framework: the **specification layer** (what the code should accomplish), the **implementation layer** (how it achieves that), and the **annotation layer** (connections between intent and execution). When reviewing AI-generated code, the specification layer is especially fragile—the reviewer didn't write the prompt and may not fully understand the AI's interpretation.

Expert programmers organize changes semantically rather than syntactically. They recognize **programming plans**—stereotypical implementations of common goals—and chunk by meaning rather than by line. A study comparing expert and novice programmers found that experts recall code in semantic entities while novices remember line-by-line. This explains why file-alphabetical ordering (the default in most tools) feels disjointed: it fragments semantic coherence.

The most effective mental models for organizing modifications follow predictable patterns:

| Pattern | Structure | Best for |
|---------|-----------|----------|
| Cause → Effect | Bug report → Fix → Test | Bugfixes, issue resolution |
| Core → Periphery | Central logic → Supporting changes | Feature additions |
| Entry → Implementation | API/interface → Internal logic → Helpers | Understanding flow |
| Before → After | Old approach → Transformation → New approach | Refactoring |

A diff viewer should present changes following these natural mental patterns, not arbitrary file ordering. When a reviewer can predict "what comes next," comprehension accelerates.

## Semantic chunking beats file-based grouping

Traditional diff tools group changes by filename, presenting them alphabetically. This approach is simple to implement but **semantically incoherent**—a single rename across 40 files appears as 40 unrelated changes. Semantic chunking groups by logical operation instead.

Tools like SemanticDiff demonstrate the power of AST-aware analysis: they parse code structure rather than raw text, identify invariances (changes that look different but are semantically equivalent), detect moved code automatically, and group refactorings separately from logic changes. Difftastic takes a similar approach with tree-sitter parsing, eliminating noise from formatting changes that would otherwise clutter the diff.

For an AI-agent diff viewer, semantic chunking should operate at multiple levels:

**Operation-level grouping:**
- **Renames/refactors**: Same pattern across many files, mechanical transformation
- **New features**: New files plus integration points, high reviewer attention needed
- **Bugfixes**: Targeted changes plus tests, focused scope
- **Formatting/cleanup**: No semantic impact, should be collapsible

**Noise reduction:**
- Auto-collapse trivial changes (imports, formatting, generated files)
- Distinguish "12 formatting changes (hidden)" from substantive modifications
- Allow expansion on demand without leaving the current view

**Progressive disclosure:**
```
Level 1: High-level summary
  "This task added user authentication and updated related tests"

Level 2: Change categories  
  - Core logic: 3 files, +45/-12 lines [REQUIRES ATTENTION]
  - Tests: 2 files, +89 lines [AUTO-GENERATED]
  - Dependencies: package.json [TRIVIAL]

Level 3: Individual files with semantic context
  - auth.service.ts: "New OAuth2 provider integration"
  - auth.controller.ts: "Updated routes to use new auth"

Level 4: Line-by-line diff with full context
```

This hierarchy matches how the brain organizes information—hierarchical representations enable efficient planning and flexibility through abstraction.

## AI agent context must be surfaced alongside diffs

AI coding agents produce metadata that traditional diff viewers ignore entirely. Claude Code, for example, operates on a **"gather context → take action → verify work → repeat"** loop, generating planning artifacts, TODO lists, extended thinking traces, and tool call sequences along the way. This context is essential for reviewers but typically invisible.

**What Claude Code produces that reviewers need:**
- **Original prompt**: The task that initiated the changes
- **Subagent reasoning**: Exploration and planning outputs
- **Tool call sequences**: What files were examined, what commands were run
- **Extended thinking**: Chain-of-thought reasoning (triggered by "think harder")
- **Commit messages**: Auto-generated with intent explanation

**What's typically missing:**
| Gap | Impact on Review |
|-----|------------------|
| Original prompt | Reviewer can't verify code matches intent |
| Rejected alternatives | No visibility into what was NOT chosen |
| Confidence levels | AI presents uncertain code identically to high-confidence code |
| Hallucination markers | No indication when AI invents non-existent APIs |
| Iteration history | Final result hides false starts |

Research on AI code comprehension found that students achieved only a **32.5% success rate** when comprehending LLM-generated code, primarily due to unfamiliar code style, incorrect assumptions of correctness (automation bias), and missing architectural intent. Microsoft Research found that reviewers miss **40% more bugs** when reviewing AI code compared to human code.

A diff viewer designed for AI output should implement **layered information architecture**:

```
┌─────────────────────────────────────────────────────────┐
│ 📝 Original Request:                                   │
│ "Add authentication using JWT tokens"                  │
├─────────────────────────────────────────────────────────┤
│ 🧠 AI Reasoning (expandable):                          │
│ 1. Identified existing auth pattern in auth.py         │
│ 2. Chose HMAC-SHA256 for token signing                 │
│ 3. Added token refresh to handle long sessions         │
├─────────────────────────────────────────────────────────┤
│ ⚠️ Attention Areas:                                    │
│ • Edge case: token expiry during active session        │
│ • No rate limiting implemented (deferred)              │
│ • [Low confidence] Error handling in refresh logic     │
└─────────────────────────────────────────────────────────┘
```

The original prompt should **always be visible**—it's the specification layer that makes everything else comprehensible. Reasoning traces should be collapsible but accessible. Confidence indicators should be surfaced visually, perhaps with color-coding for uncertainty.

## Cross-file dependencies must be visualized

AI agents often make changes that span many files, with complex interdependencies that are invisible in flat file lists. A function added in one file is called from another, which requires an import in a third. Without dependency visualization, reviewers must hold these relationships in working memory—exactly the cognitive resource that's already strained.

Effective approaches include:

- **Topological ordering**: Show dependencies before dependents, so reviewers see foundations before buildings
- **Entry-point-first**: Start with public APIs and controllers, drill down to implementations
- **Impact-based ordering**: High-impact core changes first, peripheral changes later
- **Visual dependency graphs**: Show caller/callee relationships with interactive filtering

NDepend research found that **matrix views** handle large, complex dependencies better than node-link diagrams, which become unreadable at scale. For a TUI tool, consider a simplified representation:

```
auth.service.ts → auth.controller.ts → routes/auth.ts
      ↓
    user.model.ts
```

Color-coding relationships—callers in green, callees in blue, bidirectional in amber—helps reviewers trace data flow without rebuilding mental models repeatedly.

## The reviewer needs answers to specific questions for each change

Based on the cognitive science research and code review comprehension models, an effective diff viewer should help reviewers answer these questions for each change:

**Understanding questions:**
1. What was the original task/intent? (Shows the prompt)
2. What approach did the AI take? (Shows high-level summary)
3. What are the core changes vs. supporting changes? (Shows hierarchy)
4. What files were examined but not changed? (Shows AI's context)
5. What alternatives were considered? (Shows reasoning)

**Verification questions:**
1. Does this match my expectation of how it should be done?
2. Are there edge cases the AI didn't handle?
3. Does this follow project conventions?
4. Is the code actually correct, or just plausible-looking?
5. What would break if this is wrong?

**Feedback questions:**
1. What specific changes do I want to request?
2. How do I reference this code precisely to the AI?
3. What context does the AI need to understand my feedback?
4. Is this blocking or a suggestion?

The viewer should make answering these questions efficient, not require the reviewer to hunt for information.

## Feedback-optimized presentation enables high-quality human-AI collaboration

The goal isn't just understanding—it's providing actionable feedback that the AI agent can use to improve. This requires presentation choices that facilitate structured feedback.

**Conventional Comments format** provides a standardized, machine-parseable structure:

```
<label> [decorations]: <subject>

[discussion]
```

Core labels include `praise:`, `nitpick:`, `suggestion:`, `issue:`, `question:`, and `todo:`. Decorations like `(blocking)` or `(non-blocking)` clarify severity. This format is both human-readable and structured enough for AI agents to parse reliably.

**Copy formats for AI feedback** should include context:
```
📍 src/utils/auth.ts:42-48
```typescript
function validateUser(user) {
  return user.email && user.verified;
}
```

💬 Feedback:
suggestion (security): Missing null check before accessing user.email.
Consider using optional chaining: `user?.email`
```

**Review state tracking** is essential for iterative workflows:
- `[ ]` Unreviewed
- `[?]` Needs discussion  
- `[✓]` Reviewed and understood
- `[!]` Has pending feedback
- `[x]` Changes requested

Session persistence should save: reviewed file list with timestamps, cursor position, pending unsent comments, and diff bounds (what version was reviewed). This enables resuming interrupted reviews—critical given the 60-90 minute fatigue window.

## Keyboard-driven TUI design enables efficient review workflows

TUI tools for code review have converged on consistent patterns. Lazygit, tig, and gitui all use **panel-based layouts** with vim-style navigation. The most effective keyboard mappings:

| Key | Action |
|-----|--------|
| `j/k` | Move down/up through hunks or files |
| `n/N` | Next/previous file with changes |
| `Enter` | Expand/drill down into section |
| `r` | Mark current file as reviewed |
| `c` | Add comment at current line |
| `q` | Close view or go back |
| `/` | Search within diff |
| `?` | Show available keybindings |

**Modal vs. modeless interaction** presents a tradeoff: vim-style modes are efficient for power users but increase learning curve. The recommended approach is to start modeless with clear keybinding hints, offer vim-mode as optional configuration, and always show current mode in the status bar.

**Syntax highlighting** matters enormously for diff readability. Tools like delta use tree-sitter for accurate highlighting and word-level diff detection (via Levenshtein distance) to show exactly which characters changed within a line. A diff viewer for AI code should combine syntax highlighting with diff highlighting, support both dark and light themes (with automatic detection), and offer high-contrast modes for accessibility.

**Split-pane layout** with file tree navigation is essential for multi-file changes:

```
┌──────────────────────────────────────────────────────────┐
│  Status: [3/7 files reviewed] [Mode: NORMAL] [60m]      │
├─────────────────┬────────────────────────────────────────┤
│  📁 Files       │  📄 auth.service.ts                   │
│  ────────       │  ─────────────────                    │
│  [✓] auth.ts    │  @@ -15,7 +15,12 @@                   │
│  [?] utils.ts   │  - const token = sign(payload);       │
│  [ ] api.ts     │  + const token = sign(payload, {      │
│  [ ] routes.ts  │  +   expiresIn: '1h',                 │
│                 │  +   algorithm: 'HS256'               │
│                 │  + });                                │
│                 │                                        │
│  💭 AI Context  │  🧠 Reasoning:                        │
│  ────────────   │  Added expiry and algorithm for       │
│  Prompt: "Add.. │  security best practices              │
├─────────────────┴────────────────────────────────────────┤
│  [j/k] nav  [c] comment  [r] reviewed  [?] help         │
└──────────────────────────────────────────────────────────┘
```

## Concrete design recommendations for a Claude Code companion tool

Based on the synthesized research, here are specific recommendations:

**Information architecture:**
1. **Always-visible**: Original prompt, summary of changes, progress indicator, session timer
2. **Collapsible but accessible**: Full reasoning trace, files examined, tool calls
3. **On-demand**: Raw conversation, token usage, iteration history

**Change organization:**
1. Group by semantic operation, not file location
2. Order core changes before peripheral; dependencies before dependents
3. Auto-collapse formatting, imports, and generated files
4. Show "one rename across 40 files" as a single collapsible operation
5. Lead with the "protagonist" files where the main change happens

**AI-specific adaptations:**
1. Explicit visual marking for AI-generated sections
2. Confidence indicators (even if heuristic-based initially)
3. Link reasoning to specific code blocks (hover to see "why")
4. Verification checklists for security, edge cases, conventions
5. Show what the AI examined but didn't change

**Feedback workflow:**
1. Implement Conventional Comments for structured feedback
2. Track reviewed/pending/needs-discussion states per file
3. Generate copy-ready snippets with file:line references
4. Persist session state for interrupted reviews
5. Support interdiff view ("what changed since I last reviewed")

**Navigation and interaction:**
1. Keyboard-first with vim-style defaults (configurable)
2. Split-pane layout with collapsible file tree
3. Progressive disclosure at all levels
4. External editor integration with `+line` argument support
5. Git integration for context (blame, history, branches)

## What matters most for comprehension and feedback quality

The research converges on a clear priority ordering:

1. **Semantic coherence** over file-alphabetical ordering—group related changes, minimize context switches
2. **Intent visibility**—the original prompt and AI reasoning must be accessible, not buried
3. **Progressive disclosure**—summary first, details on demand, trivial changes hidden
4. **Review state persistence**—sessions are interrupted; state must survive
5. **Structured feedback affordances**—make it easy to provide actionable, well-referenced feedback

The fundamental insight is that reviewing AI code is a **collaborative dialogue**, not a gatekeeping checkpoint. The diff viewer should optimize for understanding and iteration speed, not approval mechanics. When the human can quickly comprehend what the AI did, verify it matches intent, and provide specific feedback, the human-AI loop tightens—and the AI becomes genuinely useful.

## Conclusion

Building an intelligent diff viewer for AI-generated code requires rethinking assumptions inherited from traditional code review tools. The cognitive constraints are real: 200-400 lines per session, 60-90 minutes before fatigue, 23 minutes to recover from interruption. Semantic chunking, hierarchical organization, and dependency visualization work with the brain's natural information processing. AI-specific metadata—especially the original prompt and reasoning traces—fills the comprehension gaps that make AI code harder to review than human code.

The key innovation opportunity lies in **narrative presentation**: transforming scattered file changes into a coherent story that answers "what happened and why." When the viewer surfaces intent, groups related changes, orders by cognitive flow, and provides structured feedback affordances, it transforms reviewing AI output from cognitive burden into efficient collaboration. The reviewer's job becomes understanding and refining, not spelunking through disconnected fragments.
