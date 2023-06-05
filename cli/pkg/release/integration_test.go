package release

import (
	"embed"
	"fmt"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

//go:embed testdata/*
var templatesFS embed.FS

func init() {
	render.TemplateFS = templatesFS
}

func TestGenBodyRenderer(t *testing.T) {

	t.Run("Generates a function to render a PR body", func(t *testing.T) {
		version := "1.2.3"
		renderBody := genBodyRenderer(version, "testdata/integrationPrBody.md")

		gbmPr := repo.PullRequest{
			Url: "https://wordpress.com",
		}

		want := `## Description
This PR incorporates the 1.2.3 release of gutenberg-mobile.
For more information about this release and testing instructions, please see the related Gutenberg-Mobile PR: https://wordpress.com
`
		got := renderBody(gbmPr)

		fmt.Println(got)

		if got != want {
			t.Fatalf("strings are not equal: %s", diff.LineDiff(got, want))
		}
	})
}
