package render

import (
	"fmt"
	"time"

	"github.com/wordpress-mobile/gbm-cli/pkg/render"
)

func renderAztecSteps(conditional bool) (string, error) {
	return render.RenderTasks(render.Template{
		Path: "templates/checklist/aztec.html",
		Json: fmt.Sprintf(`{"conditional": %v}`, conditional),
	})
}

func nextReleaseDate() string {
	weekday := time.Now().Weekday()
	daysUntilThursday := 4 - weekday

	nextThursday := time.Now().AddDate(0, 0, int(daysUntilThursday))

	return nextThursday.Format("Monday January 2, 2006")
}
