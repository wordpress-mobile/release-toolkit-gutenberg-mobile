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
	Data       interface{}
}
type TaskArgs struct {
	Description string
}

func RenderTasks(t Template) (string, error) {
	if t.Funcs == nil {
		t.Funcs = template.FuncMap{}
	}
	t.Funcs["Task"] = Task

	return Render(t)
}

func RenderJSON(t Template) (string, error) {
	if t.Json != "" {
		if err := json.Unmarshal([]byte(t.Json), &t.Data); err != nil {
			return "", err
		}
	}

	return Render(t)
}

func Render(t Template) (string, error) {
	tmp, err := template.New(filepath.Base(t.Path)).
		Funcs(t.Funcs).
		ParseFS(TemplateFS, t.Path) // Parse the template file

	if err != nil {
		return "", err
	}

	var result bytes.Buffer
	if err := tmp.Execute(&result, t.Data); err != nil {
		return "", err
	}

	return result.String(), nil
}

func Task(format string, args ...interface{}) string {
	t := Template{
		Path: "templates/checklist/task.html",
		Data: TaskArgs{Description: fmt.Sprintf(format, args...)},
	}

	res, err := Render(t)
	if err != nil {
		fmt.Println(err)
		os.Exit(13)
	}
	return res
}
