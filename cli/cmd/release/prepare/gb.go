package prepare

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

var gbCmd = &cobra.Command{
	Use:   "gb",
	Short: "Prepare Gutenberg for a mobile release",
	Long:  `Use this command to prepare a Gutenberg release PR`,
	Run: func(cc *cobra.Command, args []string) {

		version, err := utils.GetVersionArg(args)
		exitIfError(err, 1)

		// Validate Aztec version
		if valid := gbm.ValidateAztecVersions(); !valid {
			exitIfError(errors.New("invalid Aztec versions found"), 1)
		}

		console.Info("Preparing Gutenberg for release %s", version)

		tempDir, err := utils.SetTempDir()
		exitIfError(err, 1)
		cleanup := func() {
			utils.CleanupTempDir(tempDir)
		}
		defer cleanup()

		// Reset the exitIfError to handle the cleanup
		exitIfError = utils.ExitIfErrorHandler(cleanup)

		console.Info("Created temporary directory %s", tempDir)

		pr, err := release.CreateGbPR(version, tempDir)
		exitIfError(err, 1)

		console.Info("Created PR %s", pr.Url)
	},
}
