package release

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
	"github.com/wordpress-mobile/gbm-cli/pkg/shell"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

type ReleaseIntegration struct {
	Android    bool
	Ios        bool
	Version    string
	BaseBranch string
	HeadBranch string
}

func Integrate(dir string, ri ReleaseIntegration) (gh.PullRequest, error) {

	pr := gh.PullRequest{}

	if !ri.Android && !ri.Ios {
		return pr, errors.New("no platform specified")
	}

	sp := shell.CmdProps{Dir: dir, Verbose: true}
	git := shell.NewGitCmd(sp)

	rpo := getRepo(ri)

	repoPath := repo.GetRepoPath(rpo)

	// clone repo
	base := ri.BaseBranch
	if base == "" {
		base = "trunk"
	}

	if err := git.Clone("-b", base, "--depth=1", repoPath, dir); err != nil {
		return pr, err
	}

	// @TODO: check if branch exists

	// Create release branch
	console.Info("Creating release branch in %s", ri.HeadBranch, rpo)
	branch := "gutenberg/integrate_release_" + ri.Version
	if err := git.Switch("-c", branch); err != nil {
		return pr, err
	}

	// Update gutenberg config
	console.Info("Updating gutenberg config")
	if err := updateGutenbergConfig(dir, git, ri); err != nil {
		return pr, err
	}
	org, _ := repo.GetOrg(rpo)
	prompt := fmt.Sprintf("\nReady to create the PR on %s/%s?", org, rpo)
	cont := console.Confirm(prompt)
	if !cont {
		console.Info("Bye 👋")
		return pr, errors.New("exiting before creating PR")
	}

	// Push branch
	console.Info("Pushing branch %s to %s", branch, rpo)
	if err := git.Push(); err != nil {
		return pr, err
	}

	// Create PR
	pr, err := createPR(dir, ri.Version, ri)
	if err != nil {
		return pr, err
	}

	// Create after branch
	if err := git.Switch(base); err != nil {
		return pr, err
	}
	afterBranch := "gutenberg/after_" + ri.Version
	console.Info("Creating after branch %s in %s", afterBranch, rpo)
	if err := git.Switch("-c", afterBranch); err != nil {
		return pr, err
	}
	if err := git.Push(); err != nil {
		return pr, err
	}

	return pr, nil
}

func getRepo(ri ReleaseIntegration) string {
	if ri.Android {
		return "WordPress-Android"
	}
	return "WordPress-iOS"
}

func updateGutenbergConfig(dir string, git shell.GitCmds, ri ReleaseIntegration) error {
	if ri.Android {
		return updateAndroid(dir, git, ri)
	}
	return updateIos(dir, git, ri)
}

func updateIos(dir string, git shell.GitCmds, ri ReleaseIntegration) error {
	// TODO update github org although not sure it's useful here
	console.Info("Update gutenberg-mobile ref in Gutenberg/config.yml")

	configPath := filepath.Join(dir, "Gutenberg/config.yml")
	buf, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// perform updates using the yq syntax
	updates := []string{".ref.commit = \"v" + ri.Version + "\"", "del(.ref.tag)"}
	config, err := utils.YqEvalAll(updates, string(buf))
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return err
	}
	sp := shell.CmdProps{Dir: dir, Verbose: true}
	bundle := shell.NewBundlerCmd(sp)

	console.Info("Running bundle install")
	if err := bundle.Install(); err != nil {
		return err
	}

	console.Info("Running rake dependencies")
	rake := shell.NewRakeCmd(sp)
	if err := rake.Dependencies(); err != nil {
		return err
	}

	return git.CommitAll("Release script: Update gutenberg-mobile ref", ri.Version)
}

func updateAndroid(dir string, git shell.GitCmds, ri ReleaseIntegration) error {
	// Find the gutenberg-mobile release PR
	filter := gh.BuildRepoFilter("gutenberg-mobile", "is:open", "is:pr", `label:"release-process"`, fmt.Sprintf("%s in:title", ri.Version))

	pr, err := gh.SearchPr(filter)
	if err != nil {
		return err
	}

	prId := pr.Number
	prSha := pr.Head.Sha

	configPath := filepath.Join(dir, "build.gradle")
	config, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`(gutenbergMobileVersion\s*=\s*)'(?:.*)'`)
	if match := re.Match(config); !match {
		return errors.New("cannot find a version in the gradle file")
	}

	repl := fmt.Sprintf(`$1'%s-%s'`, fmt.Sprint(prId), prSha)
	config = re.ReplaceAll(config, []byte(repl))

	if err := os.WriteFile(configPath, config, 0644); err != nil {
		return err
	}
	return git.CommitAll("Release script: Update build.gradle gutenbergMobileVersion to ref")
}

func createPR(dir, version string, ri ReleaseIntegration) (gh.PullRequest, error) {
	pr := gh.PullRequest{}
	console.Info("Creating PR")
	pr.Title = fmt.Sprint("Integrate gutenberg-mobile release v", ri.Version)
	pr.Base.Ref = ri.BaseBranch
	pr.Head.Ref = ri.HeadBranch

	if err := renderPrBody(ri.Version, &pr); err != nil {
		console.Info("Unable to render the GB PR body (err %s)", err)
	}

	pr.Labels = []gh.Label{{
		Name: "Gutenberg",
	}}

	gh.PreviewPr("gutenberg", dir, pr)

	rpo := getRepo(ri)
	org, _ := repo.GetOrg(rpo)
	prompt := fmt.Sprintf("\nReady to create the PR on %s/%s?", org, rpo)
	cont := console.Confirm(prompt)

	if !cont {
		console.Info("Bye 👋")
		return pr, errors.New("exiting before creating PR")
	}

	if err := gh.CreatePr(rpo, &pr); err != nil {
		return pr, err
	}
	return pr, nil
}

func renderPrBody(version string, pr *gh.PullRequest) error {
	t := render.Template{
		Path: "templates/release/integrate_pr_body.md",
		Data: struct {
			Version  string
			GbmPrUrl string
		}{
			Version:  version,
			GbmPrUrl: "TBD",
		},
	}

	body, err := render.Render(t)
	if err != nil {
		return err
	}
	pr.Body = body
	return nil
}
