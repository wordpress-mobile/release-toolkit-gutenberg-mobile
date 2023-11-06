package prepare

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/cli/pkg/console"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/cli/pkg/gh"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/cli/pkg/release"
)

var gbmCmd = &cobra.Command{
	Use:   "gbm",
	Short: "prepare Gutenberg Mobile release",
	Long:  `Use this command to prepare a Gutenberg Mobile release PR`,
	Run: func(cmd *cobra.Command, args []string) {
		preflight(args)
		defer workspace.Cleanup()

		console.Info("Preparing Gutenberg Mobile for release %s", version)

		build := release.Build{
			Dir:     tempDir,
			Version: version,
			Base: gh.Repo{
				Ref: "trunk",
			},
		}

		pr, err := release.CreateGbmPR(build)
		exitIfError(err, 1)

		console.Info("Created PR %s", pr.Url)
	},
}
