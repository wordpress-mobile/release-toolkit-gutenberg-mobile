package release

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
)

var (
	TempDir string
	Verbose bool
	Quite   bool

	// Used by `integrate` and `prepare`
	Ios     bool
	Android bool

	// Used by `integrate`
	Update     bool
	BaseBranch string

	// Used by `prepare`
	Gbm  bool
	Apps bool
)

type releaseResult struct {
	repo string
	pr   repo.PullRequest
	err  error
}

func cleanup() {
	os.RemoveAll(TempDir)
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
		os.Exit(1)
	}()
}

func setTempDir() {
	var err error
	if TempDir, err = os.MkdirTemp("", "gbm-"); err != nil {
		fmt.Println("Error creating temp dir")
		os.Exit(1)
	}
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
}
