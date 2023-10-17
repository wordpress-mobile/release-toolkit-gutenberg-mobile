package release

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/release/prepare"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
)

var exitIfError func(error, int)

var ReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release Gutenberg Mobile",
}

func Execute() {
	err := ReleaseCmd.Execute()
	if err != nil {
		console.Error(err)
		os.Exit(1)
	}
}

func init() {
	exitIfError = utils.ExitIfErrorHandler(func() {})
	ReleaseCmd.AddCommand(prepare.PrepareCmd)
	ReleaseCmd.AddCommand(IntegrateCmd)
}
