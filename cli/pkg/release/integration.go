package release

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/wordpress-mobile/gbm-cli/internal/integration"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

func CreateAndroidPr(version, baseBranch, dir string, gbmPr repo.PullRequest, verbose bool) (repo.PullRequest, error) {

	t := integration.Target{
		Repo:        "WordPress-Android",
		HeadBranch:  fmt.Sprintf("gutenberg/integrate_release_%s", version),
		BaseBranch:  baseBranch,
		RenderTitle: buildTitleRenderer(version),
		RenderBody:  buildBodyRenderer(version, "templates/release/integrationPrBody.md"),
		Labels:      []repo.Label{{Name: "Gutenberg"}},
		Draft:       true,
		Dir:         dir,
		VersionFile: "build.gradle",
		// The initial PR will be created with the prNumber-sha format
		UpdateVersion: buildUpdateAndroidVersion(fmt.Sprintf("%d-%s", gbmPr.Number, gbmPr.Head.Sha)),
	}

	return integration.CreateIntegrationPr(t, gbmPr, verbose)
}

func CreateIosPr(version, baseBranch, dir string, gbmPr repo.PullRequest, verbose bool) (repo.PullRequest, error) {

	t := integration.Target{
		Repo:        "WordPress-iOS",
		HeadBranch:  fmt.Sprintf("gutenberg/integrate_release_%s", version),
		BaseBranch:  baseBranch,
		RenderTitle: buildTitleRenderer(version),
		RenderBody:  buildBodyRenderer(version, "templates/release/integrationPrBody.md"),
		Labels:      []repo.Label{{Name: "Gutenberg"}},
		Draft:       true,
		Dir:         dir,
		VersionFile: "Gutenberg/version.rb",
		// The initial PR will be created with a commit version
		UpdateVersion: buildUpdateIosVersion(gbmPr.Head.Sha),
	}

	return integration.CreateIntegrationPr(t, gbmPr, verbose)
}

// Build a closure around the release version. This is later used to render the PR title.
// We don't need the gbm PR here, but we need to keep the same signature as the other renderers.
func buildTitleRenderer(version string) func(repo.PullRequest) string {
	return func(_ repo.PullRequest) string {
		return fmt.Sprintf("Integrate Gutenberg Mobile %s", version)
	}
}

// Build a closure around the release version and template path.
// This is later used to render the PR body.
// The template path is passed in to make it easier to test.
func buildBodyRenderer(version, templatePath string) func(repo.PullRequest) string {
	return func(gbmPr repo.PullRequest) string {

		data := struct {
			Version  string
			GbmPrUrl string
		}{
			Version:  version,
			GbmPrUrl: gbmPr.Url,
		}

		body, err := render.Render(templatePath, data, nil)
		if err != nil {
			fmt.Println(err)
		}
		return body

	}
}

func buildUpdateAndroidVersion(version string) integration.VersionUpdaterFunc {
	return func(config []byte, _ repo.PullRequest) ([]byte, error) {
		re := regexp.MustCompile(`(gutenbergMobileVersion\s*=\s*)'(?:.*)'`)

		if match := re.Match(config); !match {
			return nil, errors.New("cannot find a version in the gradle file")
		}

		repl := fmt.Sprintf(`$1'%s'`, version)
		return re.ReplaceAll(config, []byte(repl)), nil
	}
}

func buildUpdateIosVersion(version string) integration.VersionUpdaterFunc {

	return func(config []byte, _ repo.PullRequest) ([]byte, error) {
		// Set up regexps for tag or commit
		tagRe := regexp.MustCompile(`v\d+\.\d+\.\d+`)
		tagLineRe := regexp.MustCompile(`([\r\n]\s*)#*(tag:.*)`)
		commitLineRe := regexp.MustCompile(`([\r\n]\s*)#*(commit:.*)`)

		var (
			updated []byte
		)

		// TODO return an error if we can't find a tag or a commit line
		// If matching a version tag, replace the tag line with the new tag
		if tagRe.MatchString(version) {
			updated = commitLineRe.ReplaceAll(config, []byte("${1}#${2}"))
			tagRepl := []byte(fmt.Sprintf(`${1}tag: '%s'`, version))
			updated = tagLineRe.ReplaceAll(updated, tagRepl)
		} else {
			updated = tagLineRe.ReplaceAll(config, []byte("${1}#${2}"))
			commitRepl := []byte(fmt.Sprintf(`${1}commit: '%s'`, version))
			updated = commitLineRe.ReplaceAll(updated, commitRepl)
		}

		return updated, nil
	}
}
