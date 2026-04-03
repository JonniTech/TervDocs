package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"tervdocs/internal/config"
)

type FreeClient struct {
	httpClient *http.Client
	cfg        config.ProviderConfig
	template   string
}

func NewFreeClient(httpClient *http.Client, cfg config.ProviderConfig, template string) *FreeClient {
	return &FreeClient{httpClient: httpClient, cfg: cfg, template: template}
}

func (c *FreeClient) Name() string { return "free" }

func (c *FreeClient) Validate() error {
	if c.cfg.APIKey == "" {
		return ErrMissingAPIKey
	}
	return nil
}

func (c *FreeClient) Generate(ctx context.Context, req Request) (Response, error) {
	if err := c.Validate(); err != nil {
		return Response{}, err
	}
	model := req.Model
	if model == "" {
		model = c.cfg.Model
	}
	if model == "" {
		model = "glm-4.7-flash"
	}
	baseURL := strings.TrimRight(c.cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.z.ai/api/paas/v4"
	}

	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	payload := map[string]any{
		"model": model,
		"messages": []message{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.UserPrompt},
		},
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
	}
	buf, _ := json.Marshal(payload)

	attemptCtx, cancel := c.withFastTimeout(ctx)
	out, err := c.callOnce(attemptCtx, baseURL, buf, model)
	cancel()
	if err == nil || IsRateLimited(err) {
		return out, err
	}
	select {
	case <-ctx.Done():
		return Response{}, ctx.Err()
	case <-time.After(350 * time.Millisecond):
	}
	attemptCtx, cancel = c.withFastTimeout(ctx)
	defer cancel()
	return c.callOnce(attemptCtx, baseURL, buf, model)
}

func (c *FreeClient) callOnce(ctx context.Context, baseURL string, buf []byte, model string) (Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(buf))
	if err != nil {
		return Response{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return Response{}, StatusError{
			Provider:   c.Name(),
			StatusCode: resp.StatusCode,
			Body:       strings.TrimSpace(string(body)),
		}
	}
	var body struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Response{}, err
	}
	if len(body.Choices) == 0 {
		return Response{}, ErrEmptyResponse
	}
	out := Response{
		Content:  body.Choices[0].Message.Content,
		Provider: c.Name(),
		Model:    model,
	}
	return out, EnsureNonEmpty(out)
}

func (c *FreeClient) withFastTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := 12 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < timeout {
			timeout = remaining
		}
	}
	if timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}
