package release

import (
	"fmt"
	"path/filepath"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
	"github.com/wordpress-mobile/gbm-cli/pkg/shell"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

func CreateGbPR(version, dir string, noTag bool) (gh.PullRequest, error) {
	var pr gh.PullRequest

	shellProps := shell.CmdProps{Dir: dir, Verbose: true}
	git := shell.NewGitCmd(shellProps)
	npm := shell.NewNpmCmd(shellProps)

	org := repo.GetOrg("gutenberg")
	branch := "rnmobile/release_" + version

	exists, _ := gh.SearchBranch("gutenberg", branch)

	if (exists != gh.Branch{}) {
		console.Warn("Branch %s already exists", branch)

		cont := console.Confirm("Do you wish to continue? (The remote branch will be deleted.)")

		if !cont {
			console.Info("Bye 👋")
			return pr, fmt.Errorf("exiting before creating PR")
		}

		// Delete the branch on the GitHub repo
		if err := gh.DeleteBranch("gutenberg", branch); err != nil {
			return pr, fmt.Errorf("error deleting the branch: %v", err)
		}
	}

	console.Info("Cloning Gutenberg to %s", dir)

	// Let's clone into the current directory so that the git client can find the .git directory
	err := git.Clone(repo.GetRepoPath("gutenberg"), "--depth=1", ".")

	if err != nil {
		return pr, fmt.Errorf("error cloning the Gutenberg repository: %v", err)
	}

	console.Info("Checking out branch %s", branch)
	err = git.Switch("-c", branch)
	if err != nil {
		return pr, fmt.Errorf("error checking out the branch: %v", err)
	}

	console.Info("Updating package versions")
	pkgs := []string{"react-native-aztec", "react-native-bridge", "react-native-editor"}
	for _, pkg := range pkgs {
		editorPackPath := filepath.Join(dir, "packages", pkg)
		if err := npm.VersionIn(editorPackPath, version); err != nil {
			return pr, fmt.Errorf("error updating the package version: %v", err)
		}
	}

	if err := git.CommitAll("Release script: Update react-native-editor version to %s", version); err != nil {
		return pr, err
	}

	console.Info("Update the CHANGELOG in the react-native-editor package")
	chnPath := filepath.Join(dir, "packages", "react-native-editor", "CHANGELOG.md")
	if err := UpdateChangeLog(version, chnPath); err != nil {
		return pr, fmt.Errorf("error updating the CHANGELOG: %v", err)
	}
	if err := git.CommitAll("Release script: Update CHANGELOG for version %s", version); err != nil {
		return pr, fmt.Errorf("error committing the CHANGELOG updates: %v", err)
	}

	console.Info("Setting up Gutenberg node environment")

	if err := utils.SetupNode(dir); err != nil {
		return pr, fmt.Errorf("error setting up the node environment: %v", err)
	}

	if err := npm.Install(); err != nil {
		return pr, fmt.Errorf("error running npm ci: %v", err)
	}

	console.Info("Running preios script")

	// Run bundle install directly since the preios script sometimes fails
	editorIosPath := filepath.Join(dir, "packages", "react-native-editor", "ios")

	iosShellProps := shell.CmdProps{Dir: editorIosPath, Verbose: true}
	bundle := shell.NewBundlerCmd(iosShellProps)
	if err := bundle.Install(); err != nil {
		return pr, fmt.Errorf("error running bundle install: %v", err)
	}

	if err := npm.RunIn(editorIosPath, "preios"); err != nil {
		return pr, fmt.Errorf("error running npm run core preios: %v", err)
	}

	if err := git.CommitAll("Release script: Update podfile"); err != nil {
		return pr, fmt.Errorf("error committing the Podfile changes: %v", err)
	}

	console.Info("🎉 Gutenberg preparations succeeded.")

	// Prepare the GB PR
	console.Info("Creating PR")
	pr.Title = fmt.Sprint("Mobile Release v", version)
	pr.Base.Ref = "trunk"
	pr.Head.Ref = branch

	if err := renderGbPrBody(version, &pr); err != nil {
		return pr, fmt.Errorf("error rendering the GB pull body: %v", err)
	}

	pr.Labels = []gh.Label{
		{
			Name: "Mobile App - i.e. Android or iOS",
		},
		{
			Name: "[Type] Build Tooling",
		},
	}

	gh.PreviewPr("gutenberg", dir, pr)

	prompt := fmt.Sprintf("\nReady to create the PR on %s/gutenberg?", org)
	cont := console.Confirm(prompt)

	if !cont {
		return pr, fmt.Errorf("exiting before creating PR")
	}

	if err := git.Push(); err != nil {
		return pr, fmt.Errorf("error pushing the PR: %v", err)
	}

	if err := gh.CreatePr("gutenberg", &pr); err != nil {
		return pr, fmt.Errorf("error creating the PR: %v", err)
	}

	if pr.Number == 0 {
		return pr, fmt.Errorf("pr was not created successfully")
	}

	if !noTag {
		console.Info("Adding release tag")
		if err := git.PushTag("rnmobile/" + version); err != nil {
			console.Warn("Error tagging the release: %v", err)
		}
	} else {
		console.Warn("Skipping tag creation")
	}

	return pr, nil
}

func renderGbPrBody(version string, pr *gh.PullRequest) error {

	t := render.Template{
		Path: "templates/release/gb_pr_body.md",
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
