package utils

import (
	"errors"
	"fmt"
	"io"
	"os"
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

func SetTempDir() (string, error) {
	tempDir, err := os.MkdirTemp("", "gbm-")
	if err != nil {
		return "", err
	}
	return tempDir, nil
}

func CleanupTempDir(tempDir string) error {
	err := os.RemoveAll(tempDir)
	if err != nil {
		return err
	}
	return nil
}

func UpdatePackageVersion(version, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	packJson, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	update, err := updatePackageJsonVersion(version, packJson)
	if err != nil {
		return err
	}

	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()
	if _, err := w.Write(update); err != nil {
		return err
	}
	return nil
}

func updatePackageJsonVersion(version string, packJson []byte) ([]byte, error) {

	re := regexp.MustCompile(`("version"\s*:\s*)"(?:.*)"`)

	if match := re.Match(packJson); !match {
		return nil, errors.New("cannot find a version in the json file")
	}
	repl := fmt.Sprintf(`$1"%s"`, version)
	return re.ReplaceAll(packJson, []byte(repl)), nil
}
