package gbm

import (
	"regexp"
	"time"
)

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
