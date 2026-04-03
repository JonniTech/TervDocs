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

type OpenAIClient struct {
	httpClient *http.Client
	cfg        config.ProviderConfig
	model      string
}

func NewOpenAIClient(httpClient *http.Client, cfg config.ProviderConfig, model string) *OpenAIClient {
	if model == "" {
		model = cfg.Model
	}
	return &OpenAIClient{httpClient: httpClient, cfg: cfg, model: model}
}

func (c *OpenAIClient) Name() string { return "openai" }
func (c *OpenAIClient) Validate() error {
	if c.cfg.APIKey == "" {
		return ErrMissingAPIKey
	}
	return nil
}

func (c *OpenAIClient) Generate(ctx context.Context, req Request) (Response, error) {
	if err := c.Validate(); err != nil {
		return Response{}, err
	}
	baseURL := strings.TrimRight(c.cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	model := req.Model
	if model == "" {
		model = c.model
	}

	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	body := map[string]any{
		"model": model,
		"messages": []message{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.UserPrompt},
		},
		"temperature": req.Temperature,
	}
	buf, _ := json.Marshal(body)

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
			return fmt.Errorf("status %d from openai", resp.StatusCode)
		}
		var payload struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return err
		}
		if len(payload.Choices) == 0 {
			return ErrEmptyResponse
		}
		out = Response{
			Content:  payload.Choices[0].Message.Content,
			Provider: c.Name(),
			Model:    model,
		}
		return EnsureNonEmpty(out)
	})
	return out, err
}
