package utils

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	l *log.Logger
)

func init() {
	l = log.New(os.Stderr, "", 0)
}

func LogInfo(format string, args ...interface{}) {
	l.Printf(InfoString(format, args...))
	color.Unset()
}

func LogDebug(format string, args ...interface{}) {
	l.Printf(DebugString(format, args...))
	color.Unset()
}

func LogWarn(format string, args ...interface{}) {
	l.Printf(WarnString(format, args...))
	color.Unset()
}

func LogError(format string, args ...interface{}) {
	l.Printf(ErrorString(format, args...))
	color.Unset()
}

func InfoString(format string, args ...interface{}) string {
	c := color.New(color.FgCyan, color.Bold).SprintfFunc()
	return c(format, args...)
}

func WarnString(format string, args ...interface{}) string {
	c := color.New(color.FgYellow, color.Bold).SprintfFunc()
	return c(format, args...)
}

func DebugString(format string, args ...interface{}) string {
	c := color.New(color.FgGreen, color.Bold).SprintfFunc()
	return c("[DEBUG] "+format, args...)
}

func ErrorString(format string, args ...interface{}) string {
	c := color.New(color.FgRed, color.Bold).SprintfFunc()
	return c("[ERROR]"+format, args...)
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
