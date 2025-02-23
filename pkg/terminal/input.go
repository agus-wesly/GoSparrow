package terminal

import (
	"bufio"
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Input struct {
	Renderer
	Message   string
	Default   string
	Validator func(x interface{}) error
}

type InputTemplateData struct {
	Input
	Config     *PromptConfig
	Answer     string
	ShowAnswer bool
}

var InputQuestionTemplate = `
    {{- color .Config.Icons.Question.Format }}{{ .Config.Icons.Question.Text }} {{color "reset"}}
    {{- color "default+hb"}}{{ .Message }}: {{color "reset"}}
{{- if .ShowAnswer}}
  {{- color "cyan"}}{{.Answer}}{{color "reset"}}{{"\n"}}
{{- end}}`

func (input *Input) Ask(resp interface{}) error {
	option := defaultAskOption()
	input.Renderer.WithStd(option.Stdio)
	var validationError error
	var answer interface{}
	var err error
	for {
		if validationError != nil {
			if err := input.Error(&option.PromptConfig, validationError); err != nil {
				return err
			}
		}
		answer, err = input.Prompt(&option.PromptConfig)
		if err != nil {
			return err
		}
		if input.Validator != nil {
			validationError = input.Validator(answer)
		}
		if validationError == nil {
			break
		}
	}
	err = input.WriteAnswer(resp, answer)
	if err != nil {
		return err
	}
	err = input.Clear(answer, &option.PromptConfig)
	if err != nil {
		return err
	}
	return nil
}

func (input *Input) WriteAnswer(target interface{}, val interface{}) error {
	t := reflect.ValueOf(target)
	if t.Kind() != reflect.Ptr {
		return errors.New("You must pass a pointer")
	}
	var res interface{}
	var err error
	elem := t.Elem()
	switch elem.Kind() {
	case reflect.Int:
		res, err = strconv.Atoi(val.(string))
		break
	case reflect.String:
		res = val
		break
	default:
		err = errors.New("Unsupported type")
	}
	if err != nil {
		return err
	}
	elem.Set(reflect.ValueOf(res))
	return nil
}

func (input *Input) Clear(resp interface{}, config *PromptConfig) error {
	c := input.NewCursor()
	c.Restore()
	inp := InputTemplateData{
		Input:      *input,
		Config:     config,
		ShowAnswer: true,
		Answer:     resp.(string),
	}
	err := input.Render(InputQuestionTemplate, &inp)
	if err != nil {
		return err
	}
	return nil
}

func (input *Input) Prompt(config *PromptConfig) (interface{}, error) {
	inp := InputTemplateData{
		Input:      *input,
		Config:     config,
		ShowAnswer: false,
	}
	c := input.NewCursor()
	c.Save()
	defer c.Restore()
	err := input.Render(InputQuestionTemplate, &inp)
	if err != nil {
		return nil, err
	}
	inputQuery, err := input.Read()
	if IsEmpty(inputQuery) {
		return input.Default, nil
	}
	return inputQuery, nil
}

func (input *Input) Read() (string, error) {
	in := bufio.NewReader(os.Stdin)
	inputQuery, err := in.ReadString('\n')
	if err != nil {
		return "", err
	}
	inputQuery = strings.TrimSpace(inputQuery)
	return inputQuery, nil
}

func IsEmpty(val interface{}) bool {
	if val == "" || val == nil {
		return true
	}
	return false
}
