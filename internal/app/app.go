package app

import "tervdocs/internal/generate"

type Container struct {
	Generator *generate.Service
}

func New() *Container {
	return &Container{
		Generator: generate.NewService(),
	}
}
