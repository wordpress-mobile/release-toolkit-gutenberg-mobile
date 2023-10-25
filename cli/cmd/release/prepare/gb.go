package prepare

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

var gbCmd = &cobra.Command{
	Use:   "gb",
	Short: "Prepare Gutenberg for a mobile release",
	Long:  `Use this command to prepare a Gutenberg release PR`,
	Run: func(cc *cobra.Command, args []string) {
		preflight(args)
		defer workspace.Cleanup()

		console.Info("Preparing Gutenberg for release %s", version)

		pr, err := release.CreateGbPR(version, tempDir, noTag)
		exitIfError(err, 1)

		console.Info("Created PR %s", pr.Url)
	},
}
