package render

import (
	"embed"
	"testing"
)

//go:embed testdata/*
var templatesFS embed.FS

func init() {
	TemplateFS = templatesFS
}

func TestRenderJSON(t *testing.T) {

	t.Run("It renders a template with the given JSON", func(t *testing.T) {
		tmplt := Template{
			Path: "testdata/test_template.txt",
			Json: `{"world": "World"}`,
		}

		got, err := RenderJSON(tmplt)
		assertNoError(t, err)

		if got != "Hello World" {
			t.Fatalf("Expected %s, got %s", "Hello World\n", got)
		}
	})

	t.Run("It returns an error if the JSON is invalid", func(t *testing.T) {
		tmplt := Template{
			Path: "testdata/test_template.txt",
			Json: `{"world": "World`,
		}

		_, err := RenderJSON(tmplt)
		assertError(t, err)
	})

	t.Run("It returns an error if the template is invalid", func(t *testing.T) {
		tmplt := Template{
			Path: "testdata/invalid_template.txt",
			Json: `{"world": "World`,
		}

		_, err := RenderJSON(tmplt)
		assertError(t, err)
	})

	t.Run("It returns an error if the template is missing", func(t *testing.T) {

		tmplt := Template{
			Path: "testdata/missing_template.txt",
			Json: `{"world": "World`,
		}
		_, err := RenderJSON(tmplt)
		assertError(t, err)
	})

	t.Run("It renders with custom functions", func(t *testing.T) {

		tmplt := Template{
			Path: "testdata/func_template.txt",
			Json: `{}`,
			Funcs: map[string]any{
				"echo": func(str string) string {
					return str
				},
			},
		}

		got, err := RenderJSON(tmplt)
		assertNoError(t, err)

		if got != "Hello Custom" {
			t.Fatalf("Expected %s, got %s", "Hello Custom\n", got)
		}
	})

	t.Run("It renders with no json data", func(t *testing.T) {
		tmplt := Template{
			Path: "testdata/basic_template.txt",
		}

		got, err := RenderJSON(tmplt)
		assertNoError(t, err)

		if got != "Hello World!" {
			t.Fatalf("Expected %s, got %s", "Hello World!\n", got)
		}
	})
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
