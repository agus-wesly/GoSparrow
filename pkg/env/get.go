// stolen from : https://github.com/joho/godotenv
package env

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"unicode"
)

func Get(key string) (string, error) {
	file, err := os.Open(".env")
	defer file.Close()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return "", err
	}
	bts := buf.Bytes()
	bts = bytes.Replace(bts, []byte("\r\n"), []byte("\n"), -1)
	bts = bytes.TrimLeftFunc(bts, func(r rune) bool {
		switch r {
		case '\t', '\v', '\f', '\r', ' ', 0x85, 0xA0:
			return true
		}
		return false
	})

	ch := bts
	for {
		if ch == nil {
			break
		}
		k, left := extractKey(ch)
		val, left, err := extractVal(left)
		if err != nil {
			return "", err
		}
		if k == key {
			return val, nil
		}
		ch = left
	}
	return "", errors.New("Not found")
}

func extractKey(str []byte) (key string, left []byte) {
	offset := 0
	for i, r := range str {
		rune_r := rune(r)
		if rune_r == '=' || rune_r == ':' {
			key = string(str[0:i])
			offset = i + 1
			break
		}
	}
	key = strings.TrimRightFunc(key, unicode.IsSpace)
	left = str[offset:]
	return key, left
}

func extractVal(str []byte) (value string, left []byte, err error) {
	idx := bytes.IndexFunc(str, func(r rune) bool {
		return r == '\n'
	})
	if idx == -1 {
		return "", nil, errors.New("Not found")
	}
	value = string(str[0:idx])
	left = str[idx+1:]
	return value, left, nil
}
