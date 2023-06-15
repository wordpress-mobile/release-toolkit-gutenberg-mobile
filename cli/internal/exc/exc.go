package exc

import (
	"os"
	"os/exec"
)

func Npm(dir string, verbose bool, args ...string) error {
	return exc(verbose, dir, "npm", args...)
}

func NpmCi(dir string, verbose bool) error {
	return exc(verbose, dir, "npm", "ci")
}

func NpmRunBundle(dir string, verbose bool) error {
	return exc(verbose, dir, "npm", "run", "bundle")
}

func NpmRunCorePreios(dir string, verbose bool) error {
	return exc(verbose, dir, "npm", "run", "core", "preios")
}

func SetupNode(dir string, verbose bool) error {

	// Check for nvm
	_, ok := os.LookupEnv("NVM_DIR")
	if ok {
		exc(verbose, dir, "nvm", "use")
	}

	return nil
}

func BundleInstall(dir string, verbose bool) error {
	return exc(verbose, dir, "bundle", "install")
}

func PodInstall(dir string, verbose bool) error {
	return exc(verbose, dir, "bundle", "exec", "pod", "install")
}

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

func exc(verbose bool, dir, cmd string, args ...string) error {
	exc := exec.Command(cmd, args...)

	exc.Dir = dir

	if verbose {
		exc.Stdout = os.Stdout
		exc.Stderr = os.Stderr
	}

	return exc.Run()
}
