package gh

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
)

// Branch represents a GitHub branch API schema.
type Branch struct {
	Name   string
	Commit struct {
		Sha string
	}
}

type Label struct {
	Name string
}

type Repo struct {
	Ref   string
	Sha   string
	Owner struct{ Login string }
}

// PullRequest represents a GitHub pull request.
// Not all fields are populated by all API calls.

type User struct {
	Login string
}

type PullRequest struct {
	Number             int
	Url                string `json:"html_url"`
	ApiUrl             string `json:"url"`
	Body               string
	Title              string
	Labels             []Label `json:"labels"`
	State              string
	User               User
	Draft              bool
	Mergeable          bool
	Head               Repo
	Base               Repo
	RequestedReviewers []User `json:"requested_reviewers"`

	// This field is not part of the GH api but is useful
	// to get the context of the PR when passing it around
	Repo string

	// This field is not part of the GH api
	// It's used to suggest if a PR is for a release
	// or not.
	ReleaseVersion string
}

// SearchBranch returns a branch for the given repo and branch name.
func SearchBranch(rpo, branch string) (Branch, error) {
	org, err := repo.GetOrg(rpo)
	if err != nil {
		return Branch{}, err
	}
	response := Branch{}
	client := getClient()
	endpoint := fmt.Sprintf("repos/%s/%s/branches/%s", org, rpo, branch)
	if err := client.Get(endpoint, &response); err != nil {
		return Branch{}, err
	}
	return response, nil
}

func getClient() *api.RESTClient {
	client, err := api.DefaultRESTClient()
	if err != nil {
		fmt.Printf("Error getting client: %v", err)
		os.Exit(1)
	}
	return client
}
