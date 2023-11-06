package utils

import (
	"os"
	"os/exec"

	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/console"
)

func SetupNode(dir string) error {
	var cmd *exec.Cmd

	// Check for nvm
	if os.Getenv("NVM_DIR") != "" {
		console.Info("Setting up node via nvm in %s", dir)
		cmd = exec.Command("bash", "-l", "-c", "$NVM_DIR/nvm.sh", "use")
		cmd.Path = "/bin/bash"
	}

	// @TODO check for asdf and set up the command accordingly

	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
