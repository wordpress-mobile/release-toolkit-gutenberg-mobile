package prepare

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	wp "github.com/wordpress-mobile/gbm-cli/cmd/workspace"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/gbm-cli/pkg/semver"
)

var exitIfError func(error, int)
var keepTempDir, noTag bool
var workspace wp.Workspace
var tempDir string
var version semver.SemVer
var prs []string

var PrepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare for a release",
}

func Execute() {
	err := PrepareCmd.Execute()
	exitIfError(err, 1)
}

// Set up the temp directory and version
// Also validate Aztec versions
func preflight(args []string) {
	var err error
	tempDir = workspace.Dir()

	version, err = utils.GetVersionArg(args)
	exitIfError(err, 1)

	// Validate Aztec version
	if valid := gbm.ValidateAztecVersions(); !valid {
		exitIfError(errors.New("invalid Aztec versions found"), 1)
	}
	if keepTempDir {
		workspace.Keep()
	}
}

func init() {
	var err error
	workspace, err = wp.NewWorkspace()
	utils.ExitIfError(err, 1)

	exitIfError = func(err error, code int) {
		if err != nil {
			console.Error(err)
			utils.Exit(code, workspace.Cleanup)
		}
	}

	PrepareCmd.AddCommand(gbmCmd)
	PrepareCmd.AddCommand(gbCmd)
	PrepareCmd.AddCommand(allCmd)
	PrepareCmd.PersistentFlags().BoolVar(&keepTempDir, "k", false, "Keep temporary directory after running command")
	PrepareCmd.PersistentFlags().BoolVar(&noTag, "no-tag", false, "Prevent tagging the release")
	PrepareCmd.PersistentFlags().StringArrayVar(&prs, "prs", []string{}, "prs to include in the release. Only used with patch releases")
}
