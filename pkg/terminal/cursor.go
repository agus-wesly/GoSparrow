//go:build !windows
// +build !windows

package terminal

import "fmt"

type Cursor struct {
	In  FileReader
	Out FileWriter
}

func (c *Cursor) Restore() error {
	_, err := fmt.Fprint(c.Out, "\x1b8")
	return err
}

func (c *Cursor) Save() error {
	_, err := fmt.Fprint(c.Out, "\x1b7")
	return err
}

func (c *Cursor) Hide() error {
	_, err := fmt.Fprint(c.Out, "\x1b[?25l")
	return err
}

func (c *Cursor) Show() error {
	_, err := fmt.Fprintf(c.Out, "\x1b[?25h")
	return err
}

func (c *Cursor) HorizontalAbsolute(x int) error {
	_, err := fmt.Fprintf(c.Out, "\x1b[%dG", x)
	return err
}

func (c *Cursor) Up(n int) error {
	_, err := fmt.Fprintf(c.Out, "\x1b[%dA", n)
	return err
}

func (c *Cursor) PreviousLine(n int) error {
	if err := c.Up(1); err != nil {
		return err
	}
	return c.HorizontalAbsolute(0)
}
