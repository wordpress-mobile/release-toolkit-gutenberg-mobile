package release

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/pkg/integration"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

type IntegrateOp func(string, string, string, repo.PullRequest, bool) (*repo.PullRequest, error)

func CreateAndroidPr(version, baseBranch, dir string, gbmPr repo.PullRequest, verbose bool) (*repo.PullRequest, error) {
	pointTo := fmt.Sprintf("%d-%s", gbmPr.Number, gbmPr.Head.Sha)
	t := androidTarget(version, pointTo, baseBranch, dir, gbmPr)
	return createPr(t, gbmPr, verbose)
}

func CreateIosPr(version, baseBranch, dir string, gbmPr repo.PullRequest, verbose bool) (*repo.PullRequest, error) {
	t := iosTarget(version, gbmPr.Head.Sha, baseBranch, dir, gbmPr)
	return createPr(t, gbmPr, verbose)
}

func UpdateAndroidPr(version, baseBranch, dir string, gbmPr repo.PullRequest, verbose bool) (*repo.PullRequest, error) {
	aPr, err := GetAndroidReleasePr(version)
	if err != nil {
		return nil, err
	}
	pointTo := updateVersion(version, &gbmPr)
	re := regexp.MustCompile(`^v`)

	// Check if it's a release tag
	if !re.MatchString(pointTo) {
		pointTo = fmt.Sprintf("%d-%s", gbmPr.Number, pointTo)
	}
	t := androidTarget(version, pointTo, baseBranch, dir, gbmPr)
	return aPr, updatePr(t, gbmPr, verbose)
}

func UpdateIosPr(version, baseBranch, dir string, gbmPr repo.PullRequest, verbose bool) (*repo.PullRequest, error) {
	iPr, err := GetIosReleasePr(version)
	if err != nil {
		return nil, err
	}
	pointTo := updateVersion(version, &gbmPr)
	t := iosTarget(version, pointTo, baseBranch, dir, gbmPr)
	return iPr, updatePr(t, gbmPr, verbose)
}

func createPr(target *integration.Target, gbmPr repo.PullRequest, verbose bool) (*repo.PullRequest, error) {
	rpo, err := integration.PrepareBranch(target, gbmPr, verbose)
	if err != nil {
		return nil, err
	}

	pr := repo.PullRequest{
		Title:  target.Title,
		Body:   target.Body,
		Head:   repo.Repo{Ref: target.HeadBranch},
		Base:   repo.Repo{Ref: target.BaseBranch},
		Labels: []repo.Label{{Name: "Gutenberg"}},
		Draft:  true,
	}

	err = integration.CreatePr(target.Repo, rpo, &pr, verbose)
	return &pr, err
}

// Returns either the gbm PR sha or the release tag if the release was
// published. If it can't reach the release then it returns the gbm PR sha.
func updateVersion(version string, gbmPr *repo.PullRequest) string {
	vVersion := "v" + version
	sha := gbmPr.Head.Sha
	release, err := repo.GetRelease("gutenberg-mobile", vVersion)
	if err != nil {
		return sha
	}

	if release.PublishedDate == "" {
		return vVersion
	}

	return sha
}

func updatePr(target *integration.Target, gbmPr repo.PullRequest, verbose bool) error {
	rpo, err := integration.PrepareBranch(target, gbmPr, verbose)
	if err != nil {
		return err
	}
	l("Pushing changes")
	return repo.Push(rpo, verbose)
}

func androidTarget(version, pointTo, baseBranch, dir string, gbmPr repo.PullRequest) *integration.Target {
	return &integration.Target{
		Repo:          "WordPress-Android",
		HeadBranch:    fmt.Sprintf("gutenberg/integrate_release_%s", version),
		BaseBranch:    baseBranch,
		Title:         fmt.Sprintf("Integrate Gutenberg Mobile %s", version),
		Body:          renderIntegrationBody(version, "templates/release/integrationPrBody.md", gbmPr),
		Labels:        []repo.Label{{Name: "Gutenberg"}},
		Draft:         true,
		Dir:           dir,
		VersionFile:   "build.gradle",
		UpdateVersion: buildUpdateAndroidVersion(pointTo),
	}
}

func iosTarget(version, pointTo, baseBranch, dir string, gbmPr repo.PullRequest) *integration.Target {
	return &integration.Target{
		Repo:          "WordPress-iOS",
		HeadBranch:    fmt.Sprintf("gutenberg/integrate_release_%s", version),
		BaseBranch:    baseBranch,
		Title:         fmt.Sprintf("Integrate Gutenberg Mobile %s", version),
		Body:          renderIntegrationBody(version, "templates/release/integrationPrBody.md", gbmPr),
		Labels:        []repo.Label{{Name: "Gutenberg"}},
		Draft:         true,
		Dir:           dir,
		VersionFile:   "Gutenberg/version.rb",
		UpdateVersion: buildUpdateIosVersion(pointTo),
	}
}

func renderIntegrationBody(version, templatePath string, gbmPr repo.PullRequest) string {
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
		tagLineRe := regexp.MustCompile(`([\r\n]\s*)#?\s*(tag:.*)`)
		commitLineRe := regexp.MustCompile(`([\r\n]\s*)#?\s*(commit:.*)`)

		var (
			updated []byte
		)

		// TODO return an error if we can't find a tag or a commit line
		// If matching a version tag, replace the tag line with the new tag
		if tagRe.MatchString(version) {
			updated = commitLineRe.ReplaceAll(config, []byte("${1}# commit: '',"))
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
