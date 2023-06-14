package repo

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
)

var (
	WpMobileOrg   string
	WordPressOrg  string
	AutomatticOrg string
)

func init() {
	initOrgs()
}

func initOrgs() {
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

func PreviewPr(repo, dir string, pr *PullRequest) {
	org, _ := GetOrg(repo)
	boldUnder := color.New(color.Bold, color.Underline).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Println(boldUnder("\nPr Preview"))
	fmt.Println(bold("Local:"), "\t", cyan(dir))
	fmt.Println(bold("Repo:"), "\t", cyan(fmt.Sprintf("%s/%s", org, repo)))
	fmt.Println(bold("Title:"), "\t", cyan(pr.Title))
	fmt.Print(bold("Body:\n"), cyan(pr.Body))
	fmt.Println(bold("Commits:"))

	git := execGit(dir, true)

	git("log", pr.Base.Ref+"...HEAD", "--oneline", "--no-merges", "-10")
}

// Use this to drop down to `git` when go-git is not playing well.
func execGit(dir string, verbose bool) func(...string) error {
	return func(cmds ...string) error {
		cmd := exec.Command("git", cmds...)
		cmd.Dir = dir

		if verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		return cmd.Run()
	}
}

func l(f string, a ...interface{}) {
	utils.LogInfo(f, a...)
}
