package workspace

import (
	"io"
	"text/template"
)

type File interface {
	Generate(io.Writer, Cfg) error
	Name() string
}

type fileTemplate struct {
	name, tpl string
}

func (t fileTemplate) Name() string {
	return t.name
}

func (t fileTemplate) Generate(writer io.Writer, cfg Cfg) error {
	tpl, err := template.New("wsTemplateFile").Parse(t.tpl)
	if err != nil {
		return err
	}

	return tpl.Execute(writer, cfg)
}

func newFileTemplate(tpl, name string) File {
	return &fileTemplate{
		tpl:  tpl,
		name: name,
	}
}
