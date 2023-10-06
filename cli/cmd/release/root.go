package release

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
)

var ReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release Gutenberg Mobile",
}

func Execute() {
	err := ReleaseCmd.Execute()
	console.ExitIfError(err)
}

func init() {
	ReleaseCmd.AddCommand(PrepareCmd)
}
