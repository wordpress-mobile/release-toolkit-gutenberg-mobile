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

func TestUpdateNotes(t *testing.T) {

	t.Run("It updates changelogs", func(t *testing.T) {

		want :=
			`<!-- Learn how to maintain this file at https://github.com/WordPress/gutenberg/tree/HEAD/packages#maintaining-changelogs. -->

<!--
For each user feature we should also add a importance categorization label  to indicate the relevance of the change for end users of GB Mobile. The format is the following:
[***] → Major new features, significant updates to core flows, or impactful fixes (e.g. a crash that impacts a lot of users) — things our users should be aware of.

[**] → Changes our users will probably notice, but doesn’t impact core flows. Most fixes.

[*] → Minor enhancements and fixes that address annoyances — things our users can miss.
-->

## Unreleased

## 1.97.0
-   [*] [internal] Upgrade compile and target sdk version to Android API 33 [#50731]

## 1.96.0
-   [**] Tapping on all nested blocks gets focus directly instead of having to tap multiple times depending on the nesting levels. [#50672]
-   [**] Fix undo/redo history when inserting a link configured to open in a new tab [#50460]
-   [*] [List block] Fix an issue when merging a list item into a Paragraph would remove its nested list items. [#50701]

## 1.95.0
-   [**] Fix Android-only issue related to block toolbar not being displayed on some blocks in UBE [#51131]`

		td := readTestdata(t, "testdata/CHANGELOG.md")

		got := changeLogUpdater("1.97.0", td)

		assertNoDiff(t, []byte(want), got)
	})

	t.Run("It updates release notes", func(t *testing.T) {

		want :=
			`Unreleased
---

1.97.0
---
* [**] [iOS] Fix dictation regression, in which typing/dictating at the same time caused content loss. [https://github.com/WordPress/gutenberg/pull/49452]
* [*] [internal] Upgrade compile and target sdk version to Android API 33 [https://github.com/wordpress-mobile/gutenberg-mobile/pull/5789]
* [*] Show "No title"/"No description" placeholder for not belonged videos in VideoPress block [https://github.com/wordpress-mobile/gutenberg-mobile/pull/5840]

1.96.1
---
* [**] Fix Android-only issue related to block toolbar not being displayed on some blocks in UBE [https://github.com/WordPress/gutenberg/pull/51131]

1.96.0
---
* [*] Add disabled style to 'Cell' component [https://github.com/WordPress/gutenberg/pull/50665]`

		td := readTestdata(t, "testdata/RELEASE-NOTES.txt")

		got := releaseNotesUpdater("1.97.0", td)

		assertNoDiff(t, []byte(want), got)
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
