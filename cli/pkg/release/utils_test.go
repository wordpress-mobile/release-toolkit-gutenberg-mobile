package release

import (
	"io"
	"os"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

func TestGetConfig(t *testing.T) {

	t.Run("It returns a slice from a local file", func(t *testing.T) {

		testdata := readTestdata(t, "testdata/aztec/build.gradle")

		b, err := getConfig("testdata/aztec/build.gradle")
		assertNoError(t, err)

		assertNoDiff(t, testdata, b)
	})

	t.Run("It returns a slice from a remote file", func(t *testing.T) {
		testdata := readTestdata(t, "testdata/aztec/build.gradle")

		gock.New("https://example.com").
			Get("aztec/build.gradle").
			Reply(200).
			BodyString(string(testdata))
		defer gock.Off()

		b, err := getConfig("https://example.com/aztec/aztec/build.gradle")

		assertNoError(t, err)
		assertNoDiff(t, testdata, b)
	})

	t.Run("It returns an error if the file is not found", func(t *testing.T) {
		_, err := getConfig("testdata/aztec/does-not-exist.gradle")
		assertError(t, err)
	})
}

func TestVerifyVersion(t *testing.T) {

	t.Run("It can verify a valid Android version", func(t *testing.T) {
		testdata := readTestdata(t, "testdata/aztec/build.gradle")

		valid, err := verifyVersion(testdata, getLineRegexp().android)
		assertNoError(t, err)
		if !valid {
			t.Fatal("Expected android version to be valid")
		}
	})

	t.Run("It can verify a valid iOS version", func(t *testing.T) {
		testdata := readTestdata(t, "testdata/aztec/RNTAztecView.podspec")
		valid, err := verifyVersion(testdata, getLineRegexp().ios)
		assertNoError(t, err)
		if !valid {
			t.Fatal("Expected ios version to be valid")
		}
	})

	t.Run("It can verify an invalid Android version", func(t *testing.T) {
		testdata := readTestdata(t, "testdata/aztec/build.invalid.gradle")

		valid, err := verifyVersion(testdata, getLineRegexp().android)
		assertNoError(t, err)
		if valid {
			t.Fatal("Expected android version to be invalid")
		}
	})

	t.Run("It can verify an invalid iOS version", func(t *testing.T) {
		testdata := readTestdata(t, "testdata/aztec/RNTAztecView.invalid.podspec")

		valid, err := verifyVersion(testdata, getLineRegexp().ios)
		assertNoError(t, err)
		if valid {
			t.Fatal("Expected android version to be invalid")
		}
	})
}

func readTestdata(t *testing.T, path string) []byte {
	t.Helper()
	f, err := os.Open(path)
	assertNoError(t, err)
	defer f.Close()

	testfile, err := io.ReadAll(f)
	assertNoError(t, err)
	return testfile
}
