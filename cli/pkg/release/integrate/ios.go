package integrate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
	"github.com/wordpress-mobile/gbm-cli/pkg/shell"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

type IosIntegration struct {
	Repo string
}

func (ii IosIntegration) UpdateGutenbergConfig(dir string, gbmPr gh.PullRequest) error {
	sp := shell.CmdProps{Dir: dir, Verbose: true}
	git := shell.NewGitCmd(sp)

	// @TODO update github org although not sure it's useful here

	var updates []string

	// Check to see if there is a published release for the version
	if releaseAvailable, err := useRelease(gbmPr.ReleaseVersion); err != nil {
		return fmt.Errorf("unable to check for a release: %s", err)
	} else if releaseAvailable {
		console.Info("Updating gutenberg-mobile ref to the tag v%s", gbmPr.ReleaseVersion)
		updates = []string{".ref.tag = \"v" + gbmPr.ReleaseVersion + "\"", "del(.ref.commit)"}
	} else {
		console.Info("Updating gutenberg-mobile ref to the commit %s", gbmPr.Head.Sha)
		updates = []string{".ref.commit = \"v" + gbmPr.Head.Sha + "\"", "del(.ref.tag)"}
	}

	configPath := filepath.Join(dir, "Gutenberg/config.yml")
	buf, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// perform updates using the yq syntax
	config, err := utils.YqEvalAll(updates, string(buf))
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return err
	}
	sp = shell.CmdProps{Dir: dir, Verbose: true}
	bundle := shell.NewBundlerCmd(sp)

	console.Info("Running bundle install")
	if err := bundle.Install(); err != nil {
		return err
	}

	console.Info("Running rake dependencies")
	rake := shell.NewRakeCmd(sp)
	if err := rake.Dependencies(); err != nil {
		return err
	}

	return git.CommitAll("Release script: Update gutenberg-mobile ref %s", gbmPr.ReleaseVersion)
}

func (ii IosIntegration) GetRepo() string {
	return repo.WordPressIosRepo
}

func (ia IosIntegration) GetPr(ri ReleaseIntegration) (gh.PullRequest, error) {
	// @TODO: add support for finding non release PRs
	if ri.Version != "" {
		return gbm.FindIosReleasePr(ri.Version)
	}
	return gh.PullRequest{}, nil
}

func (ia IosIntegration) GbPublished(version string) (bool, error) {
	published, err := gbm.IosGbmBuildPublished(version)
	if err != nil {
		console.Warn("Error checking if GBM build is published: %v", err)
	}
	return published, nil
}
