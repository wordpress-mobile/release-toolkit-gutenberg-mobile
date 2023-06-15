package release

import (
	"fmt"
	"io"
	"net/http"

	"github.com/wordpress-mobile/gbm-cli/internal/gh"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

// Checks if the release PRs: exist, mergable, approved, and are passing.
// It returns early for non-existent PRs, otherwise it collects the reasons for
// not being ready to publish.
func IsReadyToPublish(version string, skipChecks, verbose bool) (bool, []string) {
	prs := GetReleasePrs(version, "gutenberg-mobile", "gutenberg")
	if len(prs) == 0 {
		return false, []string{"No release PRs found"}
	}
	ok := true
	reasons := []string{}
	gbmPr := prs["gutenberg-mobile"]
	gbPr := prs["gutenberg"]

	if gbmPr == nil {
		return false, []string{"GBM PR does not exist"}
	}
	if gbPr == nil {
		return false, []string{"GB PR does not exist"}
	}

	// From now on, collect the reasons for not being ready to publish
	if !gbmPr.Mergeable {
		ok = false
		reasons = append(reasons, "GBM PR is not mergeable")
	}
	if !gbPr.Mergeable {
		ok = false
		reasons = append(reasons, "GB PR is not mergeable")
	}

	if !gh.IsPrApproved(gbPr) {
		ok = false
		reasons = append(reasons, "GB PR is not approved")
	}

	if !gh.IsPrApproved(gbmPr) {
		ok = false
		reasons = append(reasons, "GBM PR is not approved")
	}

	if skipChecks {
		l(utils.WarnString("Skipping check runs"))
	} else {
		if !gh.IsPrPassing(gbPr, nil, verbose) {
			ok = false
			reasons = append(reasons, "GB PR is not passing")
		}

		if !gh.IsPrPassing(gbmPr, nil, verbose) {
			ok = false
			reasons = append(reasons, "GBM PR is not passing")
		}
	}

	return ok, reasons
}

func TagGb(version string, verbose bool) error {
	pr, err := GetGbReleasePr(version)
	if err != nil {
		return fmt.Errorf("unable to get the GB release PR: %w", err)
	}

	if _, err := gh.CreateAnnotatedTag("gutenberg", pr.Head.Sha, "rnmobile/"+version, pr.Title); err != nil {
		return fmt.Errorf("unable to create the GB release tag: %w", err)
	}

	return nil
}

func PublishGbmRelease(version string, verbose bool) error {
	// Get the new release notes for the GBM release
	org, _ := repo.GetOrg("gutenberg-mobile")
	rnUrl := fmt.Sprintf("https://raw.githubusercontent.com/%s/gutenberg-mobile/release/%s/RELEASE-NOTES.txt", org, version)
	resp, err := http.Get(rnUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	relNotes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Collect the changes from the GBM release notes
	changes, err := CollectReleaseChanges(version, nil, relNotes)
	if err != nil {
		return err
	}

	// Render the release body
	data := struct {
		Version string
		Changes []ReleaseChanges
	}{
		Version: "v" + version,
		Changes: changes,
	}

	body, err := render.Render("templates/release/gbmReleaseBody.md", data, nil)

	if err != nil {
		return err
	}

	// Create the release
	rp := &gh.ReleaseProps{
		TagName:         "v" + version,
		TargetCommitish: "release/" + version,
		Name:            "Release " + version,
		Body:            body,
	}

	return gh.CreateRelease("gutenberg-mobile", rp)
}
