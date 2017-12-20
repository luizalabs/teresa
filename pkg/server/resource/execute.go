package resource

import (
	"io"
	"io/ioutil"
	"strings"
	"text/template"
)

type TemplateExecuter interface {
	Execute(w io.Writer, r io.ReadCloser, settings []*Setting) error
}

type goTemplateExecuter struct{}

func settingsToMap(s []*Setting) map[string]string {
	m := make(map[string]string, len(s))

	for i, _ := range s {
		key := strings.Replace(s[i].Key, "-", "_", -1)
		m[key] = s[i].Value
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

func NewTemplateExecuter() TemplateExecuter {
	return &goTemplateExecuter{}
}
