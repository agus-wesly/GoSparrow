package terminal

import "fmt"


type EraseLineMode int 

const (
    ERASE_LINE_END  EraseLineMode = iota;
    ERASE_LINE_START
    ERASE_LINE_ALL
)

func EraseLine(out FileWriter, mode EraseLineMode) error {
	_, err := fmt.Fprintf(out, "\x1b[%dK", mode)
	return err
}
