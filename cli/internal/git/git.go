package git

import (
	"fmt"
	"os"
	"path/filepath"

	g "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/wordpress-mobile/gbm-cli/internal/exc"
	"github.com/wordpress-mobile/gbm-cli/internal/gh"
	rpo "github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
)

type SubmoduleRef struct {
	Repo   string
	Tag    string
	Branch string
	Sha    string
}

func Clone(url, branch, path string, verbose bool) (*g.Repository, error) {
	opts := &g.CloneOptions{
		Auth:              rpo.Auth(),
		URL:               url,
		ReferenceName:     plumbing.ReferenceName(branch),
		Depth:             1,
		RecurseSubmodules: 1,
		SingleBranch:      true,
	}
	if verbose {
		opts.Progress = os.Stdout
	}

	r, err := g.PlainClone(path, false, opts)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func Open(path string) (*g.Repository, error) {
	r, err := g.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open repo at %s (err %s)", path, err)
	}
	return r, nil
}

// go-git has an issue with cloning submodules https://github.com/go-git/go-git/issues/488
// Dropping down to git for now
func CloneGBM(dir string, pr gh.PullRequest, verbose bool) (*g.Repository, error) {
	git := exc.ExecGit(dir, verbose)

	org, _ := rpo.GetOrg("gutenberg-mobile")
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

	git := exc.ExecGit(dir, verbose)

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
	git := exc.ExecGit(dir, verbose)

	org, _ := rpo.GetOrg("gutenberg-mobile")
	url := fmt.Sprintf("git@github.com:%s/%s.git", org, "gutenberg-mobile")
	if err := git("ls-remote", "--exit-code", "--heads", url, ref); err != nil {
		return false
	}
	return true
}

func Checkout(r *g.Repository, branch string) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	return w.Checkout(&g.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: true,
	})
}

func IsPorcelain(r *g.Repository) (bool, error) {
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

func Add(r *g.Repository, files ...string) error {
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

func Commit(r *g.Repository, message string, files ...string) error {
	return CommitOptions(r, message, g.CommitOptions{}, files...)
}

func CommitOptions(r *g.Repository, message string, opts g.CommitOptions, files ...string) error {
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
		opts.Author = rpo.Signature()
	}
	_, err = w.Commit(message, &opts)

	return err
}

func CommitAll(r *g.Repository, message string) error {
	return CommitOptions(r, message, g.CommitOptions{All: true})
}

func GetSubmodule(r *g.Repository, path string) (*g.Submodule, error) {
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

	git := exc.ExecGit(dir, verbose)

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

func IsSubmoduleCurrent(s *g.Submodule, expectedHash string) (bool, error) {

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

func Tag(r *g.Repository, tag string, push bool) (*plumbing.Reference, error) {
	h, err := r.Head()
	if err != nil {
		return nil, err
	}
	ref, err := r.CreateTag(tag, h.Hash(), &g.CreateTagOptions{
		Message: tag,
		Tagger:  rpo.Signature(),
	})
	if err != nil {
		return ref, err
	}
	if push {
		return ref, PushTag(r, true)
	}
	return ref, err
}

func Push(r *g.Repository, verbose bool) error {
	opts := &g.PushOptions{
		RemoteName: "origin",
		Auth:       rpo.Auth(),
	}
	if verbose {
		opts.Progress = os.Stdout
	}

	err := r.Push(opts)

	if err == g.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

func PushTag(r *g.Repository, verbose bool) error {
	opts := &g.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec("refs/tags/*:refs/tags/*")},
		Auth:       rpo.Auth(),
	}
	if verbose {
		opts.Progress = os.Stdout
	}
	err := r.Push(opts)

	if err == g.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}
