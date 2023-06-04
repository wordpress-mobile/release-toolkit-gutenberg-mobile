package gbm

import (
	"testing"
)

func TestUpdateIosVersion(t *testing.T) {

	testWriter := func(buf []byte, file string) ([]byte, error) {
		return buf, nil
	}

	t.Run("Updates a tag to a commit", func(t *testing.T) {

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

		got, err := updateIosVersion("123", "test", config, testWriter)
		assertNoError(t, err)
		assertStringEqual(t, string(want), string(got))

	})

	t.Run("Updates a commit to a tag", func(t *testing.T) {
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

		got, err := updateIosVersion("v1.0.0", "test", config, testWriter)
		assertNoError(t, err)
		assertStringEqual(t, string(want), string(got))
	})

	t.Run("Updates a commit to a commit", func(t *testing.T) {
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

		got, err := updateIosVersion("456", "test", config, testWriter)
		assertNoError(t, err)
		assertStringEqual(t, string(want), string(got))
	})

	t.Run("Updates a tag to a tag", func(t *testing.T) {
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

		got, err := updateIosVersion("v1.97.0-alpha", "testdata/wtf", config, testWriter)
		assertNoError(t, err)
		assertStringEqual(t, string(want), string(got))
	})
}
