package utils

import (
	"os"
	"os/exec"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
)

func SetupNode(dir string) error {
	var cmd *exec.Cmd

	// Check for nvm
	if os.Getenv("NVM_DIR") != "" {
		cmd = exec.Command("bash", "-l", "-c", ". $NVM_DIR/nvm.sh && nvm use")
		cmd.Path = "/bin/bash"
	}

	// @TODO check for asdf and set up the command accordingly

	if cmd == nil {
		console.Warn("No node version manager found. Using system node.")
		return nil
	}
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
