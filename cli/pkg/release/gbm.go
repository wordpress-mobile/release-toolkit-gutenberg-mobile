package release

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

/*
 This will use internal/gbm/CreateGbmPr to create the PR
 it will just do the extra release specific stuff
 like updating the version number an release notes

 It can also post the testing instructions as comments to the PR
*/

func CreateGbmPr(version, dir string, verbose bool) (repo.PullRequest, error) {

	l := logger(verbose)

	l("\nPreparing Gutenberg Mobile Release PR")

	headBranch := "release/" + version
	pr := repo.PullRequest{
		Head:           repo.Repo{Ref: headBranch},
		Base:           repo.Repo{Ref: "trunk"},
		Draft:          true,
		Labels:         []repo.Label{{Name: "release-process"}},
		ReleaseVersion: version,
		Title:          "Release " + version,
		Repo:           "gutenberg-repo",
	}

	gbmr, err := gbm.PrepareBranch(dir, &pr, verbose)
	if err != nil {
		return pr, err
	}

	renderBody(dir, &pr)

	repo.PreviewPr("gutenberg-mobile", filepath.Join(dir, "gutenberg-mobile"), &pr)
	org, _ := repo.GetOrg("gutenberg-mobile")

	prompt := fmt.Sprintf("\nReady to create the PR on %s/gutenberg?", org)
	cont := utils.Confirm(prompt)
	if !cont {
		l("Bye 👋")
		os.Exit(0)
	}

	if err := gbm.CreatePr(gbmr, &pr, verbose); err != nil {
		return pr, err
	}

	return pr, nil
}

func renderBody(dir string, pr *repo.PullRequest) {
	version := pr.ReleaseVersion

	// Read in the change log
	clPath := filepath.Join(dir, "gutenberg-mobile", "gutenberg", "packages", "react-native-editor", "CHANGELOG.md")
	cl := []byte{}
	if clf, err := os.Open(clPath); err != nil {
		utils.LogError("unable to open the change log (err %s)", err)
	} else {
		defer clf.Close()
		if data, err := io.ReadAll(clf); err != nil {
			utils.LogError("unable to read the change log (err %s)", err)
		} else {
			cl = data
		}
	}

	// Read in the release notes
	rnPath := filepath.Join(dir, "gutenberg-mobile", "RELEASE-NOTES.txt")
	rn := []byte{}
	if rnf, err := os.Open(rnPath); err != nil {
		utils.LogError("unable to open the release notes (err %err)", err)
	} else {
		defer rnf.Close()
		if data, err := io.ReadAll(rnf); err != nil {
			utils.LogError("unable to read the release notes (err %s)", err)
		} else {
			rn = data
		}
	}

	rc, err := CollectReleaseChanges(version, cl, rn)
	if err != nil {
		utils.LogError("unable to collect release changes (err %s)", err)
	}
	rfs := []repo.RepoFilter{
		repo.BuildRepoFilter("gutenberg", "is:open", "is:pr", `label:"Mobile App - i.e. Android or iOS"`),
		repo.BuildRepoFilter("WordPress-Android", "is:open", "is:pr"),
		repo.BuildRepoFilter("WordPress-iOS", "is:open", "is:pr"),
	}

	synced, err := repo.FindGbmSyncedPrs(*pr, rfs)
	if err != nil {
		utils.LogError("unable to find synced Prs")
	}

	prs := []repo.PullRequest{}
	for _, s := range synced {
		prs = append(prs, s.Items...)
	}

	data := struct {
		Version    string
		Changes    []ReleaseChanges
		RelatedPRs []repo.PullRequest
	}{
		Version:    version,
		Changes:    rc,
		RelatedPRs: prs,
	}

	body, err := render.Render("templates/release/gbmPrBody.md", data, nil)

	if err != nil {
		utils.LogError("unable to render the GBM pr body (err %s)", err)
		pr.Body = "TBD"
	}

	pr.Body = body
}
