package release

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/internal/gh"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

// checklistCmd represents the checklist command
var IntegrateCmd = &cobra.Command{
	Use:   "integrate <version>",
	Short: "generate the release integration Pr",
	Long: `
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := normalizeVersion(args[0])

		// Set up the integration operations based on the update flag\

		var (
			results []releaseResult
			message string
		)
		if Update {
			results = updateIntegration(version)
			message = "updating"
		} else {
			results = createIntegration(version)
			message = "creating"
		}

		for _, r := range results {
			if r.err != nil {
				utils.LogError("Error %s %s PR: %s", message, r.repo, r.err)
			} else {
				utils.LogInfo("Finished %s %s PR: %s", message, r.repo, r.pr.Url)
			}
		}
	},
}

func init() {
	IntegrateCmd.Flags().BoolVarP(&Ios, "ios", "i", false, "target ios release")
	IntegrateCmd.Flags().BoolVarP(&Android, "android", "a", false, "target android release")
	IntegrateCmd.Flags().BoolVarP(&Update, "update", "u", false, "update existing PR")
	IntegrateCmd.Flags().StringVarP(&BaseBranch, "base-branch", "b", "trunk", "base branch for the PR")
	IntegrateCmd.Flags().BoolVarP(&Quite, "quite", "q", false, "silence output")
}

func createIntegration(version string) (results []releaseResult) {
	return integrate(version, release.CreateAndroidPr, release.CreateIosPr, true)
}

func updateIntegration(version string) (results []releaseResult) {
	return integrate(version, release.UpdateAndroidPr, release.UpdateIosPr, false)
}

func integrate(version string, androidOp, iosOp release.IntegrateOp, updateGBM bool) (results []releaseResult) {

	gbmPr, err := release.GetGbmReleasePr(version)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rChan := make(chan releaseResult)

	setTempDir()
	defer cleanup()

	// if neither ios or android are specified, default to both
	if !Ios && !Android {
		Ios = true
		Android = true
	}

	numPr := 0 // number of PRs to create

	// Use goroutines to create the PRs concurrently
	if Ios {
		numPr++
		utils.LogInfo("Working on iOS PR at %s/Wordpress-iOS", repo.WpMobileOrg)
		go func() {
			pr, err := iosOp(version, BaseBranch, TempDir, *gbmPr, !Quite)
			rChan <- releaseResult{"WordPress-iOS", pr, err}
		}()
	}

	if Android {
		numPr++
		utils.LogInfo("Working on Android PR at %s/WordPress-Android", repo.WpMobileOrg)
		go func() {
			pr, err := androidOp(version, BaseBranch, TempDir, *gbmPr, !Quite)
			rChan <- releaseResult{"WordPress-Android", pr, err}
		}()
	}

	for i := 0; i < numPr; i++ {
		r := <-rChan
		results = append(results, r)
	}

	// if we're updating GBM, do that now
	if updateGBM {
		l("Updating GBM PR with integration PRs")
		if err := release.RenderGbmBody("", gbmPr); err != nil {
			l(utils.WarnString("Unable to update the GBM PR body: %s", err))
			return results
		}

		if err := gh.UpdatePr(gbmPr); err != nil {
			l(utils.WarnString("Unable to update the GBM PR: %s", err))
		}

	}

	return results
}
