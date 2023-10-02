package render

import (
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

var Version string
var HostVersion string
var Message string
var ReleaseDate string
var CheckAztec bool
var Quite bool

// checklistCmd represents the checklist command
var ChecklistCmd = &cobra.Command{
	Use:   "checklist",
	Short: "render the content for the release checklist",
	Long: `
`,
	Run: func(cmd *cobra.Command, args []string) {

		vv := gbm.ValidateVersion(Version)
		if !vv {
			fmt.Printf("%v is not a valid version. Versions must have a `Major.Minor.Patch` form\n", Version)
			os.Exit(1)
		}

		// For now let's assume we should include the Aztec steps unless explicitly checking if the versions are valid.
		// We'll render the aztec steps with the optional
		includeAztec := true
		if CheckAztec {
			includeAztec = gbm.ValidateAztecVersions()
		}

		scheduled := gbm.IsScheduledRelease(Version)

		if ReleaseDate == "" {
			ReleaseDate = gbm.NextReleaseDate()
		}

		releaseUrl := fmt.Sprintf("https://github.com/wordpress-mobile/gutenberg-mobile/releases/new?tag=v%s&target=release/%s&title=Release+%s", Version, Version, Version)

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
				Version, scheduled, ReleaseDate, Message, releaseUrl, HostVersion, !includeAztec, CheckAztec),
			Funcs: template.FuncMap{"RenderAztecSteps": renderAztecSteps},
		}

		result, err := render.RenderTasks(t)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(result)

	},
}

func init() {
	ChecklistCmd.Flags().StringVarP(&Version, "version", "v", "", "release version")
	ChecklistCmd.MarkFlagRequired("version")
	ChecklistCmd.Flags().StringVarP(&Message, "message", "m", "", "release message")
	ChecklistCmd.Flags().StringVarP(&ReleaseDate, "date", "d", "", "release date")
	ChecklistCmd.Flags().BoolVar(&CheckAztec, "a", false, "Check if Aztec config is valid before adding the optional update Aztec section")
	ChecklistCmd.Flags().StringVarP(&HostVersion, "host-version", "V", "X.XX", "host app version")
	ChecklistCmd.Flags().BoolVar(&Quite, "q", true, "Silence command info logging")
}
