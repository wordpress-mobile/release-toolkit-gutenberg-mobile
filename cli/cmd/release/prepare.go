package release

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
)

var PrepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare a release",
	Long:  `Use this command to prepare a release`,
	Run: func(cmd *cobra.Command, args []string) {
		version, err := getVersionArg(args)
		console.ExitIfError(err)

		console.Info("Preparing release for version %s", version)
	},
}

func init() {

}
