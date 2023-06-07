package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
)

var (
	l *log.Logger
)

func init() {
	l = log.New(os.Stderr, "", 0)
}

func LogInfo(format string, args ...interface{}) {
	c := color.New(color.FgCyan, color.Bold)
	l.Printf(c.Sprintf(format, args...))
	// color can leave the cursor in a weird place if not calling Unset()
	color.Unset()
}

func LogDebug(format string, args ...interface{}) {
	c := color.New(color.FgGreen, color.Bold)
	l.Printf(c.Sprintf(fmt.Sprint("DEBUG ", format), args...))
	color.Unset()
}

// TODO: Allow calling with just the error  (LogError(err))
func LogError(format string, args ...interface{}) {
	c := color.New(color.FgRed, color.Bold)
	l.Printf(c.Sprintf(fmt.Sprint("ERROR ", format), args...))
	color.Unset()
}

func LogWarn(format string, args ...interface{}) {
	c := color.New(color.FgYellow, color.Bold)
	l.Printf(c.Sprintf(fmt.Sprint("WARN ", format), args...))
	color.Unset()
}

func IsScheduledRelease(version string) bool {
	re := regexp.MustCompile(`^v*(\d+)\.(\d+)\.0$`)
	return re.MatchString(version)
}

func NextReleaseDate() string {
	weekday := time.Now().Weekday()
	daysUntilThursday := 4 - weekday

	nextThursday := time.Now().AddDate(0, 0, int(daysUntilThursday))

	return nextThursday.Format("Monday 01, 2006")
}

func GetGbmReleasePr(version string) (repo.PullRequest, error) {
	filter := repo.BuildRepoFilter("gutenberg-mobile", "is:pr", fmt.Sprintf("%s in:title", version))

	res, err := repo.SearchPrs(filter)
	if err != nil {
		return repo.PullRequest{}, nil
	}

	if res.TotalCount == 0 {
		return repo.PullRequest{}, fmt.Errorf("no release PRs found for `%s`", version)
	}
	if res.TotalCount != 1 {
		return repo.PullRequest{}, fmt.Errorf("found multiple prs for %s", version)
	}
	return res.Items[0], nil
}

func Confirm(ask string) bool {
	reader := bufio.NewReader(os.Stdin)
	cyan := color.New(color.FgCyan, color.Bold).PrintfFunc()

	for {
		cyan("%s [y/n]: ", ask)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
