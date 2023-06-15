package repo

import (
	"fmt"
	"os"
	"time"

	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
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

func Auth() *http.BasicAuth {
	// load host and auth from 'gh'
	host, _ := auth.DefaultHost()
	token, _ := auth.TokenForHost(host)
	user := Signature()
	return &http.BasicAuth{
		Username: user.Name, // this can be anything since we are using a token
		Password: token,
	}
}

func Signature() *object.Signature {
	config, _ := config.LoadConfig(config.GlobalScope)
	u := config.User
	s := object.Signature{
		Name:  u.Name,
		Email: u.Email,
		When:  time.Now(),
	}
	return &s
}
