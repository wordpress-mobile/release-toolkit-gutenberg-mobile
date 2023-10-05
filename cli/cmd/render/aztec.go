package render

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
)

var AztecCmd = &cobra.Command{
	Use:   "aztec",
	Short: "render the steps for upgrading Aztec",
	Long: `
`,
	Run: func(cmd *cobra.Command, args []string) {
		result, err := renderAztecSteps(false)

		console.ExitIfError(err)

		if writeToClipboard {
			console.Clipboard(result)
		} else {
			console.Out(result)
		}
	},
}
