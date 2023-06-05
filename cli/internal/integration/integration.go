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

// Integration Target interface
type Target interface {
	UpdateVersion(string, repo.PullRequest) error
	GetVersion(repo.PullRequest) string
	Title(repo.PullRequest) string
	Body(repo.PullRequest) string
	GetRepo() string
	GetBaseBranch() string
	GetHeadBranch() string
	GetLabels() []repo.Label
}

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

func logger(v bool) func(string, ...interface{}) {
	return func(f string, a ...interface{}) {
		if v {
			utils.LogInfo(f, a...)
		}
	}
}

// Creates an integration PR for the given target
// It will return an ExitingPrError if the branch already exists
func CreateIntegrationPr(target Target, gbmPr repo.PullRequest, verbose bool) (repo.PullRequest, error) {

	l := logger(verbose)

	targetRepo := target.GetRepo()
	targetOrg, _ := repo.GetOrg(targetRepo)
	baseBranch := target.GetBaseBranch()
	headBranch := target.GetHeadBranch()

	exBranch, _ := repo.SearchBranch(targetRepo, headBranch)

	pr := repo.PullRequest{}

	// TODO - Should also check if the PR already exists ???
	// Right now we are just checking if the branch exists
	// But we could push successfully and then fail to create the PR
	if (exBranch != repo.Branch{}) {
		return pr, &repo.BranchError{Err: errors.New("branch already exists"), Type: "exists"}
	}

	setTempDir()
	dir := filepath.Join(tempDir, targetRepo)
	defer cleanup()
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

	if err := target.UpdateVersion(dir, gbmPr); err != nil {
		return pr, err
	}

	l("Committing changes")
	if err := repo.Commit(r, "Update Gutenberg version", git.CommitOptions{All: true}); err != nil {
		return pr, err
	}

	l("Pushing changes")
	if err := repo.Push(r); err != nil {
		return pr, err
	}

	l("Creating PR")
	pr = repo.PullRequest{
		Title:  target.Title(gbmPr),
		Body:   target.Body(gbmPr),
		Head:   repo.Repo{Ref: headBranch},
		Base:   repo.Repo{Ref: baseBranch},
		Labels: target.GetLabels(),
	}

	err = repo.CreatePr(targetRepo, &pr)
	return pr, err
}
