package repo

import (
	"fmt"
	"os"
)

var (
	WpMobileOrg   string
	WordPressOrg  string
	AutomatticOrg string
)

func init() {
	initOrgs()
}

func initOrgs() {
	if gbmWpMobileOrg, ok := os.LookupEnv("GBM_WPMOBILE_ORG"); !ok {
		WpMobileOrg = "wordpress-mobile"
	} else {
		WpMobileOrg = gbmWpMobileOrg
	}

	if gbmWordPressOrg, ok := os.LookupEnv("GBM_WORDPRESS_ORG"); !ok {
		WordPressOrg = "WordPress"
	} else {
		WordPressOrg = gbmWordPressOrg
	}

	if gbmAutomatticOrg, ok := os.LookupEnv("GBM_AUTOMATTIC_ORG"); !ok {
		AutomatticOrg = "Automattic"
	} else {
		AutomatticOrg = gbmAutomatticOrg
	}
}

func GetOrg(repo string) (string, error) {
	switch repo {
	case "gutenberg":
		return WordPressOrg, nil
	case "jetpack":
		return AutomatticOrg, nil
	case "gutenberg-mobile":
		fallthrough
	case "WordPress-Android":
		fallthrough
	case "WordPress-iOS":
		return WpMobileOrg, nil
	default:
		return "", fmt.Errorf("unknown repo: %s", repo)
	}
}

func GetGbmReleasePr(version string) (PullRequest, error) {
	return getReleasePr("gutenberg-mobile", version)
}

func GetGbReleasePr(version string) (PullRequest, error) {
	return getReleasePr("gutenberg", version)
}

func getReleasePr(repo, version string) (PullRequest, error) {
	filter := BuildRepoFilter(repo, "is:pr", fmt.Sprintf("%s in:title", version))

	res, err := SearchPrs(filter)
	if err != nil {
		return PullRequest{}, nil
	}

	if res.TotalCount == 0 {
		return PullRequest{}, fmt.Errorf("no release PRs found for `%s`", version)
	}
	if res.TotalCount != 1 {
		return PullRequest{}, fmt.Errorf("found multiple prs for %s", version)
	}
	pr := res.Items[0]
	pr.Repo = repo
	return pr, nil
}
