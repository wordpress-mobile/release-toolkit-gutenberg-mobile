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
