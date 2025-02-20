package terminal

import (
	"errors"
	"strconv"
)

func IsNumber(arg interface{}) error {
	_, err := strconv.Atoi(arg.(string))
	if err != nil {
		return errors.New("Value is not a number")
	}
	return nil
}

func Required(arg interface{}) error {
    if arg == nil || arg == "" {
        return errors.New("Value is required")
    }
    return nil
}
