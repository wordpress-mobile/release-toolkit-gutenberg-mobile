package prepare

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
)

var exitIfError func(error, int)

var PrepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare for a release",
}

func Execute() {
	err := PrepareCmd.Execute()
	exitIfError(err, 1)
}

func init() {
	exitIfError = utils.ExitIfErrorHandler(func() {})

	PrepareCmd.AddCommand(gbmCmd)
	PrepareCmd.AddCommand(gbCmd)
}
