package release

import (
	"os"

	"github.com/cli/go-gh/pkg/tableprinter"
	"github.com/cli/go-gh/pkg/term"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
)

// renderCmd represents the render command
var StatusCmd = &cobra.Command{
	Use:   "status <version>",
	Short: "Show the status of a release",
	Long: `
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := normalizeVersion(args[0])
		gbmPr, err := repo.GetGbmReleasePr(version)
		if err != nil {
			utils.LogError("%v", err)
			os.Exit(1)
		}

		// Get the PRs linked to the release pr
		rfs := []repo.RepoFilter{
			repo.BuildRepoFilter("gutenberg", "is:open", "is:pr", `label:"Mobile App - i.e. Android or iOS"`),
			repo.BuildRepoFilter("WordPress-Android", "is:open", "is:pr"),
			repo.BuildRepoFilter("WordPress-iOS", "is:open", "is:pr"),
		}
		results, err := repo.FindGbmSyncedPrs(gbmPr, rfs)

		if err != nil {
			utils.LogError("%v", err)
		}

		// Setup table renderer
		terminal := term.FromEnv()
		termWidth, _, _ := terminal.Size()
		t := tableprinter.New(terminal.Out(), terminal.IsTerminalOutput(), termWidth)

		cyan := color.New(color.FgCyan).Add(color.Underline).SprintFunc()
		gray := color.New(color.FgHiBlack).SprintFunc()

		renderRow(t, cyan("Repo"), cyan("PR"), cyan("State"), cyan("Draft"), cyan("Mergeable"))
		renderRow(t, "gutenberg-mobile", gbmPr.Url, gbmPr.State, draftStr(gbmPr), mergeableStr(gbmPr))

		// Render results in a table
		for _, res := range results {
			for _, pr := range res.Items {
				renderRow(t, res.Filter.Repo, pr.Url, pr.State, draftStr(pr), mergeableStr(pr))
			}
			if len(res.Items) == 0 {
				renderRow(t, gray(res.Filter.Repo), gray("N/A"), gray("N/A"), gray("N/A"), gray("N/A"))
			}
		}
		t.Render()
	},
}

func init() {

}

func renderRow(t tableprinter.TablePrinter, rp, url, state, draft, mergeable string) {
	t.AddField(rp)
	t.AddField(url)
	t.AddField(state)
	t.AddField(draft)
	t.AddField(mergeable)
	t.EndRow()
}

func mergeableStr(repo repo.PullRequest) string {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	if repo.Mergeable {
		return green("✔️")
	}
	return red("❌")
}

func draftStr(repo repo.PullRequest) string {
	green := color.New(color.FgGreen).SprintFunc()
	gray := color.New(color.FgHiBlack).SprintFunc()
	if repo.Draft {
		return gray("draft")
	}
	return green("open")
}
