package generate

import (
	"context"
	"fmt"
	"path/filepath"

	"tervdocs/internal/config"
	"tervdocs/internal/output"
	"tervdocs/internal/prompt"
	"tervdocs/internal/providers"
	"tervdocs/internal/scan"
	"tervdocs/internal/summarize"
	"tervdocs/internal/templates"
)

type Options struct {
	RootDir  string
	DryRun   bool
	Preview  bool
	Provider string
	Model    string
	Template string
	Output   string
}

type Result struct {
	Markdown   string
	OutputPath string
	BackupPath string
	Scan       scan.RepoSummary
}

type ProviderFactory func(name string, cfg config.Config) (providers.Provider, error)

type Service struct {
	newProvider ProviderFactory
}

func NewService() *Service {
	return &Service{
		newProvider: providers.New,
	}
}

func NewServiceWithProviderFactory(factory ProviderFactory) *Service {
	if factory == nil {
		factory = providers.New
	}
	return &Service{newProvider: factory}
}

func (s *Service) Run(ctx context.Context, cfg config.Config, opts Options) (Result, error) {
	if opts.Provider != "" {
		cfg.Provider = opts.Provider
	}
	if opts.Model != "" {
		cfg.Model = opts.Model
	}
	if opts.Template != "" {
		cfg.Template = opts.Template
	}
	if opts.Output != "" {
		cfg.Output.File = opts.Output
	}
	if err := config.Validate(cfg); err != nil {
		return Result{}, err
	}

	scanner := scan.New(cfg.Scan)
	repo, err := scanner.Scan(ctx, opts.RootDir)
	if err != nil {
		return Result{}, err
	}
	contextDoc := summarize.Build(repo)
	template, err := templates.Get(cfg.Template)
	if err != nil {
		return Result{}, err
	}
	systemPrompt, userPrompt := prompt.Build(template, contextDoc)

	provider, err := s.newProvider(cfg.Provider, cfg)
	if err != nil {
		return Result{}, err
	}
	if err := provider.Validate(); err != nil {
		return Result{}, err
	}
	resp, err := provider.Generate(ctx, providers.Request{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Model:        cfg.Model,
		Temperature:  cfg.Temperature,
		MaxTokens:    2000,
	})
	if err != nil {
		return Result{}, fmt.Errorf("generation failed: %w", err)
	}

	outPath := config.OutputAbsPath(opts.RootDir, cfg.Output.File)
	writeRes, err := output.Write(outPath, resp.Content, cfg.Output.Backup, opts.DryRun || opts.Preview)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Markdown:   resp.Content,
		OutputPath: filepath.Clean(writeRes.Path),
		BackupPath: writeRes.BackupPath,
		Scan:       repo,
	}, nil
}
