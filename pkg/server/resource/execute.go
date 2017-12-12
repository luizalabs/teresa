package resource

import (
	"io"
	"io/ioutil"
	"text/template"
)

type Executer interface {
	Execute(w io.Writer, r io.ReadCloser, settings []*Setting) error
}

type goTemplateExecuter struct{}

func settingsToMap(s []*Setting) map[string]string {
	m := make(map[string]string, len(s))

	for i, _ := range s {
		m[s[i].Key] = s[i].Value
	}

	return m
}

func (g *goTemplateExecuter) Execute(w io.Writer, r io.ReadCloser, settings []*Setting) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	err = r.Close()
	if err != nil {
		return err
	}

	tmpl, err := template.New("").Parse(string(b))
	if err != nil {
		return err
	}

	return tmpl.Execute(w, settingsToMap(settings))
}

func NewExecuter() Executer {
	return &goTemplateExecuter{}
}
