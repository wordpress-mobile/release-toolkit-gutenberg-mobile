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
	if err := client.Get(endpoint, &response); err != nil {
		return Branch{}, err
	}
	return response, nil
}

// GetPr returns a PullRequest struct for the given repo and PR number.
func GetPr(rpo string, id int) (*PullRequest, error) {
	org, err := repo.GetOrg(rpo)
	if err != nil {
		return nil, err
	}
	return GetPrOrg(org, rpo, id)
}

func GetPrOrg(org, repo string, id int) (*PullRequest, error) {
	client := getClient()

	endpoint := fmt.Sprintf("repos/%s/%s/pulls/%d", org, repo, id)
	pr := &PullRequest{}
	if err := client.Get(endpoint, pr); err != nil {
		return nil, err
	}

	if pr.Number == 0 {
		return nil, fmt.Errorf("pr not found %s", endpoint)
	}

	pr.Repo = repo

	return pr, nil
}

func SearchPrs(filter RepoFilter) (SearchResult, error) {
	console.Info("Searching for PRs matching %s", filter.QueryString)
	client := getClient()
	endpoint := fmt.Sprintf("search/issues?q=%s", filter.Query)
	response := SearchResult{Filter: filter}

	if err := client.Get(endpoint, &response); err != nil {
		return SearchResult{}, err
	}
	return response, nil
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

func FindGbmSyncedPrs(gbmPr PullRequest, filters []RepoFilter) ([]SearchResult, error) {
	var synced []SearchResult
	prChan := make(chan SearchResult)

	// Search for PRs in parallel
	for _, rf := range filters {
		go func(rf RepoFilter) {
			res, err := SearchPrs(rf)

			// just log the error and continue
			if err != nil {
				console.Warn("could not search for %s", err)
			}
			prChan <- res
		}(rf)
	}

	// Wait for all the PRs to be returned
	for i := 0; i < len(filters); i++ {
		resp := <-prChan
		sItems := []PullRequest{}

		for _, pr := range resp.Items {
			if strings.Contains(pr.Body, gbmPr.Url) {
				pr.Repo = resp.Filter.Repo
				sItems = append(sItems, pr)
			}
		}
		resp.Items = sItems
		synced = append(synced, resp)
	}

	return synced, nil
}
