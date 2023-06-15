package gbm

import (
	"fmt"
	"os"
	"path/filepath"

	g "github.com/go-git/go-git/v5"
	"github.com/wordpress-mobile/gbm-cli/internal/exc"
	"github.com/wordpress-mobile/gbm-cli/internal/gh"
	"github.com/wordpress-mobile/gbm-cli/internal/git"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
)

func l(f string, a ...interface{}) {
	utils.LogInfo("\nGBM "+f, a...)
}

func excNpm(dir string, verbose bool) func(cmds ...string) error {
	return func(cmds ...string) error {
		return exc.Npm(dir, verbose, cmds...)
	}
}

func PrepareBranch(dir string, pr *gh.PullRequest, gbPr *gh.PullRequest, verbose bool) (*g.Repository, error) {

	gbmDir := filepath.Join(dir, "gutenberg-mobile")
	npm := excNpm(gbmDir, verbose)

	version := pr.ReleaseVersion
	var (
		gbmr *g.Repository
		err  error
	)
	// If we already have a copy of GBM, initialize the repo
	// Otherwise clone at pr.Base
	if _, ok := os.Stat(gbmDir); ok != nil {
		os.MkdirAll(gbmDir, os.ModePerm)
		l("Cloning Gutenberg Mobile")
		gbmr, err = git.CloneGBM(dir, *pr, verbose)
		if err != nil {
			return nil, err
		}
	} else {
		l("Initializing Gutenberg Mobile Repo at %s", gbmDir)
		gbmr, err = git.Open(gbmDir)
		if err != nil {
			return nil, fmt.Errorf("issue opening gutenberg-mobile (err %s)", err)
		}
	}

	if err := git.Switch(gbmDir, "gutenberg-mobile", pr.Head.Ref, verbose); err != nil {
		return nil, err
	}
	// Set up GB
	if err := setupGb(gbmDir, gbmr, gbPr, verbose); err != nil {
		return nil, err
	}

	l("Set up Node")
	if err := exc.SetupNode(gbmDir, verbose); err != nil {
		return nil, fmt.Errorf("failed to update node (err %s)", err)
	}

	l("Installing npm packages")
	if err := npm("ci"); err != nil {
		return nil, fmt.Errorf("failed to update node packages (err %s)", err)
	}

	l("Update XCFramework builders project Podfile.lock")
	xcfDir := filepath.Join(gbmDir, "ios-xcframework")
	if err := exc.BundleInstall(xcfDir, verbose); err != nil {
		return nil, err
	}
	if err := exc.PodInstall(xcfDir, verbose); err != nil {
		return nil, err
	}

	if err := git.CommitAll(gbmr, "Release script: Sync XCFramework `Podfile.lock`"); err != nil {
		return nil, err
	}

	// If there is a version we should update the package json
	if version != "" {

		l("Updating the version")
		if err := npm("--no-git-tag-version", "version", version); err != nil {
			return nil, err
		}

		if err := git.Commit(gbmr, "Update Version", "package.json", "package-lock.json"); err != nil {
			return nil, err
		}

	} else {
		// Otherwise just update the bundle
		if err := npm("run", "bundle"); err != nil {
			return nil, err
		}
	}

	l("Updating the bundle")
	if err := git.CommitAll(gbmr, "Release script: Update bundle for "+version); err != nil {
		return nil, err
	}

	return gbmr, nil
}

func setupGb(gbmDir string, gbmr *g.Repository, gbPr *gh.PullRequest, verbose bool) error {

	l("Checking Gutenberg")

	gb, err := git.GetSubmodule(gbmr, "gutenberg")
	if err != nil {
		return err
	}
	if curr, err := git.IsSubmoduleCurrent(gb, gbPr.Head.Sha); err != nil {
		return fmt.Errorf("issue checking the submodule (err %s)", err)
	} else if !curr {
		if err := git.Switch(filepath.Join(gbmDir, "gutenberg"), "gutenberg", gbPr.Head.Ref, verbose); err != nil {
			return fmt.Errorf("unable to switch to %s (err %s)", gbPr.Head.Ref, err)
		}
	}

	l("Updating Gutenberg")
	if clean, _ := git.IsPorcelain(gbmr); !clean {
		if err = git.CommitSubmodule(gbmDir, "Release script: Updating gutenberg ref", "gutenberg", verbose); err != nil {
			return fmt.Errorf("failed to update gutenberg: %s", err)
		}
	}

	return nil
}

func CreatePr(gbmr *g.Repository, pr *gh.PullRequest, verbose bool) error {

	// TODO: make sure we are not on trunk before pushing
	l("Pushing the branch")
	if err := git.Push(gbmr, verbose); err != nil {
		return err
	}

	l("Creating the PR")
	return gh.CreatePr("gutenberg-mobile", pr)
}
