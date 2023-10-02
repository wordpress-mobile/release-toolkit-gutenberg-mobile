package render

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

var TemplateFS embed.FS

type Template struct {
	Path, Json string
	Funcs      template.FuncMap
}

func RenderTasks(t Template) (string, error) {
	if t.Funcs == nil {
		t.Funcs = template.FuncMap{}
	}
	t.Funcs["Task"] = Task

	return RenderJSON(t)
}

func RenderJSON(t Template) (string, error) {

	var data map[string]interface{}

	if t.Json != "" {
		if err := json.Unmarshal([]byte(t.Json), &data); err != nil {
			return "", err
		}
	}

	return Render(t.Path, data, t.Funcs)
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

func Task(format string, args ...interface{}) string {
	t := struct {
		Description string
	}{
		Description: fmt.Sprintf(format, args...),
	}
	res, err := Render("templates/checklist/task.html", t, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(13)
	}
	return res
}
