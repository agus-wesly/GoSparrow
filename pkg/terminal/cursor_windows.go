package terminal

import (
	"bytes"
	"syscall"
	"unsafe"
)

var cursorLoc Coord

type Cursor struct {
	In  FileReader
	Out FileWriter
}

func (c *Cursor) Restore() error {
	handle := syscall.Handle(c.Out.Fd())
	// restore it to the original position
	_, _, err := procSetConsoleCursorPosition.Call(uintptr(handle), uintptr(*(*int32)(unsafe.Pointer(&cursorLoc))))
	return normalizeError(err)
}

func (c *Cursor) Save() error {
	loc, err := c.Location(nil)
	if err != nil {
		return err
	}
	cursorLoc = *loc
	return nil
}

func (c *Cursor) Hide() error {
	handle := syscall.Handle(c.Out.Fd())

	var cci consoleCursorInfo
	if _, _, err := procGetConsoleCursorInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&cci))); normalizeError(err) != nil {
		return err
	}
	cci.visible = 0

	_, _, err := procSetConsoleCursorInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&cci)))
	return normalizeError(err)
}

func (c *Cursor) Show() error {
	handle := syscall.Handle(c.Out.Fd())

	var cci consoleCursorInfo
	if _, _, err := procGetConsoleCursorInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&cci))); normalizeError(err) != nil {
		return err
	}
	cci.visible = 1

	_, _, err := procSetConsoleCursorInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&cci)))
	return normalizeError(err)
}

func (c *Cursor) HorizontalAbsolute(x int) error {
	handle := syscall.Handle(c.Out.Fd())

	var csbi consoleScreenBufferInfo
	if _, _, err := procGetConsoleScreenBufferInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&csbi))); normalizeError(err) != nil {
		return err
	}

	var cursor Coord
	cursor.X = Short(x)
	cursor.Y = csbi.cursorPosition.Y

	if csbi.size.X < cursor.X {
		cursor.X = csbi.size.X
	}

	_, _, err := procSetConsoleCursorPosition.Call(uintptr(handle), uintptr(*(*int32)(unsafe.Pointer(&cursor))))
	return normalizeError(err)
}

func (c *Cursor) Up(n int) error {
	return c.cursorMove(0, n)
}

func (c *Cursor) Down(n int) error {
	return c.cursorMove(0, -1*n)
}

func (c *Cursor) PreviousLine(n int) error {
	if err := c.Down(n); err != nil {
		return err
	}
	return c.HorizontalAbsolute(0)
}

func (c *Cursor) Location(buf *bytes.Buffer) (*Coord, error) {
	handle := syscall.Handle(c.Out.Fd())

	var csbi consoleScreenBufferInfo
	if _, _, err := procGetConsoleScreenBufferInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&csbi))); normalizeError(err) != nil {
		return nil, err
	}

	return &csbi.cursorPosition, nil
}

func (c *Cursor) cursorMove(x int, y int) error {
	handle := syscall.Handle(c.Out.Fd())

	var csbi consoleScreenBufferInfo
	if _, _, err := procGetConsoleScreenBufferInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&csbi))); normalizeError(err) != nil {
		return err
	}

	var cursor Coord
	cursor.X = csbi.cursorPosition.X + Short(x)
	cursor.Y = csbi.cursorPosition.Y + Short(y)

	_, _, err := procSetConsoleCursorPosition.Call(uintptr(handle), uintptr(*(*int32)(unsafe.Pointer(&cursor))))
	return normalizeError(err)
}

func normalizeError(err error) error {
	if syserr, ok := err.(syscall.Errno); ok && syserr == 0 {
		return nil
	}
	return err
}
