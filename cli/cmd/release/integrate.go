package release

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/cmd/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/release/integrate"
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

		cleanup := tempDirCleaner(tempDir)
		defer cleanup()

		// reassign exitIfError to handle the cleanup
		exitIfError = utils.ExitIfErrorHandler(cleanup)

		console.Info("Created temporary directory %s", tempDir)

		ri := integrate.ReleaseIntegration{
			Version:    version,
			BaseBranch: "trunk",
			HeadBranch: fmt.Sprintf("gutenberg/integrate_release_%s", version),
		}

		results := []gh.PullRequest{}

		createAndroidPr := func() {
			androidDir := filepath.Join(tempDir, "android")
			err := os.MkdirAll(androidDir, os.ModePerm)
			exitIfError(err, 1)
			androidRi := ri
			target := integrate.AndroidIntegration{}
			androidRi.Target = target
			pr, err := androidRi.Run(filepath.Join(tempDir, "android"))
			warnIfError(err)
			results = append(results, pr)
		}

		createIosPr := func() {
			iosDir := filepath.Join(tempDir, "ios")
			err = os.MkdirAll(iosDir, os.ModePerm)
			exitIfError(err, 1)

			iosRi := ri
			target := integrate.IosIntegration{}
			iosRi.Target = target
			pr, err := iosRi.Run(filepath.Join(tempDir, "ios"))
			warnIfError(err)
			results = append(results, pr)
		}

		// Integrate GBM into Android and iOS if both flags are set or neither flag is set
		both = !android && !ios || ios && android

		switch {
		case both:
			console.Info("Integrating GBM version %s into both iOS and Android", version)
			createAndroidPr()
			createIosPr()

		case android:
			console.Info("Integrating GBM version %s into Android", version)
			createAndroidPr()

		case ios:
			console.Info("Integrating GBM version %s into iOS", version)
			createIosPr()
		}

		if len(results) == 0 {
			exitIfError(errors.New("no PRs were created"), 1)
		}
		for _, pr := range results {
			console.Info("Created PR %s", pr.Url)
		}
	},
}

func init() {
	IntegrateCmd.Flags().BoolVarP(&android, "android", "a", false, "Only integrate Android")
	IntegrateCmd.Flags().BoolVarP(&ios, "ios", "i", false, "Only integrate iOS")
}
