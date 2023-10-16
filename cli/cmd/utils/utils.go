package utils

import (
	"fmt"
	"os"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

func GetVersionArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("missing version")
	}
	return utils.NormalizeVersion(args[0])
}

func SetTempDir() (string, error) {
	tempDir, err := os.MkdirTemp("", "gbm-")
	if err != nil {
		return "", err
	}
	return tempDir, nil
}

func CleanupTempDir(tempDir string) error {
	console.Info("Cleaning up temporary directory %s", tempDir)
	err := os.RemoveAll(tempDir)
	if err != nil {
		return err
	}
	return nil
}

func ExitIfErrorHandler(deferred func()) func(error, int) {
	return func(err error, code int) {
		if err != nil {
			console.Error(err)

			os.Exit(func() int {
				defer deferred()
				return code
			}())

		}
	}
}
