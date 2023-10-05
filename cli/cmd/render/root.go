package render

import (
	"os"

	"github.com/spf13/cobra"
)

var writeToClipboard bool

// renderCmd represents the render command
var RootCmd = &cobra.Command{
	Use:   "render",
	Short: "renders various GBM templates",
	Long: `Use this command to render:
	- release checklists
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	RootCmd.AddCommand(ChecklistCmd)

	ChecklistCmd.Flags().BoolVar(&writeToClipboard, "c", false, "Send output to clipboard")
}
