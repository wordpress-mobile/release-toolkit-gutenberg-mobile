package release

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

var PublishCmd = &cobra.Command{
	Use:   "publish <version>",
	Short: "generate the gutenberg release Prs",
	Long: `
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := normalizeVersion(args[0])

		prs := release.GetReleasePrs(version, "gutenberg-mobile", "gutenberg")
		if len(prs) == 0 {
			utils.LogError("No release PRs found")
			os.Exit(1)
		}
		gbmPr := prs["gutenberg-mobile"]
		gbPr := prs["gutenberg"]

		l("Checking: ")
		l("  - Gutenberg Mobile: %s", gbmPr.Url)
		l("  - Gutenberg: %s\n", gbPr.Url)

		if ready, reasons := release.IsReadyToPublish(version, SkipChecks, !Quite); !ready {
			lWarn("🚨 The release is not ready to be published\n\n Reasons:")
			for _, reason := range reasons {
				lWarn("  - %s", reason)
			}
			os.Exit(1)
		} else {
			utils.LogInfo("🎉 The release is ready to be published")
		}

		l("\nTagging Gutenberg")
		err := release.TagGb(version, !Quite)
		if err != nil {
			l(utils.WarnString("Tagging Gutenberg failed: %s", err))
		}

		l("\nPublishing the release")
		err = release.PublishGbmRelease(version, !Quite)
		if err != nil {
			l(utils.ErrorString("Error publishing the release: %s", err))
			os.Exit(1)
		}

		if Integrate {
			Ios = true
			Android = true
		}

		if Ios || Android {
			l("\nUpdating the integration PRs with the release tag")
			updateIntegration(version)
		}

		org, _ := repo.GetOrg("gutenberg-mobile")
		l("\n🏁 The release has been published, check it out: https://github.com/%s/gutenberg-mobile/releases/tag/v%s", org, version)

	},
}

func init() {
	PublishCmd.Flags().BoolVarP(&Quite, "quite", "q", false, "silence output")
	PublishCmd.Flags().BoolVarP(&SkipChecks, "skip-checks", "", false, "Skip the Check runs on the PR")
	PublishCmd.Flags().BoolVarP(&Integrate, "integrate", "i", false, "update integration PRs with release tag")
	PublishCmd.Flags().BoolVarP(&Android, "android", "", false, "update android pr with release tag")
	PublishCmd.Flags().BoolVarP(&Ios, "ios", "", false, "update ios pr with release tag")
}
