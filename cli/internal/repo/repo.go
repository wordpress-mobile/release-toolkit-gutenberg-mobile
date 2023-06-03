package repo

import (
	"fmt"
	"os"
)

var (
	wpMobileOrg   string
	wordPressOrg  string
	automatticOrg string
)

func init() {
	initOrgs()
}

func initOrgs() {
	if gbmWpMobileOrg, ok := os.LookupEnv("GBM_WPMOBILE_ORG"); !ok {
		wpMobileOrg = "wordpress-mobile"
	} else {
		wpMobileOrg = gbmWpMobileOrg
	}

	if gbmWordPressOrg, ok := os.LookupEnv("GBM_WORDPRESS_ORG"); !ok {
		wordPressOrg = "WordPress"
	} else {
		wordPressOrg = gbmWordPressOrg
	}

	if gbmAutomatticOrg, ok := os.LookupEnv("GBM_AUTOMATTIC_ORG"); !ok {
		automatticOrg = "Automattic"
	} else {
		automatticOrg = gbmAutomatticOrg
	}
}

func getOrg(repo string) (string, error) {
	switch repo {
	case "gutenberg":
		return wordPressOrg, nil
	case "jetpack":
		return automatticOrg, nil
	case "gutenberg-mobile":
		fallthrough
	case "WordPress-Android":
		fallthrough
	case "WordPress-iOS":
		return wpMobileOrg, nil
	default:
		return "", fmt.Errorf("unknown repo: %s", repo)
	}
}
