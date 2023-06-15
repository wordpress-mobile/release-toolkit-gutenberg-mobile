package release

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

// checklistCmd represents the checklist command
var CutCmd = &cobra.Command{
	Use:   "cut <version>",
	Short: "generate the gutenberg release Prs",
	Long: `
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := normalizeVersion(args[0])

		setTempDir()
		defer cleanup()

		results := []releaseResult{}

		var err error

		runAnyIntegration := Android || Ios

		// Before we start let's make sure the someone didn't forget a flag
		if runAnyIntegration && !Gbm {
			cont := utils.Confirm("🤔 You didn't specify --gbm but also included an integration flag. Continuing will only create the Gutenberg PR, are you sure?")
			if !cont {
				utils.LogInfo("👋 Bye!")
				os.Exit(0)
			}
		}

		if All {
			utils.LogInfo("📦 Running full release pipeline. Let's go! 🚀")
		}

		gbPr, err := release.CreateGbPR(version, TempDir, !Quite)
		results = append(results, releaseResult{
			pr:   &gbPr,
			err:  err,
			repo: "gutenberg",
		})

		utils.LogInfo("🏁 Gutenberg release ready to go, check it out: %s", gbPr.Url)

		if Gbm || All {
			// Try sleeping for a second to avoid rate limiting
			// Too fast and the GB Pr might not be ready
			time.Sleep(time.Second)
			gbmPr, err := release.CreateGbmPr(version, TempDir, !Quite)

			if err != nil {
				utils.LogError("Error creating gbm PR: %s", err)
				os.Exit(1)
			}

			results = append(results, releaseResult{
				pr:   &gbmPr,
				err:  err,
				repo: "gutenberg-mobile",
			})

			utils.LogInfo("🏁 Gutenberg Mobile release ready to go, check it out: %s", gbmPr.Url)

			// Run the integrations if we are preparing all or any integration PRs
			if All || runAnyIntegration {
				if cont := utils.Confirm("Ready to create the integration PRs?"); cont {
					intResults := createIntegration(version)
					results = append(results, intResults...)
				}
			}
		}

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
	CutCmd.Flags().BoolVarP(&Gbm, "gbm", "", false, "Cut gutenberg mobile PR")
	CutCmd.Flags().BoolVarP(&All, "all", "", false, "Cut all release PRs")
	CutCmd.Flags().BoolVarP(&Android, "android", "", false, "Cut android PR - requires --gbm")
	CutCmd.Flags().BoolVarP(&Ios, "ios", "", false, "Cut ios pr - requires --gbm")
	CutCmd.Flags().BoolVarP(&Quite, "quite", "q", false, "silence output")
}
