package release

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

var UpdateCmd = &cobra.Command{
	Use:   "update <version>",
	Short: "Update the Gutenberg and Gutenberg Mobile release PRs",
	Long: `
  `,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := normalizeVersion(args[0])

		// TODO: Might need to make sure the preios step is current on Gutenberg

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

		} else {
			utils.LogInfo("The Gutenberg Mobile PR is current")
		}
	},
}

func init() {
	UpdateCmd.Flags().BoolVarP(&Quite, "quite", "q", false, "silence output")
}
