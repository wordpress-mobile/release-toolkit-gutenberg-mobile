package repo

import (
	"bytes"
	"encoding/json"
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

type PrUpdate struct {
	Title string
	Body  string
	State string
	Base  string
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

// GetPr returns a PullRequest struct for the given repo and PR number.
func GetPr(repo string, id int) (PullRequest, error) {
	org, err := getOrg(repo)
	if err != nil {
		return PullRequest{}, err
	}
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

func CreatePr(repo string, pr *PullRequest) error {
	client := getClient()
	org, err := getOrg(repo)
	if err != nil {
		return err
	}

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
	if pr.Labels != nil {
		if err := AddLabels(repo, pr); err != nil {
			return err
		}
	}
	return nil
}

func UpdatePr(repo string, pr *PullRequest, update PrUpdate) error {
	org, err := getOrg(repo)
	if err != nil {
		return err
	}
	client := getClient()

	endpoint := fmt.Sprintf("repos/%s/%s/pulls/%d", org, repo, pr.Number)

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

func labelRequest(repo string, prNum int, labels []string) ([]struct{ Name string }, error) {
	org, err := getOrg(repo)
	if err != nil {
		return nil, err
	}

	client := getClient()

	endpoint := fmt.Sprintf("repos/%s/%s/issues/%d/labels", org, repo, prNum)

	pbody := struct{ Labels []string }{Labels: labels}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(pbody); err != nil {
		return nil, err
	}

	resp := []struct{ Name string }{}

	if err := client.Post(endpoint, &buf, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}
