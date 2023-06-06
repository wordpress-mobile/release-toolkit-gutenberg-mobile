package release

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/wordpress-mobile/gbm-cli/internal/repo"
	"github.com/wordpress-mobile/gbm-cli/internal/utils"
	"github.com/wordpress-mobile/gbm-cli/pkg/release"
)

var (
	Ios        bool
	Android    bool
	Update     bool
	BaseBranch string
	Verbose    bool
	tempDir    string
)

func cleanup() {
	os.RemoveAll(tempDir)
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
	if tempDir, err = ioutil.TempDir("", "gbm-"); err != nil {
		fmt.Println("Error creating temp dir")
		os.Exit(1)
	}
}

// checklistCmd represents the checklist command
var IntegrateCmd = &cobra.Command{
	Use:   "integrate <version>",
	Short: "generate the release integration Pr",
	Long: `
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := args[0]
		gbmPr, err := utils.GetGbmReleasePr(version)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		type result struct {
			repo string
			pr   repo.PullRequest
			err  error
		}
		rChan := make(chan result)

		s := spinner.New(spinner.CharSets[23], 200*time.Millisecond)

		setTempDir()
		defer cleanup()

		// if neither ios or android are specified, default to both
		if !Ios && !Android {
			Ios = true
			Android = true
		} else {
			// if we are only doing one, set verbose to true
			Verbose = true
		}
		numPr := 0 // number of PRs to create

		// Use goroutines to create the PRs concurrently
		if Ios {
			numPr++
			utils.LogInfo("Creating iOS PR at %s/Wordpress-iOS", repo.WpMobileOrg)
			go func() {
				pr, err := release.CreateIosPr(version, BaseBranch, tempDir, gbmPr, Verbose)
				rChan <- result{"WordPress-iOS", pr, err}
			}()
		}

		if Android {
			numPr++
			utils.LogInfo("Creating Android PR at %s/WordPress-Android", repo.WpMobileOrg)
			go func() {
				pr, err := release.CreateAndroidPr(version, BaseBranch, tempDir, gbmPr, Verbose)
				rChan <- result{"WordPress-Android", pr, err}
			}()
		}

		if !Verbose {
			s.Start()
			defer s.Stop()
		}

		success := true
		for i := 0; i < numPr; i++ {
			result := <-rChan
			if result.err != nil {
				if repo.IsExistingBranchError(result.err) {
					utils.LogWarn("%s : Release branch already exists, try updating", result.repo)
				} else {
					utils.LogError("%v", result.err)
				}
			}

			if result.pr.Number == 0 {
				// There might be an error but let's consider
				// creating the pr is a success
				// TODO: Consider an existing branch as a success ?
				success = false
			} else {
				utils.LogInfo("PR created: %s", result.pr.Url)
			}
		}

		s.Stop()
		if success {
			utils.LogInfo("PRs created successfully")
		} else {
			utils.LogError("Some PRs failed to create")
		}

	},
}

func init() {
	IntegrateCmd.Flags().BoolVarP(&Ios, "ios", "i", false, "target ios release")
	IntegrateCmd.Flags().BoolVarP(&Android, "android", "a", false, "target android release")
	IntegrateCmd.Flags().BoolVarP(&Update, "update", "u", false, "update existing PR")
	IntegrateCmd.Flags().StringVarP(&BaseBranch, "base-branch", "b", "trunk", "base branch for the PR")
	IntegrateCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
}
