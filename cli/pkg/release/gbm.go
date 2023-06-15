package release

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wordpress-mobile/gbm-cli/internal/gh"
	"github.com/wordpress-mobile/gbm-cli/internal/git"
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

func CreateGbmPr(version, dir string, verbose bool) (gh.PullRequest, error) {
	l("\nPreparing Gutenberg Mobile Release PR")

	headBranch := "release/" + version
	pr := gh.PullRequest{
		Head:           gh.Repo{Ref: headBranch},
		Base:           gh.Repo{Ref: "trunk"},
		Draft:          true,
		Labels:         []gh.Label{{Name: "release-process"}},
		ReleaseVersion: version,
		Title:          "Release " + version,
		Repo:           "gutenberg-repo",
	}

	// TODO: Sometimes it can't find the GB pr right away
	gbPr, err := GetGbReleasePr(version)
	if err != nil {
		return pr, fmt.Errorf("unable to get the GB release PR (err %s)", err)
	}
	gbmr, err := gbm.PrepareBranch(dir, &pr, gbPr, verbose)
	if err != nil {
		return pr, err
	}

	l("Update the release notes")
	rnPath := filepath.Join(dir, "gutenberg-mobile", "RELEASE-NOTES.txt")
	if err := UpdateReleaseNotes(version, rnPath); err != nil {
		return pr, err
	}
	if err := git.CommitAll(gbmr, fmt.Sprintf("Release script: Update release notes for version %s", version)); err != nil {
		return pr, err
	}

	RenderGbmBody(dir, &pr)

	gh.PreviewPr("gutenberg-mobile", filepath.Join(dir, "gutenberg-mobile"), &pr)
	org, _ := repo.GetOrg("gutenberg-mobile")

	prompt := fmt.Sprintf("\nReady to create the PR on %s/gutenberg-mobile?", org)
	cont := utils.Confirm(prompt)
	if !cont {
		l("Bye 👋")
		os.Exit(0)
	}

	if err := gbm.CreatePr(gbmr, &pr, verbose); err != nil {
		return pr, err
	}

	// Update the gb release pr
	if err := renderGbPrBody(version, pr.Url, gbPr); err != nil {
		utils.LogWarn("unable to render the GB Pr body to update (err %s)", err)
	}

	if err := gh.UpdatePr(gbPr); err != nil {
		utils.LogWarn("unable to update the GB release pr (err %s)", err)
	}

	return pr, nil
}

func UpdateGbmPr(version, dir string, verbose bool) (*gh.PullRequest, error) {
	prs := GetReleasePrs(version, "gutenberg-mobile", "gutenberg")
	gbPr := prs["gutenberg"]
	gbmPr := prs["gutenberg-mobile"]
	if gbPr == nil {
		return nil, fmt.Errorf("unable to find the GB release PR")
	}

	if gbmPr == nil {
		return nil, fmt.Errorf("unable to find the GBM release PR")
	}

	gbmPr.ReleaseVersion = version

	rpo, err := gbm.PrepareBranch(dir, gbmPr, gbPr, verbose)
	if err != nil {
		return gbmPr, fmt.Errorf("issue preparing the branc (err %s)", err)
	}

	err = git.Push(rpo, verbose)
	return gbmPr, err

}

func RenderGbmBody(dir string, pr *gh.PullRequest) error {
	version := pr.ReleaseVersion

	cl := getChangeLog(dir, pr)
	rn := getReleaseNotes(dir, pr)

	rc, err := CollectReleaseChanges(version, cl, rn)
	if err != nil {
		utils.LogError("unable to collect release changes (err %s)", err)
	}
	rfs := []gh.RepoFilter{
		gh.BuildRepoFilter("gutenberg", "is:open", "is:pr", `label:"Mobile App - i.e. Android or iOS"`, fmt.Sprintf("v%s in:title", version)),
		gh.BuildRepoFilter("WordPress-Android", "is:open", "is:pr", version+" in:title"),
		gh.BuildRepoFilter("WordPress-iOS", "is:open", "is:pr", version+" in:title"),
	}

	synced, err := gh.FindGbmSyncedPrs(*pr, rfs)
	if err != nil {
		utils.LogError("unable to find synced Prs")
	}

	prs := []gh.PullRequest{}
	for _, s := range synced {
		prs = append(prs, s.Items...)
	}

	data := struct {
		Version    string
		Changes    []ReleaseChanges
		RelatedPRs []gh.PullRequest
	}{
		Version:    version,
		Changes:    rc,
		RelatedPRs: prs,
	}

	body, err := render.Render("templates/release/gbmPrBody.md", data, nil)

	if err != nil {
		utils.LogError("unable to render the GBM pr body (err %s)", err)
		pr.Body = "TBD"
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
		gbPr, err := GetGbReleasePr(gbmPr.ReleaseVersion)
		if err != nil {
			utils.LogError("unable to get the GB release PR (err %s)", err)
			return []byte{}
		}
		endpoint := fmt.Sprintf("https://raw.githubusercontent.com/%s/gutenberg/%s/packages/react-native-editor/CHANGELOG.md", org, gbPr.Head.Sha)

		if resp, err := http.Get(endpoint); err != nil {
			utils.LogError("unable to get the change log (err %s)", err)
		} else {
			defer resp.Body.Close()
			buff = resp.Body
		}
	} else {
		// Read in the change log
		clPath := filepath.Join(dir, "gutenberg-mobile", "gutenberg", "packages", "react-native-editor", "CHANGELOG.md")
		if clf, err := os.Open(clPath); err != nil {
			utils.LogError("unable to open the change log (err %s)", err)
		} else {
			defer clf.Close()
			buff = clf

		}
	}
	if data, err := io.ReadAll(buff); err != nil {
		utils.LogError("unable to read the change log (err %s)", err)
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
			utils.LogError("unable to get the change log (err %s)", err)
		} else {
			defer resp.Body.Close()
			buff = resp.Body
		}
	} else {
		// Read in the release notes
		rnPath := filepath.Join(dir, "gutenberg-mobile", "RELEASE-NOTES.txt")

		if rnf, err := os.Open(rnPath); err != nil {
			utils.LogError("unable to open the release notes (err %err)", err)
		} else {
			defer rnf.Close()
			buff = rnf
		}
	}
	if data, err := io.ReadAll(buff); err != nil {
		utils.LogError("unable to read the release notes (err %s)", err)
	} else {
		rn = data
	}

	return rn
}
