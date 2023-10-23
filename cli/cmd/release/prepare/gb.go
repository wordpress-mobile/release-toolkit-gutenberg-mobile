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
		tempDir := workspace.Dir()
		version, err := utils.GetVersionArg(args)
		exitIfError(err, 1)

		// Validate Aztec version
		if valid := gbm.ValidateAztecVersions(); !valid {
			exitIfError(errors.New("invalid Aztec versions found"), 1)
		}

		console.Info("Preparing Gutenberg for release %s", version)

		pr, err := release.CreateGbPR(version, tempDir)
		exitIfError(err, 1)

		console.Info("Created PR %s", pr.Url)
	},
}
