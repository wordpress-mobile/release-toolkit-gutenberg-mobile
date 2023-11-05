package shell

import (
	"os"
	"os/exec"
)

type CmdProps struct {
	Dir     string
	Verbose bool
}

type client struct {
	cmd       func(...string) error
	cmdInPath func(string, ...string) error
	dir       string
}

func execute(cmd *exec.Cmd, dir string, verbose bool) error {
	cmd.Dir = dir
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}

func NewNpmCmd(cp CmdProps) NpmCmds {
	return &client{
		cmd: func(cmds ...string) error {
			cmd := exec.Command("npm", cmds...)
			return execute(cmd, cp.Dir, cp.Verbose)
		},
		cmdInPath: func(path string, cmds ...string) error {
			cmd := exec.Command("npm", cmds...)
			return execute(cmd, path, cp.Verbose)
		},
		dir: cp.Dir,
	}
}

func NewGitCmd(cp CmdProps) GitCmds {
	return &client{
		cmd: func(cmds ...string) error {
			cmd := exec.Command("git", cmds...)
			return execute(cmd, cp.Dir, cp.Verbose)
		},
		cmdInPath: func(path string, cmds ...string) error {
			cmd := exec.Command("git", cmds...)
			return execute(cmd, path, cp.Verbose)
		},
		dir: cp.Dir,
	}
}

func NewBundlerCmd(cp CmdProps) BundlerCmds {
	return &client{
		cmd: func(cmds ...string) error {
			cmd := exec.Command("bundle", cmds...)
			return execute(cmd, cp.Dir, cp.Verbose)
		},
		cmdInPath: func(path string, cmds ...string) error {
			cmd := exec.Command("bundle", cmds...)
			return execute(cmd, path, cp.Verbose)
		},
		dir: cp.Dir,
	}
}

func NewRakeCmd(cp CmdProps) RakeCmds {
	return &client{
		cmd: func(cmds ...string) error {
			cmd := exec.Command("rake", cmds...)
			return execute(cmd, cp.Dir, cp.Verbose)
		},
		cmdInPath: func(path string, cmds ...string) error {
			cmd := exec.Command("rake", cmds...)
			return execute(cmd, path, cp.Verbose)
		},
		dir: cp.Dir,
	}
}

// common commands
// Install is used by npm and bundler
func (c *client) Install(args ...string) error {
	install := append([]string{"install"}, args...)
	return c.cmd(install...)
}
