package release

import (
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/semver"
)

type Build struct {
	Version semver.SemVer
	Dir     string
	UseTag  bool
	Repo    string
	Prs     []gh.PullRequest
	Shas    []string
	Base    gh.Repo
	Depth   string
}

type ReleaseChanges struct {
	Title  string
	Number int
	PrUrl  string
	Issues []string
}
