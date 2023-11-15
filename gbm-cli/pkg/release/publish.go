package release

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/repo"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/semver"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/shell"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/utils"
)

func Publish(version semver.SemVer, dir string) (gh.Release, error) {

	rel := gh.Release{}

	// Check if the release already exists
	console.Info("Checking if release already exists")
	release, _ := gh.GetReleaseByTag(repo.GutenbergMobileRepo, version.Vstring())

	if release.TagName != "" {
		return rel, errors.New("release already exists")
	}
	// Get the release PR
	console.Info("Getting release PR")
	pr, err := FindGbmReleasePr(version.String())
	if err != nil {
		return rel, err
	}

	// Check to see if the CI tests passed
	console.Info("Checking if CI tests passed")
	status, err := gh.GetStatusChecks(repo.GutenbergMobileRepo, pr.Head.Sha)
	if err != nil {
		return rel, err
	}

	if !status.Passed {
		console.Print(console.Heading, "The release PR has not passed the CI tests")
		console.Print(console.Heading, "URL: %s", pr.Url)
		console.Print(console.Heading, "Status: %s", status.State)
		console.Print(console.HeadingRow, "%-10s %-10s", "Check", "Status")
		for _, check := range status.Statuses {
			console.Print(console.Row, "%-10s %-10s", check.Description, check.State)
		}

		if !console.Confirm("Continue?") {
			return rel, errors.New("release cancelled")
		}
	}

	sp := shell.CmdProps{Dir: dir, Verbose: true}
	git := shell.NewGitCmd(sp)

	// Clone the Gutenberg Mobile repository
	console.Info("Cloning the Gutenberg Mobile repository")
	err = git.Clone(repo.GetRepoHttpsPath("gutenberg-mobile"), "--branch", pr.Head.Ref, "--depth=1", "--recursive", ".")
	if err != nil {
		return rel, fmt.Errorf("error cloning the Gutenberg Mobile repository: %v", err)
	}

	// Set up Gutenberg Mobile node environment
	console.Info("Setting up Gutenberg Mobile node environment")
	if err := utils.SetupNode(dir); err != nil {
		return rel, fmt.Errorf("error setting up Node environment: %v", err)
	}
	npm := shell.NewNpmCmd(shell.CmdProps{Dir: dir, Verbose: true})

	if err := npm.Ci(); err != nil {
		return rel, fmt.Errorf("error running npm ci: %v", err)
	}

	// Build the js maps
	console.Info("Building the js maps")
	runCmd := func(cmd *exec.Cmd) error {
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	if err := npm.Run("bundle:android"); err != nil {
		return rel, fmt.Errorf("error running npm run bundle:android: %v", err)
	}

	cmd := exec.Command("node", "gutenberg/node_modules/react-native/scripts/compose-source-maps.js", "bundle/android/App.text.js.map", "bundle/android/App.js.map", "-o ./android-App.js.map")
	if err := runCmd(cmd); err != nil {
		return rel, fmt.Errorf("error composing the Android source maps: %v", err)
	}

	if err := npm.Run("bundle:ios"); err != nil {
		return rel, fmt.Errorf("error running npm run bundle:ios: %v", err)
	}
	cmd = exec.Command("node", "gutenberg/node_modules/react-native/scripts/compose-source-maps.js", "bundle/ios/App.text.js.map", "bundle/ios/App.js.map", "-o ./ios-App.js.map")
	if err := runCmd(cmd); err != nil {
		return rel, fmt.Errorf("error composing the iOS source maps: %v", err)
	}

	// Create the release
	rel.TagName = version.Vstring()
	rel.Name = version.Vstring()
	// TODO: get the changes from the release PR body
	rel.Body = "Release v" + version.String()
	rel.Target = pr.Head.Sha

	if !console.Confirm(fmt.Sprintf("Create release %s on %s?", rel.TagName, repo.GetOrg(repo.GutenbergMobileRepo))) {
		return rel, errors.New("release cancelled")
	}

	console.Info("Creating release")
	err = gh.CreateRelease(repo.GutenbergMobileRepo, &rel)
	if err != nil {
		return rel, err
	}

	// Upload the js maps to the release
	console.Info("Uploading the js maps to the release")
	if err := gh.UploadReleaseAssets(repo.GutenbergMobileRepo, dir, rel, "./android-App.js.map", "./ios-App.js.map"); err != nil {
		return rel, fmt.Errorf("error uploading the js maps: %v", err)
	}

	return rel, nil
}
