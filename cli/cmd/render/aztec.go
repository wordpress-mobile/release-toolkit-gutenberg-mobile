package render

import (
	"github.com/spf13/cobra"
)

var AztecCmd = &cobra.Command{
	Use:   "aztec",
	Short: "render the steps for upgrading Aztec",
	Long: `
`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
}
