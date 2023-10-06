package gh

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
