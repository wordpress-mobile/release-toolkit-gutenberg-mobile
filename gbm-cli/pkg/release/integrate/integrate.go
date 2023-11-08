package integrate

import (
	"errors"
	"fmt"

	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/release"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/render"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/repo"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/shell"
)

type Integration interface {
	Run(dir string, ri ReleaseIntegration) (gh.PullRequest, error)
	cloneRepo(git shell.GitCmds) error
	createAfterBranch(git shell.GitCmds) error
	getRepo() string
	updateGutenbergConfig(dir string, gbmPr gh.PullRequest) error
	createPR(dir string, gbmPr gh.PullRequest) (gh.PullRequest, error)
}

type ReleaseIntegration struct {
	Version    string
	BaseBranch string
	HeadBranch string
	Target     Target
	GbmPr      gh.PullRequest
}

type Target interface {
	UpdateGutenbergConfig(dir string, gbmPr gh.PullRequest) error
	GetRepo() string
	GetPr(ri ReleaseIntegration) (gh.PullRequest, error)
	GbPublished(gh.PullRequest) (bool, error)
}

func (ri *ReleaseIntegration) Run(dir string) (gh.PullRequest, error) {
	rpo := ri.Target.GetRepo()
	org := repo.GetOrg(rpo)

	// Check if the GBM build is published
	// Only if the target is wordpress-mobile
	if org == "wordpress-mobile" {
		published, err := ri.Target.GbPublished(ri.GbmPr)
		if err != nil {
			return gh.PullRequest{}, err
		}
		if !published {
			console.Info("GBM build not published yet")
			return gh.PullRequest{}, nil
		}
	}

	pr := gh.PullRequest{}
	if rpo == "" {
		return pr, errors.New("no platform specified")
	}

	gbmPr := ri.GbmPr

	git := shell.NewGitCmd(shell.CmdProps{Dir: dir, Verbose: true})

	// Clone repo
	if err := ri.cloneRepo(git); err != nil {
		return pr, fmt.Errorf("error cloning the %s repository: %v", rpo, err)
	}

	// Update gutenberg config
	if err := ri.Target.UpdateGutenbergConfig(dir, gbmPr); err != nil {
		return pr, fmt.Errorf("error updating the gutenberg config: %v", err)
	}

	if err := git.Push(); err != nil {
		return pr, fmt.Errorf("error pushing changes: %v", err)
	}

	// Check if the PR already exists
	pr, err := ri.Target.GetPr(*ri)
	if err != nil {
		return pr, fmt.Errorf("error getting the PR: %v", err)
	}

	if pr.Number != 0 {
		console.Info("PR already exists: %s", pr.Url)
		return pr, nil
	}

	// Confirm PR creation
	prompt := fmt.Sprintf("\nReady to create the PR on %s/%s?", org, rpo)
	if cont := console.Confirm(prompt); !cont {
		console.Info("Bye 👋")
		return pr, errors.New("exiting before creating PR")
	}

	pr, err = ri.createPR(dir, gbmPr)
	if err != nil {
		return pr, fmt.Errorf("error creating the PR: %v", err)
	}

	// Create after branch
	if err := ri.createAfterBranch(git); err != nil {
		return pr, err
	}

	return pr, nil
}

// Clone the repo at the configured base branch or at the release branch if it already exists.
func (ri *ReleaseIntegration) cloneRepo(git shell.GitCmds) error {
	// Check if release branch already exists
	rpo := ri.Target.GetRepo()
	repoPath := repo.GetRepoHttpsPath(rpo)

	branch := fmt.Sprintf(release.IntegrateBranchName, ri.Version)
	exists, err := gh.SearchBranch(rpo, branch)
	if err != nil {
		return err
	}

	if (exists != gh.Branch{}) {
		console.Info("Cloning repo at release branch %s", branch)
		if err := git.Clone(repoPath, "-b", branch, "--depth=1", "."); err != nil {
			return err
		}
	} else {
		// clone repo
		base := ri.BaseBranch
		if base == "" {
			base = "trunk"
		}

		console.Info("Cloning repo at base branch %s", base)
		if err := git.Clone("-b", base, "--depth=1", repoPath, "."); err != nil {
			return err
		}

		// Create release branch
		console.Info("Creating release branch in %s", ri.HeadBranch)
		if err := git.Switch("-c", branch); err != nil {
			return err
		}
	}
	return nil
}

func (ri *ReleaseIntegration) createAfterBranch(git shell.GitCmds) error {
	rpo := ri.Target.GetRepo()
	afterBranch := fmt.Sprintf(release.IntegrateAfterBranchName, ri.Version)
	// Check if branch exits
	exists, err := gh.SearchBranch(rpo, afterBranch)
	if err != nil {
		return err
	}
	if (exists != gh.Branch{}) {
		console.Info("Branch %s already exists", afterBranch)
		return nil
	}

	// Switch to the base branch
	base := ri.BaseBranch
	if err := git.Fetch(base); err != nil {
		return err
	}
	if err := git.Switch(base); err != nil {
		return err
	}

	// Create the branch
	console.Info("Creating after branch %s in %s", afterBranch, rpo)
	if err := git.Switch("-c", afterBranch); err != nil {
		return err
	}
	if err := git.Push(); err != nil {
		return err
	}
	return nil
}

func (ri *ReleaseIntegration) createPR(dir string, gbmPr gh.PullRequest) (gh.PullRequest, error) {
	version := ri.Version
	pr := gh.PullRequest{}
	console.Info("Creating PR")
	pr.Title = fmt.Sprintf(release.IntegratePrTitle, ri.Version)
	pr.Base.Ref = ri.BaseBranch
	pr.Head.Ref = ri.HeadBranch

	if err := renderPrBody(version, &pr, gbmPr); err != nil {
		console.Info("Unable to render the GB PR body (err %s)", err)
	}

	pr.Labels = []gh.Label{{
		Name: release.IntegratePrLabel,
	}}

	rpo := ri.Target.GetRepo()
	gh.PreviewPr(rpo, dir, ri.BaseBranch, pr)

	if err := gh.CreatePr(rpo, &pr); err != nil {
		return pr, err
	}
	return pr, nil
}

func renderPrBody(version string, pr *gh.PullRequest, gbmPr gh.PullRequest) error {
	t := render.Template{
		Path: "templates/release/integrate_pr_body.md",
		Data: struct {
			Version  string
			GbmPrUrl string
		}{
			Version:  version,
			GbmPrUrl: gbmPr.Url,
		},
	}

	body, err := render.Render(t)
	if err != nil {
		return err
	}
	pr.Body = body
	return nil
}

func useRelease(version string) (bool, error) {
	release, err := release.GetGbmRelease(version)
	if err != nil {
		console.Warn("Unable to check for a release: %s", err)
		return false, nil
	}

	if release.PublishedAt == "" {
		return false, nil
	}

	console.Info("Found release v%s – %s", version, release.Url)
	return true, nil
}
