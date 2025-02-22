package terminal

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

type Log struct {
	Cursor
	config       *LogConfig
	baseTemplate *template.Template
}

type LogConfig struct {
	Error   Icon
	Info    Icon
	Success Icon
}

type LogData struct {
	Text string
	Icon Icon
}

var LogTemplate = `{{color .Icon.Format}}{{.Icon.Text}} {{color "default+hb"}}{{.Text}}{{"\n"}}`

func createConfig() *LogConfig {
	return &LogConfig{
		Error: Icon{
			Text:   "X",
			Format: "red+hb",
		},
		Info: Icon{
			Text:   "!",
			Format: "yellow+hb",
		},
		Success: Icon{
			Text:   "✔",
			Format: "green+hb",
		},
	}
}

func (l *Log) NewCursor() {
	c := Cursor{In: os.Stdin, Out: os.Stdout}
	l.Cursor = c
	config := createConfig()
	l.config = config
	templateWithColor, err := template.New("Log").Funcs(TemplateFuncsWithColor).Parse(LogTemplate)
	if err != nil {
		panic(err)
	}
	l.baseTemplate = templateWithColor
}

func (l *Log) Info(msg ...any) error {
	userBuf := bytes.NewBufferString("")
	l.baseTemplate.Execute(userBuf,
		LogData{
			Text: fmt.Sprint(msg...),
			Icon: l.config.Info,
		},
	)
	l.Cursor.Save()
	defer func() {
		l.Cursor.Restore()
	}()
	_, err := fmt.Fprint(l.Out, userBuf.String())
	if err != nil {
		return err
	}
	return nil
}

func (l *Log) Error(msg ...any) error {
	userBuf := bytes.NewBufferString("")
	l.baseTemplate.Execute(userBuf,
		LogData{
			Text: fmt.Sprint(msg...),
			Icon: l.config.Error,
		},
	)
	l.Cursor.Restore()
	EraseLine(l.Out, ERASE_LINE_ALL)
    defer l.Cursor.Save()
	_, err := fmt.Fprint(l.Out, userBuf.String())
	if err != nil {
		return err
	}
	return nil

}
func (l *Log) Success(msg ...any) error {
	userBuf := bytes.NewBufferString("")
	l.baseTemplate.Execute(userBuf,
		LogData{
			Text: fmt.Sprint(msg...),
			Icon: l.config.Success,
		},
	)
	l.Cursor.Restore()
	EraseLine(l.Out, ERASE_LINE_ALL)
	defer l.Cursor.Save()
	_, err := fmt.Fprint(l.Out, userBuf.String())
	if err != nil {
		return err
	}
	return nil
}
