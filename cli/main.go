package main

import (
	"embed"

	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

//go:embed templates/*
var templatesFS embed.FS

func main() {
	render.TemplateFS = templatesFS
}
