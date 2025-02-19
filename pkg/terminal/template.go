package terminal

import (
	"bytes"
	"sync"
	"text/template"

	"github.com/mgutz/ansi"
)

// Todo : Figure out how this works
var TemplateFuncsWithColor = map[string]interface{}{
	// Templates with Color formatting. See Documentation: https://github.com/mgutz/ansi#style-format
	"color": ansi.ColorCode,
}

var TemplateFuncsNoColor = map[string]interface{}{
	// Templates without Color formatting. For layout/ testing.
	"color": func(color string) string {
		return ""
	},
}

var (
	memoizedGetTemplate = map[string][2]*template.Template{}
	memoMutex           = &sync.RWMutex{}
)

func RunTemplate(tmpl string, data interface{}) (string, string, error) {
	tPair, err := GetTemplatePair(tmpl)
	if err != nil {
		return "", "", err
	}
	userBuf := bytes.NewBufferString("")
	err = tPair[0].Execute(userBuf, data)
	if err != nil {
		return "", "", err
	}
	layoutBuf := bytes.NewBufferString("")
	err = tPair[1].Execute(layoutBuf, data)
	if err != nil {
		return "", "", err
	}
    return userBuf.String(), layoutBuf.String(), nil
}

func GetTemplatePair(tmpl string) ([2]*template.Template, error) {
	memoMutex.RLock()
	if t, ok := memoizedGetTemplate[tmpl]; ok {
		memoMutex.RUnlock()
		return t, nil
	}
	memoMutex.RUnlock()
	templatePair := [2]*template.Template{nil, nil}
	templateNoColor, err := template.New("Prompt").Funcs(TemplateFuncsNoColor).Parse(tmpl)
	if err != nil {
		return [2]*template.Template{nil, nil}, err
	}
	templatePair[1] = templateNoColor
	templateWithColor, err := template.New("Prompt").Funcs(TemplateFuncsWithColor).Parse(tmpl)
	if err != nil {
		return [2]*template.Template{nil, nil}, err
	}
	templatePair[0] = templateWithColor
	memoMutex.Lock()
	memoizedGetTemplate[tmpl] = templatePair
	memoMutex.Unlock()
	return templatePair, nil
}
