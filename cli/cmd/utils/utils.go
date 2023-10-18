package utils

import (
	"fmt"
	"os"

	"github.com/wordpress-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/gbm-cli/pkg/utils"
)

func GetVersionArg(args []string) (string, error) {
	if len(args) == 0 {
		return "", fmt.Errorf("missing version")
	}
	if !utils.ValidateVersion(args[0]) {
		return "", fmt.Errorf("invalid version %s.  Versions must have a `Major.Minor.Patch` form", args[0])
	}
	return utils.NormalizeVersion(args[0])
}

func ExitIfError(err error, code int) {
	if err != nil {
		console.Error(err)
		Exit(code)
	}
}

func Exit(code int, deferred ...func()) {
	os.Exit(func() int {
		for _, d := range deferred {
			d()
		}
		return code
	}())
}
