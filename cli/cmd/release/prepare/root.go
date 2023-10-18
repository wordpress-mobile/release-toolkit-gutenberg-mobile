package prepare

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
)

var exitIfError func(error, int)
var tempDirCleaner func(string) func()
var keepTempDir bool

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
	tempDirCleaner = func(tempDir string) func() {
		return func() {
			if keepTempDir {
				console.Info("Keeping temporary directory %s", tempDir)
				return
			}
			utils.CleanupTempDir(tempDir)
		}
	}
	PrepareCmd.AddCommand(gbmCmd)
	PrepareCmd.AddCommand(gbCmd)
	PrepareCmd.PersistentFlags().BoolVar(&keepTempDir, "k", false, "Keep temporary directory after running command")
}
