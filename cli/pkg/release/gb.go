package release

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wordpress-mobile/gbm-cli/internal/exc"
	"github.com/wordpress-mobile/gbm-cli/internal/gh"
	"github.com/wordpress-mobile/gbm-cli/internal/git"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

func CreateGbPR(version, dir string, verbose bool) (gh.PullRequest, error) {
	pr := gh.PullRequest{}

	gbBranchName := fmt.Sprintf("rnmobile/release_%s", version)
	org, _ := repo.GetOrg("gutenberg-mobile")
	l("Preparing the %s release from %s/%s", version, org, "gutenberg-mobile")

	l("Checking if branch %s exists", gbBranchName)
	existing, _ := gh.SearchBranch("gutenberg", gbBranchName)

	if (existing != gh.Branch{}) {
		l("Branch %s already exists", gbBranchName)
		pr, err := GetGbReleasePr(version)
		if err != nil {
			utils.LogWarn("Unable to get the GB release PR (err %s)", err)
		}
		return *pr, &gh.BranchError{Err: errors.New("branch already exists"), Type: "exists"}
	}

	gbmDir := filepath.Join(dir, "gutenberg-mobile")
	gbDir := filepath.Join(gbmDir, "gutenberg")

	l("Cloning GBM repo to %s", gbmDir)
	_, err := git.CloneGBM(dir, pr, verbose)
	if err != nil {

		return pr, err
	}

	l("Validating aztec")

	valid, err := ValidateAztecVersions(AztecSrc{GbmDir: gbmDir})
	if err != nil {
		return pr, err
	}
	if !valid {
		return pr, errors.New("aztec versions are not valid")
	}
	l("Aztec version validated")
	valid, err = ValidateVersion(version)
	if err != nil {
		return pr, err
	}
	if !valid {
		return pr, errors.New("version is not valid")
	}
	l("Release version validated")

	l("Switching gutenberg to %s", gbBranchName)
	gbr, err := git.Open(gbDir)
	if err != nil {
		return pr, err
	}
	if err := git.Checkout(gbr, gbBranchName); err != nil {
		return pr, err
	}

	l("Set up Node")
	if err := exc.SetupNode(gbDir, verbose); err != nil {
		return pr, err
	}

	l("Installing npm packages")
	if err := exc.NpmCi(gbDir, verbose); err != nil {
		return pr, err
	}

	l("Update the version in the Gutenberg packages")

	pkgs := []string{"react-native-aztec", "react-native-bridge", "react-native-editor"}
	for _, pkg := range pkgs {
		editorPackPath := filepath.Join(gbDir, "packages", pkg, "package.json")
		if err := UpdatePackageVersion(version, editorPackPath); err != nil {
			return pr, err
		}
	}

	if err := git.CommitAll(gbr, fmt.Sprintf("Release script: Update react-native-editor version to %s", version)); err != nil {
		return pr, err
	}

	l("Update bundle")

	// Run bundle install directly since the preios script sometimes fails
	bundlePath := filepath.Join(gbDir, "packages", "react-native-editor", "ios")
	if err := exc.BundleInstall(bundlePath, verbose); err != nil {
		return pr, err
	}

	if err := exc.NpmRunCorePreios(gbmDir, verbose); err != nil {
		return pr, err
	}
	if clean, err := git.IsPorcelain(gbr); err != nil {
		utils.LogWarn("Could not check if the repo is clean: %s", err)
	} else if !clean {
		if err := git.CommitAll(gbr, "Release script: Update podfile"); err != nil {
			return pr, err
		}
	}

	l("Update the change notes in the mobile editor package")
	chnPath := filepath.Join(gbmDir, "gutenberg", "packages", "react-native-editor", "CHANGELOG.md")
	if err := UpdateChangeLog(version, chnPath); err != nil {
		return pr, err
	}
	if err := git.CommitAll(gbr, fmt.Sprintf("Release script: Update changelog for version %s", version)); err != nil {
		return pr, err
	}

	l("\n 🎉 Gutenberg preparations succeeded.")

	// Prepare the PR
	pr.Title = fmt.Sprint("Mobile Release v", version)
	pr.Base.Ref = "trunk"
	pr.Head.Ref = gbBranchName

	if err := renderGbPrBody(version, "", &pr); err != nil {
		utils.LogWarn("Unable to render the GB PR body (err %s)", err)
	}

	pr.Labels = []gh.Label{{
		Name: "Mobile App - i.e. Android or iOS",
	}}

	gh.PreviewPr("gutenberg", gbDir, &pr)

	prompt := fmt.Sprintf("\nReady to create the PR on %s/gutenberg?", org)
	cont := utils.Confirm(prompt)

	if !cont {
		l("Bye 👋")
		os.Exit(0)
	}

	// TODO: Warn if the submodule remote is not set to the script config
	// Right now it will use what ever is in the Gutenberg Mobile gitmodules file
	l("Creating the PR")
	if err := git.Push(gbr, verbose); err != nil {
		return pr, err
	}
	if err := gh.CreatePr("gutenberg", &pr); err != nil {
		return pr, err
	}

	if pr.Number == 0 {
		return pr, fmt.Errorf("unable to create the pr")
	}

	/*
		l("Pushing the release tag")
		if err := repo.Tag(gbr, fmt.Sprint("rnmobile/", version), fmt.Sprint("Mobile Release v", version), true); err != nil {
			utils.LogWarn("Unable to push the release tag: %s", err)
		}
	*/
	return pr, nil
}

func renderGbPrBody(version, gbmPRUrl string, pr *gh.PullRequest) error {
	pd := struct {
		Version  string
		GbmPrUrl string
	}{
		Version:  version,
		GbmPrUrl: gbmPRUrl,
	}

	body, err := render.Render("templates/release/gbPrBody.md", pd, nil)
	if err != nil {
		return err
	}
	pr.Body = body
	return nil
}
