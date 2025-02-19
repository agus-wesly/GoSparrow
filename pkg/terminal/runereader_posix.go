package terminal

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"syscall"
	"unsafe"
)

type BufferedReader struct {
	In     io.Reader
	Buffer *bytes.Buffer
}

type FileWriter interface {
	io.Writer
	Fd() uintptr
}

// FileReader provides a minimal interface for Stdout.
type FileReader interface {
	io.Reader
	Fd() uintptr
}

type Stdio struct {
	In  FileReader
	Out FileWriter
	Err io.Writer
}

type RuneReader struct {
	stdio Stdio
	state runeReaderState
}

type runeReaderState struct {
	term   syscall.Termios
	reader *bufio.Reader
	buf    *bytes.Buffer
}

func NewRuneReader(stdio Stdio) *RuneReader {
	return &RuneReader{
		stdio: stdio,
		state: newRuneReaderState(stdio.In),
	}
}

func newRuneReaderState(input FileReader) runeReaderState {
	buf := new(bytes.Buffer)
	return runeReaderState{
		reader: bufio.NewReader(&BufferedReader{
			In:     input,
			Buffer: buf,
		}),
		buf: buf,
	}
}

func (br *BufferedReader) Read(p []byte) (int, error) {
	n, err := br.Buffer.Read(p)
	if err != nil && err != io.EOF {
		return n, err
	} else if err == nil {
		return n, nil
	}
	return br.In.Read(p[n:])
}

func (rr *RuneReader) Buffer() *bytes.Buffer {
	return rr.state.buf
}

func (rr *RuneReader) ReadRune() (rune, int, error) {
	r, size, err := rr.state.reader.ReadRune()
	if err != nil {
		return r, size, err
	}

	if r != KeyEscape {
		return r, size, err
	}

	if rr.state.reader.Buffered() == 0 {
		// no more characters so must be `Esc` key
		return KeyEscape, 1, nil
	}

	r, size, err = rr.state.reader.ReadRune()
	if err != nil {
		return r, size, err
	}

	// ESC O ... or ESC [ ...?
	if r != normalKeypad && r != applicationKeypad {
		return r, size, fmt.Errorf("unexpected escape sequence from terminal: %q", []rune{KeyEscape, r})
	}

	keypad := r

	r, size, err = rr.state.reader.ReadRune()
	if err != nil {
		return r, size, err
	}

	switch r {
	case 'A': // ESC [ A or ESC O A
		return KeyArrowUp, 1, nil
	case 'B': // ESC [ B or ESC O B
		return KeyArrowDown, 1, nil
	case 'C': // ESC [ C or ESC O C
		return KeyArrowRight, 1, nil
	case 'D': // ESC [ D or ESC O D
		return KeyArrowLeft, 1, nil
	case 'F': // ESC [ F or ESC O F
		return SpecialKeyEnd, 1, nil
	case 'H': // ESC [ H or ESC O H
		return SpecialKeyHome, 1, nil
	case '3': // ESC [ 3
		if keypad == normalKeypad {
			// discard the following '~' key from buffer
			_, _ = rr.state.reader.Discard(1)
			return SpecialKeyDelete, 1, nil
		}
	}

	// discard the following '~' key from buffer
	_, _ = rr.state.reader.Discard(1)
	return IgnoreKey, 1, nil
}

func (rr *RuneReader) SetTermMode() error {
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(rr.stdio.In.Fd()), ioctlReadTermios, uintptr(unsafe.Pointer(&rr.state.term)), 0, 0, 0); err != 0 {
		return err
	}

	newState := rr.state.term
	newState.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG
	// Because we are clearing canonical mode, we need to ensure VMIN & VTIME are
	// set to the values we expect. This combination puts things in standard
	// "blocking read" mode (see termios(3)).
	newState.Cc[syscall.VMIN] = 1
	newState.Cc[syscall.VTIME] = 0

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(rr.stdio.In.Fd()), ioctlWriteTermios, uintptr(unsafe.Pointer(&newState)), 0, 0, 0); err != 0 {
		return err
	}

	return nil
}


func (rr *RuneReader) RestoreTermMode() error {
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(rr.stdio.In.Fd()), ioctlWriteTermios, uintptr(unsafe.Pointer(&rr.state.term)), 0, 0, 0); err != 0 {
		return err
	}
	return nil
}
