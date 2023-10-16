package prepare

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
)

var gbmCmd = &cobra.Command{
	Use:   "gbm",
	Short: "Prepare Gutenberg Mobile release",
	Long:  `Use this command to prepare a Gutenberg Mobile release PR`,
	Run: func(cmd *cobra.Command, args []string) {
		version, err := utils.GetVersionArg(args)
		exitIfError(err, 1)

		// Validate Aztec version
		if valid := gbm.ValidateAztecVersions(); !valid {
			exitIfError(errors.New("the Aztec versions are not valid"), 1)
		}

		console.Info("Preparing Gutenberg Mobile for release %s", version)

		tempDir, err := utils.SetTempDir()
		exitIfError(err, 1)

		defer utils.CleanupTempDir(tempDir)

		console.Info("Created temporary directory %s", tempDir)

		exitIfError(errors.New("not implemented"), 1)
	},
}
