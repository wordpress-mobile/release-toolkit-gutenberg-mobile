package prepare

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

var PrepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "Prepare for a release",
}

func Execute() {
	err := PrepareCmd.Execute()
	console.ExitIfError(err)
}

func init() {
	PrepareCmd.AddCommand(gbmCmd)
	PrepareCmd.AddCommand(gbCmd)
}

func getVersionArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("missing version")
	}

	return utils.NormalizeVersion(args[0])
}
