package templates

import (
	"bytes"
	"io/ioutil"
	"text/template"
)

func ParseCadenceTemplateV2(templatePath string, data interface{}) []byte {
	fb, err := ioutil.ReadFile(templatePath)
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New("Template").Parse(string(fb))
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, data)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}
