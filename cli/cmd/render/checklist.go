package render

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

var Version string
var Message string
var ReleaseDate string

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
		var scheduled string
		if s := utils.IsScheduledRelease(Version); s {
			scheduled = "true"
		} else {
			scheduled = "false"
		}

		if ReleaseDate == "" {
			ReleaseDate = utils.NextReleaseDate()
		}

		releaseUrl := fmt.Sprintf("https://github.com/wordpress-mobile/gutenberg-mobile/releases/new?tag=v%s&target=release/%s&title=Release+%s", Version, Version, Version)

		jsonData := fmt.Sprintf(`
			{
				"version": "%s",
				"scheduled": %s,
				"date": "%s",
				"message" : "%s",
				"releaseUrl": "%s",
				"mobileVersion": ""
			}
			`,
			Version, scheduled, ReleaseDate, Message, releaseUrl)

		renderTask := func(format string, args ...interface{}) string {
			t := struct{ Task string }{
				Task: fmt.Sprintf(format, args...),
			}

			res, err := render.Render("templates/checklist/task.html", t, nil)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			return res
		}
		result, err := render.RenderJSON("templates/checklist/checklist.html", jsonData, map[string]any{"renderTask": renderTask})
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
}
