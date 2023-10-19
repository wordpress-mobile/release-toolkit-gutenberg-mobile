package release

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/exec"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/shell"

	"github.com/wordpress-mobile/gbm-cli/pkg/render"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

func CreateGbmPR(version, dir string) (gh.PullRequest, error) {
	var pr gh.PullRequest

	sp := shell.CmdProps{Dir: dir, Verbose: true}
	git := shell.NewGitCmd(sp)

	// Set Gutenberg Mobile repository and org
	org, err := repo.GetOrg("gutenberg-mobile")
	console.ExitIfError(err)

	// Set Gutenberg Mobile branch name e.g., (release/x.xx.x)
	branch := "release/" + version

	// Check if branch already exists
	// Return if it does
	// Otherwise, clone the repo and checkout the branch
	console.Info("Checking if branch %s exists", branch)
	exists, _ := gh.SearchBranch("gutenberg-mobile", branch)

	if (exists != gh.Branch{}) {
		console.Info("Branch %s already exists", branch)
		return pr, nil
	} else {
		console.Info("Cloning Gutenberg Mobile to %s", dir)
		err := git.Clone(repo.GetRepoPath("gutenberg-mobile"), "--depth=1", "--recursive", ".")
		if err != nil {
			return pr, err
		}

		console.Info("Add remote for %s", org)
		err = git.AddRemote("upstream", repo.GetRepoPath("gutenberg-mobile"))
		if err != nil {
			return pr, err
		}

		console.Info("Set upstream to trunk", org)
		err = git.SetUpstreamTo("trunk")
		if err != nil {
			return pr, err
		}

		console.Info("Checking out branch %s", branch)
		err = git.Switch("-c", branch)
		if err != nil {
			return pr, err
		}
	}

	// Set up Gutenberg Mobile node environment
	console.Info("Setting up Node environment")
	npm := shell.NewNpmCmd(sp)

	if err := exec.SetupNode(dir, true); err != nil {
		return pr, err
	}
	// Run npm ci and npm run bundle
	if err := npm.Ci(); err != nil {
		return pr, err
	}

	// Commit package.json and package-lock.json
	// Update package versions for package.json and package-lock.json
	console.Info("Updating package versions")
	updatePackageJson(dir, version, "package.json", "package-lock.json")

	// Create a git client for Gutenberg submodule so the Gutenberg ref can be updated to the correct branch
	console.Info("Updating Gutenberg submodule")
	gbBranch := "rnmobile/release_" + version
	if org != repo.WpMobileOrg {
		console.Warn("You are not using the %s org. Check the .gitmodules file to make sure the gutenberg submodule is pointing to %s/gutenberg.", repo.WpMobileOrg, org)
	}
	if exists, _ := gh.SearchBranch("gutenberg", gbBranch); (exists == gh.Branch{}) {
		return pr, fmt.Errorf("the Gutenberg branch %s does not exist on %s/gutenberg-mobile", gbBranch, org)
	}

	gbSp := sp
	gbSp.Dir = filepath.Join(dir, "gutenberg")
	gbGit := shell.NewGitCmd(gbSp)


	if err := gbGit.Fetch(gbBranch); err != nil {
		return pr, err
	}

	if err := gbGit.Switch(gbBranch); err != nil {
		return pr, err
	}

	if err := git.CommitAll("Release script: Update gutenberg submodule"); err != nil {
		return pr, err
	}

	console.Info("Bundling Gutenberg Mobile")
	if err := npm.Run("bundle"); err != nil {
		return pr, err
	}

	// Commit the updated Gutenberg submodule ref
	if git.IsPorcelain() {
		console.Info("Nothing to commit after bundling")
	} else {
		// Commit the updated bundle output
		if err := git.CommitAll("Release script: Update bundle for %s", version); err != nil {
			return pr, err
		}
	}

	// Update XCFramework builders project Podfile.lock
	console.Info("Update XCFramework builders project Podfile.lock")

	// set up a shell command for the ios-xcframework directory
	xcSp := sp
	xcSp.Dir = fmt.Sprintf("%s/ios-xcframework", dir)
	bundle := shell.NewBundlerCmd(xcSp)


	// Run `bundle install`
	if err := bundle.Install(); err != nil {
		return pr, err
	}

	// Run `bundle exec pod install``
	if err := bundle.PodInstall(); err != nil {
		return pr, err
	}

	// Commit output of bundle commands
	if err := git.CommitAll("Release script: Sync XCFramework `Podfile.lock` with %s", version); err != nil {
		return pr, err
	}

	// Update the RELEASE-NOTES.txt and commit output
	console.Info("Update the release-notes in the mobile package")
	chnPath := filepath.Join(dir, "RELEASE-NOTES.txt")
	if err := utils.UpdateReleaseNotes(version, chnPath); err != nil {
		return pr, err
	}

	if err := git.CommitAll("Release script: Update release notes for version %s", version); err != nil {
		return pr, err
	}

	console.Info("\n 🎉 Gutenberg Mobile preparations succeeded.")

	// Create Gutenberg Mobile PR
	console.Info("Creating PR for %s", branch)
	pr.Title = fmt.Sprint("Release ", version)
	pr.Base.Ref = "trunk"
	pr.Head.Ref = branch

	if err := renderGbmPrBody(version, &pr); err != nil {
		console.Info("Unable to render the GB PR body (err %s)", err)
	}

	// Add PR labels
	pr.Labels = []gh.Label{{
		Name: "release-process",
	}}

	// Display PR preview
	gh.PreviewPr("gutenberg-mobile", dir, pr)

	// Add prompt to confirm PR creation
	prompt := fmt.Sprintf("\nReady to create the PR on %s/gutenberg-mobile?", org)
	cont := console.Confirm(prompt)

	if !cont {
		console.Info("Bye 👋")
		return pr, fmt.Errorf("exiting before creating PR")
	}

	// Push the branch
	if err := git.Push(); err != nil {
		return pr, err
	}

	// Create the PR
	if err := gh.CreatePr("gutenberg-mobile", &pr); err != nil {
		return pr, err
	}

	if pr.Number == 0 {
		return pr, fmt.Errorf("failed to create the PR")
	}

	return pr, nil
}

func renderGbmPrBody(version string, pr *gh.PullRequest) error {
	t := render.Template{
		Path: "templates/release/gbm_pr_body.md",
		Data: struct {
			Version  string
			GbmPrUrl string
		}{
			Version: version,
		},
	}
	cl := getChangeLog(dir, pr)
	rn := getReleaseNotes(dir, pr)

	body, err := render.Render(t)
	if err != nil {
		return err
	}
	pr.Body = body
	return nil
}

func getChangeLog(dir string, gbmPr *gh.PullRequest) []byte {
	var buff io.ReadCloser
	cl := []byte{}

	if dir == "" {
		org, _ := repo.GetOrg("gutenberg")
		endpoint := fmt.Sprintf("https://raw.githubusercontent.com/%s/gutenberg/%s/packages/react-native-editor/CHANGELOG.md", org, gbPr.Head.Sha)

		if resp, err := http.Get(endpoint); err != nil {
			fmt.Errorf("unable to get the changelog (err %s)", err)
		} else {
			defer resp.Body.Close()
			buff = resp.Body
		}
	} else {
		// Read in the change log
		clPath := filepath.Join(dir, "gutenberg-mobile", "gutenberg", "packages", "react-native-editor", "CHANGELOG.md")
		if clf, err := os.Open(clPath); err != nil {
			fmt.Errorf("unable to open the changelog %s", err)
		} else {
			defer clf.Close()
			buff = clf

		}
	}
	if data, err := io.ReadAll(buff); err != nil {
		fmt.Errorf("unable to read the changelog %s", err)
	} else {
		cl = data
	}

	return cl
}

func getReleaseNotes(dir string, gbmPr *gh.PullRequest) []byte {
	var buff io.ReadCloser
	rn := []byte{}

	if dir == "" {
		org, _ := repo.GetOrg("gutenberg-mobile")
		endpoint := fmt.Sprintf("https://raw.githubusercontent.com/%s/gutenberg-mobile/%s/RELEASE-NOTES.txt", org, gbmPr.Head.Sha)

		if resp, err := http.Get(endpoint); err != nil {
			fmt.Errorf("unable to get the changelog (err %s)", err)
		} else {
			defer resp.Body.Close()
			buff = resp.Body
		}
	} else {
		// Read in the release notes
		rnPath := filepath.Join(dir, "gutenberg-mobile", "RELEASE-NOTES.txt")

		if rnf, err := os.Open(rnPath); err != nil {
			fmt.Errorf("unable to open the release notes %s", err)
		} else {
			defer rnf.Close()
			buff = rnf
		}
	}
	if data, err := io.ReadAll(buff); err != nil {
		fmt.Errorf("unable to read the release notes %s", err)
	} else {
		rn = data
	}

	return rn
}
