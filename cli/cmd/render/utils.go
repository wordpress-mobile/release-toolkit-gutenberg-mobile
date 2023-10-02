package render

import (
	"fmt"

	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

func renderAztecSteps(conditional bool) (string, error) {
	return render.RenderTasks(render.Template{
		Path: "templates/checklist/aztec.html",
		Json: fmt.Sprintf(`{"conditional": %v}`, conditional),
	})
}
