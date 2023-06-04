package render

import (
	"fmt"

	"github.com/spf13/cobra"
)

// renderCmd represents the render command
var RootCmd = &cobra.Command{
	Use:   "render",
	Short: "renders various GBM templates",
	Long: `Use this command to render:
	- release checklists
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("render called")
	},
}

func init() {
	RootCmd.AddCommand(ChecklistCmd)
}
