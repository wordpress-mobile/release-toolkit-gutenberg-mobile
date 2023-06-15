package release

import (
	"fmt"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/wordpress-mobile/gbm-cli/internal/gh"
)

func TestUpdateIosVersion(t *testing.T) {

	t.Run("Updates a tag to a commit", func(t *testing.T) {

		updateIosVersion := buildUpdateIosVersion("123")

		config := []byte(`
# frozen_string_literal: true
GUTENBERG_CONFIG = {
  #commit: '',
  tag: 'v1.96.0'
}

GITHUB_ORG = 'wordpress-mobile'
REPO_NAME = 'gutenberg-mobile'
`)

		want := []byte(`
# frozen_string_literal: true
GUTENBERG_CONFIG = {
  commit: '123'
  #tag: 'v1.96.0'
}

GITHUB_ORG = 'wordpress-mobile'
REPO_NAME = 'gutenberg-mobile'
`)

		got, err := updateIosVersion(config, gh.PullRequest{})
		assertNoError(t, err)
		assertNoDiff(t, want, got)

	})

	t.Run("Updates a commit to a tag", func(t *testing.T) {

		updateIosVersion := buildUpdateIosVersion("v1.0.0")
		config := []byte(`
# frozen_string_literal: true
GUTENBERG_CONFIG = {
  commit: '123'
  #tag: 'v1.96.0'
}

GITHUB_ORG = 'wordpress-mobile'
REPO_NAME = 'gutenberg-mobile'
`)

		want := []byte(`
# frozen_string_literal: true
GUTENBERG_CONFIG = {
  # commit: '',
  tag: 'v1.0.0'
}

GITHUB_ORG = 'wordpress-mobile'
REPO_NAME = 'gutenberg-mobile'
`)

		got, err := updateIosVersion(config, gh.PullRequest{})
		assertNoError(t, err)
		assertNoDiff(t, want, got)
	})

	t.Run("Updates a commit to a commit", func(t *testing.T) {
		updateIosVersion := buildUpdateIosVersion("456")
		config := []byte(`
# frozen_string_literal: true
GUTENBERG_CONFIG = {
  commit: '123'
  #tag: 'v1.96.0'
}

GITHUB_ORG = 'wordpress-mobile'
REPO_NAME = 'gutenberg-mobile'
`)

		want := []byte(`
# frozen_string_literal: true
GUTENBERG_CONFIG = {
  commit: '456'
  #tag: 'v1.96.0'
}

GITHUB_ORG = 'wordpress-mobile'
REPO_NAME = 'gutenberg-mobile'
`)

		got, err := updateIosVersion(config, gh.PullRequest{})
		assertNoError(t, err)
		assertNoDiff(t, want, got)
	})

	t.Run("Updates a tag to an alpha tag", func(t *testing.T) {
		updateIosVersion := buildUpdateIosVersion("v1.97.0-alpha")
		config := []byte(`
# frozen_string_literal: true
GUTENBERG_CONFIG = {
  # commit: '',
  tag: 'v1.96.0'
}

GITHUB_ORG = 'wordpress-mobile'
REPO_NAME = 'gutenberg-mobile'
`)

		want := []byte(`
# frozen_string_literal: true
GUTENBERG_CONFIG = {
  # commit: '',
  tag: 'v1.97.0-alpha'
}

GITHUB_ORG = 'wordpress-mobile'
REPO_NAME = 'gutenberg-mobile'
`)

		got, err := updateIosVersion(config, gh.PullRequest{})
		assertNoError(t, err)
		assertNoDiff(t, want, got)
	})

	t.Run("Updates an alpha tag to a commit", func(t *testing.T) {
		updateIosVersion := buildUpdateIosVersion("123")
		config := []byte(`
# frozen_string_literal: true
GUTENBERG_CONFIG = {
  # commit: '',
  tag: 'v1.97.0-alpha1'
}

GITHUB_ORG = 'wordpress-mobile'
REPO_NAME = 'gutenberg-mobile'
`)

		want := []byte(`
# frozen_string_literal: true
GUTENBERG_CONFIG = {
  commit: '123'
  #tag: 'v1.97.0-alpha1'
}

GITHUB_ORG = 'wordpress-mobile'
REPO_NAME = 'gutenberg-mobile'
`)

		got, err := updateIosVersion(config, gh.PullRequest{})
		assertNoError(t, err)
		fmt.Println(string(got))
		assertNoDiff(t, want, got)
	})

}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func assertNoDiff(t testing.TB, got, want []byte) {
	t.Helper()
	if g, w := string(got), string(want); g != w {
		t.Errorf("Result not as expected:\n%v", diff.LineDiff(g, w))
	}
}
