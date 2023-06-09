package release

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

var (
	Gbm  bool
	Apps bool
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

		var err error

		gbpr, _ := release.CreateGbPR(version, TempDir, true)

		utils.LogInfo("🏁 Gutenberg release ready to go, check it out: %s", gbpr.Url)

		utils.LogDebug("✔️ Done with %s", TempDir)

		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	IntegrateCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}
