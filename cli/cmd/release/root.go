package release

import (
	"os"

	"github.com/spf13/cobra"
)

// renderCmd represents the render command
var RootCmd = &cobra.Command{
	Use:   "release",
	Short: "release related commands",
	Long: `
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	RootCmd.AddCommand(IntegrateCmd)
	RootCmd.AddCommand(StatusCmd)
}
