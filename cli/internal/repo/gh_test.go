package repo

import (
	"encoding/json"
	"fmt"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

func TestGetPr(t *testing.T) {

	t.Run("It returns an error using a non GBM repo", func(t *testing.T) {
		_, err := GetPr("wp-calypso", 123)
		assertError(t, err)
	})

	t.Run("It determines the org from the repo", func(t *testing.T) {
		repos := []struct {
			org  string
			repo string
		}{
			{org: "WordPress", repo: "gutenberg"},
			{org: "Automattic", repo: "jetpack"},
			{org: "wordpress-mobile", repo: "gutenberg-mobile"},
			{org: "wordpress-mobile", repo: "WordPress-Android"},
			{org: "wordpress-mobile", repo: "WordPress-iOS"},
		}

		for _, r := range repos {
			t.Run(fmt.Sprintf("It checks %s given %s", r.org, r.repo), func(t *testing.T) {
				t.Cleanup(gock.Off)
				prNumber := 123
				path := fmt.Sprintf("/repos/%s/%s/pulls/%d", r.org, r.repo, prNumber)
				gock.New("https://api.github.com").
					Get(path).
					Reply(200).
					JSON(fmt.Sprintf(`{"number": %d}`, prNumber))

				pr, err := GetPr(r.repo, prNumber)
				assertNoError(t, err)
				if pr.Number != prNumber {
					t.Errorf("expected pr.Number to be 123 but got %d", pr.Number)
				}
			})
		}
	})

	t.Run("It allows overriding the org with env settings", func(t *testing.T) {
		repos := []struct {
			env  string
			org  string
			repo string
		}{
			{env: "GBM_WORDPRESS_ORG", org: "my-WordPress", repo: "gutenberg"},
			{env: "GBM_AUTOMATTIC_ORG", org: "my-automattic", repo: "jetpack"},
			{env: "GBM_WPMOBILE_ORG", org: "my-wordpress-mobile", repo: "gutenberg-mobile"},
			{env: "GBM_WPMOBILE_ORG", org: "my-wordpress-mobile", repo: "WordPress-Android"},
			{env: "GBM_WPMOBILE_ORG", org: "my-wordpres-mobile", repo: "WordPress-iOS"},
		}

		prNumber := 123
		t.Cleanup(gock.Off)

		for _, r := range repos {
			t.Run(fmt.Sprintf("It uses the env %s for the %s org", r.env, r.repo), func(t *testing.T) {

				// Set the mock orgs and reset the orgs
				t.Setenv(r.env, r.org)
				initOrgs()

				path := fmt.Sprintf("/repos/%s/%s/pulls/%d", r.org, r.repo, prNumber)
				gock.New("https://api.github.com").
					Get(path).
					Reply(200).
					JSON(fmt.Sprintf(`{"number": %d}`, prNumber))
			})
			_, err := GetPr(r.repo, prNumber)
			assertNoError(t, err)
		}
	})

	t.Run("It returns an error when the PR is not found", func(t *testing.T) {
		t.Skip()
		t.Cleanup(gock.Off)
		gock.New("https://api.github.com").
			Get("/repos/WordPress/gutenberg/pulls/1").
			Reply(404).
			JSON(`{"message": "Not Found"}`)

		_, err := GetPr("gutenberg", 1)
		assertError(t, err)
	})
}

func TestCreatePR(t *testing.T) {

	t.Run("It populates the passed in PR struct on success ", func(t *testing.T) {

		setupMockOrg(t, "TEST")

		pr := PullRequest{
			Title: "Test PR",
			Body:  "This is a test PR",
			Head: struct {
				Ref string
				Sha string
			}{Ref: "try/sentry"},
			Base: struct {
				Ref string
				Sha string
			}{Ref: "trunk"},
			Draft: true,
		}

		want := pr
		want.Number = 123

		resp, _ := json.Marshal(&want)

		gock.New("https://api.github.com").
			Post("/repos/TEST/gutenberg-mobile/pulls").
			Reply(200).
			JSON(string(resp))
		t.Cleanup(gock.Off)

		err := CreatePr("gutenberg-mobile", &pr)
		assertNoError(t, err)
		assertEqual(t, pr.Number, want.Number)

	})

	t.Run("It returns an error if the request fails", func(t *testing.T) {
		setupMockOrg(t, "TEST")

		gock.New("https://api.github.com").
			Post("/repos/TEST/gutenberg-mobile/pulls").
			Reply(422)
		t.Cleanup(gock.Off)

		pr := PullRequest{}

		err := CreatePr("gutenberg-mobile", &pr)
		assertError(t, err)
	})
}

func setupMockOrg(t *testing.T, org string) {
	t.Helper()
	t.Setenv("GBM_WPMOBILE_ORG", org)
	initOrgs()
	t.Cleanup(func() {
		t.Setenv("GBM_WPMOBILE_ORG", "")
		initOrgs()
	})
}
