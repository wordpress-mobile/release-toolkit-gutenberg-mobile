package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
)

type SubmoduleRef struct {
	Repo   string
	Tag    string
	Branch string
	Sha    string
}

func getAuth() *http.BasicAuth {
	// load host and auth from 'gh'
	host, _ := auth.DefaultHost()
	token, _ := auth.TokenForHost(host)
	user := getSignature()
	return &http.BasicAuth{
		Username: user.Name, // this can be anything since we are using a token
		Password: token,
	}
}

func getSignature() *object.Signature {
	// Load the config from 'gh'
	config, _ := config.LoadConfig(config.GlobalScope)
	u := config.User
	s := object.Signature{
		Name:  u.Name,
		Email: u.Email,
		When:  time.Now(),
	}

	return &s
}

func Clone(url, branch, path string, verbose bool) (*git.Repository, error) {
	opts := &git.CloneOptions{
		Auth:              getAuth(),
		URL:               url,
		ReferenceName:     plumbing.ReferenceName(branch),
		Depth:             1,
		RecurseSubmodules: 1,
		SingleBranch:      true,
	}
	if verbose {
		opts.Progress = os.Stdout
	}

	r, err := git.PlainClone(path, false, opts)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func Open(path string) (*git.Repository, error) {
	r, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open repo at %s (err %s)", path, err)
	}
	return r, nil
}

// go-git has an issue with cloning submodules https://github.com/go-git/go-git/issues/488
// Dropping down to git for now
func CloneGBM(dir string, pr PullRequest, verbose bool) (*git.Repository, error) {
	git := execGit(dir, verbose)

	org, _ := GetOrg("gutenberg-mobile")
	url := fmt.Sprintf("git@github.com:%s/%s.git", org, "gutenberg-mobile")

	cmd := []string{"clone", "--recurse-submodules", "--depth", "1"}

	fmt.Println("Checking remote branch...")
	// check to see if the remote branch exists
	if err := git("ls-remote", "--exit-code", "--heads", url, pr.Head.Ref); err != nil {
		cmd = append(cmd, url)
	} else {
		cmd = append(cmd, "--branch", pr.Head.Ref, url)
	}

	if err := git(cmd...); err != nil {
		return nil, fmt.Errorf("unable to clone gutenberg mobile %s", err)
	}
	return Open(filepath.Join(dir, "gutenberg-mobile"))
}

func Switch(dir, repo, branch string, verbose bool) error {

	git := execGit(dir, verbose)

	create := !remoteExists(dir, repo, branch, verbose)

	if create {
		return git("switch", "-c", branch)
	}

	// We do shallow checkouts so we need to fetch the branch
	err := git("remote", "set-branches", "origin", branch)
	if err != nil {
		return fmt.Errorf("unable to set remote branches (err %s)", err)
	}
	err = git("fetch", "origin", "--depth", "1")
	if err != nil {
		return fmt.Errorf("unable to fetch branch (err %s)", err)
	}

	return git("switch", branch)
}

func remoteExists(dir, repo, ref string, verbose bool) bool {
	git := execGit(dir, verbose)

	org, _ := GetOrg("gutenberg-mobile")
	url := fmt.Sprintf("git@github.com:%s/%s.git", org, "gutenberg-mobile")
	if err := git("ls-remote", "--exit-code", "--heads", url, ref); err != nil {
		return false
	}
	return true
}

func Checkout(r *git.Repository, branch string) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	return w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: true,
	})
}

func IsPorcelain(r *git.Repository) (bool, error) {
	w, err := r.Worktree()
	if err != nil {
		return false, err
	}
	status, err := w.Status()
	if err != nil {
		return false, err
	}
	return status.IsClean(), nil
}

func Add(r *git.Repository, files ...string) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	for _, f := range files {
		_, err := w.Add(f)
		if err != nil {
			utils.LogWarn("Error adding the file %s :%s", f, err)
		}
	}
	return nil
}

func Commit(r *git.Repository, message string, files ...string) error {
	return CommitOptions(r, message, git.CommitOptions{}, files...)
}

func CommitOptions(r *git.Repository, message string, opts git.CommitOptions, files ...string) error {
	w, err := r.Worktree()

	for _, f := range files {
		_, err := w.Add(f)
		if err != nil {
			utils.LogWarn("Error adding the file %s :%s", f, err)
		}
	}

	if err != nil {
		return err
	}

	if opts.Author == nil {
		opts.Author = getSignature()
	}
	_, err = w.Commit(message, &opts)

	return err
}

func CommitAll(r *git.Repository, message string) error {
	return CommitOptions(r, message, git.CommitOptions{All: true})
}

func GetSubmodule(r *git.Repository, path string) (*git.Submodule, error) {
	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	return w.Submodule(path)
}

// go-git has an open issue about committing submodules
// https://github.com/go-git/go-git/issues/248
// https://stackoverflow.com/a/71263056/1373043
// This drops dow to `git` to commit the submodule update
func CommitSubmodule(dir, message, submodule string, verbose bool) error {

	git := execGit(dir, verbose)

	if err := git("add", submodule); err != nil {
		return fmt.Errorf("unable to add submodule %s in %s :%s", submodule, dir, err)
	}

	if err := git("commit", "-m", message); err != nil {
		return fmt.Errorf("unable to commit submodule update %s : %s", submodule, err)
	}
	return nil
}

type NotPorcelainError struct {
	Err error
}

func (r *NotPorcelainError) Error() string {
	return r.Err.Error()
}

func IsSubmoduleCurrent(s *git.Submodule, expectedHash string) (bool, error) {

	// Check if the submodule is porcelain
	sr, err := s.Repository()
	if clean, err := IsPorcelain(sr); err != nil {
		return false, err
	} else if !clean {
		return false, &NotPorcelainError{fmt.Errorf("submodule %s is not clean", s.Config().Name)}
	}

	if err != nil {
		return false, err
	}
	stat, err := s.Status()
	if err != nil {
		return false, err
	}
	eh := plumbing.NewHash(expectedHash)

	return stat.Current == eh, nil
}

func Tag(r *git.Repository, tag string, push bool) (*plumbing.Reference, error) {
	h, err := r.Head()
	if err != nil {
		return nil, err
	}
	ref, err := r.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Message: tag,
		Tagger:  getSignature(),
	})
	if err != nil {
		return ref, err
	}
	if push {
		return ref, PushTag(r, true)
	}
	return ref, err
}

func Push(r *git.Repository, verbose bool) error {
	opts := &git.PushOptions{
		RemoteName: "origin",
		Auth:       getAuth(),
	}
	if verbose {
		opts.Progress = os.Stdout
	}

	err := r.Push(opts)

	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func PushTag(r *git.Repository, verbose bool) error {
	opts := &git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth:       getAuth(),
	}
	if verbose {
		opts.Progress = os.Stdout
	}
	err := r.Push(opts)

	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

// Use this to drop down to `git` when go-git is not playing well.
func execGit(dir string, verbose bool) func(...string) error {
	return func(cmds ...string) error {
		cmd := exec.Command("git", cmds...)
		cmd.Dir = dir

		if verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		return cmd.Run()
	}
}
