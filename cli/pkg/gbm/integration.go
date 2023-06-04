package gbm

import (
	"errors"
	"fmt"

	"os"
	"path/filepath"
	"regexp"

	"github.com/wordpress-mobile/gbm-cli/internal/integration"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
)

// Creates an integration PR for the given target
// It will return an ExitingPrError if the branch already exists
// If successful, it will populate the gbmPR with the newly created PR info
func CreateIntegrationPr(target integration.Target, gbmPR *repo.PullRequest) error {
	return integration.CreateIntegrationPr(target, gbmPR)
}

type writerFunc func([]byte, string) ([]byte, error)

// Android Integration Target
type AndroidInPr struct {
	HeadBranch  string
	BaseBranch  string
	RenderTitle func(gbmPR repo.PullRequest) string
	RenderBody  func(gbmPR repo.PullRequest) string
	Version     string
	Tags        []string
}

func (a AndroidInPr) UpdateVersion(dir string, gbmPR repo.PullRequest) error {
	file := filepath.Join(dir, "WordPress-Android", "build.gradle")
	config, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	_, err = updateAndroidVersion(dir, a.GetVersion(gbmPR), config, writeUpdate)
	return err

}

func updateAndroidVersion(version, file string, config []byte, writer writerFunc) ([]byte, error) {
	re := regexp.MustCompile(`(gutenbergMobileVersion\s*=\s*)'(?:.*)'`)

	if match := re.Match(config); !match {
		return nil, errors.New("cannot find a version in the gradle file")
	}

	repl := fmt.Sprintf(`$1'%s'`, version)
	updated := re.ReplaceAll(config, []byte(repl))
	return writer(updated, file)
}

func (a AndroidInPr) GetVersion(gbmPR repo.PullRequest) string {
	if a.Version == "" {
		return fmt.Sprintf("%d-%s", gbmPR.Number, gbmPR.Head.Ref)
	}
	return a.Version
}

func (a AndroidInPr) Title(gbmPR repo.PullRequest) string {
	return a.RenderTitle(gbmPR)
}

func (a AndroidInPr) Body(gbmPR repo.PullRequest) string {
	return a.RenderBody(gbmPR)
}

func (a AndroidInPr) GetTags() []string {
	return a.Tags
}

func (a AndroidInPr) GetRepo() string {
	return "WordPress-Android"
}

func (a AndroidInPr) GetHeadBranch() string {
	if a.HeadBranch == "" {
		return fmt.Sprintf("gutenberg/%s", a.Version)
	}
	return a.HeadBranch
}

func (a AndroidInPr) GetBaseBranch() string {
	return a.BaseBranch
}

// iOS integration Target
type IosInPr struct {
	HeadBranch  string
	BaseBranch  string
	RenderTitle func(gbmPR repo.PullRequest) string
	RenderBody  func(gbmPR repo.PullRequest) string
	Version     string
	Tags        []string
}

func (i IosInPr) UpdateVersion(dir string, gbmPr repo.PullRequest) error {
	file := filepath.Join(dir, "WordPress-iOS", "Gutenberg", "version.rb")
	config, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	_, err = updateIosVersion(i.GetVersion(gbmPr), file, config, writeUpdate)
	return err

}

// Update the version in the Gutenberg version.rb file
// It will result in either a tag or a commit in the version.rb file
// even though a tag can take precedence over a commit.
// dir is assumed to be inside the WordPress-iOS repo
func updateIosVersion(version, file string, config []byte, writer writerFunc) ([]byte, error) {

	// Set up regexs for tag or commit
	tagRe := regexp.MustCompile(`v\d+\.\d+\.\d+`)
	tagLineRe := regexp.MustCompile(`([\r\n]\s*)#*(tag:.*)`)
	commitLineRe := regexp.MustCompile(`([\r\n]\s*)#*(commit:.*)`)

	var (
		updated []byte
	)

	// If matching a version tag, replace the tag line with the new tag
	if tagRe.MatchString(version) {
		updated = commitLineRe.ReplaceAll(config, []byte("${1}#${2}"))
		tagRepl := []byte(fmt.Sprintf(`${1}tag: '%s'`, version))
		updated = tagLineRe.ReplaceAll(updated, tagRepl)
	} else {
		updated = tagLineRe.ReplaceAll(config, []byte("${1}#${2}"))
		commitRepl := []byte(fmt.Sprintf(`${1}commit: '%s'`, version))
		updated = commitLineRe.ReplaceAll(updated, commitRepl)
	}

	return writer(updated, file)
}

func (i IosInPr) GetVersion(gbmPR repo.PullRequest) string {
	if i.Version == "" {
		return gbmPR.Head.Ref
	}
	return i.Version
}

func (i IosInPr) Title(gbmPR repo.PullRequest) string {

	if i.RenderTitle == nil {
		return "Update Gutenberg Mobile"
	}
	return i.RenderTitle(gbmPR)
}

func (i IosInPr) Body(gbmPR repo.PullRequest) string {
	if i.RenderBody == nil {
		return fmt.Sprintf("See %s", gbmPR.Url)
	}
	return i.RenderBody(gbmPR)
}

func (i IosInPr) GetTags() []string {
	return i.Tags
}

func (i IosInPr) GetRepo() string {
	return "WordPress-iOS"
}

func (i IosInPr) GetHeadBranch() string {
	if i.HeadBranch == "" {
		return fmt.Sprintf("gutenberg/%s", i.Version)
	}
	return i.HeadBranch
}

func (i IosInPr) GetBaseBranch() string {
	return i.BaseBranch
}

type ExitingPrError struct {
	Err error
}

func (r *ExitingPrError) Error() string {
	return r.Err.Error()
}

func writeUpdate(update []byte, file string) ([]byte, error) {
	f, err := os.Create(file)

	if err != nil {
		return update, err
	}
	defer f.Close()
	if _, err := f.Write(update); err != nil {
		return update, err
	}

	return update, nil
}
