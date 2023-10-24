package prepare

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

var gbmCmd = &cobra.Command{
	Use:   "gbm",
	Short: "Prepare Gutenberg Mobile release",
	Long:  `Use this command to prepare a Gutenberg Mobile release PR`,
	Run: func(cmd *cobra.Command, args []string) {
		preflight(args)

		console.Info("Preparing Gutenberg Mobile for release %s", version)

		pr, err := release.CreateGbmPR(version, tempDir)
		exitIfError(err, 1)

		console.Info("Created PR %s", pr.Url)
	},
}
