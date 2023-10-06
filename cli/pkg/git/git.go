package git

import (
	"github.com/wordpress-mobile/gbm-cli/pkg/exec"
)

func Clone(repo, dir string, shallow bool) error {
	cmd := exec.ExecGit(dir, true)
	if shallow {
		return cmd("clone", "--depth", "1", repo)
	}
	return cmd("clone", repo)
}

func Switch(dir, branch string, create bool) error {
	cmd := exec.ExecGit(dir, true)
	if create {
		return cmd("switch", "-c", branch)
	}
	return cmd("switch", branch)
}

func CommitAll(dir, msg string) error {
	return exec.ExecGit(dir, true)("commit", "-am", msg)
}
