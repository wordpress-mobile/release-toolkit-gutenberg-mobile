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
			var cmd *exec.Cmd

			// If we are running on a CI and NVM is available we run `nvm use` before each npm command
			// to make sure we are using the correct node version
			ci := os.Getenv("CI")
			if ci == "true" && os.Getenv("NVM_DIR") != "" {
				withNvmUse := append([]string{"-l", "-c", "nvm use && npm"}, cmds...)
				cmd = exec.Command("bash", withNvmUse...)
			} else {
				cmd = exec.Command("npm", cmds...)
			}

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
