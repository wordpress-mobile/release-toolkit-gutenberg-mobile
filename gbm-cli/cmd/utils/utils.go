package utils

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/inconshreveable/go-update"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/semver"
)

func GetVersionArg(args []string) (semver.SemVer, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("missing version")
	}
	version, err := semver.NewSemVer(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid version %s.  Versions must have a `Major.Minor.Patch` form", args[0])
	}

	return version, nil
}

func ExitIfError(err error, code int) {
	if err != nil {
		console.Error(err)
		Exit(code)
	}
}

func Exit(code int, deferred ...func()) {
	os.Exit(func() int {
		for _, d := range deferred {
			d()
		}
		return code
	}())
}

// Checks if running from a temp directory (go build)
// Useful for checking if running via `go run main.go`
// We ignore errors since this only relevant to local development
func CheckIfTempRun() bool {
	ex, _ := os.Executable()
	dir := filepath.Dir(ex)
	return strings.Contains(dir, "go-build")
}

// Updates the currently running executable.
func UpdateExe(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := update.Apply(resp.Body, update.Options{}); err != nil {
		return err
	}

	return nil
}

// Gets the download url for the gbm-cli executable
func exeDownloadUrl(release gh.Release) string {
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, "gbm-cli") {
			return asset.DownloadUrl
		}
	}
	return ""
}

// Checks if the currently running executable is the latest version
// If not, prompts the user to update.
// If update is confirmed, the executable is updated and the process is restarted
func CheckExeVersion(version string) {
	latestRelease, err := gh.GetLatestRelease("release-toolkit-gutenberg-mobile")
	console.ExitIfError(err)

	if latestRelease.TagName != version {
		if console.Confirm("You are running an older version of the CLI. Would you like to update?") {

			if url := exeDownloadUrl(latestRelease); url != "" {
				if err := UpdateExe(url); err != nil {
					console.ExitError("Could not update the CLI: %v", err)
				} else {
					console.Info("CLI updated successfully")
					reStart()
				}
			} else {
				console.ExitError("Could not find download url for latest release")
			}
		}
	}
}

// Restarts the process
func reStart() {
	args := os.Args
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
