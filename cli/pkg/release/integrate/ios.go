package integrate

import (
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
	// TODO update github org although not sure it's useful here
	console.Info("Update gutenberg-mobile ref in Gutenberg/config.yml")

	configPath := filepath.Join(dir, "Gutenberg/config.yml")
	buf, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	version := gbmPr.ReleaseVersion
	// perform updates using the yq syntax
	updates := []string{".ref.commit = \"v" + version + "\"", "del(.ref.tag)"}
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

	return git.CommitAll("Release script: Update gutenberg-mobile ref", version)
}

func (ii IosIntegration) GetRepo() string {
	return repo.WordPressIosRepo
}

func (ia IosIntegration) GetPr(version string) (gh.PullRequest, error) {
	return gbm.FindIosReleasePr(version)
}

func (ia IosIntegration) GbPublished(version string) (bool, error) {
	published, err := gbm.IosGbmBuildPublished(version)
	if err != nil {
		console.Warn("Error checking if GBM build is published: %v", err)
	}
	return published, nil
}
