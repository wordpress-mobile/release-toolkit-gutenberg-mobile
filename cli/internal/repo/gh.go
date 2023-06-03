package repo

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

func GetPr(repo string, prNumber int) (PullRequest, error) {

	return PullRequest{}, nil
}
