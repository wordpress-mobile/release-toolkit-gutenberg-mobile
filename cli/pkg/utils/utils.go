package utils

import (
	"fmt"
	"regexp"
	"time"
)

func ValidateVersion(version string) bool {
	re := regexp.MustCompile(`v*(\d+)\.(\d+)\.(\d+)$`)
	return re.MatchString(version)
}

func IsScheduledRelease(version string) bool {
	re := regexp.MustCompile(`^v*(\d+)\.(\d+)\.0$`)
	return re.MatchString(version)
}

func NextReleaseDate() string {
	weekday := time.Now().Weekday()
	daysUntilThursday := 4 - weekday

	nextThursday := time.Now().AddDate(0, 0, int(daysUntilThursday))

	return nextThursday.Format("Monday January 2, 2006")
}

func NormalizeVersion(version string) (string, error) {
	v := version
	if version[0] == 'v' {
		v = version[1:]
	}

	re := regexp.MustCompile(`\d+\.\d+\.\d+`)
	if !re.MatchString(v) {
		return "", fmt.Errorf("invalid version")
	}
	return v, nil
}
