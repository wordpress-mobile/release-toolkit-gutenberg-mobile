package release

import (
	"fmt"

	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

type renderFunc func(repo.PullRequest) string

func GenBodyRenderer(version, templatePath string) renderFunc {
	return func(gbmPR repo.PullRequest) string {
		data := struct {
			Version  string
			GbmPrUrl string
		}{
			Version:  version,
			GbmPrUrl: gbmPR.Url,
		}

		body, err := render.Render(templatePath, data, nil)
		if err != nil {
			fmt.Println(err)
		}
		return body
	}
}
