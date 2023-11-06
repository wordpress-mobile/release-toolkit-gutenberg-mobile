package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/cli/cmd/release"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/cli/cmd/render"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/cli/pkg/console"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gbm",
	Short: "Gutenberg Mobile CLI",
}

func Execute() {
	err := rootCmd.Execute()
	console.ExitIfError(err)
}

func init() {
	// Add the render command
	rootCmd.AddCommand(render.RenderCmd)
	rootCmd.AddCommand(release.ReleaseCmd)
}
