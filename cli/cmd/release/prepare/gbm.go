package prepare

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

var gbmCmd = &cobra.Command{
	Use:   "gbm",
	Short: "Prepare Gutenberg Mobile release",
	Long:  `Use this command to prepare a Gutenberg Mobile release PR`,
	Run: func(cmd *cobra.Command, args []string) {
		version, err := getVersionArg(args)
		console.ExitIfError(err)

		// Validate Aztec version
		if valid := gbm.ValidateAztecVersions(); !valid {
			console.ExitError("Aztec versions are not valid")
		}

		console.Info("Preparing Gutenberg Mobile for release %s", version)

		tempDir, err := utils.SetTempDir()
		console.ExitIfError(err)

		defer utils.CleanupTempDir(tempDir)

		console.Info("Created temporary directory %s", tempDir)

		// console.ExitError("not implemented")
	},
}
