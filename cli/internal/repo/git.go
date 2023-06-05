package repo

import (
	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func getAuth() *http.BasicAuth {
	token, _ := auth.TokenForHost("github")
	return &http.BasicAuth{
		Username: "gbm-cli", // this can be anything since we are using a token
		Password: token,
	}
}

func Clone(url, branch, path string) (*git.Repository, error) {
	r, err := git.PlainClone(path, false, &git.CloneOptions{
		Auth:              getAuth(),
		URL:               url,
		ReferenceName:     plumbing.ReferenceName(branch),
		Depth:             1,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
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

func Commit(r *git.Repository, message string, opts git.CommitOptions) error {
	w, err := r.Worktree()

	if err != nil {
		return err
	}
	_, err = w.Commit(message, &opts)

	return err
}

func Push(r *git.Repository) error {
	return r.Push(&git.PushOptions{
		Auth: getAuth(),
	})
}
