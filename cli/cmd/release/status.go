package release

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "get the status of a release",
	Long:  `Use this command to get the status of a release.`,
	Run: func(cmd *cobra.Command, args []string) {
		version, err := utils.GetVersionArg(args)
		exitIfError(err, 1)

		// Get the GBM Pr
		pr, err := gbm.FindGbmReleasePr(version)
		exitIfError(err, 1)
		console.Info("Checking: %s", pr.Title)

		// Get the status for the head sha
		sha := pr.Head.Sha
		console.Info("Getting Gutenberg Builds for sha: %s", sha)
		exitIfError(err, 1)

		androidReady, err := gbm.AndroidGbmBuildPublished(version)
		exitIfError(err, 1)

		iosReady, err := gbm.IosGbmBuildPublished(version)
		exitIfError(err, 1)

		console.Info("Android Build Ready: %v", androidReady)
		console.Info("iOS Build Ready: %v", iosReady)

	},
}
