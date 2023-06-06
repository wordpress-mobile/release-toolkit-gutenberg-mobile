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
				defer gock.Off()

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

				_, err := GetPr(r.repo, prNumber)
				assertNoError(t, err)
			})
			defer gock.Off()
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

		pr := createTestPr(t)

		want := pr
		want.Number = 123

		// Confirm that the PR is not equal to the want
		assertNotEqual(t, pr, want)

		resp, _ := json.Marshal(&want)

		gock.New("https://api.github.com").
			Post("/repos/TEST/gutenberg-mobile/pulls").
			Reply(200).
			JSON(string(resp))
		defer gock.Off()

		err := CreatePr("gutenberg-mobile", &pr)
		assertNoError(t, err)
		assertEqual(t, pr, want)

	})

	t.Run("It returns an error if the request fails", func(t *testing.T) {
		setupMockOrg(t, "TEST")

		gock.New("https://api.github.com").
			Post("/repos/TEST/gutenberg-mobile/pulls").
			Reply(422)
		defer gock.Off()

		pr := PullRequest{}

		err := CreatePr("gutenberg-mobile", &pr)
		assertError(t, err)
	})

	t.Run("It adds labels to the PR if provided", func(t *testing.T) {
		setupMockOrg(t, "TEST")

		pr := createTestPr(t)

		labels := []Label{
			{Name: "awesome-sauce"},
		}

		pr.Labels = labels

		want := pr
		want.Number = 123

		// Confirm that the PR is not equal to the want
		assertNotEqual(t, pr, want)

		resp, _ := json.Marshal(&want)

		gock.New("https://api.github.com").
			Post("/repos/TEST/gutenberg-mobile/pulls").
			Reply(200).
			JSON(string(resp))

		gock.New("https://api.github.com").
			Post("/repos/TEST/gutenberg-mobile/issues/123/labels").
			Reply(200).
			JSON(labels)
		defer gock.Off()

		err := CreatePr("gutenberg-mobile", &pr)
		assertNoError(t, err)
		assertEqual(t, pr, want)

	})
}

func TestUpdatePR(t *testing.T) {

	t.Run("It updates the passed in PR struct on success ", func(t *testing.T) {
		setupMockOrg(t, "TEST")

		pr := createTestPr(t)
		pr.Number = 123

		labels := []Label{
			{Name: "awesome-sauce"},
		}
		pr.Labels = labels
		want := pr
		want.Title = "Updated TEST PR"

		resp, _ := json.Marshal(&want)

		gock.New("https://api.github.com").
			Patch("/repos/TEST/gutenberg-mobile/pulls/123").
			Reply(200).
			JSON(string(resp))

		gock.New("https://api.github.com").
			Post("/repos/TEST/gutenberg-mobile/issues/123/labels").
			Reply(200).
			JSON(labels)
		defer gock.Off()

		update := PrUpdate{
			Title: "Updated TEST PR",
		}

		err := UpdatePr("gutenberg-mobile", &pr, update)
		assertNoError(t, err)
		assertEqual(t, pr, want)
	})

	t.Run("It updates labels if present", func(t *testing.T) {
		setupMockOrg(t, "TEST")

		pr := createTestPr(t)
		pr.Number = 123

		want := pr
		want.Title = "Updated TEST PR"

		resp, _ := json.Marshal(&want)

		gock.New("https://api.github.com").
			Patch("/repos/TEST/gutenberg-mobile/pulls/123").
			Reply(200).
			JSON(string(resp))
		defer gock.Off()

		update := PrUpdate{
			Title: "Updated TEST PR",
		}

		err := UpdatePr("gutenberg-mobile", &pr, update)
		assertNoError(t, err)
		assertEqual(t, pr, want)
	})
}

func TestAddLabels(t *testing.T) {

	t.Run("It adds labels to a PR", func(t *testing.T) {
		setupMockOrg(t, "TEST")

		pr := createTestPr(t)
		pr.Number = 123

		labels := []Label{
			{Name: "foo"},
			{Name: "bar"},
		}

		pr.Labels = labels

		gock.New("https://api.github.com").
			Post("/repos/TEST/gutenberg-mobile/issues/123/labels").
			Reply(200).
			JSON(labels)
		defer gock.Off()

		err := AddLabels("gutenberg-mobile", &pr)
		assertNoError(t, err)
		assertEqual(t, pr.Labels, labels)
	})

	t.Run("It returns an error if the request fails", func(t *testing.T) {
		setupMockOrg(t, "TEST")

		pr := createTestPr(t)
		pr.Number = 123

		gock.New("https://api.github.com").
			Post("/repos/TEST/gutenberg-mobile/issues/123/labels").
			Reply(422)

		defer gock.Off()

		err := AddLabels("gutenberg-mobile", &pr)
		assertError(t, err)
	})

	t.Run("It returns an error if the labels are empty", func(t *testing.T) {
		setupMockOrg(t, "TEST")

		pr := createTestPr(t)
		pr.Number = 123

		err := AddLabels("gutenberg-mobile", &pr)
		assertError(t, err)
	})
}

func TestRemoveAllLabel(t *testing.T) {
	t.Run("It removes a labels from a PR if the request is successful", func(t *testing.T) {
		setupMockOrg(t, "TEST")

		pr := createTestPr(t)
		pr.Number = 123

		pr.Labels = []Label{
			{Name: "foo"},
			{Name: "bar"},
		}

		gock.New("https://api.github.com").
			Post("/repos/TEST/gutenberg-mobile/issues/123/labels").
			Reply(200).
			JSON("[]")
		defer gock.Off()

		err := RemoveAllLabels("gutenberg-mobile", &pr)
		assertNoError(t, err)
		if len(pr.Labels) != 0 {
			t.Errorf("Expected labels to be removed")
		}
	})

	t.Run("It does not remove the label from the PR if the request fails", func(t *testing.T) {
	})
}

func TestSearchPrs(t *testing.T) {
	t.Run("It returns a list of PRs", func(t *testing.T) {
		setupMockOrg(t, "TEST")

		prs := []PullRequest{
			{Title: "Test PR"},
			{Title: "Another PR"},
		}

		gock.New("https://api.github.com").
			Get("/search/issues").
			Reply(200).
			JSON(SearchResult{
				Items:      prs,
				TotalCount: len(prs),
			})
		defer gock.Off()

		got, err := SearchPrs(RepoFilter{})
		assertNoError(t, err)
		assertEqual(t, got.Items, prs)
		assertEqual(t, got.TotalCount, len(prs))
	})

	t.Run("It returns an error if the request fails", func(t *testing.T) {
		setupMockOrg(t, "TEST")

		gock.New("https://api.github.com").
			Get("/search/issuesXX").
			Reply(422)
		defer gock.Off()

		_, err := SearchPrs(RepoFilter{})
		assertError(t, err)
	})
}

func TestBuildRepoFilter(t *testing.T) {

	t.Run("It returns a filter with the org", func(t *testing.T) {
		setupMockOrg(t, "TEST")

		filter := BuildRepoFilter("gutenberg-mobile", "")
		assertEqual(t, filter.Repo, "TEST/gutenberg-mobile")
	})

	t.Run("It encodes the queries", func(t *testing.T) {
		setupMockOrg(t, "TEST")

		filter := BuildRepoFilter("gutenberg", "is:open", "is:pr", `label:"Mobile App - i.e. Android or iOS"`)
		assertEqual(t, filter.Query, "is%3Aopen+is%3Apr+label%3A%22Mobile+App+-+i.e.+Android+or+iOS%22+repo%3ATEST%2Fgutenberg")
	})

}

// Helpers

func setupMockOrg(t *testing.T, org string) {
	t.Helper()
	t.Setenv("GBM_WPMOBILE_ORG", org)
	t.Setenv("GBM_WORDPRESS_ORG", org)
	initOrgs()
	t.Cleanup(func() {
		t.Setenv("GBM_WPMOBILE_ORG", "")
		t.Setenv("GBM_WORDPRESS_ORG", "")
		initOrgs()
	})
}

func createTestPr(t *testing.T) PullRequest {
	t.Helper()

	return PullRequest{
		Title: "Test PR",
		Body:  "This is a test PR",
		Head: struct {
			Ref string
			Sha string
		}{Ref: "test/branch"},
		Base: struct {
			Ref string
			Sha string
		}{Ref: "trunk"},
		Draft: true,
	}
}
