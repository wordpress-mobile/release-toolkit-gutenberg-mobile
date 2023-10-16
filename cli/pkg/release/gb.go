package release

import (
	"fmt"
	"path/filepath"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/exec"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	g "github.com/wordpress-mobile/gbm-cli/pkg/git"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

func CreateGbPR(version, dir string) (gh.PullRequest, error) {
	var pr gh.PullRequest

	git := g.NewClient(dir, true)

	org, err := repo.GetOrg("gutenberg")
	console.ExitIfError(err)

	branch := "rnmobile/release_" + version

	exists, _ := gh.SearchBranch("gutenberg", branch)

	if (exists != gh.Branch{}) {
		console.Warn("Branch %s already exists", branch)

		cont := console.Confirm("Do you wish to continue?")

		if !cont {
			console.Info("Bye 👋")
			return pr, fmt.Errorf("exiting before creating PR")
		}

	} else {
		console.Info("Cloning Gutenberg to %s", dir)

		// Let's clone into the current directory so that the git client can find the .git directory
		err := git.Clone(repo.GetRepoPath("gutenberg"), "--depth=1", ".")

		if err != nil {
			console.ExitError("There was an error cloning the Gutenberg repository.")
		}

		console.Info("Checking out branch %s", branch)
		err = git.Switch("-c", branch)
		if err != nil {
			console.ExitError("There was an error checking out the branch.")
		}
	}

	console.Info("Updating package versions")
	pkgs := []string{"react-native-aztec", "react-native-bridge", "react-native-editor"}
	for _, pkg := range pkgs {
		editorPackPath := filepath.Join(dir, "packages", pkg, "package.json")
		if err := utils.UpdatePackageVersion(version, editorPackPath); err != nil {
			console.ExitError("There was an error updating the package version.")
		}
	}

	if err := git.CommitAll("Release script: Update react-native-editor version to %s", version); err != nil {
		return pr, err
	}

	console.Info("Update the CHANGELOG in the react-native-editor package")
	chnPath := filepath.Join(dir, "packages", "react-native-editor", "CHANGELOG.md")
	if err := utils.UpdateChangeLog(version, chnPath); err != nil {
		console.ExitError("There was an error updating CHANGELOG in the react-native-editor package.")
	}
	if err := git.CommitAll("Release script: Update CHANGELOG for version %s", version); err != nil {
		console.ExitError("There was an error committing the CHANGELOG updates.")
	}

	console.Info("Setting up Gutenberg node environment")

	if err := exec.SetupNode(dir, true); err != nil {
		console.ExitError("There was an error setting up the node environment.")
	}

	if err := exec.NpmCi(dir, true); err != nil {
		console.ExitError("There was an error running npm ci.")
	}

	console.Info("Running preios script")

	// Run bundle install directly since the preios script sometimes fails
	editorIosPath := filepath.Join(dir, "packages", "react-native-editor", "ios")

	if err := exec.BundleInstall(editorIosPath, true); err != nil {
		console.ExitError("There was an error running bundle install.")
	}

	if err := exec.NpmRun(editorIosPath, true, "preios"); err != nil {
		console.ExitError("There was an error running `npm run core preios`.")
	}

	if err := git.CommitAll("Release script: Update podfile"); err != nil {
		console.ExitError("There was an error commiting Podfile changes.")
	}

	console.Info("\n 🎉 Gutenberg preparations succeeded.")

	// Prepare the GB PR
	console.Info("Creating PR")
	pr.Title = fmt.Sprint("Mobile Release v", version)
	pr.Base.Ref = "trunk"
	pr.Head.Ref = branch

	if err := renderGbPrBody(version, &pr); err != nil {
		console.ExitError("There was an error rendering the GB pull body.")
	}

	pr.Labels = []gh.Label{{
		Name: "Mobile App - i.e. Android or iOS",
	}}

	gh.PreviewPr("gutenberg", dir, &pr)

	prompt := fmt.Sprintf("\nReady to create the PR on %s/gutenberg?", org)
	cont := console.Confirm(prompt)

	if !cont {
		console.ExitError("Exiting without creating the PR.")
	}

	if err := git.Push(); err != nil {
		console.ExitError("There was an error pushing the pull request.")
	}

	if err := gh.CreatePr("gutenberg", &pr); err != nil {
		console.ExitError("There was an error creating the PR.")
	}

	if pr.Number == 0 {
		console.ExitError("There was an error creating the PR.")
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
