package render

import (
	"bytes"
	"embed"
	"encoding/json"
	"text/template"
)

var TemplateFS embed.FS

func Render(templatePath string, rawJSON string) (string, error) {

	var data map[string]interface{}

	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
		return "", err
	}

	tmpl, err := template.ParseFS(TemplateFS, templatePath)
	if err != nil {
		return "", err
	}

	var result bytes.Buffer
	if err := tmpl.Execute(&result, data); err != nil {
		return "", nil
	}

	return result.String(), nil
}
