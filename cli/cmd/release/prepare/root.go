package prepare

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
)

var tempDir string
var cleanup func()
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
	cleanup = func() {
		if tempDir != "" {
			utils.CleanupTempDir(tempDir)
		}
	}
	exitIfError = utils.ExitIfErrorHandler(cleanup)

	PrepareCmd.AddCommand(gbmCmd)
	PrepareCmd.AddCommand(gbCmd)
}
