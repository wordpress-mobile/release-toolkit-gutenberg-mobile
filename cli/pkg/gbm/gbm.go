package gbm

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/wordpress-mobile/gbm-cli/internal/exc"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
)

func logger(v bool) func(string, ...interface{}) {
	return func(f string, a ...interface{}) {
		if v {
			utils.LogInfo("\nGBM "+f, a...)
		}
	}
}

func excNpm(dir string, verbose bool) func(cmds ...string) error {
	return func(cmds ...string) error {
		return exc.Npm(dir, verbose, cmds...)
	}
}

func PrepareBranch(dir string, pr *repo.PullRequest, verbose bool) (*git.Repository, error) {
	l := logger(verbose)

	gbmDir := filepath.Join(dir, "gutenberg-mobile")
	npm := excNpm(gbmDir, verbose)

	d := utils.LogDebug
	version := pr.ReleaseVersion
	var gbmr *git.Repository
	var err error

	// If we already have a copy of GBM, initialize the repo
	// Otherwise clone at pr.Base
	if _, ok := os.Stat(gbmDir); ok != nil {
		l("Cloning Gutenberg Mobile")
		gbmr, err = repo.CloneGBM(gbmDir, verbose)
		if err != nil {
			return nil, err
		}
	} else {
		l("Initializing Gutenberg Mobile Repo at %s", gbmDir)
		gbmr, err = repo.Open(gbmDir)
		if err != nil {
			return nil, fmt.Errorf("issue opening gutenberg-mobile (err %s)", err)
		}
	}

	if err := repo.Switch(gbmDir, pr.Head.Ref, verbose); err != nil {
		return nil, fmt.Errorf("unable to switch to %s (err %s)", pr.Head.Ref, err)
	}

	l("Checking Gutenberg")
	// Check if Gutenberg is porcelain
	gbr, err := repo.Open(filepath.Join(gbmDir, "gutenberg"))
	if err != nil {
		return nil, err
	}
	if clean, err := repo.IsPorcelain(gbr); err != nil {
		return nil, err
	} else if !clean {
		return nil, errors.New("gutenberg has some uncommitted changes")
	}

	l("Updating Gutenberg")
	if err = repo.CommitSubmodule(gbmDir, "Release script: Updating gutenberg ref", "gutenberg", verbose); err != nil {
		return nil, fmt.Errorf("failed to update gutenberg: %s", err)
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
	d("about to commit xcf")
	if err := repo.CommitAll(gbmr, "Release script: Sync XCFramework `Podfile.lock`"); err != nil {
		return nil, err
	}

	// If there is a version we should update the package json
	if version != "" {

		l("Updating the version")
		if err := npm("--no-git-tag-version", "version", version); err != nil {
			return nil, err
		}

		if err := repo.Commit(gbmr, "Update Version", "package.json", "package-lock.json"); err != nil {
			return nil, err
		}

	} else {
		// Otherwise just update the bundle
		if err := npm("run", "bundle"); err != nil {
			return nil, err
		}
	}

	l("Updating the bundle")
	if err := repo.CommitAll(gbmr, "Release script: Update bundle for"+version); err != nil {
		return nil, err
	}

	return gbmr, nil
}

func CreatePr(gbmr *git.Repository, pr *repo.PullRequest, verbose bool) error {

	l := logger(verbose)

	l("Pushing the branch")
	if err := repo.Push(gbmr, verbose); err != nil {
		return err
	}

	l("Creating the PR")
	if err := repo.CreatePr("gutenberg-mobile", pr); err != nil {
		return err
	}

	return nil
}
