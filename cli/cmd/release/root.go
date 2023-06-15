package release

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
)

var (
	TempDir string
	Quite   bool

	// Used by `integrate` and `prepare`
	Ios     bool
	Android bool

	// Used by `integrate`
	Update     bool
	BaseBranch string

	// Used by `prepare`
	Gbm bool
	All bool

	// Used by `publish`
	SkipChecks bool
	Integrate  bool
)

type releaseResult struct {
	repo string
	pr   *repo.PullRequest
	err  error
}

func cleanup() {
	os.RemoveAll(TempDir)
}

func l(f string, a ...interface{}) {
	utils.LogInfo(fmt.Sprintf(f, a...))
}

func lWarn(f string, a ...interface{}) {
	utils.LogWarn(fmt.Sprintf(f, a...))
}

func init() {
	// Make sure we clean up temp files on early exits
	// Use a buffered channel so we don't miss the signal.
	// see https://go.dev/tour/concurrency/5 and https://gobyexample.com/signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(0)
	}()
}

func setTempDir() {
	var err error
	if TempDir, err = os.MkdirTemp("", "gbm-"); err != nil {
		fmt.Println("Error creating temp dir")
		os.Exit(1)
	}
}

func normalizeVersion(version string) string {
	v := version
	if version[0] == 'v' {
		v = version[1:]
	}

	re := regexp.MustCompile(`\d+\.\d+\.\d+`)
	if !re.MatchString(v) {
		fmt.Println("Invalid version")
		os.Exit(1)
	}
	return v
}

// renderCmd represents the render command
var RootCmd = &cobra.Command{
	Use:   "release",
	Short: "release related commands",
	Long: `
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func init() {
	RootCmd.AddCommand(PrepareCmd)
	RootCmd.AddCommand(IntegrateCmd)
	RootCmd.AddCommand(StatusCmd)
	RootCmd.AddCommand(UpdateCmd)
	RootCmd.AddCommand(PublishCmd)
}
