package release

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

var android, ios, both bool

var IntegrateCmd = &cobra.Command{
	Use:   "integrate",
	Short: "integrate a release",
	Long:  `Use this command to integrate a release. If the android or ios flags are set, only that platform will be integrated. Otherwise, both will be integrated.`,
	Run: func(cmd *cobra.Command, args []string) {
		version, err := utils.GetVersionArg(args)
		exitIfError(err, 1)

		tempDir, err := utils.SetTempDir()
		exitIfError(err, 1)

		cleanup := func() {
			utils.CleanupTempDir(tempDir)
		}

		defer cleanup()
		// reassign exitIfError to handle the cleanup
		exitIfError = utils.ExitIfErrorHandler(cleanup)

		console.Info("Created temporary directory %s", tempDir)

		androidRi := release.ReleaseIntegration{
			Android:    true,
			Version:    version,
			BaseBranch: "trunk",
			HeadBranch: fmt.Sprintf("gutenberg/integrate_release_%s", version),
		}

		iosRi := release.ReleaseIntegration{
			Ios:        true,
			Version:    version,
			BaseBranch: "trunk",
			HeadBranch: fmt.Sprintf("gutenberg/integrate_release_%s", version),
		}

		createPr := func(ri release.ReleaseIntegration) {
			pr, err := release.Integrate(tempDir, ri)
			exitIfError(err, 1)
			console.Info("Created PR %s", pr.Url)
		}

		// Integrate GBM into Android and iOS if both flags are set or neither flag is set
		both = !android && !ios || ios && android

		switch {
		case both:
			console.Info("Integrating GBM version %s into both iOS and Android", version)

			createPr(androidRi)
			createPr(iosRi)

		case android:
			console.Info("Integrating GBM version %s into Android", version)
			createPr(androidRi)

		case ios:
			console.Info("Integrating GBM version %s into iOS", version)
			createPr(iosRi)
		}
	},
}

func init() {
	IntegrateCmd.Flags().BoolVarP(&android, "android", "a", false, "Only integrate Android")
	IntegrateCmd.Flags().BoolVarP(&ios, "ios", "i", false, "Only integrate iOS")
}
