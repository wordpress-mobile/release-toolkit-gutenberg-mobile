package integration

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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

	// The function that will be used to render the PR title.
	RenderTitle func(gbmPr repo.PullRequest) string

	// The function that will be used to render the PR body.
	RenderBody func(gbmPr repo.PullRequest) string

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
	tempDir string
)

func cleanup() {
	os.RemoveAll(tempDir)
}

func init() {
	// Make sure we clean up temp files on early exits
	// Use a buffered channel so we don't miss the signal.
	// see https://go.dev/tour/concurrency/5 and https://gobyexample.com/signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()
}

func setTempDir() {
	var err error
	if tempDir, err = ioutil.TempDir("", "gbm-"); err != nil {
		fmt.Println("Error creating temp dir")
		os.Exit(1)
	}
}

func logger(v bool, repo string) func(string, ...interface{}) {
	return func(f string, a ...interface{}) {
		if v {
			utils.LogInfo(fmt.Sprint(repo, ": ", f), a...)
		}
	}
}

// Creates an integration PR for the given target
// It will return an ExitingPrError if the branch already exists
func CreateIntegrationPr(target Target, gbmPr repo.PullRequest, verbose bool) (repo.PullRequest, error) {

	targetRepo := target.Repo
	targetOrg, _ := repo.GetOrg(targetRepo)
	baseBranch := target.BaseBranch
	headBranch := target.HeadBranch

	// Since functions can be nil we need to check if the version updater exists
	if target.UpdateVersion == nil {
		return repo.PullRequest{}, fmt.Errorf("%s UpdateVersion function is nil", targetRepo)
	}

	l := logger(verbose, targetRepo)

	exBranch, _ := repo.SearchBranch(targetRepo, headBranch)

	pr := repo.PullRequest{}

	// TODO - Should also check if the PR already exists ???
	// Right now we are just checking if the branch exists
	// But we could push successfully and then fail to create the PR
	if (exBranch != repo.Branch{}) {
		return pr, &repo.BranchError{Err: errors.New("branch already exists"), Type: "exists"}
	}

	dir := filepath.Join(target.Dir, targetRepo)

	l("Cloning %s into %s", targetRepo, dir)

	repoUrl := fmt.Sprintf("git@github.com:%s/%s.git", targetOrg, targetRepo)
	r, err := repo.Clone(repoUrl, baseBranch, dir)
	if err != nil {
		return pr, err
	}

	l("Checking out %s", headBranch)
	if err := repo.Checkout(r, headBranch); err != nil {
		return pr, err
	}

	l("Updating Gutenberg Mobile version")
	configPath := filepath.Join(dir, target.VersionFile)
	config, err := os.ReadFile(configPath)
	if err != nil {
		return pr, fmt.Errorf("%s error reading version file: %w", targetRepo, err)
	}
	update, err := target.UpdateVersion(config, gbmPr)
	if err != nil {
		return pr, fmt.Errorf("%s error updating version file: %w", targetRepo, err)
	}

	// We just overwrite the file with the new bytes
	f, err := os.Create(configPath)
	if err != nil {
		return pr, fmt.Errorf("%s error creating version file: %w", targetRepo, err)
	}
	defer f.Close()
	if _, err := f.Write(update); err != nil {
		return pr, fmt.Errorf("%s error writing version file file: %w", targetRepo, err)
	}

	l("Committing changes")
	if err := repo.Commit(r, "Update Gutenberg version", git.CommitOptions{All: true}); err != nil {
		return pr, err
	}

	l("Pushing changes")
	if err := repo.Push(r); err != nil {
		return pr, err
	}

	l("Creating the PR")
	pr = repo.PullRequest{
		Title:  target.RenderTitle(gbmPr),
		Body:   target.RenderBody(gbmPr),
		Head:   repo.Repo{Ref: headBranch},
		Base:   repo.Repo{Ref: baseBranch},
		Labels: target.Labels,
		Draft:  target.Draft,
	}

	err = repo.CreatePr(targetRepo, &pr)
	return pr, err
}
