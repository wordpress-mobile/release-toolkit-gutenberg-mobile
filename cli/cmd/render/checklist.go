package render

import (
	"fmt"
	"os"

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

type Checklist struct {
	Version   string
	Scheduled bool
}

func (c *Checklist) Task(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

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

		if CheckAztec {
			vav := gbm.ValidateAztecVersions()

			if !Quite {
				fmt.Fprintln(os.Stderr, "Checking Aztec versions...")

				if vav {
					fmt.Fprintln(os.Stderr, "Aztec looks good. Omitting the optional Aztec release section.")
				} else {

					fmt.Fprintln(os.Stderr, "NOTE: Adding update Aztec section")
				}
			}
		}

		var scheduled string
		if s := gbm.IsScheduledRelease(Version); s {
			scheduled = "true"
		} else {
			scheduled = "false"
		}

		if ReleaseDate == "" {
			ReleaseDate = gbm.NextReleaseDate()
		}

		releaseUrl := fmt.Sprintf("https://github.com/wordpress-mobile/gutenberg-mobile/releases/new?tag=v%s&target=release/%s&title=Release+%s", Version, Version, Version)

		jsonData := fmt.Sprintf(`
			{
				"version": "%s",
				"scheduled": %s,
				"date": "%s",
				"message" : "%s",
				"releaseUrl": "%s",
				"hostVersion": "%s"
			}
			`,
			Version, scheduled, ReleaseDate, Message, releaseUrl, HostVersion)

		renderTask := func(checked bool, format string, args ...interface{}) string {
			t := struct {
				Task    string
				Checked bool
			}{
				Task:    fmt.Sprintf(format, args...),
				Checked: checked,
			}

			res, err := render.Render("templates/checklist/task.html", t, nil)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			return res
		}
		CheckedTask := func(format string, args ...interface{}) string {
			return renderTask(true, format, args)
		}
		Task := func(format string, args ...interface{}) string {
			return renderTask(false, format, args)
		}

		result, err := render.RenderJSON("templates/checklist/checklist.html", jsonData, map[string]any{"Task": Task, "CheckedTask": CheckedTask})
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
