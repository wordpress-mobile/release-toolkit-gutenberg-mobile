package integrate

import (
	"os"
	"path/filepath"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/shell"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

func IosIntegration(ri ReleaseIntegration) ReleaseIntegration {
	ri.Type = iosIntegration{}
	return ri
}

func updateIos(dir string, ri ReleaseIntegration) error {

	sp := shell.CmdProps{Dir: dir, Verbose: true}
	git := shell.NewGitCmd(sp)
	// TODO update github org although not sure it's useful here
	console.Info("Update gutenberg-mobile ref in Gutenberg/config.yml")

	configPath := filepath.Join(dir, "Gutenberg/config.yml")
	buf, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// perform updates using the yq syntax
	updates := []string{".ref.commit = \"v" + ri.Version + "\"", "del(.ref.tag)"}
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

	return git.CommitAll("Release script: Update gutenberg-mobile ref", ri.Version)
}
