package repo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/fatih/color"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
)

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
type PullRequest struct {
	Number int
	Url    string `json:"html_url"`
	Body   string
	Title  string
	Labels []Label `json:"labels"`
	State  string
	User   struct {
		Login string
	}
	Draft     bool
	Mergeable bool
	// Org       string
	Head Repo
	Base Repo

	// This field is not part of the GH api but is useful
	// to get the context of the PR when passing it around
	Repo string

	// This field is not part of the GH api
	// It's used to suggest if a PR is for a release
	// or not.
	ReleaseVersion string
}

type Status struct {
	State       string
	Description string
	Context     string
	Created     string `json:"created_at"`
	Updated     string `json:"updated_at"`
}

type RefStatus struct {
	State    string
	Statuses []Status
}

// Used to send PR update requests to the GitHub API.
type PrUpdate struct {
	Title string
	Body  string
	State string
	Base  string
}

// RepoFilter is used to filter PRs by repo and query.
type RepoFilter struct {
	Repo  string
	Query string
}

// SearchResult is used to return a list of PRs from a search.
type SearchResult struct {
	Filter     RepoFilter
	TotalCount int `json:"total_count"`
	Items      []PullRequest
}

// Branch represents a GitHub branch API schema.
type Branch struct {
	Name   string
	Commit struct {
		Sha string
	}
}

type BranchError struct {
	Err  error
	Type string
}

func (r *BranchError) Error() string {
	return r.Err.Error()
}

func IsBranchError(err error) bool {
	_, ok := err.(*BranchError)
	return ok
}

func IsExistingBranchError(err error) bool {
	if !IsBranchError(err) {
		return false
	}
	return err.(*BranchError).Type == "exists"
}

type GhContents struct {
	Name string
	Sha  string
	Type string
	Url  string
}

// GetPr returns a PullRequest struct for the given repo and PR number.
func GetPr(repo string, id int) (PullRequest, error) {
	org, err := GetOrg(repo)
	if err != nil {
		return PullRequest{}, err
	}
	return GetPrOrg(org, repo, id)
}

func GetPrOrg(org, repo string, id int) (PullRequest, error) {
	client := getClient()

	endpoint := fmt.Sprintf("repos/%s/%s/pulls/%d", org, repo, id)
	response := PullRequest{}
	if err := client.Get(endpoint, &response); err != nil {
		return PullRequest{}, err
	}

	if response.Number == 0 {
		return PullRequest{}, fmt.Errorf("pr not found %s", endpoint)
	}

	response.Repo = repo

	return response, nil
}

func PreviewPr(repo, dir string, pr *PullRequest) {
	org, _ := GetOrg(repo)
	boldUnder := color.New(color.Bold, color.Underline).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Println(boldUnder("\nPr Preview"))
	fmt.Println(bold("Local:"), "\t", cyan(dir))
	fmt.Println(bold("Repo:"), "\t", cyan(fmt.Sprintf("%s/%s", org, repo)))
	fmt.Println(bold("Title:"), "\t", cyan(pr.Title))
	fmt.Print(bold("Body:\n"), cyan(pr.Body))
	fmt.Println(bold("Commits:"))
	exc := exec.Command(
		"git",
		"log",
		"trunk...HEAD",
		"--oneline",
		"--no-merges",
		"-10",
	)
	exc.Dir = dir
	exc.Stdout = os.Stdout

	if err := exc.Run(); err != nil {
		fmt.Println(err)
	}
}

func CreatePr(repo string, pr *PullRequest) error {
	client := getClient()
	org, err := GetOrg(repo)
	if err != nil {
		return err
	}

	labels := pr.Labels
	endpoint := fmt.Sprintf("repos/%s/%s/pulls", org, repo)

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
		if err := AddLabels(repo, pr); err != nil {
			return err
		}
	}
	return nil
}

func UpdatePr(pr *PullRequest) error {
	org, repo, err := getOrgRepo(pr)
	if err != nil {
		return err
	}
	client := getClient()

	endpoint := fmt.Sprintf("repos/%s/%s/pulls/%d", org, repo, pr.Number)

	update := struct {
		Title string `json:"title,omitempty"`
		Body  string `json:"body,omitempty"`
		State string `json:"state,omitempty"`
		Base  string `json:"base,omitempty"`
	}{
		Title: pr.Title,
		Body:  pr.Body,
		State: pr.State,
		Base:  pr.Base.Ref,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(update); err != nil {
		return err
	}

	if err := client.Patch(endpoint, &buf, &pr); err != nil {
		return err
	}

	if pr.Labels != nil {
		if err := AddLabels(repo, pr); err != nil {
			return err
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

// Removes all labels from a PR
func RemoveAllLabels(repo string, pr *PullRequest) error {
	_, err := labelRequest(repo, pr.Number, []string{})
	if err != nil {
		return err
	}
	pr.Labels = nil
	return nil
}

// Build a RepoFilter from a repo name and a list of queries.
func BuildRepoFilter(repo string, queries ...string) RepoFilter {

	// We just need to warn if the org is not found.
	org, err := GetOrg(repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", err)
	}
	var encoded []string
	queries = append(queries, fmt.Sprintf("repo:%s/%s", org, repo))

	for _, q := range queries {
		encoded = append(encoded, url.QueryEscape(q))
	}

	return RepoFilter{
		Repo:  repo,
		Query: strings.Join(encoded, "+"),
	}
}

// SearchPRs returns a list of PRs matching the given filter.
func SearchPrs(filter RepoFilter) (SearchResult, error) {
	client := getClient()
	endpoint := fmt.Sprintf("search/issues?q=%s", filter.Query)
	response := SearchResult{Filter: filter}

	if err := client.Get(endpoint, &response); err != nil {
		return SearchResult{}, err
	}
	return response, nil
}

// SearchBranch returns a branch for the given repo and branch name.
func SearchBranch(repo, branch string) (Branch, error) {
	org, err := GetOrg(repo)
	if err != nil {
		return Branch{}, err
	}
	response := Branch{}
	client := getClient()
	endpoint := fmt.Sprintf("repos/%s/%s/branches/%s", org, repo, branch)
	if err := client.Get(endpoint, &response); err != nil {
		return Branch{}, err
	}
	return response, nil
}

// Find all the PRs that are synced with the given Gutenberg Mobile PR
// A PR is considered synced if the PR body contains the GBM pr url
func FindGbmSyncedPrs(gbmPr PullRequest, filters []RepoFilter) ([]SearchResult, error) {
	var synced []SearchResult
	prChan := make(chan SearchResult)

	// Search for PRs in parallel
	for _, rf := range filters {
		go func(rf RepoFilter) {
			res, err := SearchPrs(rf)

			// just log the error and continue
			if err != nil {
				fmt.Println(err)
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

func GetPrStatus(pr *PullRequest) (RefStatus, error) {
	org, repo, err := getOrgRepo(pr)
	if err != nil {
		utils.LogError("%s", err)
		return RefStatus{}, err
	}
	client := getClient()
	ref := pr.Head.Ref

	endpoint := fmt.Sprintf("repos/%s/%s/commits/%s/status", org, repo, ref)
	utils.LogDebug(endpoint)
	fs := RefStatus{}
	if err := client.Get(endpoint, &fs); err != nil {
		fmt.Println(err)
		return RefStatus{}, err
	}

	return fs, nil
}

func GetContents(repo, file, branch string) (GhContents, error) {
	org, err := GetOrg(repo)
	if err != nil {
		return GhContents{}, err
	}
	client := getClient()
	endpoint := fmt.Sprintf("repos/%s/%s/contents/%s?ref=%s", org, repo, file, branch)
	response := GhContents{}
	if err := client.Get(endpoint, &response); err != nil {
		return GhContents{}, err
	}
	return response, nil
}

// getClient returns a REST client for the GitHub API.
func getClient() *api.RESTClient {
	client, err := api.DefaultRESTClient()
	if err != nil {
		fmt.Printf("Error getting client: %v", err)
		os.Exit(1)
	}
	return client
}

func labelRequest(repo string, prNum int, labels []string) ([]Label, error) {
	org, err := GetOrg(repo)
	if err != nil {
		return nil, err
	}

	client := getClient()

	endpoint := fmt.Sprintf("repos/%s/%s/issues/%d/labels", org, repo, prNum)

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

func getOrgRepo(pr *PullRequest) (org string, repo string, err error) {

	if repo = pr.Repo; repo == "" {
		return "", "", errors.New("Pr is missing a repo")
	}
	if org, err = GetOrg(repo); err != nil {
		return "", "", fmt.Errorf("Unable to determine the org for the %s repo", repo)
	}
	return org, repo, nil
}
