package prepare

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Prepare Gutenberg and Gutenberg Mobile for a mobile release",
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

		gbPr, err = release.CreateGbPR(version, gbDir)
		exitIfError(err, 1)
		console.Info("Finished preparing Gutenberg PR")

		console.Info("Preparing Gutenberg Mobile for release %s", version)

		pr, err := release.CreateGbmPR(version, gbmDir)
		exitIfError(err, 1)
		console.Info("Finished preparing Gutenberg Mobile PR")

		console.Info("\nFinished preparing PRs:\n%s\n%s", gbPr.Url, pr.Url)
	},
}
