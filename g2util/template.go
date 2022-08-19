package g2util

import (
	"bytes"
	"text/template"
)

// TextTemplateMustParse ...
func TextTemplateMustParse(text string, data interface{}) (result string) {
	tp, err := template.New("t").Parse(text)
	if err != nil {
		return
	}
	var buf bytes.Buffer
	if err = tp.Execute(&buf, data); err != nil {
		return
	}
	return buf.String()
}
