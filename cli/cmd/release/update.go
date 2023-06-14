package release

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

var UpdateCmd = &cobra.Command{
	Use:   "update <version>",
	Short: "Update the release PRs",
	Long: `Updates the Gutenberg Mobile PRs for the given version. Also updates the integration PRs per the integration flags
  `,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := normalizeVersion(args[0])

		// TODO: Might want to sure the preios step is current on Gutenberg

		utils.LogInfo("Checking if the Gutenberg Mobile PR is current")
		gbmCurrent := release.IsGbmPrCurrent(version)

		if !gbmCurrent {
			utils.LogInfo("🚨 The Gutenberg Mobile PR is not current, updating Gutenberg")
			setTempDir()
			defer cleanup()
			utils.LogDebug("Directory: %s", TempDir)
			if gbPr, err := release.UpdateGbmPr(version, TempDir, true); err != nil {
				utils.LogError("Error updating gbm PR: %s", err)
				os.Exit(1)
			} else {
				utils.LogInfo("🏁 Gutenberg Mobile release updated, check it out: %s", gbPr.Url)
			}

			if All || Ios || Android {
				updateIntegration(version)
			}

		} else {
			utils.LogInfo("The Gutenberg Mobile PR is current")
		}
	},
}

func init() {
	UpdateCmd.Flags().BoolVarP(&Quite, "quite", "q", false, "silence output")
	UpdateCmd.Flags().BoolVarP(&Ios, "ios", "i", false, "update iOS release")
	UpdateCmd.Flags().BoolVarP(&Android, "android", "a", false, "update Android release")
	UpdateCmd.Flags().BoolVarP(&All, "all", "", false, "prepare both integration PRs")
}
