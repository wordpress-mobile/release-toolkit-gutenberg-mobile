package release

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

var PrepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare a release",
	Long:  `Use this command to prepare a release`,
	Run: func(cmd *cobra.Command, args []string) {
		version, err := getVersionArg(args)
		console.ExitIfError(err)

		console.Info("Preparing release for version %s", version)

		tempDir, err := utils.SetTempDir()
		console.ExitIfError(err)

		console.Info("Created temporary directory %s", tempDir)

		// defer utils.CleanupTempDir(tempDir)

		_, err = release.CreateGbPR(version, tempDir)
		console.ExitIfError(err)

	},
}

func init() {

}
