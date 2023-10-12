package exec

import (
	"os"
	"os/exec"
)

func Git(dir string, verbose bool) func(...string) error {
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

// Deprecated: Use Git instead
func ExecGit(dir string, verbose bool) func(...string) error {
	return Git(dir, verbose)
}

func SetupNode(dir string, verbose bool) error {
	// Check for nvm
	_, ok := os.LookupEnv("NVM_DIR")
	if ok {
		exc(verbose, dir, "nvm", "use")
	}

	return nil
}

func NpmCi(dir string, verbose bool) error {
	return exc(verbose, dir, "npm", "ci")
}

func NpmRun(dir string, verbose bool, args ...string) error {
	return exc(verbose, dir, "npm", append([]string{"run"}, args...)...)
}

func Bundle(dir string, verbose bool, args ...string) error {
	return exc(verbose, dir, "bundle", args...)
}

func BundleInstall(dir string, verbose bool, args ...string) error {
	return Bundle(dir, true, append([]string{"install"}, args...)...)
}

func exc(verbose bool, dir, cmd string, args ...string) error {
	exc := exec.Command(cmd, args...)

	exc.Dir = dir

	if verbose {
		exc.Stdout = os.Stdout
		exc.Stderr = os.Stderr
	}

	return exc.Run()
}
