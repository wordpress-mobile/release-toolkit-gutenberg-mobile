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
