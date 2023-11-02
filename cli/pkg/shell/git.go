package shell

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
)

type GitCmds interface {
	Clone(...string) error
	Switch(...string) error
	CommitAll(string, ...interface{}) error
	Push() error
	RemoteExists(string, string) bool
	Submodule(...string) error
	Fetch(...string) error
	SetRemoteBranches(...string) error
	AddRemote(...string) error
	SetUpstreamTo(...string) error
	IsPorcelain() bool
	PushTag(string, ...string) error
	Log(...string) error
	CherryPick(string) error
	StatConflicts() ([]string, error)
}

func (c *client) Clone(args ...string) error {
	clone := append([]string{"clone"}, args...)
	return c.cmd(clone...)
}

func (c *client) Switch(args ...string) error {
	swtch := append([]string{"switch"}, args...)
	return c.cmd(swtch...)
}

func (c *client) CommitAll(format string, args ...interface{}) error {

	if c.IsPorcelain() {
		console.Warn("No changes to commit")
		return nil
	}
	message := fmt.Sprintf(format, args...)
	return c.cmd("commit", "-am", message)
}

func (c *client) Push() error {
	return c.cmd("push", "origin", "HEAD")
}

func (c *client) RemoteExists(remote, branch string) bool {
	err := c.cmd("ls-remote", "--exit-code", "--heads", remote, branch)
	return err == nil
}

func (c *client) Submodule(args ...string) error {
	submodule := append([]string{"submodule"}, args...)
	return c.cmd(submodule...)
}

func (c *client) Fetch(args ...string) error {
	// Let's make sure we can fetch the branch by setting the remote branches
	c.cmd("remote", "set-branches", "origin", "*")
	fetch := append([]string{"fetch", "origin"}, args...)
	return c.cmd(fetch...)
}

func (c *client) SetRemoteBranches(args ...string) error {
	checkout := append([]string{"remote", "set-branches", "origin"}, args...)
	return c.cmd(checkout...)
}

func (c *client) AddRemote(args ...string) error {
	clone := append([]string{"remote", "add"}, args...)
	return c.cmd(clone...)
}

func (c *client) SetUpstreamTo(args ...string) error {
	branch := append([]string{"branch", "--set-upstream-to", "origin"}, args...)
	return c.cmd(branch...)
}

func (c *client) IsPorcelain() bool {
	err := c.cmd("diff", "--exit-code")
	return err == nil
}

func (c *client) PushTag(tag string, annotate ...string) error {
	args := []string{"tag"}
	if len(annotate) > 0 {
		args = append([]string{"-a", tag, "-m"}, annotate...)
	} else {
		args = append(args, tag)
	}

	if err := c.cmd(args...); err != nil {
		return err
	}
	return c.cmd("push", "origin", tag)
}

func (c *client) Log(args ...string) error {
	log := append([]string{"log"}, args...)
	return c.cmd(log...)
}

func (c *client) CherryPick(commitOrContinue string) error {
	if commitOrContinue == "--continue" {
		c.cmd("add", "--all")
	}
	pick := append([]string{"cherry-pick"}, commitOrContinue)
	return c.cmd(pick...)
}

func (c *client) StatConflicts() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--relative", "--diff-filter=U")
	cmd.Dir = c.dir

	out, err := cmd.Output()
	if err != nil {
		return []string{}, err
	}

	files := strings.Split(string(out), "\n")

	conflicts := []string{}
	for _, file := range files {
		if file != "" {
			conflicts = append(conflicts, file)
		}
	}

	return conflicts, nil
}
