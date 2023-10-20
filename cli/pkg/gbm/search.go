package gbm

import (
	"fmt"

	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
)

func FindGbmReleasePr(version string) (gh.PullRequest, error) {
	label := fmt.Sprintf("label:%s", GbmReleasePrLabel)
	title := fmt.Sprintf("%s in:title", version)

	filter := gh.BuildRepoFilter(repo.GutenbergMobileRepo, "is:open", "is:pr", label, title)
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
