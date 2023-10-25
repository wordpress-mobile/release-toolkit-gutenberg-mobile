package console

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/fatih/color"
	"golang.design/x/clipboard"
)

var (
	l          *log.Logger
	Heading    *color.Color
	HeadingRow *color.Color
	Row        *color.Color
)

func init() {
	l = log.New(os.Stderr, "", 0)
	Heading = color.New(color.FgWhite, color.Bold)
	HeadingRow = color.New(color.FgGreen, color.Bold)
	Row = color.New(color.FgGreen)
}

// Deprecated
func ExitIfError(err error) {
	if err != nil {
		ExitError(err.Error() + "\n")
	}
}

// Deprecated
func ExitError(format string, args ...interface{}) {
	if len(args) == 0 {
		Exit(1, format)
	} else {
		Exit(1, format, args...)
	}
}

// Deprecated
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
	l.Printf(cyan("\n[INFO] "+format, args...))
	color.Unset()
}

func Log(format string, args ...interface{}) {
	l.Printf(format+"\n", args...)
	color.Unset()
}

func Debug(format string, args ...interface{}) {
	blue := color.New(color.FgHiBlue).SprintfFunc()
	l.Printf(blue("\n[DEBUG] "+format, args...))
	color.Unset()
}

func Print(c *color.Color, format string, args ...interface{}) {
	styled := c.SprintfFunc()
	l.Printf(styled(format, args...))
	color.Unset()
}

func Warn(format string, args ...interface{}) {
	yellow := color.New(color.FgYellow).SprintfFunc()
	l.Printf(yellow("\n[WARN] "+format, args...))
	color.Unset()
}

func Inspect(i interface{}) {
	spew.Dump(i)
}

func Error(err error) {
	red := color.New(color.FgRed).SprintfFunc()
	l.Printf(red("\n[ERROR] " + err.Error()))
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
