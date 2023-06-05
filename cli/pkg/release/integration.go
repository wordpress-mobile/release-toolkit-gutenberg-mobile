package release

import (
	"fmt"

	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

func CreateAndroidScheduledPr(version string) error {
	return createAndroidPr(version, "trunk")
}

func CreateAndroidPatchPr(version, baseBranch string) error {
	return createAndroidPr(version, baseBranch)
}

func CreateIosScheduledPr(version string) error {
	return createIosPr(version, "trunk")
}

func CreateIosPatchPr(version, baseBranch string) error {
	return createIosPr(version, baseBranch)
}

type renderFunc func(gbmPR repo.PullRequest) string

func genBodyRenderer(version, templatePath string) renderFunc {
	return func(gbmPR repo.PullRequest) string {
		data := struct {
			Version  string
			GbmPrUrl string
		}{
			Version:  version,
			GbmPrUrl: gbmPR.Url,
		}

		body, err := render.Render(templatePath, data, nil)
		fmt.Println(err)
		return body
	}
}

func createAndroidPr(version, baseBranch string) error {
	gbmPR, err := GetGbmReleasePr(version)
	if err != nil {
		return err
	}

	renderTitle := func(gbmPR repo.PullRequest) string {
		return fmt.Sprintf("Integrate Gutenberg Mobile %s", version)
	}

	apr := gbm.AndroidInPr{
		BaseBranch:  baseBranch,
		HeadBranch:  fmt.Sprintf("gutenberg/integrate_release_%s", version),
		RenderTitle: renderTitle,
		RenderBody:  genBodyRenderer(version, "templates/release/integrationPrBody.md"),
		Tags:        []string{"Gutenberg"},
	}

	return gbm.CreateIntegrationPr(apr, gbmPR)
}

func createIosPr(version, baseBranch string) error {
	gbmPR, err := GetGbmReleasePr(version)
	if err != nil {
		return err
	}

	renderTitle := func(gbmPR repo.PullRequest) string {
		return fmt.Sprintf("Integrate Gutenberg Mobile %s", version)
	}

	ipr := gbm.IosInPr{
		BaseBranch:  baseBranch,
		HeadBranch:  fmt.Sprintf("gutenberg/integrate_release_%s", version),
		RenderTitle: renderTitle,
		RenderBody:  genBodyRenderer(version, "templates/release/integrationPrBody.md"),
		Tags:        []string{"Gutenberg"},
	}

	return gbm.CreateIntegrationPr(ipr, gbmPR)
}
