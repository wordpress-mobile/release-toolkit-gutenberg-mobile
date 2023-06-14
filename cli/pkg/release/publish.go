package release

import (
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
)

// Checks if the release PRs: exist, mergable, approved, and are passing.
// It returns early for non-existent PRs, otherwise it collects the reasons for
// not being ready to publish.
func IsReadyToPublish(version string, skipChecks []repo.CheckRun, verbose bool) (bool, []string) {
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

	if !repo.IsPrApproved(gbPr) {
		ok = false
		reasons = append(reasons, "GB PR is not approved")
	}

	if !repo.IsPrApproved(gbmPr) {
		ok = false
		reasons = append(reasons, "GBM PR is not approved")
	}

	if !repo.IsPrPassing(gbPr, skipChecks, verbose) {
		ok = false
		reasons = append(reasons, "GB PR is not passing")
	}

	if !repo.IsPrPassing(gbmPr, skipChecks, verbose) {
		ok = false
		reasons = append(reasons, "GBM PR is not passing")
	}

	return ok, reasons
}
