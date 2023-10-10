package release

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
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

		// Validate Aztec version
		if valid := gbm.ValidateAztecVersions(); !valid {
			console.ExitError("Aztec versions are not valid")
		}

		console.Info("Preparing release for version %s", version)

		tempDir, err := utils.SetTempDir()
		console.ExitIfError(err)

		defer utils.CleanupTempDir(tempDir)

		console.Info("Created temporary directory %s", tempDir)

		pr, err := release.CreateGbPR(version, tempDir)
		console.ExitIfError(err)

		console.Info("Created PR  %s", pr.Url)
	},
}
