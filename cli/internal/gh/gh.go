package gh

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/fatih/color"

	"github.com/wordpress-mobile/gbm-cli/internal/exc"
	rpo "github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
)

func l(f string, a ...interface{}) {
	utils.LogInfo(f, a...)
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

type Release struct {
	Url           string `json:"url"`
	TagName       string `json:"tag_name"`
	PublishedDate string `json:"published_at"`
	Draft         bool   `json:"draft"`
	Prerelease    bool   `json:"prerelease"`
	Target        string `json:"target_commitish"`
}

type ReleaseProps struct {
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish"`
	Name            string `json:"name"`
	Body            string `json:"body"`
}

// GetPr returns a PullRequest struct for the given repo and PR number.
func GetPr(repo string, id int) (*PullRequest, error) {
	org, err := rpo.GetOrg(repo)
	if err != nil {
		return nil, err
	}
	return GetPrOrg(org, repo, id)
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

func CreatePr(repo string, pr *PullRequest) error {
	client := getClient()
	org, err := rpo.GetOrg(repo)
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
	org, err := rpo.GetOrg(repo)
	if err != nil {
		l(utils.WarnString("could not find org for %s", err))
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
	org, err := rpo.GetOrg(repo)
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
				l(utils.WarnString("could not search for %s", err))
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
		return RefStatus{}, err
	}
	client := getClient()
	ref := pr.Head.Ref

	endpoint := fmt.Sprintf("repos/%s/%s/commits/%s/status", org, repo, ref)
	fs := RefStatus{}
	if err := client.Get(endpoint, &fs); err != nil {
		return RefStatus{}, err
	}

	return fs, nil
}

func GetContents(repo, file, branch string) (GhContents, error) {
	org, err := rpo.GetOrg(repo)
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

func GetRelease(repo, version string) (Release, error) {
	org, err := rpo.GetOrg(repo)
	if err != nil {
		return Release{}, err
	}
	client := getClient()
	endpoint := fmt.Sprintf("repos/%s/%s/releases/tags/%s", org, repo, version)
	response := Release{}
	if err := client.Get(endpoint, &response); err != nil {
		return Release{}, err
	}
	return response, nil
}

func CreateRelease(repo string, rp *ReleaseProps) error {
	client := getClient()
	org, err := rpo.GetOrg(repo)
	if err != nil {
		return err
	}
	endpoint := fmt.Sprintf("repos/%s/%s/releases", org, repo)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(rp); err != nil {
		return err
	}

	resp := http.Response{}

	if err := client.Post(endpoint, &buf, resp); err != nil {
		return err
	}
	return nil
}

type Tagger struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}
type AnnotatedTag struct {
	Tag     string `json:"tag"`
	Message string `json:"message"`
	Type    string `json:"type"`
	Tagger  Tagger `json:"tagger"`
	Sha     string `json:"sha"`
}

func CreateLightweightTag(repo string, tag *AnnotatedTag) error {

	if tag.Sha == "" {
		return errors.New("sha is required")
	}
	body := struct {
		Ref string `json:"ref"`
		Sha string `json:"sha"`
	}{
		Ref: fmt.Sprintf("refs/tags/%s", tag.Tag),
		Sha: tag.Sha,
	}
	org, _ := rpo.GetOrg(repo)
	client := getClient()
	endpoint := fmt.Sprintf("repos/%s/%s/git/refs", org, repo)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return err
	}

	resp := http.Response{}

	return client.Post(endpoint, &buf, resp)
}

func CreateAnnotatedTag(repo, sha, tag, message string) (*AnnotatedTag, error) {
	org, _ := rpo.GetOrg(repo)
	client := getClient()

	endpoint := fmt.Sprintf("repos/%s/%s/git/tags", org, repo)

	sig := rpo.Signature()

	t := struct {
		Tag     string `json:"tag"`
		Message string `json:"message"`
		Object  string `json:"object"`
		Type    string `json:"type"`
		Tagger  Tagger `json:"tagger"`
	}{
		Tag:     tag,
		Message: message,
		Object:  sha,
		Type:    "commit",
		Tagger: Tagger{
			Name:  sig.Name,
			Email: sig.Email,
			Date:  sig.When.Format(time.RFC3339),
		},
	}

	resp := &AnnotatedTag{}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(t); err != nil {
		return resp, err
	}
	if err := client.Post(endpoint, &buf, resp); err != nil {
		return resp, err
	}
	if err := CreateLightweightTag(repo, resp); err != nil {
		return resp, err
	}
	return resp, nil
}

type Review struct {
	State string `json:"state"`
}

// The PR is considered approved if the last review is approved and there are no pending reviews.
// We don't care which commit was approved, just that the PR is approved.
func IsPrApproved(pr *PullRequest) bool {
	client := getClient()
	org, repo, err := getOrgRepo(pr)
	if err != nil {
		return false
	}
	endpoint := fmt.Sprintf("repos/%s/%s/pulls/%d/reviews", org, repo, pr.Number)

	reviews := []Review{}

	if err := client.Get(endpoint, &reviews); err != nil {
		l(utils.WarnString("Error getting reviews: %v", err))
		return false
	}

	numR := len(reviews)

	if numR == 0 {
		return false
	}

	// If the last review is not approved, not approved
	if reviews[numR-1].State != "APPROVED" {
		return false
	}

	// Check to see if there is an additional review request
	if pr.RequestedReviewers != nil && len(pr.RequestedReviewers) > 0 {
		return false
	}

	return true
}

type CheckRun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}
type CheckRuns struct {
	TotalCount int        `json:"total_count"`
	CheckRuns  []CheckRun `json:"check_runs"`
}

// Checks to see if the run checks are passing.
// Use verbose to output which checks are failing.
// Also accepts a list of checks to skip.
func IsPrPassing(pr *PullRequest, skipRuns []CheckRun, verbose bool) bool {
	org, repo, err := getOrgRepo(pr)
	if err != nil {
		return false
	}
	client := getClient()
	endpoint := fmt.Sprintf("repos/%s/%s/commits/%s/check-runs", org, repo, pr.Head.Sha)
	checkRuns := CheckRuns{}
	if err := client.Get(endpoint, &checkRuns); err != nil {
		l(utils.WarnString("Error getting check runs: %v", err))
		return false
	}

	ok := func(c string) bool {
		if c == "neutral" || c == "skipped" || c == "success" {
			return true
		}
		return false
	}

	skipped := func(n string) bool {
		for _, sr := range skipRuns {
			if sr.Name == n {
				return true
			}
		}
		return false
	}

	if verbose {
		l(utils.InfoString("Checking runs checks for %s/%s@%s", org, repo, pr.Head.Sha))
	}
	passed := true
	for _, checkRun := range checkRuns.CheckRuns {
		if checkRun.Status == "completed" && !ok(checkRun.Conclusion) {
			if !skipped(checkRun.Name) {
				passed = false
				if verbose {
					l(utils.WarnString("Check run '%s' failed with conclusion %s", checkRun.Name, checkRun.Conclusion))
				}
			}
		}
	}
	if verbose {
		l("")
	}

	return passed
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
	org, err := rpo.GetOrg(repo)
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
		return "", "", errors.New("pr is missing a repo")
	}
	if org, err = rpo.GetOrg(repo); err != nil {
		return "", "", fmt.Errorf("unable to determine the org for the %s repo", repo)
	}
	return org, repo, nil
}

func PreviewPr(repo, dir string, pr *PullRequest) {
	org, _ := rpo.GetOrg(repo)
	boldUnder := color.New(color.Bold, color.Underline).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Println(boldUnder("\nPr Preview"))
	fmt.Println(bold("Local:"), "\t", cyan(dir))
	fmt.Println(bold("Repo:"), "\t", cyan(fmt.Sprintf("%s/%s", org, repo)))
	fmt.Println(bold("Title:"), "\t", cyan(pr.Title))
	fmt.Print(bold("Body:\n"), cyan(pr.Body))
	fmt.Println(bold("Commits:"))

	git := exc.ExecGit(dir, true)

	git("log", pr.Base.Ref+"...HEAD", "--oneline", "--no-merges", "-10")
}
