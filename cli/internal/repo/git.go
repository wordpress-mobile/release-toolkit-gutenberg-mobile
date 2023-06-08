package repo

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type SubmoduleRef struct {
	Repo   string
	Tag    string
	Branch string
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
		return nil, err
	}
	return r, nil
}

// go-git has an issue with cloning submodules https://github.com/go-git/go-git/issues/488
// Dropping down to exec.Command for now
func CloneGBM(path string, verbose bool) (*git.Repository, error) {
	org, _ := GetOrg("gutenberg-mobile")
	url := fmt.Sprintf("git@github.com:%s/%s.git", org, "gutenberg-mobile")

	cmd := exec.Command("git", "clone", url, "--recursive", "--depth", "1", path)

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return Open(path)
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

func CheckoutTag(r *git.Repository, tag string) error {
	co := git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/tags/" + tag),
	}
	return checkout(r, &co)
}

func CheckoutBranch(r *git.Repository, branch string) error {
	co := git.CheckoutOptions{
		Branch: plumbing.NewRemoteReferenceName("origin", branch),
	}
	return checkout(r, &co)
}

func CheckoutSha(r *git.Repository, sha string) error {
	co := git.CheckoutOptions{
		Hash: plumbing.NewHash(sha),
	}
	return checkout(r, &co)
}

func checkout(r *git.Repository, o *git.CheckoutOptions) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	return w.Checkout(o)
}

func SubmoduleUpdate(r *git.Repository, sref SubmoduleRef) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	sub, err := w.Submodule(sref.Repo)
	if err != nil {
		return err
	}
	srep, err := sub.Repository()
	if err != nil {
		return err
	}

	if sref.Tag != "" {
		return CheckoutTag(srep, sref.Tag)
	}
	return fmt.Errorf("not sure how to update the submodule")
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

func Commit(r *git.Repository, message string, opts git.CommitOptions) error {
	if opts.Author == nil {
		opts.Author = getSignature()
	}
	w, err := r.Worktree()

	if err != nil {
		return err
	}
	_, err = w.Commit(message, &opts)

	return err
}

func CommitAll(r *git.Repository, message string) error {
	return Commit(r, message, git.CommitOptions{All: true})
}

func Tag(r *git.Repository, tag, message string, push bool) error {

	h, err := r.Head()
	if err != nil {
		return err
	}
	_, err = r.CreateTag(tag, h.Hash(), &git.CreateTagOptions{
		Message: tag,
		Tagger:  getSignature(),
	})
	if err != nil {
		return err
	}
	if push {
		return PushTag(r, true)
	}
	return err
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
