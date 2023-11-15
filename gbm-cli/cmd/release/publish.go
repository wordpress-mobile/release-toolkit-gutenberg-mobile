package release

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/cmd/utils"
	wp "github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/cmd/workspace"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/release"
)

var PublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "publish a release",
	Long:  `Use this command to publish a release.`,
	Run: func(cmd *cobra.Command, args []string) {

		if keepTempDir {
			workspace.Keep()
		}
		version, err := utils.GetVersionArg(args)
		exitIfError(err, 1)

		r, err := release.Publish(version, tempDir)
		exitIfError(err, 1)

		if r.Url != "" {
			console.Info("Release published: %s", r.Url)
		}
	},
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
	tempDir = workspace.Dir()
}
