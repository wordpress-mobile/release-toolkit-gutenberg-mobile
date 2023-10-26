package gbm

import (
	"strings"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
)

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
