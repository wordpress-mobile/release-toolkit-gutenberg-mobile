package release

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

// checklistCmd represents the checklist command
var PrepareCmd = &cobra.Command{
	Use:   "prepare <version>",
	Short: "generate the gutenberg release Prs",
	Long: `
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := args[0]

		setTempDir()

		results := []releaseResult{}

		var err error

		runIntegration := Apps || Android || Ios

		if Gbm && runIntegration {
			utils.LogInfo("📦 Running full release pipeline. Let's go! 🚀")
		}

		gbpr, err := release.CreateGbPR(version, TempDir, Verbose)
		results = append(results, releaseResult{
			pr:   gbpr,
			err:  err,
			repo: "gutenberg",
		})

		utils.LogInfo("🏁 Gutenberg release ready to go, check it out: %s", gbpr.Url)

		if Gbm {
			gbmpr, _ := release.CreateGbmPr(version, TempDir, Verbose)

			results = append(results, releaseResult{
				pr:   gbmpr,
				err:  err,
				repo: "gutenberg-mobile",
			})

			utils.LogInfo("🏁 Gutenberg Mobile release ready to go, check it out: %s", gbmpr.Url)
		}

		if runIntegration {
			intResults := integrate(version)
			results = append(results, intResults...)
		}

		for _, r := range results {
			if r.err != nil {
				utils.LogError("Error creating %s PR: %s", r.repo, r.err)
			} else {
				utils.LogInfo("Created %s PR: %s", r.repo, r.pr.Url)
			}
		}

		utils.LogDebug("✔️ Done with %s", TempDir)
	},
}

func init() {
	PrepareCmd.Flags().BoolVarP(&Gbm, "gbm", "", false, "prepare gutenberg mobile pr")
	PrepareCmd.Flags().BoolVarP(&Apps, "integrate", "", false, "prepare ios and android prs")
	PrepareCmd.Flags().BoolVarP(&Android, "android", "", false, "prepare android pr")
	PrepareCmd.Flags().BoolVarP(&Ios, "ios", "", false, "prepare ios pr")
	PrepareCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}
