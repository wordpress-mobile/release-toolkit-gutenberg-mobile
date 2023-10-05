package render

import (
	"fmt"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

var version string
var hostVersion string
var message string
var releaseDate string
var checkAztec bool

// checklistCmd represents the checklist command
var ChecklistCmd = &cobra.Command{
	Use:   "checklist",
	Short: "render the content for the release checklist",
	Long: `
`,
	Run: func(cmd *cobra.Command, args []string) {

		vv := gbm.ValidateVersion(version)
		if !vv {
			console.ExitError(1, "%v is not a valid version. Versions must have a `Major.Minor.Patch` form", version)
		}

		// For now let's assume we should include the Aztec steps unless explicitly checking if the versions are valid.
		// We'll render the aztec steps with the optional
		includeAztec := true
		if checkAztec {
			includeAztec = !gbm.ValidateAztecVersions()

			if includeAztec {
				console.Info("Aztec is not set to a stable version. Including the Update Aztec steps.")
			}
		}

		scheduled := gbm.IsScheduledRelease(version)

		if releaseDate == "" {
			releaseDate = gbm.NextReleaseDate()
		}

		releaseUrl := fmt.Sprintf("https://github.com/wordpress-mobile/gutenberg-mobile/releases/new?tag=v%s&target=release/%s&title=Release+%s", version, version, version)

		t := render.Template{
			Path: "templates/checklist/checklist.html",
			Json: fmt.Sprintf(`
				{
					"version": "%s",
					"scheduled": %v,
					"date": "%s",
					"message" : "%s",
					"releaseUrl": "%s",
					"hostVersion": "%s",
					"includeAztec": %v,
					"checkAztec": %v}
				`,
				version, scheduled, releaseDate, message, releaseUrl, hostVersion, includeAztec, checkAztec),
			Funcs: template.FuncMap{"RenderAztecSteps": renderAztecSteps},
		}

		result, err := render.RenderTasks(t)
		console.ExitIfError(err)

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
