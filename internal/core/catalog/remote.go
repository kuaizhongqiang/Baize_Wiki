package catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
)

// RemoteSummarizer calls an OpenAI-compatible API for summarization.
type RemoteSummarizer struct {
	endpoint string
	model    string
	apiKey   string
	lang     string
}

// openAIRequest is the request body for the OpenAI chat completions API.
type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIResponse is the response from the OpenAI chat completions API.
type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// summarizerOutput is the expected JSON output format from the LLM.
type summarizerOutput struct {
	Summary  string   `json:"summary"`
	Keywords []string `json:"keywords"`
}

// Summarize implements Summarizer for remote backend.
func (r *RemoteSummarizer) Summarize(ctx context.Context, page *model.Page, lang string) (*CatalogResult, error) {
	ext := detectLanguage(page)
	prompt := buildRemotePrompt(page, ext, lang)

	reqBody := openAIRequest{
		Model: r.model,
		Messages: []openAIMessage{
			{Role: "system", Content: "You are a code and documentation analyst. Output only valid JSON."},
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if r.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+r.apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http call: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBytes))
	}

	var apiResp openAIResponse
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	content := apiResp.Choices[0].Message.Content
	content = strings.TrimSpace(content)
	// Strip markdown code fences if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var output summarizerOutput
	if err := json.Unmarshal([]byte(content), &output); err != nil {
		return nil, fmt.Errorf("parse LLM output: %w", err)
	}

	return &CatalogResult{
		Abstract: output.Summary,
		Keywords: output.Keywords,
	}, nil
}

// buildRemotePrompt constructs the prompt for the remote summarizer.
func buildRemotePrompt(page *model.Page, ext, lang string) string {
	var b strings.Builder

	switch ext {
	case "csharp", "go", "python", "javascript/typescript", "rust", "java", "c/c++":
		fmt.Fprintf(&b, "Analyze this %s source file. ", ext)
		b.WriteString("Understand its core type, main responsibilities, and dependencies.\n\n")
		if page.Title != "" {
			fmt.Fprintf(&b, "File title: %s\n", page.Title)
		}
		b.WriteString("---\n")

		// Truncate content to 500KB, take from front (key info is in header)
		content := page.Content
		if len(content) > 500*1024 {
			content = content[:500*1024]
		}
		b.WriteString(content)

		fmt.Fprintf(&b, "\n\n---\nOutput JSON: {\"summary\": \"%s\", \"keywords\": [\"keyword1\", \"keyword2\"]}", langPrompt(lang))

	default:
		// Markdown or text
		fmt.Fprintf(&b, "Summarize this %s document. ", ext)
		if page.Title != "" {
			fmt.Fprintf(&b, "Title: %s\n", page.Title)
		}
		b.WriteString("---\n")

		content := page.Content
		if len(content) > 400*1024 {
			content = content[:400*1024] + "\n... [truncated]"
		}
		b.WriteString(content)

		fmt.Fprintf(&b, "\n\n---\nOutput JSON: {\"summary\": \"%s\", \"keywords\": [\"keyword1\", \"keyword2\"]}", langPrompt(lang))
	}

	return b.String()
}

// langPrompt returns a language hint for the LLM prompt.
func langPrompt(lang string) string {
	switch lang {
	case "zh":
		return "50-100 token Chinese summary, concise and focused on the main purpose. 3-10 Chinese keywords"
	case "en":
		return "50-100 token English summary, concise and focused on the main purpose. 3-10 English keywords"
	default:
		return fmt.Sprintf("50-100 token %s summary. 3-10 %s keywords", lang, lang)
	}
}
