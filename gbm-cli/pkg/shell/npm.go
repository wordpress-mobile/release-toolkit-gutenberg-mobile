package shell

import (
	"os"
	"os/exec"
	"strings"
)

type NpmCmds interface {
	Install(...string) error
	Ci() error
	Run(...string) error
	RunIn(string, ...string) error
	Version(string) error
	VersionIn(string, string) error
}

// Check to see if a node manager is available and set up the command accordingly
func switchNodeCmd(cmds ...string) *exec.Cmd {

	// Check if nvm is installed
	if os.Getenv("NVM_DIR") != "" {
		nvmCmd := "nvm use && npm " + strings.Join(cmds, " ")
		nvmCheck := exec.Command("bash", "-l", "-c", "nvm")
		if err := nvmCheck.Run(); err != nil {
			// Load nvm before running npm
			return exec.Command("bash", "-l", "-c", ". $NVM_DIR/nvm.sh && "+nvmCmd)
		} else {
			return exec.Command("bash", "-l", "-c", nvmCmd)
		}
	}

	// Other node managers can be added here...

	// Use system node
	return exec.Command("npm", cmds...)
}

func (c *client) Ci() error {
	return c.cmd("ci")
}

func (c *client) Run(args ...string) error {
	run := append([]string{"run"}, args...)
	return c.cmd(run...)
}

func (c *client) RunIn(path string, args ...string) error {
	run := append([]string{"run"}, args...)
	return c.cmdInPath(path, run...)
}

func (c *client) Version(version string) error {
	// Let's not add the tag by default.
	// If we need it we should consider a different function.
	versionCmd := []string{"version", version, "--no-git-tag=false"}
	return c.cmd(versionCmd...)
}

func (c *client) VersionIn(packagePath, version string) error {
	versionCmd := []string{"version", version, "--no-git-tag=false"}
	return c.cmdInPath(packagePath, versionCmd...)
}
