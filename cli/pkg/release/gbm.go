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

func CreateGbmPR(version, dir string) (gh.PullRequest, error) {
	var pr gh.PullRequest

	gbmDir := fmt.Sprintf("%s/gutenberg-mobile", dir)
	git := g.NewClient(gbmDir, true)

	org, err := repo.GetOrg("gutenberg-mobile")
	console.ExitIfError(err)

	branch := "release/" + version

	console.Info("Checking if branch %s exists", branch)
	exists, _ := gh.SearchBranch("gutenberg-mobile", branch)

	if (exists != gh.Branch{}) {
		console.Info("Branch %s already exists", branch)
		return pr, nil
	} else {
		console.Info("Cloning Gutenberg Mobile to %s", dir)

		err := git.Clone(repo.GetRepoPath("gutenberg-mobile"), "--depth=1")
		if err != nil {
			return pr, err
		}

		console.Info("Checking out branch %s", branch)
		err = git.Switch(branch, "-c")
		if err != nil {
			return pr, err
		}
	}

	console.Info("Updating package versions")
	pkgs := []string{"./package.json", "./package-lock.json"}
	for _, pkg := range pkgs {
		if err := utils.UpdatePackageVersion(version, pkg); err != nil {
			return pr, err
		}
	}

	if err := git.CommitAll(gbmDir, "Release script: Update package.json version to %s", version); err != nil {
		return pr, err
	}

	if err := git.Submodule("update", "--init", "--recursive", "--depth=1", "--recommend-shallow"); err != nil {
		return pr, err
	}

	console.Info("Setting up Gutenberg Mobile node environment")
	if err := exec.SetupNode(gbmDir, true); err != nil {
		return pr, err
	}

	gbGit := g.NewClient(filepath.Join(gbmDir, "gutenberg"), true)
	if err := gbGit.Switch("rnmobile/release_" + version); err != nil {
		return pr, err
	}

	if err := git.CommitAll(gbmDir, "Release script: Update gutenberg submodule"); err != nil {
		return pr, err
	}

	if err := exec.NpmCi(gbmDir, true); err != nil {
		return pr, err
	}

	if err := exec.NpmRun(gbmDir, true, "bundle"); err != nil {
		return pr, err
	}

	if err := git.CommitAll(gbmDir, "Release script: Update bundle for %s", version); err != nil {
		return pr, err
	}

	console.Info("Update XCFramework builders project Podfile.lock")
	xcframeworkDir := fmt.Sprintf("%s/ios-xcframework", dir)

	if err := exec.BundleInstall(xcframeworkDir, true); err != nil {
		return pr, err
	}

	if err := exec.Bundle(xcframeworkDir, true, "exec", "pod", "install"); err != nil {
		return pr, err
	}

	if err := git.CommitAll(xcframeworkDir, "Release script: Sync XCFramework `Podfile.lock` with %s", version); err != nil {
		return pr, err
	}

	console.Info("Update the release-notes in the mobile package")
	chnPath := filepath.Join(gbmDir, "RELEASE-NOTES.txt")
	if err := utils.UpdateReleaseNotes(version, chnPath); err != nil {
		return pr, err
	}

	if err := git.CommitAll(gbmDir, "Release script: Update release notes for version %s", version); err != nil {
		return pr, err
	}

	console.Info("\n 🎉 Gutenberg Mobile preparations succeeded.")

	console.Info("Creating PR for %s", branch)
	pr.Title = fmt.Sprint("Release ", version)
	pr.Base.Ref = "trunk"
	pr.Head.Ref = branch

	if err := renderGbmPrBody(version, &pr); err != nil {
		console.Info("Unable to render the GB PR body (err %s)", err)
	}

	pr.Labels = []gh.Label{{
		Name: "release-process",
	}}

	gh.PreviewPr("gutenberg-mobile", gbmDir, &pr)


	prompt := fmt.Sprintf("\nReady to create the PR on %s/gutenberg?", org)
	cont := console.Confirm(prompt)

	if !cont {
		console.Info("Bye 👋")
		return pr, fmt.Errorf("exiting before creating PR")
	}
	
	if err := git.Push(); err != nil {
		return pr, err
	}

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

	body, err := render.Render(t)
	if err != nil {
		return err
	}
	pr.Body = body
	return nil
}
