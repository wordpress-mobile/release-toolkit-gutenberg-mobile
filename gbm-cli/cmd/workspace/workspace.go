package workspace

import (
	"os"
	"os/signal"
	"path"

	"github.com/wordpress-mobile/release-toolkit-gutenberg-mobile/gbm-cli/pkg/console"
)

type Workspace interface {
	Cleanup()
	Dir() string
	Keep()
	create() error
	setCleaner()
}

type workspace struct {
	dir      string
	keep     bool
	prefix   string
	cleaner  func()
	disabled bool
}

func NewWorkspace() (Workspace, error) {
	w := &workspace{prefix: "gbm-"}
	if _, noWorkspace := os.LookupEnv("GBM_NO_WORKSPACE"); noWorkspace {
		console.Info("GBM_NO_WORKSPACE is set, not creating a workspace directory")
		w.disabled = true
	}

	if _, ci := os.LookupEnv("CI"); ci {
		console.Info("CI environment detected, not creating a workspace directory")
		w.disabled = true
		w.dir = "."
	}

	if err := w.create(); err != nil {
		return nil, err
	}
	w.setCleaner()
	return w, nil
}

func (w *workspace) create() error {
	// if we're disabled, don't create a temp directory
	if w.disabled {
		return nil
	}
	tempDir, err := os.MkdirTemp("", w.prefix)
	if err != nil {
		return err
	}
	w.dir = tempDir
	return nil
}

func (w *workspace) setCleaner() {
	w.cleaner = func() {
		if w.disabled || w.dir == "" {
			return
		}

		if w.keep {
			console.Info("Keeping temporary directory %s", w.dir)
			return
		}

		if err := os.RemoveAll(w.dir); err != nil {
			console.Error(err)
		}
	}
	// register a listener for ^C, call the cleanup function
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan // wait for ^C
		w.cleaner()
		os.Exit(1)
	}()
}

func (w *workspace) Dir() string {
	// if the workspace is disabled, return the current directory
	if w.disabled {
		return path.Base(".")
	}
	return w.dir
}

func (w *workspace) Keep() {
	w.keep = true
}

func (w *workspace) Cleanup() {
	if w.keep {
		console.Info("Keeping temporary directory %s", w.dir)
	} else {
		console.Info("Cleaning up workspace directory %s", w.dir)
	}
	w.cleaner()
}
