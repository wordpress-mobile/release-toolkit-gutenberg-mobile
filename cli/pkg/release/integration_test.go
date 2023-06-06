package release

import (
	"embed"
	"fmt"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

//go:embed testdata/*
var templatesFS embed.FS

func init() {
	render.TemplateFS = templatesFS
}

func TestGenBodyRenderer(t *testing.T) {

	t.Run("Generates a function to render a PR body", func(t *testing.T) {
		version := "1.2.3"
		renderBody := buildBodyRenderer(version, "testdata/integrationPrBody.md")

		gbmPr := repo.PullRequest{
			Url: "https://wordpress.com",
		}

		want := `## Description
This PR incorporates the 1.2.3 release of gutenberg-mobile.
For more information about this release and testing instructions, please see the related Gutenberg-Mobile PR: https://wordpress.com
`
		got := renderBody(gbmPr)

		fmt.Println(got)

		if got != want {
			t.Fatalf("strings are not equal: %s", diff.LineDiff(got, want))
		}
	})
}

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

		got, err := updateIosVersion(config, repo.PullRequest{})
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
  #commit: '123'
  tag: 'v1.0.0'
}

GITHUB_ORG = 'wordpress-mobile'
REPO_NAME = 'gutenberg-mobile'
`)

		got, err := updateIosVersion(config, repo.PullRequest{})
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

		got, err := updateIosVersion(config, repo.PullRequest{})
		assertNoError(t, err)
		assertNoDiff(t, want, got)
	})

	t.Run("Updates a tag to a tag", func(t *testing.T) {
		updateIosVersion := buildUpdateIosVersion("v1.97.0-alpha")
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
  #commit: '',
  tag: 'v1.97.0-alpha'
}

GITHUB_ORG = 'wordpress-mobile'
REPO_NAME = 'gutenberg-mobile'
`)

		got, err := updateIosVersion(config, repo.PullRequest{})
		assertNoError(t, err)
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
