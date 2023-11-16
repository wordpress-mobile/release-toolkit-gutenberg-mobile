package render

import (
	"fmt"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/render"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/semver"
)

var version string
var hostVersion string
var message string
var releaseDate string
var checkAztec bool

type templateData struct {
	Version            string
	Scheduled          bool
	Date               string
	Message            string
	ReleaseUrl         string
	HostVersion        string
	IncludeAztec       bool
	CheckAztec         bool
	BuildkitReleaseUrl string
}

// checklistCmd represents the checklist command
var ChecklistCmd = &cobra.Command{
	Use:   "checklist",
	Short: "Render the content for the release checklist",
	Long: `
`,
	Run: func(cmd *cobra.Command, args []string) {

		semver, err := semver.NewSemVer(version)
		if err != nil {
			exitIfError(fmt.Errorf("invalid version %s.  Versions must have a `Major.Minor.Patch` form", version), 1)
		}
		version = semver.String()

		// For now let's assume we should include the Aztec steps unless explicitly checking if the versions are valid.
		// We'll render the aztec steps with the optional
		includeAztec := true
		if checkAztec {
			includeAztec = !gbm.ValidateAztecVersions()

			if includeAztec {
				console.Info("Aztec is not set to a stable version. Including the Update Aztec steps.")
			}
		}

		scheduled := semver.IsScheduledRelease()

		if releaseDate == "" {
			releaseDate = nextReleaseDate()
		}

		releaseUrl := fmt.Sprintf("https://github.com/wordpress-mobile/gutenberg-mobile/releases/new?tag=v%s&target=release/%s&title=Release+%s", version, version, version)

		t := render.Template{
			Path:  "templates/checklist/checklist.html",
			Funcs: template.FuncMap{"RenderAztecSteps": renderAztecSteps},
			Data: templateData{
				Version:            version,
				Scheduled:          scheduled,
				ReleaseUrl:         releaseUrl,
				Date:               releaseDate,
				Message:            message,
				HostVersion:        hostVersion,
				IncludeAztec:       includeAztec,
				CheckAztec:         checkAztec,
				BuildkitReleaseUrl: "https://buildkite.com/automattic/gutenberg-mobile/builds?branch=" + semver.Vstring(),
			},
		}

		result, err := render.RenderTasks(t)
		exitIfError(err, 1)

		if writeToClipboard {
			console.Clipboard(result)
		} else {
			console.Out(result)
		}
	},
}

func init() {
	ChecklistCmd.Flags().StringVarP(&version, "version", "v", "", "release version")
	ChecklistCmd.MarkFlagRequired("version")
	ChecklistCmd.Flags().StringVarP(&message, "message", "m", "", "release message")
	ChecklistCmd.Flags().StringVarP(&releaseDate, "date", "d", "", "release date")
	ChecklistCmd.Flags().BoolVar(&checkAztec, "a", false, "Check if Aztec config is valid before adding the optional update Aztec section")
	ChecklistCmd.Flags().StringVarP(&hostVersion, "host-version", "V", "X.XX", "host app version")
}
