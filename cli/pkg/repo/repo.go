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
