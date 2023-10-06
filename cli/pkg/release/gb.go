package release

import (
	"fmt"
	"path/filepath"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/exec"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/git"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

func CreateGbPR(version, dir string) (gh.PullRequest, error) {
	var pr gh.PullRequest
	gbDir := fmt.Sprintf("%s/gutenberg", dir)

	branch := fmt.Sprintf("rnmobile/release_%s", version)

	console.Info("Checking if branch %s exists", branch)
	exists, _ := gh.SearchBranch("gutenberg", branch)

	if (exists != gh.Branch{}) {
		console.Info("Branch %s already exists", branch)
		return pr, nil
	} else {
		console.Info("Cloning Gutenberg to %s", dir)
		err := git.Clone(repo.GetRepoPath("gutenberg"), dir, true)
		if err != nil {
			return pr, err
		}

		console.Info("Checking out branch %s", branch)
		err = git.Switch(gbDir, branch, true)
		if err != nil {
			return pr, err
		}
	}

	console.Info("Updating package versions")
	pkgs := []string{"react-native-aztec", "react-native-bridge", "react-native-editor"}
	for _, pkg := range pkgs {
		editorPackPath := filepath.Join(gbDir, "packages", pkg, "package.json")
		if err := utils.UpdatePackageVersion(version, editorPackPath); err != nil {
			return pr, err
		}
	}
	if err := git.CommitAll(gbDir, "Release script: Update react-native-editor version to %s", version); err != nil {
		return pr, err
	}

	console.Info("Update the change notes in the mobile editor package")
	chnPath := filepath.Join(gbDir, "packages", "react-native-editor", "CHANGELOG.md")
	if err := utils.UpdateChangeLog(version, chnPath); err != nil {
		return pr, err
	}
	if err := git.CommitAll(gbDir, "Release script: Update changelog for version %s", version); err != nil {
		return pr, err
	}

	console.Info("Setting up Gutenberg node environment")

	if err := exec.SetupNode(gbDir, true); err != nil {
		return pr, err
	}

	if err := exec.NpmCi(gbDir, true); err != nil {
		return pr, err
	}

	console.Info("Running preios script")

	// Run bundle install directly since the preios script sometimes fails
	editorIosPath := filepath.Join(gbDir, "packages", "react-native-editor", "ios")

	if err := exec.BundleInstall(editorIosPath, true); err != nil {
		return pr, err
	}

	if err := exec.NpmRun(editorIosPath, true, "preios"); err != nil {
		return pr, err
	}

	if err := git.CommitAll(gbDir, "Release script: Update podfile"); err != nil {
		return pr, err
	}

	console.Info("\n 🎉 Gutenberg preparations succeeded.")

	return pr, fmt.Errorf("not implemented")
}
