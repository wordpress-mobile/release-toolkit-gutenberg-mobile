package gh

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/fatih/color"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/repo"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/shell"
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
	MergeCommit        string `json:"merge_commit_sha"`

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

type Check struct {
	Id          int
	State       string
	Description string
	Context     string
}

type Status struct {
	State    string
	Passed   bool
	Statuses []Check
}

type Release struct {
	Id          int
	TagName     string `json:"tag_name"`
	Url         string `json:"html_url"`
	Name        string
	Body        string
	Draft       bool
	Prerelease  bool
	Target      string `json:"target_commitish"`
	PublishedAt string `json:"published_at"`
}

type Commit struct {
	Sha    string
	Url    string
	Commit struct {
		Message string
	}
}

type Ref struct {
	Ref    string
	Object struct {
		Sha string
		Url string
	}
}

type Tag struct {
	Sha string

	// annotated tags have a tagger
	Tagger struct {
		Date string
	}
	// lightweight tags have an author
	Author struct {
		Date string
	}

	// Lifting the date up which is not part of the api
	// but is why we can handle both annotated and lightweight tags the same way
	Date string
}

// Build a RepoFilter from a repo name and a list of queries.
func BuildRepoFilter(rpo string, queries ...string) RepoFilter {
	org := repo.GetOrg(rpo)

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
	org := repo.GetOrg(rpo)
	if org == "" {
		return Branch{}, fmt.Errorf("unable to get org for %s", rpo)
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

func GetReleaseByTag(rpo, tag string) (Release, error) {
	org := repo.GetOrg(rpo)
	client := getClient()
	endpoint := fmt.Sprintf("repos/%s/%s/releases/tags/%s", org, rpo, tag)
	release := Release{}
	if err := client.Get(endpoint, &release); err != nil {
		return Release{}, err
	}
	return release, nil
}

func GetLatestRelease(rpo string) (Release, error) {
	org := repo.GetOrg(rpo)
	client := getClient()
	endpoint := fmt.Sprintf("repos/%s/%s/releases/latest", org, rpo)
	release := Release{}
	if err := client.Get(endpoint, &release); err != nil {
		return Release{}, err
	}
	return release, nil
}

func CreateRelease(rpo string, rel *Release) error {
	org := repo.GetOrg(rpo)
	client := getClient()
	endpoint := fmt.Sprintf("repos/%s/%s/releases", org, rpo)

	nrel := struct {
		TagName         string `json:"tag_name"`
		TargetCommitish string `json:"target_commitish"`
		Name            string `json:"name"`
		Body            string `json:"body"`
	}{
		TagName:         rel.TagName,
		TargetCommitish: rel.Target,
		Name:            rel.Name,
		Body:            rel.Body,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(nrel); err != nil {
		return err
	}

	if err := client.Post(endpoint, &buf, &rel); err != nil {
		return err
	}

	return nil
}

func UploadReleaseAssets(rpo, dir string, rel Release, files ...string) error {

	org := repo.GetOrg(rpo)
	client := getClient()
	for _, file := range files {
		// Check if the file exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", file)
		}
		endpoint := fmt.Sprintf("repos/%s/%s/releases/%d/assets?name=%s", org, rpo, rel.Id, file)

		fb, err := os.Open(path.Join(dir, file))
		if err != nil {
			return err
		}
		defer fb.Close()

		if err := client.Post(endpoint, fb, struct{}{}); err != nil {
			return err
		}
	}

	return nil
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

func GetPr(rpo string, number int) (PullRequest, error) {
	pr := PullRequest{}
	org := repo.GetOrg(rpo)
	if org == "" {
		return pr, fmt.Errorf("unable to get org for %s", rpo)
	}

	client := getClient()
	endpoint := fmt.Sprintf("repos/%s/%s/pulls/%d", org, rpo, number)
	if err := client.Get(endpoint, &pr); err != nil {
		return pr, err
	}
	pr.Repo = rpo
	return pr, nil
}

func GetPrs(rpo string, numbers []string) (prs []PullRequest) {
	for _, n := range numbers {
		num, err := strconv.Atoi(n)
		if err != nil {
			console.Warn("Skipping PR %s, not a valid number", n)
			continue
		}

		if pr, err := GetPr(rpo, num); err != nil {
			console.Warn("Skipping PR %d, %s", num, err)
		} else {
			prs = append(prs, pr)
		}
	}
	return prs
}

func CreatePr(rpo string, pr *PullRequest) error {
	client := getClient()
	org := repo.GetOrg(rpo)

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

func GetTag(rpo, tag string) (Tag, error) {
	t := Tag{}

	// First we have to get the Ref for the tag
	org := repo.GetOrg(rpo)
	client := getClient()
	endpoint := fmt.Sprintf("repos/%s/%s/git/refs/tags/%s", org, rpo, tag)
	r := Ref{}
	if err := client.Get(endpoint, &r); err != nil {
		return t, err
	}

	// Then get the tag object
	if err := client.Get(r.Object.Url, &t); err != nil {
		return t, err
	}

	// Lift the date up from either the tagger (annotated) or the author (lightweight)
	if t.Tagger.Date != "" {
		t.Date = t.Tagger.Date
	}

	if t.Author.Date != "" {
		t.Date = t.Author.Date
	}
	return t, nil
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

func GetStatusChecks(rpo, sha string) (Status, error) {
	org := repo.GetOrg(rpo)

	client := getClient()
	endpoint := fmt.Sprintf("repos/%s/%s/commits/%s/status", org, rpo, sha)

	status := Status{}

	if err := client.Get(endpoint, &status); err != nil {
		return Status{}, err
	}

	status.Passed = status.State == "success"

	return status, nil
}

func GetStatusCheck(rpo, sha, context string) (Check, error) {
	status, err := GetStatusChecks(rpo, sha)
	if err != nil {
		return Check{}, err
	}

	for _, check := range status.Statuses {
		if strings.Contains(string(check.Context), context) {
			return check, nil
		}
	}

	return Check{}, fmt.Errorf("context not found")
}

func GetStatus(rpo, sha string) (string, error) {
	status, err := GetStatusChecks(rpo, sha)
	if err != nil {
		return "", err
	}

	return status.State, nil
}

func ChecksPassed(rpo, sha string) (bool, error) {
	status, err := GetStatusChecks(rpo, sha)
	if err != nil {
		return false, err
	}

	if status.State != "success" {
		return false, nil
	}
	return true, nil
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
	org := repo.GetOrg(rpo)

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

func PreviewPr(rpo, dir, branchFrom string, pr PullRequest) {
	// if not in a tty (CI) don't print the preview
	if os.Getenv("CI") == "true" {
		return
	}
	org := repo.GetOrg(rpo)
	row := console.Row

	console.Print(console.Heading, "\nPr Preview")

	white := color.New(color.FgWhite).SprintFunc()

	console.Print(row, "Repo: %s/%s", white(org), white(rpo))
	console.Print(row, "Base: %s", white(pr.Base.Ref))
	console.Print(row, "Head: %s", white(pr.Head.Ref))
	console.Print(row, "Title: %s", white(pr.Title))
	console.Print(row, "Body:\n%s", white(pr.Body))
	console.Print(row, "Commits:")

	git := shell.NewGitCmd(shell.CmdProps{Dir: dir, Verbose: true})

	git.Log(branchFrom+"...HEAD", "--oneline", "--no-merges", "-10")
}
