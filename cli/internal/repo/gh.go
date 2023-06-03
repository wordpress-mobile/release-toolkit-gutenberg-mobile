package repo

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
)

// PullRequest represents a GitHub pull request.
// Not all fields are populated by all API calls.
type PullRequest struct {
	Number int
	Url    string `json:"html_url"`
	Body   string
	Title  string
	Labels []struct{ Name string }
	State  string
	User   struct {
		Login string
	}
	Draft     bool
	Mergeable bool
	Org       string
	Head      struct {
		Ref string
		Sha string
	}
	Base struct {
		Ref string
		Sha string
	}
}

// getClient returns a REST client for the GitHub API.
func getClient() *api.RESTClient {
	client, err := api.DefaultRESTClient()
	if err != nil {
		fmt.Errorf("Error getting client: %v", err)
		os.Exit(1)
	}
	return client
}

// GetPr returns a PullRequest struct for the given repo and PR number.
func GetPr(repo string, id int) (PullRequest, error) {
	org := getOrg(repo)
	client := getClient()

	endpoint := fmt.Sprintf("repos/%s/%s/pulls/%d", org, repo, id)
	response := PullRequest{}
	if err := client.Get(endpoint, &response); err != nil {
		return PullRequest{}, err
	}

	if response.Number == 0 {
		return PullRequest{}, fmt.Errorf("pr not found %s", endpoint)
	}

	return response, nil
}
