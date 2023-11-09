package console

import (
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
	Highlight  *color.Color
)

func init() {
	l = log.New(os.Stderr, "", 0)
	Heading = color.New(color.FgWhite, color.Bold)
	HeadingRow = color.New(color.FgGreen, color.Bold)
	Row = color.New(color.FgGreen)
	Highlight = color.New(color.FgHiWhite)
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
	l.Printf(cyan("[INFO] "+format, args...))
	color.Unset()
}

func Log(format string, args ...interface{}) {
	l.Printf(format+"\n", args...)
	color.Unset()
}

func Debug(format string, args ...interface{}) {
	blue := color.New(color.FgHiBlue).SprintfFunc()
	l.Printf(blue("[DEBUG] "+format, args...))
	color.Unset()
}

func Print(c *color.Color, format string, args ...interface{}) {
	styled := c.SprintfFunc()
	l.Printf(styled(format, args...))
	color.Unset()
}

func Warn(format string, args ...interface{}) {
	yellow := color.New(color.FgYellow).SprintfFunc()
	l.Printf(yellow("[WARN] "+format, args...))
	color.Unset()
}

func Inspect(i interface{}) {
	spew.Dump(i)
}

func Error(err error) {
	red := color.New(color.FgRed).SprintfFunc()
	l.Printf(red("[ERROR] " + err.Error() + "\n"))
	color.Unset()
}

func Confirm(ask string) bool {
	// If not in a tty (CI) return true
	if os.Getenv("CI") == "true" {
		return true
	}
	var response string
	fmt.Print(Highlight.Sprintf("%s [y/n]: ", ask))

	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	response = strings.ToLower(strings.TrimSpace(response))

	if response == "y" || response == "yes" {
		return true
	}
	return false
}

func Ask(ask string) string {
	var response string
	fmt.Print(Highlight.Sprintf("%s: ", ask))

	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	return strings.TrimSpace(response)
}
