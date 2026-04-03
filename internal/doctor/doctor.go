package doctor

import (
	"context"
	"errors"

	"tervdocs/internal/config"
	"tervdocs/internal/providers"
	"tervdocs/internal/scan"
)

type Report struct {
	Checks   []string
	Warnings []string
	Errors   []string
}

func Run(ctx context.Context, cfg config.Config, root string) Report {
	r := Report{}
	if err := config.Validate(cfg); err != nil {
		r.Errors = append(r.Errors, err.Error())
	} else {
		r.Checks = append(r.Checks, "config shape is valid")
	}

	p, err := providers.New(cfg.Provider, cfg)
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	} else if err := p.Validate(); err != nil {
		if errors.Is(err, providers.ErrMissingAPIKey) {
			r.Errors = append(r.Errors, "provider API key missing")
		} else {
			r.Errors = append(r.Errors, err.Error())
		}
	} else {
		r.Checks = append(r.Checks, "provider credentials/config look valid")
	}

	sc := scan.New(cfg.Scan)
	sum, err := sc.Scan(ctx, root)
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	} else {
		r.Checks = append(r.Checks, "repository scan succeeded")
		if sum.FilesScanned >= cfg.Scan.MaxFiles {
			r.Warnings = append(r.Warnings, "scan reached max file cap; summary may be partial")
		}
		if len(sum.MonorepoHints) > 0 {
			r.Warnings = append(r.Warnings, "monorepo hints detected: "+sum.MonorepoHints[0])
		}
	}
	return r
}
