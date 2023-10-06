package release

import (
	"fmt"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/git"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
)

func CreateGbPR(version, dir string) (gh.PullRequest, error) {
	var pr gh.PullRequest

	branch := fmt.Sprintf("rnmobile/release_%s", version)

	console.Info("Checking if branch %s exists", branch)
	exists, _ := gh.SearchBranch("gutenberg", branch)

	if (exists != gh.Branch{}) {
		console.Info("Branch %s already exists", branch)
		return pr, nil
	} else {

		console.Info("Cloning Gutenberg to %s", dir)
		err := git.Clone(repo.GetRepoPath("gutenberg"), dir, true)
		if err != nil {
			return pr, err
		}
	}

	return pr, fmt.Errorf("not implemented")
}
