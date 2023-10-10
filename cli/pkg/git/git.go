package git

import (
	"fmt"

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

func CommitAll(dir, format string, args ...interface{}) error {
	return exec.ExecGit(dir, true)("commit", "-am", fmt.Sprintf(format, args...))
}

func Push(dir, branch string) error {
	return exec.ExecGit(dir, true)("push", "origin", branch)
}
