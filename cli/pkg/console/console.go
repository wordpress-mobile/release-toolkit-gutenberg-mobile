package console

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"golang.design/x/clipboard"
)

var (
	l *log.Logger
)

func init() {
	l = log.New(os.Stderr, "", 0)
}

func ExitIfError(err error) {
	if err != nil {
		ExitError(err.Error() + "\n")
	}
}

func ExitError(format string, args ...interface{}) {
	if len(args) == 0 {
		Exit(1, format)
	} else {
		Exit(1, format, args...)
	}
}

func Exit(code int, format string, args ...interface{}) {
	red := color.New(color.FgRed).SprintfFunc()
	l.Printf(red("\n"+format, args...))
	color.Unset()
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
	color.Unset()
}

/*
Use Info to log messages from the scripts. Output is sent to stderr to not muddle up pipe-able output
*/
func Info(format string, args ...interface{}) {
	cyan := color.New(color.FgCyan).SprintfFunc()
	l.Printf(cyan("\n"+format, args...))
	color.Unset()
}

func Log(format string, args ...interface{}) {
	l.Printf(format+"\n", args...)
	color.Unset()
}

func Debug(format string, args ...interface{}) {
	blue := color.New(color.FgBlue).SprintfFunc()
	l.Printf(blue("\n"+format, args...))
	color.Unset()
}

func Warn(format string, args ...interface{}) {
	yellow := color.New(color.FgYellow).SprintfFunc()
	l.Printf(yellow("\n"+format, args...))
	color.Unset()
}

func Confirm(ask string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		l.Printf("%s [y/n]: ", ask)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
