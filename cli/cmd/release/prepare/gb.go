package prepare

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

var gbCmd = &cobra.Command{
	Use:   "gb",
	Short: "prepare Gutenberg for a mobile release",
	Long:  `Use this command to prepare a Gutenberg release PR`,
	Run: func(cc *cobra.Command, args []string) {
		preflight(args)
		defer workspace.Cleanup()

		if version.IsPatchRelease() {
			console.Info("Preparing Gutenberg for patch release %s", version)
		}

		console.Info("Preparing Gutenberg for release %s", version)

		build := release.Build{
			Dir:     tempDir,
			Version: version,
			Tag:     !noTag,
		}

		pr, err := release.CreateGbPR(build)
		exitIfError(err, 1)

		console.Info("Created PR %s", pr.Url)
	},
}
