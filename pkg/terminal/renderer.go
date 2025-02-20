package terminal

import (
	"bytes"
	"fmt"
	"os"

	"golang.org/x/term"
)

type Renderer struct {
	stdio          Stdio
	renderedErrors bytes.Buffer
	renderedText   bytes.Buffer
}

type ErrorTemplateData struct {
	Error error
	Icon  Icon
}
var ErrorTemplate = `{{color .Icon.Format }}{{ .Icon.Text }} Sorry, your reply was invalid: {{ .Error.Error }}{{color "reset"}}
`

func (r *Renderer) WithStd(stdio Stdio) {
    r.stdio = stdio
}

func (r *Renderer) NewCursor() *Cursor {
	return &Cursor{
		In:  r.stdio.In,
		Out: r.stdio.Out,
	}
}

func (r *Renderer) RenderWithCursorOffset(tmpl string, data IterableOpts, opts []OptionAnswer, idx int) error {
	cursor := r.NewCursor()
	cursor.Restore()
	if err := r.Render(tmpl, data); err != nil {
        return err
	}
    cursor.Save()
	return nil
}

func (r *Renderer) Render(tmpl string, data interface{}) error {
    lineCount := r.countLines(r.renderedText)
    r.resetPrompt(lineCount)
    r.renderedText.Reset()

    userOut, layoutOut, err := RunTemplate(tmpl, data)
    if err != nil {
        return err
    }
    if _, err = fmt.Fprint(r.stdio.Out, userOut); err != nil {
        return err
    }
    r.AppendRenderedText(layoutOut)
    return nil
}

func (r *Renderer) AppendRenderedText(layoutOut string) {
    r.renderedText.WriteString(layoutOut)
}

func (r *Renderer) appendRenderedError(text string) {
	r.renderedErrors.WriteString(text)
}

func (r *Renderer) countLines(buf bytes.Buffer) int {
	w := r.termWidthSafe()

	bufBytes := buf.Bytes()

	count := 0
	curr := 0
	for curr < len(bufBytes) {
		var delim int
		// read until the next newline or the end of the string
		relDelim := bytes.IndexRune(bufBytes[curr:], '\n')
		if relDelim != -1 {
			count += 1 // new line found, add it to the count
			delim = curr + relDelim
		} else {
			delim = len(bufBytes) // no new line found, read rest of text
		}

		str := string(bufBytes[curr:delim])
		if lineWidth := StringWidth(str); lineWidth > w {
			// account for word wrapping
			count += lineWidth / w
			if (lineWidth % w) == 0 {
				// content whose width is exactly a multiplier of available width should not
				// count as having wrapped on the last line
				count -= 1
			}
		}
		curr = delim + 1
	}
	return count
}

func (r *Renderer) resetPrompt(lines int) {
	// clean out current line in case tmpl didnt end in newline
	cursor := r.NewCursor()
	cursor.HorizontalAbsolute(0)
	EraseLine(r.stdio.Out, ERASE_LINE_ALL)
	// clean up what we left behind last time
	for i := 0; i < lines; i++ {
		cursor.PreviousLine(1)
		EraseLine(r.stdio.Out, ERASE_LINE_ALL)
	}
}

func (r *Renderer) termWidthSafe() int {
	w, err := r.termWidth()
	if err != nil || w == 0 {
		// if we got an error due to GetSize not being supported
		// on current platform then just assume a very wide terminal
		w = 10000
	}
	return w
}

func (s *Renderer) NewRuneReader() *RuneReader {
	return NewRuneReader(Stdio{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	})
}

func (r *Renderer) termWidth() (int, error) {
	fd := int(r.stdio.Out.Fd())
	termWidth, _, err := term.GetSize(fd)
	return termWidth, err
}

func (r *Renderer) Error(config *PromptConfig, invalid error) error {
	// cleanup the currently rendered errors
	r.resetPrompt(r.countLines(r.renderedErrors))
	r.renderedErrors.Reset()

	// cleanup the rest of the prompt
	r.resetPrompt(r.countLines(r.renderedText))
	r.renderedText.Reset()

	userOut, layoutOut, err := RunTemplate(ErrorTemplate, &ErrorTemplateData{
		Error: invalid,
		Icon:  config.Icons.Error,
	})
	if err != nil {
		return err
	}

	// send the message to the user
	if _, err := fmt.Fprint(r.stdio.Out, userOut); err != nil {
		return err
	}

	// add the printed text to the rendered error buffer so we can cleanup later
	r.appendRenderedError(layoutOut)

	return nil
}
