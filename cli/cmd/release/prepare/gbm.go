package prepare

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
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

		cleanup := func() {
			utils.CleanupTempDir(tempDir)
		}

		// Reset the exitIfError to handle the cleanup
		exitIfError = utils.ExitIfErrorHandler(cleanup)
		defer cleanup()

		console.Info("Created temporary directory %s", tempDir)

		pr, err := release.CreateGbmPR(version, tempDir)
		console.Info("Created PR %s", pr.Url)

		console.ExitIfError(err)
	},
}
