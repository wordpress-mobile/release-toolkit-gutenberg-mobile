package prepare

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	"github.com/wordpress-mobile/gbm-cli/cmd/workspace"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
)

var exitIfError func(error, int)
var keepTempDir bool
var tempDir string

var PrepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare for a release",
}

func Execute() {
	err := PrepareCmd.Execute()
	exitIfError(err, 1)
	if keepTempDir {
		console.Debug("I should not clean up the temp dir")
	}
	defer workspace.Cleanup()
}

func init() {
	exitIfError = func(err error, code int) {
		if err != nil {
			utils.Exit(code, workspace.Cleanup)
		}
	}
	tempDir = workspace.GetTempDir()
	PrepareCmd.AddCommand(gbmCmd)
	PrepareCmd.AddCommand(gbCmd)
	PrepareCmd.PersistentFlags().BoolVar(&keepTempDir, "k", false, "Keep temporary directory after running command")
}
