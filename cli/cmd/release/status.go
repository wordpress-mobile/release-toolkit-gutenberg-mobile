package release

import (
	"os"

	"github.com/spf13/cobra"
)

// renderCmd represents the render command
var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of a release",
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

}
