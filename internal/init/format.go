package init

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/platformsh/cli/internal/init/streaming"
	"github.com/platformsh/cli/internal/tui"
)

var (
	defaultMessageColor = "default"
	defaultColorFunc    = fmt.Sprintf
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

func defaultSpinner(w io.Writer) *tui.Spinner {
	return tui.NewDefault(w)
}

func printWithSpinner(spinr *tui.Spinner, colorName, format string, args ...any) {
	spinr.Suffix = " " + colorFunc(colorName)(format, args...)
	spinr.Start()
	time.Sleep(time.Millisecond * 300)
}

type logPrinter struct {
	spinr  *tui.Spinner
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
