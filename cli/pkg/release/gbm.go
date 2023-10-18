package release

import (
	"fmt"
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
	git := shell.GitCmd(sp)

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
			return pr, fmt.Errorf("error cloning the Gutenberg Mobile repository: %v", err)
		}

		console.Info("Add remote for %s", org)
		err = git.AddRemote("upstream", repo.GetRepoPath("gutenberg-mobile"))
		if err != nil {
			return pr, fmt.Errorf("error adding remote for %s: %v", org, err)
		}

		console.Info("Set upstream to trunk", org)
		err = git.SetUpstreamTo("trunk")
		if err != nil {
			return pr, fmt.Errorf("error setting upstream to trunk: %v", err)
		}

		console.Info("Checking out branch %s", branch)
		err = git.Switch("-c", branch)
		if err != nil {
			return pr, fmt.Errorf("error checking out the branch: %v", err)
		}
	}

	// Set up Gutenberg Mobile node environment
	console.Info("Setting up Node environment")
	npm := shell.NpmCmd(sp)
	if err := exec.SetupNode(dir, true); err != nil {
		return pr, fmt.Errorf("error setting up Node environment: %v", err)
	}
	// Run npm ci and npm run bundle
	if err := npm.Ci(); err != nil {
		return pr, fmt.Errorf("error running npm ci: %v", err)
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
	gbGit := shell.GitCmd(gbSp)

	if err := gbGit.Fetch(gbBranch); err != nil {
		return pr, fmt.Errorf("error fetching the Gutenberg branch: %v", err)
	}

	if err := gbGit.Switch(gbBranch); err != nil {
		return pr, fmt.Errorf("error checking out the Gutenberg branch: %v", err)
	}

	if err := git.CommitAll("Release script: Update gutenberg submodule"); err != nil {
		return pr, fmt.Errorf("error committing the gutenberg submodule update: %v", err)
	}

	console.Info("Bundling Gutenberg Mobile")
	if err := npm.Run("bundle"); err != nil {
		return pr, fmt.Errorf("error running npm run bundle: %v", err)
	}

	// Commit the updated Gutenberg submodule ref
	if git.IsPorcelain() {
		console.Info("Nothing to commit after bundling")
	} else {
		// Commit the updated bundle output
		if err := git.CommitAll("Release script: Update bundle for %s", version); err != nil {
			return pr, fmt.Errorf("error committing the bundle update: %v", err)
		}
	}

	// Update XCFramework builders project Podfile.lock
	console.Info("Update XCFramework builders project Podfile.lock")

	// set up a shell command for the ios-xcframework directory
	xcSp := sp
	xcSp.Dir = fmt.Sprintf("%s/ios-xcframework", dir)
	bundle := shell.BundlerCmd(xcSp)

	// Run `bundle install`
	if err := bundle.Install(); err != nil {
		return pr, fmt.Errorf("error running bundle install: %v", err)
	}

	// Run `bundle exec pod install``
	if err := bundle.PodInstall(); err != nil {
		return pr, fmt.Errorf("error running bundle exec pod install: %v", err)
	}

	// Commit output of bundle commands
	if err := git.CommitAll("Release script: Sync XCFramework `Podfile.lock` with %s", version); err != nil {
		return pr, fmt.Errorf("error committing the XCFramework `Podfile.lock` update: %v", err)
	}

	// Update the RELEASE-NOTES.txt and commit output
	console.Info("Update the release-notes in the mobile package")
	chnPath := filepath.Join(dir, "RELEASE-NOTES.txt")
	if err := utils.UpdateReleaseNotes(version, chnPath); err != nil {
		return pr, fmt.Errorf("error updating the release notes: %v", err)
	}

	if err := git.CommitAll("Release script: Update release notes for version %s", version); err != nil {
		return pr, fmt.Errorf("error committing the release notes update: %v", err)
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
		return pr, fmt.Errorf("error pushing the branch: %v", err)
	}

	// Create the PR
	if err := gh.CreatePr("gutenberg-mobile", &pr); err != nil {
		return pr, fmt.Errorf("error creating the PR: %v", err)
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

	body, err := render.Render(t)
	if err != nil {
		return err
	}
	pr.Body = body
	return nil
}
