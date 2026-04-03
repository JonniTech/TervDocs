package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"tervdocs/internal/config"
	"tervdocs/internal/util"
)

type ClaudeClient struct {
	httpClient *http.Client
	cfg        config.ProviderConfig
	model      string
}

func NewClaudeClient(httpClient *http.Client, cfg config.ProviderConfig, model string) *ClaudeClient {
	if model == "" {
		model = cfg.Model
	}
	return &ClaudeClient{httpClient: httpClient, cfg: cfg, model: model}
}

func (c *ClaudeClient) Name() string { return "claude" }
func (c *ClaudeClient) Validate() error {
	if c.cfg.APIKey == "" {
		return ErrMissingAPIKey
	}
	return nil
}

func (c *ClaudeClient) Generate(ctx context.Context, req Request) (Response, error) {
	if err := c.Validate(); err != nil {
		return Response{}, err
	}
	model := req.Model
	if model == "" {
		model = c.model
	}
	baseURL := strings.TrimRight(c.cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}
	body := map[string]any{
		"model":      model,
		"max_tokens": req.MaxTokens,
		"system":     req.SystemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": req.UserPrompt},
		},
	}
	buf, _ := json.Marshal(body)

	var out Response
	err := util.Retry(ctx, 3, func() error {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/messages", bytes.NewReader(buf))
		if err != nil {
			return err
		}
		httpReq.Header.Set("x-api-key", c.cfg.APIKey)
		httpReq.Header.Set("anthropic-version", "2023-06-01")
		httpReq.Header.Set("content-type", "application/json")

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			return fmt.Errorf("status %d from claude", resp.StatusCode)
		}
		var payload struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return err
		}
		if len(payload.Content) == 0 {
			return ErrEmptyResponse
		}
		out = Response{
			Content:  payload.Content[0].Text,
			Provider: c.Name(),
			Model:    model,
		}
		return EnsureNonEmpty(out)
	})
	return out, err
}
