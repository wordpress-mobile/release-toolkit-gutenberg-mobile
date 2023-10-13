package exec

import (
	"errors"
	"os"
	"os/exec"
	"time"
)

// Deprecated: Use shell package instead
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

// Deprecated: Use shell package instead
func ExecGit(dir string, verbose bool) func(...string) error {
	return Git(dir, verbose)
}

func SetupNode(dir string, verbose bool) error {
	// Check for nvm

	exc(verbose, dir, "nvm", "use")

	return nil
}

// Deprecated: Use shell package instead
func NpmCi(dir string, verbose bool) error {
	return exc(verbose, dir, "npm", "ci")
}

// Deprecated: Use shell package instead
func NpmRun(dir string, verbose bool, args ...string) error {
	return exc(verbose, dir, "npm", append([]string{"run"}, args...)...)
}

// Deprecated: Use shell package instead
func Bundle(dir string, verbose bool, args ...string) error {
	return exc(verbose, dir, "bundle", args...)
}

// Deprecated: Use shell package instead
func BundleInstall(dir string, verbose bool, args ...string) error {
	return Bundle(dir, true, append([]string{"install"}, args...)...)
}

func Try(times int, cmd string, args ...string) error {

	for times > 0 {
		err := exc(true, "", cmd, args...)
		if err == nil {
			return nil
		}
		times--
		time.Sleep(time.Second)
	}
	return errors.New("failed to execute command")
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
