package utils

import (
	"fmt"
	"os"

	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/console"
	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/semver"
)

func GetVersionArg(args []string) (semver.SemVer, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("missing version")
	}
	version, err := semver.NewSemVer(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid version %s.  Versions must have a `Major.Minor.Patch` form", args[0])
	}

	return version, nil
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
