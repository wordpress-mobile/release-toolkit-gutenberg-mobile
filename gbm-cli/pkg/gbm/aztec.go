package gbm

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type aztecResult struct {
	err      error
	valid    bool
	platform string
}

func ValidateAztecVersions() bool {

	// Since the platform validation functions might reach out to a remote repo,
	// let's use a channel to hold the results of the goroutines
	res := make(chan aztecResult)

	bothValid := true

	go func() {
		res <- ValidateAndroidAztecVersion()
	}()

	go func() {
		res <- ValidateIosAztecVersion()
	}()

	// Wait for each result...
	for i := 0; i < 2; i++ {
		r := <-res
		if r.err != nil {
			fmt.Printf("Error validating %s aztec version: %s\n", r.platform, r.err)
			bothValid = false
		}

		// If the first returned value is false we just need to wait for the second and
		// carry on without updating `bothValid`.
		if bothValid && !r.valid {
			bothValid = false
		}
	}

	return bothValid
}

func ValidateAndroidAztecVersion() aztecResult {
	branch := "trunk"
	org := "WordPress"
	path := fmt.Sprintf("https://raw.githubusercontent.com/%s/gutenberg/%s/packages/react-native-aztec/android/build.gradle", org, branch)

	config, err := getConfig(path)

	if err != nil {
		return aztecResult{err: err, valid: false, platform: "android"}
	}

	regex := regexp.MustCompile(`(?m)^.*aztecVersion.*$`)
	valid, err := verifyVersion(config, regex)

	return aztecResult{err: err, valid: valid, platform: "android"}
}

func ValidateIosAztecVersion() aztecResult {

	branch := "trunk"
	org := "wordpress-mobile"
	path := fmt.Sprintf("https://raw.githubusercontent.com/%s/gutenberg/%s/packages/react-native-aztec/RNTAztecView.podspec", org, branch)

	config, err := getConfig(path)

	if err != nil {
		return aztecResult{err: err, valid: false, platform: "ios"}
	}

	regex := regexp.MustCompile(`(?m)^.*WordPress-Aztec-iOS.*$`)
	valid, err := verifyVersion(config, regex)

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

	resp, err := http.Get(path)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	file = resp.Body

	if file == nil {
		return nil, fmt.Errorf("file not found")
	}

	return io.ReadAll(file)
}
