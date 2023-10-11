package git

import (
	"fmt"

	"github.com/wordpress-mobile/gbm-cli/pkg/exec"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
)

func Clone(repo, dir string, shallow bool) error {
	cmd := exec.ExecGit(dir, true)
	if shallow {
		return cmd("clone", "--depth", "1", repo)
	}
	return cmd("clone", repo)
}

func CloneGBM(dir string, pr gh.PullRequest, verbose bool) (*g.Repository, error) {
	git := exec.ExecGit(dir, verbose)

	org, _ := repo.GetOrg("gutenberg-mobile")
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
	// return Open(filepath.Join(dir, "gutenberg-mobile"))

	return false, nil
}

func Switch(dir, branch string, create bool) error {
	cmd := exec.ExecGit(dir, true)
	if create {
		return cmd("switch", "-c", branch)
	}
	return cmd("switch", branch)
}

func CommitAll(dir, format string, args ...interface{}) error {
	return exec.ExecGit(dir, true)("commit", "-am", fmt.Sprintf(format, args...))
}

func Push(dir, branch string) error {
	return exec.ExecGit(dir, true)("push", "origin", branch)
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
