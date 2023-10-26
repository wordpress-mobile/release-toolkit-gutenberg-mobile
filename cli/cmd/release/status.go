package release

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
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
		basic := color.New(color.FgWhite)

		console.Print(heading, "\nRelease %s Status\n", version)

		// Check to see if the release has been published
		rel, err := release.GetGbmRelease(version)
		exitIfError(err, 1)

		if (rel != gh.Release{}) {
			console.Print(heading, "\n🎉 Release %s has been published!\n %s\n", version, rel.Url)
		}

		prs := []gh.PullRequest{}
		gbPr, gbmPr, androidPr, iosPr := gh.PullRequest{}, gh.PullRequest{}, gh.PullRequest{}, gh.PullRequest{}

		// @TODO: search for gb pr
		gbPr, err = release.FindGbReleasePr(version)
		if err != nil {
			console.Warn("Could not get Gutenberg PR: %s", err)
		}
		gbPr.Repo = repo.GetOrg("gutenberg") + "/gutenberg"
		prs = append(prs, gbPr)

		gbmPr, err = release.FindGbmReleasePr(version)
		if err != nil {
			console.Warn("Could not get Gutenberg Mobile PR: %s", err)
		}
		gbmPr.Repo = repo.GetOrg("gutenberg-mobile") + "/gutenberg-mobile"
		prs = append(prs, gbmPr)

		androidPr, err = release.FindAndroidReleasePr(version)
		if err != nil {
			console.Warn("Could not find Android PR: %s", err)
		}
		androidPr.Repo = repo.GetOrg("WordPress-Android") + "/WordPress-Android"
		prs = append(prs, androidPr)

		iosPr, err = release.FindIosReleasePr(version)
		if err != nil {
			console.Warn("Could not find iOS PR: %s", err)
		}
		iosPr.Repo = repo.GetOrg("WordPress-iOS") + "/WordPress-iOS"
		prs = append(prs, iosPr)

		console.Print(heading, "Release Prs:")
		console.Print(headingRow, "%-36s %-10s %-10v %s", "Repo", "State", "Mergeable", "Url")

		// List the PRs
		for _, pr := range prs {
			if pr.Number == 0 {
				pr.State = "…"
				pr.Url = "…"
			}
			console.Print(row, "• %-34s %-10s %-10v %s", pr.Repo, pr.State, pr.Mergeable, pr.Url)
		}

		// Get the status for the head sha
		console.Print(heading, "\nGutenberg Mobile Build Status\n")
		if gbmPr.Number == 0 {
			console.Print(row, "...Waiting for Gutenberg Mobile PR to be created before checking build status")
			return
		}
		sha := gbmPr.Head.Sha
		console.Print(basic, "Getting Gutenberg Builds for sha: %s", sha)

		console.Print(headingRow, "%-10s %-10s", "Platform", "Status")

		androidReady, err := gbm.AndroidGbmBuildPublished(gbmPr)
		if err != nil {
			console.Warn("Could not get Android build status: %s", err)
		}

		iosReady, err := gbm.IosGbmBuildPublished(gbmPr)

		if err != nil {
			console.Warn("Could not get iOS build status: %s", err)
		}
		console.Print(row, "%-10s %-10v", "Android", androidReady)
		console.Print(row, "%-10s %-10v", "iOS", iosReady)

	},
}
