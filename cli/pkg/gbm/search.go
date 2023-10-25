package gbm

import (
	"fmt"
	"strings"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
)

func FindGbmReleasePr(version string) (gh.PullRequest, error) {
	label := fmt.Sprintf("label:%s", GbmReleasePrLabel)
	title := fmt.Sprintf("%s in:title", version)

	filter := gh.BuildRepoFilter(repo.GutenbergMobileRepo, "is:pr", "is:open", label, title)
	pr, err := gh.SearchPr(filter)
	if err != nil {
		return gh.PullRequest{}, err
	}
	pr.ReleaseVersion = version
	return pr, nil
}

func FindAndroidReleasePr(version string) (gh.PullRequest, error) {
	label := fmt.Sprintf("label:%s", IntegrationPrLabel)
	title := fmt.Sprintf("v%s in:title", version)

	filter := gh.BuildRepoFilter(repo.WordPressAndroidRepo, "is:open", "is:pr", label, title)

	return gh.SearchPr(filter)
}

func FindIosReleasePr(version string) (gh.PullRequest, error) {
	label := fmt.Sprintf("label:%s", IntegrationPrLabel)
	title := fmt.Sprintf("v%s in:title", version)

	filter := gh.BuildRepoFilter(repo.WordPressIosRepo, "is:open", "is:pr", label, title)
	return gh.SearchPr(filter)
}
func FindGbmSyncedPrs(gbmPr gh.PullRequest, filters []gh.RepoFilter) ([]gh.SearchResult, error) {
	var synced []gh.SearchResult
	prChan := make(chan gh.SearchResult)

	// Search for PRs in parallel
	for _, rf := range filters {
		go func(rf gh.RepoFilter) {
			res, err := gh.SearchPrs(rf)

			// just log the error and continue
			if err != nil {
				console.Warn("could not search for %s", err)
			}
			prChan <- res
		}(rf)
	}

	// Wait for all the PRs to be returned
	for i := 0; i < len(filters); i++ {
		resp := <-prChan
		sItems := []gh.PullRequest{}

		for _, pr := range resp.Items {
			if strings.Contains(pr.Body, gbmPr.Url) {
				pr.Repo = resp.Filter.Repo
				sItems = append(sItems, pr)
			}
		}
		resp.Items = sItems
		synced = append(synced, resp)
	}

	return synced, nil
}

func GetGbmRelease(version string) (gh.Release, error) {
	return gh.GetReleaseByTag(repo.GutenbergMobileRepo, "v"+version)
}
