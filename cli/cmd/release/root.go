package release

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/release/prepare"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
)

var exitIfError func(error, int)
var tempDirCleaner func(string) func()
var keepTempDir bool

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
	tempDirCleaner = func(tempDir string) func() {
		return func() {
			if keepTempDir {
				console.Info("Keeping temporary directory %s", tempDir)
				return
			}
			utils.CleanupTempDir(tempDir)
		}
	}
	ReleaseCmd.AddCommand(prepare.PrepareCmd)
	ReleaseCmd.AddCommand(IntegrateCmd)
	ReleaseCmd.AddCommand(StatusCmd)
	ReleaseCmd.PersistentFlags().BoolVar(&keepTempDir, "k", false, "Keep temporary directory after running command")
}

func warnIfError(err error) {
	if err != nil {
		console.Warn(err.Error())
	}
}
