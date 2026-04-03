package providers

import (
	"fmt"
	"net/http"
	"time"

	"tervdocs/internal/config"
)

func New(name string, cfg config.Config) (Provider, error) {
	httpClient := &http.Client{Timeout: time.Duration(cfg.TimeoutSec) * time.Second}
	switch name {
	case "free":
		return NewFreeClient(httpClient, cfg.Providers.Free, cfg.Template), nil
	case "openai":
		return NewOpenAIClient(httpClient, cfg.Providers.OpenAI, cfg.Model), nil
	case "gemini":
		return NewGeminiClient(httpClient, cfg.Providers.Gemini, cfg.Model), nil
	case "claude":
		return NewClaudeClient(httpClient, cfg.Providers.Claude, cfg.Model), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
}
