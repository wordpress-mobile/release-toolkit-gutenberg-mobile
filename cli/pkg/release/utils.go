package release

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
)

func logger() func(string, ...interface{}) {
	return func(f string, a ...interface{}) {
		utils.LogInfo(f, a...)
	}
}

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

// Change Log and Release notes update functions

// Updates the release notes by replacing "Unreleased" with
// the new version and adding a new "Unreleased" section
func UpdateReleaseNotes(version, path string) error {
	return readWriteNotes(version, path, releaseNotesUpdater)
}

// Check to see if the Gutenberg submodule in GBM is pointing to the
// head of the Gutenberg release PR
func IsGbmPrCurrent(version string) bool {
	gbPr, err := repo.GetGbReleasePr(version)
	if err != nil {
		utils.LogWarn("Error getting Gutenberg Pr: %s", err)
		return false
	}
	pr, err := repo.GetGbmReleasePr(version)
	if err != nil {
		utils.LogWarn("Error getting GB Mobile Pr: %s", err)
		return false
	}
	submoduleStat, err := repo.GetContents("gutenberg-mobile", "gutenberg", pr.Head.Ref)
	if err != nil {
		utils.LogWarn("Error getting submodule status: %s", err)
		return false
	}

	return gbPr.Head.Sha == submoduleStat.Sha
}

// See UpdateReleaseNotes
// This handles the string replacement
func releaseNotesUpdater(version string, notes []byte) []byte {
	re := regexp.MustCompile(`(^Unreleased\s*\n)`)

	repl := fmt.Sprintf("$1---\n\n%s\n", version)

	return re.ReplaceAll(notes, []byte(repl))
}

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

type ReleaseChanges struct {
	Title  string
	Number int
	PrUrl  string
	Issues []string
}

func CollectReleaseChanges(version string, changelog, relnotes []byte) ([]ReleaseChanges, error) {
	changesRe := regexp.MustCompile(`(?s)\d+\.\d+\.\d+(.*?)\d+\.\d+\.\d+`)
	// rnChangesRe := regexp.MustCompile(`(?s)\d+\.\d+\.\d+(.*?)\d+\.\d+\.\d+`)
	prNumbRe := regexp.MustCompile(`\[#(\d+)\]`)
	prOrgRepoNumRe := regexp.MustCompile(`https://github\.com/(\w+)/(\w+)/pull/(\d+)`)
	bracketRe := regexp.MustCompile(`\[.*\]\s*-*`)

	prs := []ReleaseChanges{}

	prFoundAlready := func(num int) bool {
		for _, p := range prs {
			if p.Number == num {
				return true
			}
		}
		return false
	}

	// Get the changes from the Release notes
	match := changesRe.Find(relnotes)
	if match != nil {

		lines := strings.Split(string(match), "\n")

		for _, l := range lines {
			// first check for any prs relative to gutenberg-mobile
			matches := prNumbRe.FindAllStringSubmatch(l, -1)

			if len(matches) == 1 {

				prId, _ := strconv.Atoi(matches[0][1])

				pr, err := repo.GetPrOrg("WordPress", "gutenberg", prId)
				if err != nil {
					utils.LogWarn("There was an issue fetching a gutenberg pr #%d", prId)
					continue
				}
				// Scrub [] from title
				title := bracketRe.ReplaceAllString(pr.Title, "")
				prs = append(prs, ReleaseChanges{Title: title, PrUrl: pr.Url, Number: pr.Number})
			}

			// now look for urls
			matches = prOrgRepoNumRe.FindAllStringSubmatch(l, -1)
			if len(matches) != 0 {
				match := matches[0]
				org := match[1]
				rep := match[2]
				num := match[3]
				prId, _ := strconv.Atoi(num)
				pr, err := repo.GetPrOrg(org, rep, prId)
				if err != nil {
					utils.LogWarn("There was an issue fetching %s/%s/pull/%d", org, rep, prId)
					continue
				}
				// Scrub [] from title
				title := bracketRe.ReplaceAllString(pr.Title, "")
				rc := ReleaseChanges{
					Title:  title,
					PrUrl:  pr.Url,
					Number: pr.Number,
				}
				checkPRforIssues(pr, &rc)
				prs = append(prs, rc)
			}
		}
	}
	// Get changes from the Change log
	match = changesRe.Find(changelog)

	if match != nil {
		lines := strings.Split(string(match), "\n")

		for _, l := range lines {

			match := prNumbRe.FindAllStringSubmatch(l, -1)

			if len(match) == 1 {
				prId, _ := strconv.Atoi(match[0][1])
				if prFoundAlready(prId) {
					continue
				}
				pr, err := repo.GetPrOrg("WordPress", "gutenberg", prId)
				if err != nil {
					utils.LogWarn("There was an issue fetching a gutenberg pr #%d", prId)
					continue
				}

				title := bracketRe.ReplaceAllString(pr.Title, "")
				rc := ReleaseChanges{
					Title:  title,
					PrUrl:  pr.Url,
					Number: pr.Number,
				}
				checkPRforIssues(pr, &rc)
				prs = append(prs, rc)

			}
		}

	}

	return prs, nil
}

func checkPRforIssues(pr repo.PullRequest, rc *ReleaseChanges) {
	issueRe := regexp.MustCompile(`(https:\/\/github.com\/.*\/.*\/issues\/\d*)`)

	matches := issueRe.FindAllStringSubmatch(pr.Body, -1)

	for _, m := range matches {
		rc.Issues = append(rc.Issues, m[1])
	}
}
