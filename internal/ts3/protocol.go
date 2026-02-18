package ts3

import (
	"strings"
)

var escaper = strings.NewReplacer(
	"\\", "\\\\",
	"/", "\\/",
	" ", "\\s",
	"|", "\\p",
	"\a", "\\a",
	"\b", "\\b",
	"\f", "\\f",
	"\n", "\\n",
	"\r", "\\r",
	"\t", "\\t",
	"\v", "\\v",
)

var unescaper = strings.NewReplacer(
	"\\\\", "\\",
	"\\/", "/",
	"\\s", " ",
	"\\p", "|",
	"\\a", "\a",
	"\\b", "\b",
	"\\f", "\f",
	"\\n", "\n",
	"\\r", "\r",
	"\\t", "\t",
	"\\v", "\v",
)

// Escape escapes special characters in a string for TS3 ServerQuery.
func Escape(s string) string {
	return escaper.Replace(s)
}

// Unescape unescapes special characters from a TS3 ServerQuery string.
func Unescape(s string) string {
	return unescaper.Replace(s)
}
