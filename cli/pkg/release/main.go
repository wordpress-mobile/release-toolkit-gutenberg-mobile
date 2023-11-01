package release

import (
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/semver"
)

type Build struct {
	Version semver.SemVer
	Dir     string
	UseTag  bool
	Prs     []gh.PullRequest
	Base    gh.Repo
}

type ReleaseChanges struct {
	Title  string
	Number int
	PrUrl  string
	Issues []string
}
