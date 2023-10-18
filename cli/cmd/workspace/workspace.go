package workspace

import (
	"os"
	"os/signal"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
)

var tempDir string
var Cleanup func()

func SetTempDir() (string, error) {
	tempDir, err := os.MkdirTemp("", "gbm-")
	if err != nil {
		return "", err
	}
	return tempDir, nil
}

func GetTempDir() string {
	return tempDir
}

func UpdateCleaner(keep bool) {
	Cleanup = Cleaner(tempDir, keep)
}

func Cleaner(tempDir string, keep bool) func() {
	clean := func() {
		if keep {
			console.Info("Keeping temporary directory %s", tempDir)
			return
		}
		cleanupTempDir()
	}

	// register a listener for ^C, call the cleanup function
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan // wait for ^C
		clean()
		os.Exit(1)
	}()
	return clean
}

func cleanupTempDir() {
	// do nothing if we don't have a temp dir
	if tempDir == "" {
		return
	}

	console.Info("Cleaning up temporary directory %s", tempDir)
	err := os.RemoveAll(tempDir)
	if err != nil {
		console.Error(err)
	}
}

func init() {
	if _, noWorkspace := os.LookupEnv("GBM_NO_WORKSPACE"); noWorkspace {
		console.Info("GBM_NO_WORKSPACE is set, not creating a workspace")
		return
	}

	var err error
	tempDir, err = SetTempDir()
	if err != nil {
		console.Error(err)
		os.Exit(1)
	}

	// Set up the default cleaner
	Cleanup = Cleaner(tempDir, false)
}
