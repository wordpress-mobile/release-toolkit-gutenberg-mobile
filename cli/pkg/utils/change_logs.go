package utils

import (
	"fmt"
	"io"
	"os"
	"regexp"
)

// Updates the change log by replacing "Unreleased" with
// the new version and adding a new "Unreleased" section
func UpdateChangeLog(version, path string) error {
	return readWriteNotes(version, path, changeLogUpdater)
}

// See UpdateChangeLog
// This handles the string replacement
func changeLogUpdater(version string, notes []byte) []byte {

	re := regexp.MustCompile(`(##\s*Unreleased\s*\n)`)

	repl := fmt.Sprintf("$1\n## %s\n", version)

	return re.ReplaceAll(notes, []byte(repl))
}

func readWriteNotes(version, path string, updater func(string, []byte) []byte) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	changeNotes, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	update := updater(version, changeNotes)
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
