package console

import (
	"fmt"
	"os"

	"golang.design/x/clipboard"
)

func ExitIfError(err error) {
	if err != nil {
		ExitError(1, err.Error()+"\n")
	}
}

func ExitError(code int, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func Clipboard(m string) {
	clipboard.Write(clipboard.FmtText, []byte(m))
}

/*
Use Out for printing resulting messages that should be piped. For status logging use console.Info
*/
func Out(m string) {
	fmt.Fprintln(os.Stdout, m)
}

/*
Use Info to log messages from the scripts. Output is sent to stderr to not muddle up pipe-able output
*/
func Info(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
