package render

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
)

var writeToClipboard bool
var exitIfError func(error, int)

// rootCmd represents the render command
var RenderCmd = &cobra.Command{
	Use:   "render",
	Short: "Renders various GBM templates",
	Long: `Use this command to render:
	- Release checklists
	- Steps to update Aztec
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	exitIfError = utils.ExitIfError
	RenderCmd.AddCommand(ChecklistCmd)
	RenderCmd.AddCommand(AztecCmd)
	RenderCmd.PersistentFlags().BoolVar(&writeToClipboard, "c", false, "Send output to clipboard")
}
