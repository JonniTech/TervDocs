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
	}
	buf, _ := json.Marshal(payload)

	var out Response
	err := util.Retry(ctx, 3, func() error {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/chat/completions", bytes.NewReader(buf))
		if err != nil {
			return err
		}
		httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
		httpReq.Header.Set("Content-Type", "application/json")
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			return fmt.Errorf("status %d from free provider", resp.StatusCode)
		}
		var body struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			return err
		}
		if len(body.Choices) == 0 {
			return ErrEmptyResponse
		}
		out = Response{
			Content:  body.Choices[0].Message.Content,
			Provider: c.Name(),
			Model:    model,
		}
		return EnsureNonEmpty(out)
	})
	return out, err
}
