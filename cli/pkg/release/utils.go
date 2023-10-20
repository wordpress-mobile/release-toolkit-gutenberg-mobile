package release

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/wordpress-mobile/gbm-cli/internal/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/shell"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

func updatePackageJson(dir, version string, pkgs ...string) error {
	sp := shell.CmdProps{Dir: dir, Verbose: true}
	git := shell.NewGitCmd(sp)


	for _, pkg := range pkgs {
		if err := utils.UpdatePackageVersion(version, filepath.Join(dir, pkg)); err != nil {
			return err
		}
	}
	if err := git.CommitAll("Release script: Update package.json versions to %s", version); err != nil {
		return err
	}

	return nil
}

type ReleaseChanges struct {
	Title  string
	Number int
	PrUrl  string
	Issues []string
}

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