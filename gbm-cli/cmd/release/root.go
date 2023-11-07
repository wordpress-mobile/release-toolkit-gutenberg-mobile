package release

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/cmd/release/prepare"
	wp "github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/cmd/workspace"
)

var exitIfError func(error, int)
var keepTempDir bool
var tempDir string
var workspace wp.Workspace

var ReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release Gutenberg Mobile",
}

func Execute() {
	err := ReleaseCmd.Execute()
	exitIfError(err, 1)

	if keepTempDir {
		workspace.Keep()
	}

	defer workspace.Cleanup()
}

func init() {

	ReleaseCmd.AddCommand(prepare.PrepareCmd)
	ReleaseCmd.AddCommand(IntegrateCmd)
	ReleaseCmd.AddCommand(StatusCmd)
	ReleaseCmd.PersistentFlags().BoolVar(&keepTempDir, "keep", false, "Keep temporary directory after running command")
}
