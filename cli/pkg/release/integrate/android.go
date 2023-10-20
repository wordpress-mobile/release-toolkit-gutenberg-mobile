package integrate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/gbm-cli/pkg/gh"
	"github.com/wordpress-mobile/gbm-cli/pkg/shell"
)

type AndroidIntegration struct {
}

func (ai AndroidIntegration) UpdateGutenbergConfig(dir string, gbmPr gh.PullRequest) error {
	sp := shell.CmdProps{Dir: dir, Verbose: true}
	git := shell.NewGitCmd(sp)
	prId := gbmPr.Number
	prSha := gbmPr.Head.Sha

	configPath := filepath.Join(dir, "build.gradle")
	config, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`(gutenbergMobileVersion\s*=\s*)'(?:.*)'`)
	if match := re.Match(config); !match {
		return errors.New("cannot find a version in the gradle file")
	}

	repl := fmt.Sprintf(`$1'%s-%s'`, fmt.Sprint(prId), prSha)
	config = re.ReplaceAll(config, []byte(repl))

	if err := os.WriteFile(configPath, config, 0644); err != nil {
		return err
	}
	return git.CommitAll("Release script: Update build.gradle gutenbergMobileVersion to ref")
}

func (ia AndroidIntegration) GetRepo() string {
	return "foobar"
	// return repo.WordPressAndroidRepo
}

func (ia AndroidIntegration) GetPr(version string) (gh.PullRequest, error) {
	return gbm.FindAndroidReleasePr(version)
}
