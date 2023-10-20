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
	cmd func(...string) error
}

func NewNpmCmd(cp CmdProps) NpmCmds {
	return &client{
		cmd: func(cmds ...string) error {
			cmd := exec.Command("npm", cmds...)
			cmd.Dir = cp.Dir

			if cp.Verbose {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			}

			return cmd.Run()
		},
	}
}

func NewGitCmd(cp CmdProps) GitCmds {
	return &client{
		cmd: func(cmds ...string) error {
			cmd := exec.Command("git", cmds...)
			cmd.Dir = cp.Dir

			if cp.Verbose {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			}

			return cmd.Run()
		},
	}
}

func NewBundlerCmd(cp CmdProps) BundlerCmds {
	return &client{
		cmd: func(cmds ...string) error {
			cmd := exec.Command("bundle", cmds...)
			cmd.Dir = cp.Dir

			if cp.Verbose {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			}

			return cmd.Run()
		},
	}
}

func NewRakeCmd(cp CmdProps) RakeCmds {
	return &client{
		cmd: func(cmds ...string) error {
			cmd := exec.Command("rake", cmds...)
			cmd.Dir = cp.Dir

			if cp.Verbose {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			}

			return cmd.Run()
		},
	}
}

// common commands
// Install is used by npm and bundler
func (c *client) Install(args ...string) error {
	install := append([]string{"install"}, args...)
	return c.cmd(install...)
}
