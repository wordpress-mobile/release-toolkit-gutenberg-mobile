package release

import "github.com/spf13/cobra"

var PrepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare a release",
	Long:  `Use this command to prepare a release`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {

}
