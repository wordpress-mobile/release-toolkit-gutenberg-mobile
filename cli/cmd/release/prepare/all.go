package prepare

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/cli/pkg/console"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/cli/pkg/gh"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/cli/pkg/release"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "prepare Gutenberg and Gutenberg Mobile for a mobile release",
	Long:  `Use this command to prepare a Gutenberg and Gutenberg Mobile release PRs`,
	Run: func(cc *cobra.Command, args []string) {
		var err error

		preflight(args)
		defer workspace.Cleanup()

		// Set up separate directories for each repo
		gbDir := filepath.Join(tempDir, "gb")
		err = os.MkdirAll(gbDir, os.ModePerm)
		exitIfError(err, 1)

		gbmDir := filepath.Join(tempDir, "gbm")
		err = os.MkdirAll(gbmDir, os.ModePerm)
		exitIfError(err, 1)

		gbPr := gh.PullRequest{}

		console.Info("Preparing Gutenberg for release %s", version)
		build := release.Build{
			Dir:     gbDir,
			Version: version,
			UseTag:  !noTag,
			Base: gh.Repo{
				Ref: "trunk",
			},
		}

		gbPr, err = release.CreateGbPR(build)
		exitIfError(err, 1)
		console.Info("Finished preparing Gutenberg PR")

		console.Info("Preparing Gutenberg Mobile for release %s", version)

		build = release.Build{
			Dir:     gbmDir,
			Version: version,
		}
		pr, err := release.CreateGbmPR(build)
		exitIfError(err, 1)
		console.Info("Finished preparing Gutenberg Mobile PR")

		console.Info("\nFinished preparing PRs:\n%s\n%s", gbPr.Url, pr.Url)
	},
}
