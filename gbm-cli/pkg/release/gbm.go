package release

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/shell"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/utils"

	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/render"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/repo"
)

func CreateGbmPR(build Build) (gh.PullRequest, error) {
	var pr gh.PullRequest
	version := build.Version.String()
	dir := build.Dir

	sp := shell.CmdProps{Dir: dir, Verbose: true}
	git := shell.NewGitCmd(sp)

	// Set Gutenberg Mobile repository and org
	org := repo.GetOrg("gutenberg-mobile")

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
		err := git.Clone(repo.GetRepoHttpsPath("gutenberg-mobile"), "--branch", build.Base.Ref, "--depth=1", "--recursive", ".")
		if err != nil {
			return pr, fmt.Errorf("error cloning the Gutenberg Mobile repository: %v", err)
		}

		console.Info("Setting up the branch %s", branch)
		err = git.Switch("-c", branch)
		if err != nil {
			return pr, fmt.Errorf("error switching to the branch: %v", err)
		}
	}

	// Update the Gutenberg submodule
	gbBranch := "rnmobile/release_" + version
	if org != repo.WpMobileOrg {
		console.Warn("You are not using the %s org. Check the .gitmodules file to make sure the gutenberg submodule is pointing to %s/gutenberg.", repo.WpMobileOrg, org)
	}
	if exists, _ := gh.SearchBranch("gutenberg", gbBranch); (exists == gh.Branch{}) {
		return pr, fmt.Errorf("the Gutenberg branch %s does not exist on %s/gutenberg-mobile", gbBranch, org)
	}
	if err := updateGbSubmodule(gbBranch, dir, git); err != nil {
		return pr, fmt.Errorf("error updating the Gutenberg submodule: %v", err)
	}

	// Set up Gutenberg Mobile node environment
	console.Info("Setting up Node environment")
	npm := shell.NewNpmCmd(sp)
	if err := utils.SetupNode(dir); err != nil {
		return pr, fmt.Errorf("error setting up Node environment: %v", err)
	}

	// Run npm ci and npm run bundle
	if err := npm.Ci(); err != nil {
		return pr, fmt.Errorf("error running npm ci: %v", err)
	}

	// Update package version
	console.Info("Updating package versions")
	if err := npm.Version(version); err != nil {
		return pr, fmt.Errorf("error updating the package version: %v", err)
	}
	if err := git.CommitAll("Release script: Update package versions to %s", version); err != nil {
		return pr, fmt.Errorf("error committing the package version update: %v", err)
	}

	console.Info("Updating i18n files")
	if err := npm.Run("i18n:update"); err != nil {
		return pr, fmt.Errorf("error running npm run bundle: %v", err)
	}

	// Commit the updated strings
	if err := git.CommitAll("Release script: Update i18n files for %s", version); err != nil {
		return pr, fmt.Errorf("error committing the bundle update: %v", err)
	}

	if err := updateXcFramework(version, dir, git); err != nil {
		return pr, fmt.Errorf("error updating the XCFramework builders project: %v", err)
	}

	// Update the RELEASE-NOTES.txt and commit output
	console.Info("Update the release-notes in the mobile package")
	chnPath := filepath.Join(dir, "RELEASE-NOTES.txt")
	if err := UpdateReleaseNotes(version, chnPath); err != nil {
		return pr, fmt.Errorf("error updating the release notes: %v", err)
	}
	// If this is a patch release we should prompt for the wrangler to manually update the release notes
	if build.Version.IsPatchRelease() {
		console.Print(console.Highlight, "\nSince this is a patch release manually update the release notes")

		if err := openInEditor(dir, []string{"RELEASE-NOTES.txt"}); err != nil {
			console.Warn("There was an issue opening RELEASE-NOTES.txt in your editor: %v", err)
		}

		if cont := console.Confirm("Do you wish to continue after updating RELEASE-NOTES.txt?"); !cont {
			return pr, fmt.Errorf("exiting before creating PR, Stopping at RELEASE-NOTES.txt update")
		}
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

	if err := renderGbmPrBody(dir, version, &pr); err != nil {
		console.Info("Unable to render the GB PR body (err %s)", err)
	}

	// Add PR labels
	pr.Labels = []gh.Label{{
		Name: "release-process",
	}}

	// Display PR preview
	gh.PreviewPr("gutenberg-mobile", dir, build.Base.Ref, pr)

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

func renderGbmPrBody(dir string, version string, pr *gh.PullRequest) error {
	cl, err := getChangeLog(dir, pr)
	if err != nil {
		console.Warn(err.Error())
	}
	rn, err := getReleaseNotes(dir, pr)
	if err != nil {
		console.Warn(err.Error())
	}

	rc, err := CollectReleaseChanges(version, cl, rn)

	if err != nil {
		console.Error(err)
	}

	rfs := []gh.RepoFilter{
		gh.BuildRepoFilter("gutenberg", "is:open", "is:pr", `label:"Mobile App - i.e. Android or iOS"`, fmt.Sprintf("v%s in:title", version)),
		gh.BuildRepoFilter("WordPress-Android", "is:open", "is:pr", version+" in:title"),
		gh.BuildRepoFilter("WordPress-iOS", "is:open", "is:pr", version+" in:title"),
	}

	synced, err := gbm.FindGbmSyncedPrs(*pr, rfs)
	if err != nil {
		console.Error(err)
	}

	prs := []gh.PullRequest{}
	for _, s := range synced {
		prs = append(prs, s.Items...)
	}

	t := render.Template{
		Path: "templates/release/gbm_pr_body.md",
		Data: struct {
			Version    string
			GbmPrUrl   string
			Changes    []ReleaseChanges
			RelatedPRs []gh.PullRequest
		}{
			Version:    version,
			Changes:    rc,
			RelatedPRs: prs,
		},
	}

	body, err := render.Render(t)
	if err != nil {
		return err
	}

	pr.Body = body
	return nil
}

func getChangeLog(dir string, gbmPr *gh.PullRequest) ([]byte, error) {
	var buff io.ReadCloser
	cl := []byte{}

	if dir == "" {
		gbPr, _ := FindGbReleasePr(gbmPr.ReleaseVersion)

		org := repo.GetOrg("gutenberg")
		endpoint := fmt.Sprintf("https://raw.githubusercontent.com/%s/gutenberg/%s/packages/react-native-editor/CHANGELOG.md", org, gbPr.Head.Sha)

		if resp, err := http.Get(endpoint); err != nil {
			return cl, fmt.Errorf("unable to get the changelog (err %s)", err)
		} else {
			defer resp.Body.Close()
			buff = resp.Body
		}

	} else {
		// Read in the change log
		clPath := filepath.Join(dir, "gutenberg-mobile", "gutenberg", "packages", "react-native-editor", "CHANGELOG.md")
		if clf, err := os.Open(clPath); err != nil {
			return cl, fmt.Errorf("unable to open the changelog %s", err)
		} else {
			defer clf.Close()
			buff = clf
		}
	}

	if data, err := io.ReadAll(buff); err != nil {
		return cl, fmt.Errorf("unable to read the changelog %s", err)
	} else {
		cl = data
	}

	return cl, nil
}

func getReleaseNotes(dir string, gbmPr *gh.PullRequest) ([]byte, error) {
	var buff io.ReadCloser
	rn := []byte{}

	if dir == "" {
		org := repo.GetOrg("gutenberg-mobile")
		endpoint := fmt.Sprintf("https://raw.githubusercontent.com/%s/gutenberg-mobile/%s/RELEASE-NOTES.txt", org, gbmPr.Head.Sha)

		if resp, err := http.Get(endpoint); err != nil {
			return rn, fmt.Errorf("unable to get the changelog (err %s)", err)
		} else {
			defer resp.Body.Close()
			buff = resp.Body
		}
	} else {
		// Read in the release notes
		rnPath := filepath.Join(dir, "gutenberg-mobile", "RELEASE-NOTES.txt")

		if rnf, err := os.Open(rnPath); err != nil {
			return rn, fmt.Errorf("unable to open the release notes %s", err)
		} else {
			defer rnf.Close()
			buff = rnf
		}
	}
	if data, err := io.ReadAll(buff); err != nil {
		return rn, fmt.Errorf("unable to read the release notes %s", err)
	} else {
		rn = data
	}

	return rn, nil
}

func updateGbSubmodule(gbBranch, dir string, git shell.GitCmds) error {
	console.Info("Updating Gutenberg submodule")
	// Create a git client for Gutenberg submodule so the Gutenberg ref can be
	// updated to the correct branch
	gbSp := shell.CmdProps{Dir: filepath.Join(dir, "gutenberg"), Verbose: true}
	gbGit := shell.NewGitCmd(gbSp)

	if err := gbGit.Fetch(gbBranch); err != nil {
		return fmt.Errorf("error fetching the Gutenberg branch: %v", err)
	}

	if err := gbGit.Switch(gbBranch); err != nil {
		return fmt.Errorf("error checking out the Gutenberg branch: %v", err)
	}

	if err := git.CommitAll("Release script: Update gutenberg submodule"); err != nil {
		return fmt.Errorf("error committing the gutenberg submodule update: %v", err)
	}
	return nil
}

func updateXcFramework(version, dir string, git shell.GitCmds) error {
	console.Info("Update XCFramework builders project Podfile.lock")

	// set up a shell command for the ios-xcframework directory
	xcSp := shell.CmdProps{Dir: fmt.Sprintf("%s/ios-xcframework", dir), Verbose: true}

	bundle := shell.NewBundlerCmd(xcSp)

	// Run `bundle install`
	if err := bundle.Install(); err != nil {
		return fmt.Errorf("error running bundle install: %v", err)
	}

	// Run `bundle exec pod install``
	if err := bundle.PodInstall(); err != nil {
		return fmt.Errorf("error running bundle exec pod install: %v", err)
	}

	// Commit output of bundle commands
	if err := git.CommitAll("Release script: Sync XCFramework `Podfile.lock` with %s", version); err != nil {
		return fmt.Errorf("error committing the XCFramework `Podfile.lock` update: %v", err)
	}

	return nil
}
