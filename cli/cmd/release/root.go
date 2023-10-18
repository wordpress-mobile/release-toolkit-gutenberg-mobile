package release

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/release/prepare"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	"github.com/wordpress-mobile/gbm-cli/cmd/workspace"
)

var exitIfError func(error, int)
var keepTempDir bool
var tempDir string

var ReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release Gutenberg Mobile",
}

func Execute() {
	err := ReleaseCmd.Execute()
	exitIfError(err, 1)

	defer workspace.Cleanup()
}

func init() {
	exitIfError = utils.ExitIfError
	tempDir = workspace.GetTempDir()
	ReleaseCmd.AddCommand(prepare.PrepareCmd)
	ReleaseCmd.AddCommand(IntegrateCmd)
	ReleaseCmd.PersistentFlags().BoolVar(&keepTempDir, "k", false, "Keep temporary directory after running command")
}
