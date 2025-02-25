package terminal

import (
	"unicode"
	"golang.org/x/text/width"
)

type RuneReader struct {
	stdio Stdio
	state runeReaderState
}

func StringWidth(str string) int {
	w := 0
	ansi := false

	for _, r := range str {
		// increase width only when outside of ANSI escape sequences
		if ansi || isAnsiMarker(r) {
			ansi = !isAnsiTerminator(r)
		} else {
			w += runeWidth(r)
		}
	}
	return w
}

func isAnsiMarker(r rune) bool {
	return r == '\x1B'
}

func isAnsiTerminator(r rune) bool {
	return (r >= 0x40 && r <= 0x5a) || (r == 0x5e) || (r >= 0x60 && r <= 0x7e)
}

func runeWidth(r rune) int {
	switch width.LookupRune(r).Kind() {
	case width.EastAsianWide, width.EastAsianFullwidth:
		return 2
	}

	if !unicode.IsPrint(r) {
		return 0
	}
	return 1
}

func NewRuneReader(stdio Stdio) *RuneReader {
	return &RuneReader{
		stdio: stdio,
		state: newRuneReaderState(stdio.In),
	}
}
