package release

import (
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/pkg/console"
)

var android, ios, both bool

var IntegrateCmd = &cobra.Command{
	Use:   "integrate",
	Short: "integrate a release",
	Long:  `Use this command to integrate a release. If the android or ios flags are set, only that platform will be integrated. Otherwise, both will be integrated.`,
	Run: func(cmd *cobra.Command, args []string) {
		version, err := getVersionArg(args)
		console.ExitIfError(err)

		// Integrate GBM into Android and iOS if both flags are set or neither flag is set
		both = !android && !ios || ios && android

		switch {
		case both:
			console.Info("Integrating GBM version %s into both iOS and Android", version)
		case android:
			console.Info("Integrating GBM version %s into Android", version)
		case ios:
			console.Info("Integrating GBM version %s into iOS", version)
		}

	},
}

func init() {
	IntegrateCmd.Flags().BoolVarP(&android, "android", "a", false, "Only integrate Android")
	IntegrateCmd.Flags().BoolVarP(&ios, "ios", "i", false, "Only integrate iOS")

}
