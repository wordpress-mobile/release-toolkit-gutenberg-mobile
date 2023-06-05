package release

import (
	"fmt"

	"github.com/wordpress-mobile/gbm-cli/internal/repo"
)

func GetGbmReleasePr(version string) (repo.PullRequest, error) {
	filter := repo.BuildRepoFilter("gutenberg-mobile", "is:pr", fmt.Sprintf("%s in:title", version))

	res, err := repo.SearchPrs(filter)
	if err != nil {
		return repo.PullRequest{}, nil
	}

	if res.TotalCount == 0 {
		return repo.PullRequest{}, fmt.Errorf("no release PRs found for `%s`", version)
	}
	if res.TotalCount != 1 {
		return repo.PullRequest{}, fmt.Errorf("found multiple prs for %s", version)
	}
	return res.Items[0], nil
}
