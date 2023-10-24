package release

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "get the status of a release",
	Long:  `Use this command to get the status of a release.`,
	Run: func(cmd *cobra.Command, args []string) {
		version, err := utils.GetVersionArg(args)
		exitIfError(err, 1)

		// Print styles
		heading := console.Heading
		headingRow := console.HeadingRow
		row := console.Row

		console.Print(heading, "\nRelease %s Status\n", version)

		prs := []gh.PullRequest{}
		gbPr, gbmPr, androidPr, iosPr := gh.PullRequest{}, gh.PullRequest{}, gh.PullRequest{}, gh.PullRequest{}

		// @TODO: search for gb pr
		gbPr.Repo = repo.GetOrg("gutenberg") + "/gutenberg"
		prs = append(prs, gbPr)

		gbmPr, err = gbm.FindGbmReleasePr(version)
		exitIfError(err, 1)
		gbmPr.Repo = repo.GetOrg("gutenberg-mobile") + "/gutenberg-mobile"
		prs = append(prs, gbmPr)

		androidPr, err = gbm.FindAndroidReleasePr(version)
		exitIfError(err, 1)
		androidPr.Repo = repo.GetOrg("WordPress-Android") + "/WordPress-Android"
		prs = append(prs, androidPr)

		iosPr, err = gbm.FindIosReleasePr(version)
		exitIfError(err, 1)
		iosPr.Repo = repo.GetOrg("WordPress-iOS") + "/WordPress-iOS"
		prs = append(prs, iosPr)

		console.Print(heading, "Release Prs:")
		console.Print(headingRow, "%-27s %-10s %-10v %s", "Repo", "State", "Mergeable", "Url")

		// List the PRs
		for _, pr := range prs {
			if pr.Number == 0 {
				pr.State = "…"
				pr.Url = "…"
			}
			console.Print(row, "• %-25s %-10s %-10v %s", pr.Repo, pr.State, pr.Mergeable, pr.Url)
		}

		// Get the status for the head sha

		console.Print(heading, "\nGutenberg Mobile Build Status")
		if gbmPr.Number == 0 {
			console.Info("...Waiting for Gutenberg Mobile PR to be created before checking build status")
			return
		}
		sha := gbmPr.Head.Sha
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
