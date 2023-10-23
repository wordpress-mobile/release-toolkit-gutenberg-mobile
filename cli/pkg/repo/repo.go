package repo

import (
	"fmt"
	"os"
)

const WordPressAndroidRepo = "WordPress-Android"
const WordPressIosRepo = "WordPress-iOS"
const GutenbergMobileRepo = "gutenberg-mobile"
const GutenbergRepo = "gutenberg"
const JetpackRepo = "jetpack"

var (
	WpMobileOrg   string
	WordPressOrg  string
	AutomatticOrg string
)

func init() {
	InitOrgs()
}

func InitOrgs() {
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
	case GutenbergRepo:
		return WordPressOrg, nil
	case JetpackRepo:
		return AutomatticOrg, nil
	case GutenbergMobileRepo:
		fallthrough
	case WordPressAndroidRepo:
		fallthrough
	case WordPressIosRepo:
		return WpMobileOrg, nil
	default:
		return "", fmt.Errorf("unknown repo: %s", repo)
	}
}

func GetRepoPath(repo string) string {
	org, _ := GetOrg(repo)
	return fmt.Sprintf("git@github.com:%s/%s", org, repo)
}
