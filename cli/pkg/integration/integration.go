package integration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
)

// Integration Target struct
type Target struct {
	// The integration target repo. The allowed values are:
	// WordPress-Android, WordPress-iOS
	Repo string

	// The integration branch where the integration updates are committed.
	// The naming is the convention used by the Github api.
	HeadBranch string

	// The base branch where the integration branch is based on.
	// The HeadBranch will be branched from this branch. Also the PR will be
	// opened against this branch.
	BaseBranch string

	// PR Title
	Title string

	// PR Body
	Body string

	// The relative path to the integration config file from the integration repo root.
	VersionFile string

	// The function that updates the version in the integration config.
	UpdateVersion VersionUpdaterFunc

	// The labels that will be added to the PR.
	Labels []repo.Label

	// Sets the PR to draft if this is true.
	Draft bool

	// The directory where the integration repo is cloned into.
	Dir string
}

type VersionUpdaterFunc func([]byte, repo.PullRequest) ([]byte, error)

var (
	l func(string, ...interface{})
)

func init() {
	l = func(f string, a ...interface{}) {
		utils.LogInfo(fmt.Sprint(f), a...)
	}
}

// Creates an integration PR for the given target
// It will return an ExitingPrError if the branch already exists
func PrepareBranch(target *Target, gbmPr repo.PullRequest, verbose bool) (*git.Repository, error) {

	targetRepo := target.Repo
	targetOrg, _ := repo.GetOrg(targetRepo)
	baseBranch := target.BaseBranch
	headBranch := target.HeadBranch

	// Since functions can be nil we need to check if the version updater exists
	if target.UpdateVersion == nil {
		return nil, fmt.Errorf("%s UpdateVersion function is nil", targetRepo)
	}

	exBranch, _ := repo.SearchBranch(targetRepo, headBranch)

	dir := filepath.Join(target.Dir, targetRepo)
	repoUrl := fmt.Sprintf("git@github.com:%s/%s.git", targetOrg, targetRepo)

	var (
		r   *git.Repository
		err error
	)

	// Clone at the existing branch if it exists
	if (exBranch != repo.Branch{}) {
		l("Cloning %s ref:%s into %s", targetRepo, headBranch, dir)
		r, err = repo.Clone(repoUrl, headBranch, dir, verbose)
		if err != nil {
			return nil, err
		}
	} else {
		l("Cloning %s ref:%s into %s", targetRepo, baseBranch, dir)
		r, err = repo.Clone(repoUrl, baseBranch, dir, verbose)
		if err != nil {
			return nil, err
		}

		l("Checking out %s", headBranch)
		if err := repo.Checkout(r, headBranch); err != nil {
			return r, err
		}
	}

	l("Updating Gutenberg Mobile version")
	configPath := filepath.Join(dir, target.VersionFile)
	config, err := os.ReadFile(configPath)
	if err != nil {
		return r, fmt.Errorf("%s error reading version file: %w", targetRepo, err)
	}
	update, err := target.UpdateVersion(config, gbmPr)
	if err != nil {
		return r, fmt.Errorf("%s error updating version file: %w", targetRepo, err)
	}

	// We just overwrite the file with the new bytes
	f, err := os.Create(configPath)
	if err != nil {
		return r, fmt.Errorf("%s error creating version file: %w", targetRepo, err)
	}
	defer f.Close()
	if _, err := f.Write(update); err != nil {
		return r, fmt.Errorf("%s error writing version file file: %w", targetRepo, err)
	}

	l("Committing changes")
	if err := repo.CommitAll(r, "Update Gutenberg version"); err != nil {
		return r, err
	}

	return r, err
}

func CreatePr(target string, rpo *git.Repository, pr *repo.PullRequest, verbose bool) error {

	l("Pushing changes")
	if err := repo.Push(rpo, verbose); err != nil {
		return err
	}

	l("Creating the PR")
	return repo.CreatePr(target, pr)
}
