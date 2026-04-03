package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"tervdocs/internal/config"
	"tervdocs/internal/util"
)

type GeminiClient struct {
	httpClient *http.Client
	cfg        config.ProviderConfig
	model      string
}

func NewGeminiClient(httpClient *http.Client, cfg config.ProviderConfig, model string) *GeminiClient {
	if model == "" {
		model = cfg.Model
	}
	return &GeminiClient{httpClient: httpClient, cfg: cfg, model: model}
}

func (c *GeminiClient) Name() string { return "gemini" }
func (c *GeminiClient) Validate() error {
	if c.cfg.APIKey == "" {
		return ErrMissingAPIKey
	}
	return nil
}

func (c *GeminiClient) Generate(ctx context.Context, req Request) (Response, error) {
	if err := c.Validate(); err != nil {
		return Response{}, err
	}
	model := req.Model
	if model == "" {
		model = c.model
	}
	baseURL := strings.TrimRight(c.cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	body := map[string]any{
		"contents": []map[string]any{
			{"parts": []map[string]string{{"text": req.SystemPrompt + "\n\n" + req.UserPrompt}}},
		},
	}
	buf, _ := json.Marshal(body)

	var out Response
	err := util.Retry(ctx, 3, func() error {
		u := fmt.Sprintf("%s/models/%s:generateContent?key=%s", baseURL, model, url.QueryEscape(c.cfg.APIKey))
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(buf))
		if err != nil {
			return err
		}
		httpReq.Header.Set("Content-Type", "application/json")
		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			return fmt.Errorf("status %d from gemini", resp.StatusCode)
		}
		var payload struct {
			Candidates []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			} `json:"candidates"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return err
		}
		if len(payload.Candidates) == 0 || len(payload.Candidates[0].Content.Parts) == 0 {
			return ErrEmptyResponse
		}
		out = Response{
			Content:  payload.Candidates[0].Content.Parts[0].Text,
			Provider: c.Name(),
			Model:    model,
		}
		return EnsureNonEmpty(out)
	})
	return out, err
}
