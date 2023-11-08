package repo

import (
	"fmt"
	"os"
)

const WordPressAndroidRepo = "WordPress-Android"
const WordPressIosRepo = "WordPress-iOS"
const GutenbergMobileRepo = "gutenberg-mobile"
const ReleaseToolkitGutenbergMobileRepo = "release-toolkit-gutenberg-mobile"
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

func GetOrg(repo string) string {
	switch repo {
	case GutenbergRepo:
		return WordPressOrg
	case JetpackRepo:
		return AutomatticOrg
	case GutenbergMobileRepo:
		fallthrough
	case WordPressAndroidRepo:
		fallthrough
	case ReleaseToolkitGutenbergMobileRepo:
		fallthrough
	case WordPressIosRepo:
		return WpMobileOrg
	default:
		return ""
	}
}

func GetRepoPath(repo string) string {
	org := GetOrg(repo)
	return fmt.Sprintf("git@github.com:%s/%s", org, repo)
}

func GetRepoHttpsPath(repo string) string {
	org := GetOrg(repo)
	return fmt.Sprintf("https://github.com/%s/%s", org, repo)
}
