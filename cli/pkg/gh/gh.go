package gh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/fatih/color"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/exec"
	"github.com/wordpress-mobile/gbm-cli/pkg/repo"
)

// Branch represents a GitHub branch API schema.
type Branch struct {
	Name   string
	Commit struct {
		Sha string
	}
	StatusCode int
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

// RepoFilter is used to filter PRs by repo and query.
type RepoFilter struct {
	Repo        string
	Query       string
	QueryString string
}

// SearchResult is used to return a list of PRs from a search.
type SearchResult struct {
	Filter     RepoFilter
	TotalCount int `json:"total_count"`
	Items      []PullRequest
}

// Build a RepoFilter from a repo name and a list of queries.
func BuildRepoFilter(rpo string, queries ...string) RepoFilter {
	org, _ := repo.GetOrg(rpo)

	var encoded []string
	queries = append(queries, fmt.Sprintf("repo:%s/%s", org, rpo))

	queryString := strings.Join(queries, " ")

	for _, q := range queries {
		encoded = append(encoded, url.QueryEscape(q))
	}

	return RepoFilter{
		Repo:        rpo,
		Query:       strings.Join(encoded, "+"),
		QueryString: queryString,
	}
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

	// @TODO need to figure out how to handle 404s, right now a 404 is an error but
	// a 404 should return that the branch doesn't exist. But we should check for other network errors
	if err := client.Get(endpoint, &response); err != nil {
		return Branch{}, nil
	}
	return response, nil
}

// SearchPrs returns a list of PRs for the given repo and filter.
func SearchPrs(filter RepoFilter) (SearchResult, error) {
	client := getClient()
	endpoint := fmt.Sprintf("search/issues?q=%s", filter.Query)
	response := SearchResult{Filter: filter}

	if err := client.Get(endpoint, &response); err != nil {
		return SearchResult{}, err
	}
	return response, nil
}

// Returns a single PR with all the details given a filter.
// Returns an error if more than one PR is found.
func SearchPr(filter RepoFilter) (PullRequest, error) {
	result, err := SearchPrs(filter)
	if err != nil {
		return PullRequest{}, err
	}

	// Don't return an error if no PRs are found. This is useful for checking if the PR exists
	if result.TotalCount == 0 {
		return PullRequest{}, nil
	}
	if result.TotalCount > 1 {
		return PullRequest{}, fmt.Errorf("too many PRs found")
	}
	number := result.Items[0].Number
	return GetPr(filter.Repo, number)
}

func GetPr(rpo string, number int) (PullRequest, error) {
	pr := PullRequest{}
	org, err := repo.GetOrg(rpo)
	if err != nil {
		return pr, err
	}

	client := getClient()
	endpoint := fmt.Sprintf("repos/%s/%s/pulls/%d", org, rpo, number)
	if err := client.Get(endpoint, &pr); err != nil {
		return pr, err
	}
	pr.Repo = rpo
	return pr, nil
}

func CreatePr(rpo string, pr *PullRequest) error {
	client := getClient()
	org, err := repo.GetOrg(rpo)
	if err != nil {
		return err
	}

	labels := pr.Labels
	endpoint := fmt.Sprintf("repos/%s/%s/pulls", org, rpo)

	// We need to flatten the struct to match the API
	npr := struct {
		Title string `json:"title"`
		Body  string `json:"body"`
		Head  string `json:"head"`
		Base  string `json:"base"`
		Draft bool   `json:"draft"`
	}{
		Title: pr.Title,
		Body:  pr.Body,
		Head:  pr.Head.Ref,
		Base:  pr.Base.Ref,
		Draft: pr.Draft,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(npr); err != nil {
		return err
	}

	if err := client.Post(endpoint, &buf, &pr); err != nil {
		return err
	}

	// To date there is no way to set the labels on creation
	// so we need to send the label to the labels endpoint
	pr.Labels = labels
	if pr.Labels != nil {
		if err := AddLabels(rpo, pr); err != nil {
			console.Warn("Unable to add label '%s' to PR, are you sure it exists on the %s/%s repo?", err, org, rpo)
		}
	}
	return nil
}

// Adds labels to a PR
func AddLabels(repo string, pr *PullRequest) error {

	labels := []string{}
	for _, label := range pr.Labels {
		labels = append(labels, label.Name)
	}
	if len(labels) == 0 {
		return fmt.Errorf("no labels to add")
	}

	resp, err := labelRequest(repo, pr.Number, labels)

	if err != nil {
		return err
	}
	pr.Labels = resp
	return nil
}

func getClient() *api.RESTClient {
	client, err := api.DefaultRESTClient()
	if err != nil {
		fmt.Printf("Error getting client: %v", err)
		os.Exit(1)
	}
	return client
}

func labelRequest(rpo string, prNum int, labels []string) ([]Label, error) {
	org, err := repo.GetOrg(rpo)
	if err != nil {
		return nil, err
	}

	client := getClient()

	endpoint := fmt.Sprintf("repos/%s/%s/issues/%d/labels", org, rpo, prNum)

	type labelBody struct {
		Labels []string `json:"labels"`
	}

	pbody := labelBody{Labels: labels}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(pbody); err != nil {
		return nil, err
	}

	resp := []Label{}

	if err := client.Post(endpoint, &buf, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func PreviewPr(rpo, dir string, pr PullRequest) {
	org, _ := repo.GetOrg(rpo)
	cyan := color.New(color.FgCyan, color.Bold).SprintfFunc()
	console.Log(cyan("\nPr Preview"))
	console.Log(cyan("Local:")+" %s\n", dir)
	console.Log(cyan("Repo:")+" %s/%s\n", org, rpo)
	console.Log(cyan("Title:")+" %s\n", pr.Title)
	console.Log(cyan("Body:\n")+"%s\n", pr.Body)
	console.Log(cyan("Commits:"))

	git := exec.Git(dir, true)

	git("log", pr.Base.Ref+"...HEAD", "--oneline", "--no-merges", "-10")

	console.Info("\n")
}
