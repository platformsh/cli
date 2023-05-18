package md

import (
	"fmt"
	"strings"
)

type HeadingLevel int

const (
	L1 HeadingLevel = iota + 1
	L2
	L3
	L4
	L5
	L6
)

func Heading(level HeadingLevel, s string) string {
	if s == "" {
		return ""
	}
	return strings.Repeat("#", int(level)) + " " + s
}

func Bold(s string) string {
	if s == "" {
		return ""
	}
	return "**" + s + "**"
}

func Italic(s string) string {
	if s == "" {
		return ""
	}
	return "*" + s + "*"
}

func UnorderedListItem(s string) string {
	if s == "" {
		return ""
	}
	return "* " + s
}

func Code(s string) string {
	if s == "" {
		return ""
	}
	return "`" + s + "`"
}

func CodeBlock(s string) string {
	if s == "" {
		return ""
	}
	return "```\n" + s + "\n```"
}

func Link(text, url string) string {
	if text == "" || url == "" {
		return ""
	}
	return fmt.Sprintf("[%s](%s)", text, url)
}

func Anchor(s string) string {
	if s == "" {
		return ""
	}
	return "#" + strings.ReplaceAll(s, ":", "")
}
