package release

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
	"github.com/wordpress-mobile/gbm-cli/pkg/shell"
)

func CollectReleaseChanges(version string, changelog, relnotes []byte) ([]ReleaseChanges, error) {
	changesRe := regexp.MustCompile(`(?s)\d+\.\d+\.\d+(.*?)\d+\.\d+\.\d+`)
	// rnChangesRe := regexp.MustCompile(`(?s)\d+\.\d+\.\d+(.*?)\d+\.\d+\.\d+`)
	prNumbRe := regexp.MustCompile(`\[#(\d+)\]`)
	prOrgRepoNumRe := regexp.MustCompile(`https://github\.com/(\w+)/(\w+)/pull/(\d+)`)
	bracketRe := regexp.MustCompile(`\[.*\]\s*-*`)

	prs := []ReleaseChanges{}

	prFoundAlready := func(num int) bool {
		for _, p := range prs {
			if p.Number == num {
				return true
			}
		}
		return false
	}

	// Get the changes from the Release notes
	match := changesRe.Find(relnotes)
	if match != nil {

		lines := strings.Split(string(match), "\n")

		for _, l := range lines {
			// first check for any prs relative to gutenberg-mobile
			matches := prNumbRe.FindAllStringSubmatch(l, -1)

			if len(matches) == 1 {

				prId, _ := strconv.Atoi(matches[0][1])

				pr, err := gh.GetPr("gutenberg", prId)
				if err != nil {
					console.Warn("There was an issue fetching a gutenberg pr #%d", prId)
					continue
				}
				// Scrub [] from title
				title := bracketRe.ReplaceAllString(pr.Title, "")
				prs = append(prs, ReleaseChanges{Title: title, PrUrl: pr.Url, Number: pr.Number})
			}

			// now look for urls
			matches = prOrgRepoNumRe.FindAllStringSubmatch(l, -1)
			if len(matches) != 0 {
				match := matches[0]
				org := match[1]
				rep := match[2]
				num := match[3]
				prId, _ := strconv.Atoi(num)
				pr, err := gh.GetPrOrg(org, rep, prId)
				if err != nil {
					console.Warn("There was an issue fetching %s/%s/pull/%d", org, rep, prId)
					continue
				}
				// Scrub [] from title
				title := bracketRe.ReplaceAllString(pr.Title, "")
				rc := ReleaseChanges{
					Title:  title,
					PrUrl:  pr.Url,
					Number: pr.Number,
				}
				checkPRforIssues(*pr, &rc)
				prs = append(prs, rc)
			}
		}
	}
	// Get changes from the Change log
	match = changesRe.Find(changelog)

	if match != nil {
		lines := strings.Split(string(match), "\n")

		for _, l := range lines {

			match := prNumbRe.FindAllStringSubmatch(l, -1)

			if len(match) == 1 {
				prId, _ := strconv.Atoi(match[0][1])
				if prFoundAlready(prId) {
					continue
				}
				pr, err := gh.GetPrOrg("WordPress", "gutenberg", prId)
				if err != nil {
					console.Warn("There was an issue fetching a gutenberg pr #%d", prId)
					continue
				}

				title := bracketRe.ReplaceAllString(pr.Title, "")
				rc := ReleaseChanges{
					Title:  title,
					PrUrl:  pr.Url,
					Number: pr.Number,
				}
				checkPRforIssues(*pr, &rc)
				prs = append(prs, rc)

			}
		}
	}
	return prs, nil
}

func checkPRforIssues(pr gh.PullRequest, rc *ReleaseChanges) {
	issueRe := regexp.MustCompile(`(https:\/\/github.com\/.*\/.*\/issues\/\d*)`)

	matches := issueRe.FindAllStringSubmatch(pr.Body, -1)

	for _, m := range matches {
		rc.Issues = append(rc.Issues, m[1])
	}
}

// Updates the release notes by replacing "Unreleased" with
// the new version and adding a new "Unreleased" section
func UpdateReleaseNotes(version, path string) error {
	return readWriteNotes(version, path, releaseNotesUpdater)
}

// See UpdateReleaseNotes
// This handles the string replacement
func releaseNotesUpdater(version string, notes []byte) []byte {
	re := regexp.MustCompile(`(^Unreleased\s*\n)`)

	repl := fmt.Sprintf("$1---\n\n%s\n", version)

	return re.ReplaceAll(notes, []byte(repl))
}

// Updates the change log by replacing "Unreleased" with
// the new version and adding a new "Unreleased" section
func UpdateChangeLog(version, path string) error {
	return readWriteNotes(version, path, changeLogUpdater)
}

// See UpdateChangeLog
// This handles the string replacement
func changeLogUpdater(version string, notes []byte) []byte {

	re := regexp.MustCompile(`(##\s*Unreleased\s*\n)`)

	repl := fmt.Sprintf("$1\n## %s\n", version)

	return re.ReplaceAll(notes, []byte(repl))
}

func readWriteNotes(version, path string, updater func(string, []byte) []byte) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	changeNotes, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	update := updater(version, changeNotes)
	if err != nil {
		return err
	}

	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()
	if _, err := w.Write(update); err != nil {
		return err
	}
	return nil
}

func previewPr(rpo, dir, branchFrom string, pr gh.PullRequest) {
	org := repo.GetOrg(rpo)
	row := console.Row

	console.Print(console.Heading, "\nPr Preview")

	white := color.New(color.FgWhite).SprintFunc()

	console.Print(row, "Repo: %s/%s", white(org), white(rpo))
	console.Print(row, "Title: %s", white(pr.Title))
	console.Print(row, "Body:\n%s", white(pr.Body))
	console.Print(row, "Commits:")

	git := shell.NewGitCmd(shell.CmdProps{Dir: dir, Verbose: true})

	git.Log(branchFrom+"...HEAD", "--oneline", "--no-merges", "-10")
}

func openInEditor(dir string, files []string) error {
	editor := os.Getenv("EDITOR")

	fileArgs := strings.Join(files, " ")

	if editor == "" {
		editor = console.Ask("\nNo $EDITOR set. Enter the command to open your editor:")
	}

	if editor == "" {
		console.Warn("No editor set. Manually edit or verify the following files before continuing:")
		for _, f := range files {
			console.Print(console.Row, f)
		}
		return nil
	}
	if open := console.Confirm(fmt.Sprintf("\nOpen '%s' with `%s`?", fileArgs, editor)); !open {
		console.Warn("Canceled opening the files in the editor. Manually edit the files before continuing")
		return nil
	}

	for i, f := range files {
		files[i] = filepath.Join(dir, f)
	}
	cmd := exec.Command(editor, files...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
