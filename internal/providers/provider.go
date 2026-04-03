package providers

import (
	"context"
	"errors"
	"fmt"
)

type Request struct {
	SystemPrompt string
	UserPrompt   string
	Model        string
	Temperature  float64
	MaxTokens    int
}

type Response struct {
	Content  string
	Provider string
	Model    string
}

type Provider interface {
	Name() string
	Validate() error
	Generate(context.Context, Request) (Response, error)
}

var (
	ErrMissingAPIKey = errors.New("missing API key")
	ErrEmptyResponse = errors.New("provider returned empty response")
)

type StatusError struct {
	Provider   string
	StatusCode int
	Body       string
}

func (e StatusError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("status %d from %s provider", e.StatusCode, e.Provider)
	}
	return fmt.Sprintf("status %d from %s provider: %s", e.StatusCode, e.Provider, e.Body)
}

func IsRateLimited(err error) bool {
	var statusErr StatusError
	return errors.As(err, &statusErr) && statusErr.StatusCode == 429
}

func EnsureNonEmpty(r Response) error {
	if r.Content == "" {
		return ErrEmptyResponse
	}
	return nil
}

func WrapProviderError(name string, err error) error {
	return fmt.Errorf("%s provider error: %w", name, err)
}
