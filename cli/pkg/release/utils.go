package release

import (
	"path/filepath"

	"github.com/wordpress-mobile/gbm-cli/pkg/shell"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

func updatePackageJson(dir, version string, pkgs ...string) error {
	sp := shell.CmdProps{Dir: dir, Verbose: true}
	git := shell.NewGitCmd(sp)

	for _, pkg := range pkgs {
		if err := utils.UpdatePackageVersion(version, filepath.Join(dir, pkg)); err != nil {
			return err
		}
	}
	if err := git.CommitAll("Release script: Update package.json versions to %s", version); err != nil {
		return err
	}

	return nil
}
