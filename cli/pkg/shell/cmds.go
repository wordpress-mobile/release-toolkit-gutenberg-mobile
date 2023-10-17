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

func NpmCmd(cp CmdProps) NpmCmds {
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

func GitCmd(cp CmdProps) gitCmds {
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

func BundlerCmd(cp CmdProps) BundlerCmds {
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

// common commands

// Install is used by npm and bundler
func (c *client) Install(args ...string) error {
	install := append([]string{"install"}, args...)
	return c.cmd(install...)
}
