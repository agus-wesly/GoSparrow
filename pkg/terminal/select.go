package terminal

// In:  os.Stdin,
// Out: os.Stdout,
// Err: os.Stderr,

import (
	"errors"
	// "fmt"
	"os"
)

type AskOptions struct {
	Stdio        Stdio
	PromptConfig PromptConfig
}

type Icon struct {
	Text   string
	Format string
}

type IconSet struct {
	HelpInput      Icon
	Error          Icon
	Help           Icon
	Question       Icon
	MarkedOption   Icon
	UnmarkedOption Icon
	SelectFocus    Icon
}

func defaultAskOption() *AskOptions {
	return &AskOptions{
		Stdio: Stdio{In: os.Stdin, Out: os.Stdout, Err: os.Stderr},
		PromptConfig: PromptConfig{
			PageSize:  7,
			HelpInput: "?",
			Icons: IconSet{
				Error: Icon{
					Text:   "X",
					Format: "red",
				},
				Help: Icon{
					Text:   "?",
					Format: "cyan",
				},
				Question: Icon{
					Text:   "?",
					Format: "green+hb",
				},
				MarkedOption: Icon{
					Text:   "[x]",
					Format: "green",
				},
				UnmarkedOption: Icon{
					Text:   "[ ]",
					Format: "default+hb",
				},
				SelectFocus: Icon{
					Text:   ">",
					Format: "cyan+b",
				},
			},
		},
	}
}

var SelectQuestionTemplate = `
{{- define "option"}}
    {{- if eq .SelectedIndex .CurrentIndex }}{{color .Config.Icons.SelectFocus.Format }}{{ .Config.Icons.SelectFocus.Text }} {{else}}{{color "default"}}  {{end}}
    {{- .CurrentOpt.Value}}
    {{- color "reset"}}
{{end}}
{{- color .Config.Icons.Question.Format }}{{ .Config.Icons.Question.Text }} {{color "reset"}}
{{- color "default+hb"}}{{ .Message }}{{color "reset"}}
{{- "  "}}{{- color "cyan"}}[Use arrows to move]{{color "reset"}}
{{- "\n"}}
{{- range $ix, $option := .PageEntries}}
    {{- template "option" $.IterateOption $ix $option}}
{{- end}}`

// OptionAnswer is the return type of Selects/MultiSelects that lets the appropriate information
// get copied to the user's struct
type OptionAnswer struct {
	Value string
	Index int
}

func OptionAnswerList(incoming []string) []OptionAnswer {
	list := []OptionAnswer{}
	for i, opt := range incoming {
		list = append(list, OptionAnswer{Value: opt, Index: i})
	}
	return list
}

type Select struct {
	Renderer
	Opts          []string
	Message       string
	selectedIndex int
}

type PromptConfig struct {
	PageSize  int
	HelpInput string
	Icons     IconSet
}

type SelectTemplateData struct {
	Select
	PageEntries   []OptionAnswer
	SelectedIndex int
	Config        *PromptConfig

	// These fields are used when rendering an individual option
	CurrentOpt   OptionAnswer
	CurrentIndex int
}

type IterableOpts interface {
	IterateOption(int, OptionAnswer) interface{}
}

func (s SelectTemplateData) IterateOption(ix int, opt OptionAnswer) interface{} {
	copy := s
	copy.CurrentIndex = ix
	copy.CurrentOpt = opt
	return copy
}

func (s *Select) Ask(result *int) error {
	if result == nil {
		return errors.New("Please provide result container")
	}
	option := defaultAskOption()
	s.Renderer.WithStd(option.Stdio)
	resp, err := s.Prompt(&option.PromptConfig)
	if err != nil {
		return err
	}
	s.Clear()
	*result = resp
	return nil
}

func (s *Select) Clear() {
	cursor := s.NewCursor()
	cursor.Restore()
	r := s.Renderer
	lineCount := r.countLines(r.renderedText)
	r.resetPrompt(lineCount)
	r.renderedText.Reset()
}

func (s *Select) Prompt(config *PromptConfig) (int, error) {
	opts := OptionAnswerList(s.Opts)
	cursor := s.NewCursor()
	cursor.Save()
	cursor.Hide()
	defer cursor.Show()
	defer cursor.Restore()

	tmpData := SelectTemplateData{
		Select:        *s,
		SelectedIndex: s.selectedIndex,
		PageEntries:   opts,
		Config:        config,
	}
	err := s.RenderWithCursorOffset(SelectQuestionTemplate, tmpData, opts, 0)
	if err != nil {
		return -1, err
	}
	rr := s.NewRuneReader()
	_ = rr.SetTermMode()
	defer func() {
		rr.RestoreTermMode()
	}()
	for {
		r, _, err := rr.ReadRune()
		if err != nil {
			return -1, err
		}
		if r == KeyInterrupt {
			return -1, errors.New("Terminal interrupt")
		}
		if r == KeyEndTransmission {
			break
		}
		if s.OnKeyPressed(r, config, opts) {
			break
		}
	}
	if s.selectedIndex < len(s.Opts) {
		return s.selectedIndex, nil
	}
	return 0, nil
}

func (s *Select) OnKeyPressed(key rune, config *PromptConfig, opts []OptionAnswer) bool {
	if key == KeyEnter || key == '\n' {
		return true
	}
	if key == KeyArrowUp {
		if s.selectedIndex == 0 {
			s.selectedIndex = len(s.Opts) - 1
		} else {
			s.selectedIndex -= 1
		}
	} else if key == KeyArrowDown {
		if s.selectedIndex == len(s.Opts)-1 {
			s.selectedIndex = 0
		} else {
			s.selectedIndex += 1
		}
	}
	tmpData := SelectTemplateData{
		Select:        *s,
		SelectedIndex: s.selectedIndex,
		PageEntries:   opts,
		Config:        config,
	}
	_ = s.RenderWithCursorOffset(SelectQuestionTemplate, tmpData, opts, s.selectedIndex)
	return false
}

