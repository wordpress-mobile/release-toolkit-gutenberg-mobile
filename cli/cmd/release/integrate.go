package release

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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

		results := integrate(version)

		for _, r := range results {
			if r.err != nil {
				utils.LogError("Error creating %s PR: %s", r.repo, r.err)
			} else {
				utils.LogInfo("Created %s PR: %s", r.repo, r.pr.Url)
			}
		}

	},
}

func init() {
	IntegrateCmd.Flags().BoolVarP(&Ios, "ios", "i", false, "target ios release")
	IntegrateCmd.Flags().BoolVarP(&Android, "android", "a", false, "target android release")
	IntegrateCmd.Flags().BoolVarP(&Update, "update", "u", false, "update existing PR")
	IntegrateCmd.Flags().StringVarP(&BaseBranch, "base-branch", "b", "trunk", "base branch for the PR")
	IntegrateCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}

func integrate(version string) (results []releaseResult) {

	gbmPr, err := repo.GetGbmReleasePr(version)
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
	} else {
		// if we are only doing one, set verbose to true
		Verbose = true
	}
	numPr := 0 // number of PRs to create

	// Use goroutines to create the PRs concurrently
	if Ios {
		numPr++
		utils.LogInfo("Creating iOS PR at %s/Wordpress-iOS", repo.WpMobileOrg)
		go func() {
			pr, err := release.CreateIosPr(version, BaseBranch, TempDir, gbmPr, Verbose)
			rChan <- releaseResult{"WordPress-iOS", pr, err}
		}()
	}

	if Android {
		numPr++
		utils.LogInfo("Creating Android PR at %s/WordPress-Android", repo.WpMobileOrg)
		go func() {
			pr, err := release.CreateAndroidPr(version, BaseBranch, TempDir, gbmPr, Verbose)
			rChan <- releaseResult{"WordPress-Android", pr, err}
		}()
	}

	for i := 0; i < numPr; i++ {
		r := <-rChan
		results = append(results, r)
	}

	return results
}
