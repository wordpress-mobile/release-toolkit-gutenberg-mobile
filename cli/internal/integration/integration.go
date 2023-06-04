package integration

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-git/go-git/v5"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
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
	GetTags() []string
}

var (
	tempDir string
)

func cleanup() {
	os.RemoveAll(tempDir)
}

func init() {
	// Make sure we clean up temp files on early exits
	c := make(chan os.Signal)
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

// Creates an integration PR for the given target
// It will return an ExitingPrError if the branch already exists
func CreateIntegrationPr(target Target, gbmPr *repo.PullRequest) error {
	targetRepo := target.GetRepo()
	targetOrg, _ := repo.GetOrg(targetRepo)
	baseBranch := target.GetBaseBranch()
	headBranch := target.GetHeadBranch()

	exBranch, _ := repo.SearchBranch(targetRepo, headBranch)

	// TODO - Should also check if the PR already exists ???
	// Right now we are just checking if the branch exists
	// But we could push successfully and then fail to create the PR
	if (exBranch != repo.Branch{}) {
		return &repo.BranchError{Err: errors.New("branch already exists"), Type: "exists"}
	}

	setTempDir()
	defer cleanup()

	repoUrl := fmt.Sprintf("git@github.com:%s/%s.git", targetOrg, targetRepo)
	r, err := repo.Clone(repoUrl, baseBranch, tempDir)
	if err != nil {
		return err
	}

	if err := repo.Checkout(r, headBranch); err != nil {
		return err
	}

	if err := target.UpdateVersion(tempDir, *gbmPr); err != nil {
		return err
	}

	// TODO: Should we allow a custom commit message?
	if err := repo.Commit(r, "Update Gutenberg version", git.CommitOptions{All: true}); err != nil {
		return err
	}

	return nil
}
