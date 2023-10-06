package exec

import (
	"os"
	"os/exec"
)

func ExecGit(dir string, verbose bool) func(...string) error {
	return func(cmds ...string) error {
		cmd := exec.Command("git", cmds...)
		cmd.Dir = dir

		if verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		return cmd.Run()
	}
}
