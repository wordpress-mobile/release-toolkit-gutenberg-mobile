package git

import (
	"fmt"

	"github.com/wordpress-mobile/gbm-cli/pkg/exec"
)

type Client interface {
	Clone(...string) error
	Switch(...string) error
	CommitAll(string, ...interface{}) error
	Push() error
	RemoteExists(string, string) bool
}

type client struct {
	dir     string
	verbose bool
}

func NewClient(dir string, verbose bool) Client {
	return &client{
		dir:     dir,
		verbose: verbose,
	}
}

func (c *client) Clone(args ...string) error {
	cmd := exec.Git(c.dir, c.verbose)
	clone := append([]string{"clone"}, args...)
	return cmd(clone...)
}

func (c *client) Switch(args ...string) error {
	cmd := exec.Git(c.dir, c.verbose)
	swtch := append([]string{"switch"}, args...)
	return cmd(swtch...)
}

func (c *client) CommitAll(format string, args ...interface{}) error {
	cmd := exec.Git(c.dir, c.verbose)
	message := fmt.Sprintf(format, args...)
	return cmd("commit", "-am", message)
}

func (c *client) Push() error {
	cmd := exec.Git(c.dir, c.verbose)
	return cmd("push", "origin", "HEAD")
}

func (c *client) RemoteExists(remote, branch string) bool {
	cmd := exec.Git(c.dir, c.verbose)
	err := cmd("ls-remote", "--exit-code", "--heads", remote, branch)
	return err == nil
}

func GetSubmodule(r gh.Repo, path string) (*g.Submodule, error) {
	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	return w.Submodule(path)
}

func CommitSubmodule(dir, message, submodule string, verbose bool) error {
	git := exec.ExecGit(dir, verbose)

	if err := git("add", submodule); err != nil {
		return fmt.Errorf("unable to add submodule %s in %s :%s", submodule, dir, err)
	}

	if err := git("commit", "-m", message); err != nil {
		return fmt.Errorf("unable to commit submodule update %s : %s", submodule, err)
	}
	return nil
}

func IsSubmoduleCurrent(s gh.Submodule, expectedHash string) (bool, error) {
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

func IsPorcelain(r gh.Repo) (bool, error) {
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
