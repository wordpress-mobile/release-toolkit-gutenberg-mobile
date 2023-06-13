package release

import (
	"fmt"
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
		setTempDir()

		utils.LogInfo("Checking if the Gutenberg Mobile PR is current")
		gbmCurrent := release.IsGbmPrCurrent(version)

		if !gbmCurrent {
			utils.LogInfo("🚨 The Gutenberg Mobile PR is not current, updating Gutenberg")
			utils.LogDebug("Updating Gutenberg is not yet implemented")
			os.Exit(1)
		} else {
			utils.LogInfo("The Gutenberg Mobile PR is current")
		}
		fmt.Println()

		if gbmCurrent {
			utils.LogInfo("Both Release PRs are current, nothing to do!")
			os.Exit(0)
		}
	},
}

func init() {
	UpdateCmd.Flags().BoolVarP(&Quite, "quite", "q", false, "slience output")
}
