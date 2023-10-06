package release

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/gbm"
)

var ReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release Gutenberg Mobile",
}

func Execute() {
	err := ReleaseCmd.Execute()
	console.ExitIfError(err)
}

func init() {
	ReleaseCmd.AddCommand(PrepareCmd)
}

func getVersionArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("missing version")
	}

	return gbm.NormalizeVersion(args[0])
}
