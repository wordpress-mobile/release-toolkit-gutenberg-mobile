package prepare

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/release"
)

var gbCmd = &cobra.Command{
	Use:   "gb",
	Short: "prepare Gutenberg for a mobile release",
	Long:  `Use this command to prepare a Gutenberg release PR`,
	Run: func(cc *cobra.Command, args []string) {
		// Check if we're trying to do a pre-release
		if version.IsPreRelease() {
			console.Warn("You can't use 'gb' for pre-releases, just 'gbm")
			return
		}

		preflight(args)

		defer workspace.Cleanup()
		build := release.Build{
			Dir:     tempDir,
			Version: version,
			UseTag:  !noTag,
			Repo:    "gutenberg",
			Base: gh.Repo{
				Ref: "trunk",
			},
		}

		if version.IsPatchRelease() {
			console.Info("Preparing a patch release")
			tagName := "rnmobile/" + version.PriorVersion().String()
			setupPatchBuild(tagName, &build)
		}

		console.Info("Preparing Gutenberg for release %s", version)

		pr, err := release.CreateGbPR(build)
		exitIfError(err, 1)

		console.Info("Created PR %s", pr.Url)
	},
}
