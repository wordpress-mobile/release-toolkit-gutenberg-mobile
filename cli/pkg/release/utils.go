package release

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wordpress-mobile/gbm-cli/internal/repo"
)

type AztecSrc struct {
	// This is a local path to the Gutenberg Mobile repo. If this is set, the validator will use this path
	GbmDir string

	// This is the branch to use for the validation. If this is set, the validator will use this branch
	// If both GbmDir and Branch are empty, the validator will use Gutenberg and Gutenberg Mobile trunk branches
	Branch string
}

type aztecResult struct {
	err      error
	valid    bool
	platform string
}

func ValidateVersion(version string) (bool, error) {
	return true, nil
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

func UpdateChangeNotes(version, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	changeNotes, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	update := updateChangeNotes(version, changeNotes)
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

func updateChangeNotes(version string, chNotes []byte) []byte {

	re := regexp.MustCompile(`(##\s*Unreleased\s*\n)`)

	repl := fmt.Sprintf("$1##\n %s\n", version)

	return re.ReplaceAll(chNotes, []byte(repl))
}

// This validates the version of Aztec to make sure it's using a stable release and
// not a development version.
// Use the AztecSrc struct to point the validator to the correct location.
func ValidateAztecVersions(az AztecSrc) (bool, error) {
	res := make(chan aztecResult)

	bothValid := true
	go func() {
		res <- ValidateAndroidAztecVersion(az)
	}()

	go func() {
		res <- ValidateIosAztecVersion(az)
	}()

	for i := 0; i < 2; i++ {
		r := <-res
		if r.err != nil {
			fmt.Printf("Error validating %s aztec version: %s\n", r.platform, r.err)
			bothValid = false
		}
	}

	return bothValid, nil
}

func ValidateAndroidAztecVersion(az AztecSrc) aztecResult {
	var path string

	// check if we have a local file
	if az.GbmDir != "" {
		path = filepath.Join(az.GbmDir, "gutenberg", "packages", "react-native-aztec", "android", "build.gradle")
	} else {
		// Check which branch we should check
		branch := "trunk"
		if az.Branch != "" {
			branch = az.Branch
		}
		org, _ := repo.GetOrg("gutenberg")
		path = fmt.Sprintf("https://raw.githubusercontent.com/%s/gutenberg/%s/packages/react-native-aztec/android/build.gradle", org, branch)
	}

	config, err := getConfig(path)
	if err != nil {
		return aztecResult{err: err, valid: false, platform: "android"}
	}
	valid, err := verifyVersion(config, getLineRegexp().android)
	return aztecResult{err: err, valid: valid, platform: "android"}
}

func ValidateIosAztecVersion(az AztecSrc) aztecResult {
	var path string

	if az.GbmDir != "" {
		path = filepath.Join(az.GbmDir, "RNTAztecView.podspec")
	} else {
		// Check which branch we should check
		branch := "trunk"
		if az.Branch != "" {
			branch = az.Branch
		}
		org, _ := repo.GetOrg("gutenberg-mobile")
		path = fmt.Sprintf("https://raw.githubusercontent.com/%s/gutenberg-mobile/%s/RNTAztecView.podspec", org, branch)
	}
	config, err := getConfig(path)
	if err != nil {
		return aztecResult{err: err, valid: false, platform: "ios"}
	}
	valid, err := verifyVersion(config, getLineRegexp().ios)
	return aztecResult{err: err, valid: valid, platform: "ios"}
}

func verifyVersion(config []byte, lnRe *regexp.Regexp) (bool, error) {
	versionLine := lnRe.FindAll(config, 1)
	if versionLine == nil {
		return false, fmt.Errorf("no version line found")
	}
	semRe := regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)
	return semRe.Match(versionLine[0]), nil
}

func getConfig(path string) ([]byte, error) {
	var file io.Reader

	// Check for a local file
	if fileInfo, _ := os.Stat(path); fileInfo != nil {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		file = f
	}

	// Check for a a remote file
	if strings.Contains(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		file = resp.Body
	}

	if file == nil {
		return nil, fmt.Errorf("file not found")
	}

	return io.ReadAll(file)
}

func getLineRegexp() struct {
	ios     *regexp.Regexp
	android *regexp.Regexp
} {
	return struct {
		ios     *regexp.Regexp
		android *regexp.Regexp
	}{
		ios:     regexp.MustCompile(`(?m)^.*WordPress-Aztec-iOS.*$`),
		android: regexp.MustCompile(`(?m)^.*aztecVersion.*$`),
	}
}
