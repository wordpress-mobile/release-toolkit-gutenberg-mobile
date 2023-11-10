package prepare

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/release"
)

var gbRef string
var updateStrings bool

var gbmCmd = &cobra.Command{
	Use:   "gbm",
	Short: "prepare Gutenberg Mobile release",
	Long:  `Use this command to prepare a Gutenberg Mobile release PR`,
	Run: func(cmd *cobra.Command, args []string) {
		preflight(args)
		defer workspace.Cleanup()

		console.Info("Preparing Gutenberg Mobile for release %s", version)

		build := release.Build{
			Dir:     tempDir,
			Version: version,
			Base: gh.Repo{
				Ref: "trunk",
			},
			Repo:          "gutenberg-mobile",
			UpdateStrings: true,
		}

		if version.IsPatchRelease() {
			console.Info("Preparing a patch release")
			tagName := version.PriorVersion().Vstring()
			setupPatchBuild(tagName, &build)
		}

		if version.IsPreRelease() {

			// Check to see if the previous version is the latest release
			previousVersion := version.PriorVersion()
			latestRelease, err := gh.GetLatestRelease("gutenberg-mobile")
			if err != nil {
				console.Warn("Could not get latest release: %v", err)
			} else {
				if previousVersion.Vstring() != latestRelease.TagName {
					console.Warn("Looks like the previous version %s is not the latest release, %s. Pre releases should be on minor version higher.", previousVersion.Vstring(), latestRelease.TagName)
					os.Exit(1)
				}
			}

			console.Info("Preparing a pre-release")
			if gbRef == "" {
				exitIfError(errors.New("you must specify a Gutenberg ref via -(r)ef for a pre-release"), 1)
			}
			build.GbRef = gbRef
			build.UpdateStrings = updateStrings
		}

		pr, err := release.CreateGbmPR(build)
		exitIfError(err, 1)

		console.Info("Created PR %s", pr.Url)
	},
}

func init() {
	gbmCmd.Flags().StringVarP(&gbRef, "ref", "r", "", "Gutenberg ref (only used for test and alpha releases)")
	gbmCmd.Flags().BoolVarP(&updateStrings, "update-strings", "u", false, "Update i18n strings files (only used for test and alpha releases)")
}
