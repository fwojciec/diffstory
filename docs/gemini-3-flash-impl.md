# Optimizing Gemini Flash for Go CLI code diff classification

**Your best choice for production is Gemini 2.5 Flash-Lite** (`gemini-2.5-flash-lite`) — Google explicitly designed it for "high-volume, latency-sensitive tasks like classification" with the lowest pricing tier ($0.10/1M input tokens). All Flash models support **1M token context windows**, making your 50K-token diffs well within limits. However, your 650-token static prompt is below the minimum threshold for explicit context caching, so you'll rely on implicit caching or need to restructure your approach.

## Model selection: stability beats capability for classification

For code diff classification in production, **Gemini 2.5 Flash-Lite** is the optimal choice despite newer options being available. Three factors drive this recommendation:

| Model | Input Cost | Output Cost | Status | Best For |
|-------|-----------|-------------|--------|----------|
| **Gemini 2.5 Flash-Lite** | $0.10/1M | $0.40/1M | Stable GA | Classification, high volume |
| Gemini 2.5 Flash | $0.30/1M | $2.50/1M | Stable GA | Complex reasoning tasks |
| Gemini 3 Flash Preview | $0.50/1M | $3.00/1M | Preview | Frontier coding capabilities |
| Gemini 2.0 Flash | $0.10/1M | $0.40/1M | Stable | General purpose |

**Why 2.5 Flash-Lite wins:** It's 1.5x faster than Gemini 2.0 Flash, maintains stable GA status with predictable behavior, supports full JSON Schema structured outputs, and costs **3-7x less than higher-tier models**. The preview Gemini 3 Flash scores 78% on SWE-bench for code understanding, but preview models carry a 2-week deprecation warning risk and potential behavior changes.

For your classification task, the difference in code understanding capability between models matters less than structured output reliability. All 2.5+ models preserve property ordering and support full JSON Schema — critical for consistent classification results.

## Structured output: use responseSchema, not prompt-embedded JSON

Gemini's controlled decoding **guarantees syntactically valid JSON** when you use `responseSchema` with `responseMIMEType: "application/json"`. Critical implementation details:

**Don't duplicate schema in prompts.** Google's documentation explicitly states this "lowers quality." Your JSON schema belongs only in the `ResponseSchema` configuration field. For simple categorical classification, use `text/x.enum` MIME type for deterministic output constrained to your enum values.

```go
// For classification with multiple output fields
config := &genai.GenerateContentConfig{
    Temperature:      genai.Ptr(float32(0.3)), // Good for 2.x classification
    ResponseMIMEType: "application/json",
    ResponseSchema: &genai.Schema{
        Type: "object",
        Properties: map[string]*genai.Schema{
            "narrative_type": {
                Type: "string",
                Enum: []string{"feature", "bugfix", "refactor", "docs", "test", "chore"},
                Description: "Primary classification of the code change",
            },
            "confidence": {
                Type: "number",
                Description: "Confidence score between 0 and 1",
            },
            "affected_areas": {
                Type:  "array",
                Items: &genai.Schema{Type: "string"},
                Description: "List of affected code areas or modules",
            },
        },
        Required:         []string{"narrative_type", "confidence"},
        PropertyOrdering: []string{"narrative_type", "confidence", "affected_areas"},
    },
}
```

**Temperature considerations:** Your current **0.3 is appropriate for Gemini 2.x** classification tasks. However, if you upgrade to Gemini 3.x, Google strongly recommends keeping temperature at **1.0** — lower values can cause unexpected looping behavior and degraded performance with the newer reasoning architecture.

## Caching strategy: implicit only for your token counts

Your static content (150 token system instruction + 500 token prompt template = **650 tokens**) falls below Gemini's minimum threshold for explicit context caching:

| Model | Minimum Cache Tokens |
|-------|---------------------|
| Gemini 2.5 Flash | 1,024 |
| Gemini 2.5 Flash-Lite | 1,024 |
| Gemini 2.0 Flash | 2,048 |

**Three practical options exist:**

1. **Rely on implicit caching** (recommended): Gemini 2.5 models automatically enable implicit caching since May 2025. Put your system instruction and prompt template at the beginning of each request, and send similar requests in temporal bursts when possible. Savings are automatic but not guaranteed — you pay standard rates but get cached pricing when cache hits occur.

2. **Expand cacheable content:** If you can semantically justify adding 400+ tokens to your static prompt (additional examples, detailed classification criteria, context about your codebase), you could reach the 1,024 minimum. Cached content pricing is **$0.01/1M tokens** (90% discount) on 2.5 Flash-Lite.

3. **Batch API for offline processing:** For non-real-time classification, the Batch API provides **50% cost reduction** regardless of caching. Upload a JSONL file with multiple diff requests, and results return within 24 hours (often much faster).

Your current SHA256 file-based caching remains valuable for avoiding redundant API calls entirely — it's orthogonal to Gemini's context caching.

## Cost estimation framework

For your use case pattern (650 input tokens base + 1K-50K token diffs + ~500 output tokens):

**Per-request costs on Gemini 2.5 Flash-Lite:**
| Diff Size | Input Tokens | Input Cost | Output Cost | Total |
|-----------|-------------|------------|-------------|-------|
| 1K tokens | 1,650 | $0.000165 | $0.0002 | **$0.000365** |
| 10K tokens | 10,650 | $0.001065 | $0.0002 | **$0.001265** |
| 50K tokens | 50,650 | $0.005065 | $0.0002 | **$0.005265** |

**Monthly estimates at scale:**

| Volume | Avg 10K Diffs | Monthly Cost (2.5 Flash-Lite) | With Batch API (50% off) |
|--------|---------------|-------------------------------|--------------------------|
| 1,000/day | 30K diffs | ~$38 | ~$19 |
| 10,000/day | 300K diffs | ~$380 | ~$190 |
| 100,000/day | 3M diffs | ~$3,800 | ~$1,900 |

Using Gemini 2.5 Flash instead would cost approximately **3x more** ($0.30 input / $2.50 output), while Gemini 3 Flash Preview would cost approximately **5-7x more**.

## Go SDK implementation patterns

Here's a production-ready implementation pattern for your CLI tool:

```go
package gemini

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "math"
    "math/rand"
    "time"

    "google.golang.org/genai"
)

type DiffClassification struct {
    NarrativeType  string   `json:"narrative_type"`
    Confidence     float64  `json:"confidence"`
    AffectedAreas  []string `json:"affected_areas,omitempty"`
    Summary        string   `json:"summary,omitempty"`
}

type Classifier struct {
    client     *genai.Client
    model      string
    config     *genai.GenerateContentConfig
    maxRetries int
}

func NewClassifier(ctx context.Context, apiKey string) (*Classifier, error) {
    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        APIKey:  apiKey,
        Backend: genai.BackendGeminiAPI,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create client: %w", err)
    }

    return &Classifier{
        client:     client,
        model:      "gemini-2.5-flash-lite",
        maxRetries: 5,
        config: &genai.GenerateContentConfig{
            Temperature:      genai.Ptr(float32(0.3)),
            ResponseMIMEType: "application/json",
            SystemInstruction: &genai.Content{
                Parts: []*genai.Part{{Text: systemPrompt}},
            },
            ResponseSchema: classificationSchema(),
        },
    }, nil
}

func classificationSchema() *genai.Schema {
    return &genai.Schema{
        Type: "object",
        Properties: map[string]*genai.Schema{
            "narrative_type": {
                Type:        "string",
                Enum:        []string{"feature", "bugfix", "refactor", "docs", "test", "chore"},
                Description: "Primary classification of the code change",
            },
            "confidence": {
                Type:        "number",
                Description: "Classification confidence from 0.0 to 1.0",
            },
            "affected_areas": {
                Type:        "array",
                Items:       &genai.Schema{Type: "string"},
                Description: "Code modules or subsystems affected",
            },
            "summary": {
                Type:        "string",
                Description: "Brief description of the change narrative",
            },
        },
        Required:         []string{"narrative_type", "confidence"},
        PropertyOrdering: []string{"narrative_type", "confidence", "affected_areas", "summary"},
    }
}

func (c *Classifier) ClassifyDiff(ctx context.Context, diff string) (*DiffClassification, error) {
    contents := []*genai.Content{{
        Parts: []*genai.Part{{Text: fmt.Sprintf("Classify this code diff:\n\n%s", diff)}},
        Role:  genai.RoleUser,
    }}

    var resp *genai.GenerateContentResponse
    var lastErr error

    for attempt := 0; attempt < c.maxRetries; attempt++ {
        resp, lastErr = c.client.Models.GenerateContent(ctx, c.model, contents, c.config)
        if lastErr == nil {
            break
        }

        if !c.isRetryable(lastErr) {
            return nil, lastErr
        }

        delay := c.backoffDelay(attempt)
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(delay):
            // Continue retry loop
        }
    }

    if lastErr != nil {
        return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
    }

    var result DiffClassification
    if err := json.Unmarshal([]byte(resp.Text()), &result); err != nil {
        return nil, fmt.Errorf("JSON parse error: %w", err)
    }

    return &result, nil
}

func (c *Classifier) isRetryable(err error) bool {
    var apiErr *genai.APIError
    if errors.As(err, &apiErr) {
        switch apiErr.HTTPStatusCode {
        case 429, 500, 503:
            return true
        }
    }
    return false
}

func (c *Classifier) backoffDelay(attempt int) time.Duration {
    baseMs := 1000.0
    maxMs := 60000.0
    delay := math.Min(baseMs*math.Pow(2, float64(attempt)), maxMs)
    jitter := rand.Float64() * baseMs * 0.3
    return time.Duration(delay+jitter) * time.Millisecond
}

const systemPrompt = `You are a code change classifier. Analyze git diffs and classify them into narrative structures.

Classification categories:
- feature: New functionality or capabilities
- bugfix: Corrections to existing behavior
- refactor: Code restructuring without behavior change  
- docs: Documentation updates
- test: Test additions or modifications
- chore: Maintenance tasks, dependencies, configuration

Provide confident classifications with clear reasoning.`
```

## Streaming for perceived latency improvement

For interactive CLI use, **streaming does help perceived latency** even with JSON responses. Gemini streams valid partial JSON chunks that concatenate to form the complete response:

```go
func (c *Classifier) ClassifyDiffStream(ctx context.Context, diff string) (*DiffClassification, error) {
    contents := []*genai.Content{{
        Parts: []*genai.Part{{Text: fmt.Sprintf("Classify this code diff:\n\n%s", diff)}},
        Role:  genai.RoleUser,
    }}

    stream := c.client.Models.GenerateContentStream(ctx, c.model, contents, c.config)
    
    var fullResponse strings.Builder
    for chunk, err := range stream {
        if err != nil {
            return nil, err
        }
        text := chunk.Text()
        fullResponse.WriteString(text)
        // Optional: print progress indicator
    }

    var result DiffClassification
    if err := json.Unmarshal([]byte(fullResponse.String()), &result); err != nil {
        return nil, err
    }
    return &result, nil
}
```

Time-to-first-token for Gemini 2.5 Flash models is approximately **210-370ms** under normal load, with output speeds around 160+ tokens/second.

## Handling large diffs effectively

For 50K+ token diffs, no chunking is required — all Flash models support 1M token context. However, optimize performance by:

- **Placing the classification instruction after the diff content** — Google recommends putting the query at the end of long context for better attention
- **Using explicit section markers** in your diff presentation (file paths, line numbers) to help the model anchor its analysis
- **Pre-filtering irrelevant content** before sending (binary files, lock files, auto-generated code)

For very high-volume scenarios, the **Batch API** accepts JSONL files up to 2GB and processes them with 50% cost savings. Structure batch requests as:

```json
{"key": "diff-001", "request": {"contents": [{"parts": [{"text": "...diff..."}]}], "config": {...}}}
{"key": "diff-002", "request": {"contents": [{"parts": [{"text": "...diff..."}]}], "config": {...}}}
```

## Key implementation recommendations

Based on this research, here are your optimal implementation choices:

**Model:** Start with `gemini-2.5-flash-lite` for production stability and lowest cost. Evaluate `gemini-2.5-flash` if classification accuracy needs improvement, or `gemini-3-flash-preview` if you need cutting-edge code understanding and can tolerate preview instability.

**Structured output:** Use `ResponseSchema` with `ResponseMIMEType: "application/json"` — never embed schema in prompts. For simple single-label classification, consider `text/x.enum` for guaranteed enum outputs.

**Caching:** Your 650-token static content won't benefit from explicit context caching. Rely on implicit caching (automatic on 2.5 models) by keeping static content at prompt start. Your existing SHA256 file-based caching remains valuable for eliminating redundant API calls.

**Error handling:** Implement exponential backoff with jitter for 429/500/503 errors. Maximum 5 retries with 1-60 second delay range. JSON parsing failures with structured output should be rare — focus validation on semantic correctness.

**Temperature:** Keep 0.3 for Gemini 2.x. If upgrading to Gemini 3.x, test with the default 1.0 first.
