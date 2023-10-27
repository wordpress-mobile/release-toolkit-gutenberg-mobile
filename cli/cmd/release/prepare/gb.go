package prepare

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

var gbCmd = &cobra.Command{
	Use:   "gb",
	Short: "prepare Gutenberg for a mobile release",
	Long:  `Use this command to prepare a Gutenberg release PR`,
	Run: func(cc *cobra.Command, args []string) {
		preflight(args)

		defer workspace.Cleanup()
		build := release.Build{
			Dir:     tempDir,
			Version: version,
			Tag:     !noTag,
		}

		if version.IsPatchRelease() {
			console.Info("Preparing Gutenberg for patch release %s", version)
			build.Prs = gh.GetPrs("gutenberg", prs)

			if len(build.Prs) == 0 {
				exitIfError(errors.New("no PRs found for patch release"), 1)
				return
			}

			exitIfError(errors.New("not implemented yet"), 1)
		}

		console.Info("Preparing Gutenberg for release %s", version)

		pr, err := release.CreateGbPR(build)
		exitIfError(err, 1)

		console.Info("Created PR %s", pr.Url)
	},
}
