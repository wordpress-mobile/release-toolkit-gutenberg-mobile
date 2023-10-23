package release

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/release/prepare"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	wp "github.com/wordpress-mobile/gbm-cli/cmd/workspace"
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
	var err error
	workspace, err = wp.NewWorkspace()
	utils.ExitIfError(err, 1)

	exitIfError = func(err error, code int) {
		if err != nil {
			utils.Exit(code, workspace.Cleanup)
		}
	}
	tempDir = workspace.Dir()
	ReleaseCmd.AddCommand(prepare.PrepareCmd)
	ReleaseCmd.AddCommand(IntegrateCmd)
	ReleaseCmd.AddCommand(StatusCmd)
	ReleaseCmd.PersistentFlags().BoolVar(&keepTempDir, "k", false, "Keep temporary directory after running command")
}
