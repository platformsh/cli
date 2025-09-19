package init

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"

	"github.com/platformsh/cli/internal/init/streaming"
)

var (
	defaultMessageColor = "green"
	defaultColorFunc    = color.GreenString
	levelColors         = map[string]string{
		streaming.LogLevelDebug: "cyan",
		streaming.LogLevelInfo:  defaultMessageColor,
		streaming.LogLevelWarn:  "yellow",
		streaming.LogLevelError: "red",
	}
	colorFuncs = map[string]func(string, ...any) string{
		"blue":   color.BlueString,
		"cyan":   color.CyanString,
		"green":  color.GreenString,
		"red":    color.RedString,
		"white":  color.WhiteString,
		"yellow": color.YellowString,
	}
)

func levelColor(level string) string {
	if c, ok := levelColors[level]; ok {
		return c
	}
	return defaultMessageColor
}

func colorFunc(name string) func(string, ...any) string {
	if fn, ok := colorFuncs[name]; ok {
		return fn
	}
	return defaultColorFunc
}

func defaultSpinner(w io.Writer) *spinner.Spinner {
	return spinner.New(spinner.CharSets[23], 80*time.Millisecond, spinner.WithWriter(w))
}

func printWithSpinner(spinr *spinner.Spinner, colorName, format string, args ...any) {
	_ = spinr.Color(colorName)
	spinr.Suffix = " " + colorFunc(colorName)(format, args...)
	spinr.Start()
	time.Sleep(time.Millisecond * 300)
}

type logPrinter struct {
	spinr  *spinner.Spinner
	stderr io.Writer
}

// print a log message based on a streaming.Level constant, and tags such as "spin".
func (p *logPrinter) print(level, msg string, tags ...string) {
	lc := levelColor(level)

	if slices.Contains(tags, "spin") {
		printWithSpinner(p.spinr, lc, msg)
		return
	}

	formatter := colorFunc(lc)
	fmt.Fprint(p.stderr, formatter(strings.TrimRight(msg, "\n")+"\n"))
}
