/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"embed"

	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/cli/cmd"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/cli/pkg/render"
)

//go:embed templates/*
var templatesFS embed.FS

func main() {
	render.TemplateFS = templatesFS
	cmd.Execute()
}
