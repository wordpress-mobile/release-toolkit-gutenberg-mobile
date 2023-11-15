package release

import (
	"fmt"

	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/repo"
)

func FindGbReleasePr(version string) (gh.PullRequest, error) {
	label := fmt.Sprintf("label:\"%s\"", GbReleasePrLabel)
	title := fmt.Sprintf("v%s in:title", version)
	filter := gh.BuildRepoFilter(repo.GutenbergRepo, "is:pr", label, title)

	pr, err := gh.SearchPr(filter)
	if err != nil {
		return gh.PullRequest{}, err
	}
	pr.ReleaseVersion = version
	return pr, nil
}

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
	label := fmt.Sprintf("label:%s", IntegratePrLabel)
	title := fmt.Sprintf(IntegratePrTitle+" in:title", version)

	filter := gh.BuildRepoFilter(repo.WordPressAndroidRepo, "is:pr", label, title)

	return gh.SearchPr(filter)
}

func FindIosReleasePr(version string) (gh.PullRequest, error) {
	label := fmt.Sprintf("label:%s", IntegratePrLabel)
	title := fmt.Sprintf(IntegratePrTitle+" in:title", version)

	filter := gh.BuildRepoFilter(repo.WordPressIosRepo, "is:pr", label, title)
	return gh.SearchPr(filter)
}

func GetGbmRelease(version string) (gh.Release, error) {
	return gh.GetReleaseByTag(repo.GutenbergMobileRepo, "v"+version)
}
