package render

import (
	"bytes"
	"embed"
	"encoding/json"
	"path/filepath"
	"text/template"
)

var TemplateFS embed.FS

func RenderJSON(templatePath string, rawJSON string, funcs template.FuncMap) (string, error) {

	var data map[string]interface{}

	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
		return "", err
	}

	return Render(templatePath, data, funcs)
}

func Render(tmplPath string, data interface{}, funcs map[string]any) (string, error) {
	t, err := template.New(filepath.Base(tmplPath)).
		Funcs(funcs).
		ParseFS(TemplateFS, tmplPath) // Parse the template file

	if err != nil {
		return "", err
	}

	var result bytes.Buffer
	if err := t.Execute(&result, data); err != nil {
		return "", err
	}

	return result.String(), nil
}
