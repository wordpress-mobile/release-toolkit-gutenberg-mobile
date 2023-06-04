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

func TestRender(t *testing.T) {

	t.Run("It renders a template with the given JSON", func(t *testing.T) {
		templatePath := "testdata/test_template.txt"
		rawJSON := `{"world": "World"}`

		got, err := RenderJSON(templatePath, rawJSON, nil)
		assertNoError(t, err)

		if got != "Hello World" {
			t.Fatalf("Expected %s, got %s", "Hello World\n", got)
		}
	})

	t.Run("It returns an error if the JSON is invalid", func(t *testing.T) {
		templatePath := "testdata/test_template.txt"
		rawJSON := `{"world": "World"`

		_, err := RenderJSON(templatePath, rawJSON, nil)
		assertError(t, err)
	})

	t.Run("It returns an error if the template is invalid", func(t *testing.T) {
		templatePath := "testdata/invalid_template.txt"
		rawJSON := `{"world": "World"}`

		_, err := RenderJSON(templatePath, rawJSON, nil)
		assertError(t, err)
	})

	t.Run("It returns an error if the template is missing", func(t *testing.T) {
		templatePath := "testdata/missing_template.txt"
		rawJSON := `{"world": "World"}`
		_, err := RenderJSON(templatePath, rawJSON, nil)
		assertError(t, err)
	})

	t.Run("It renders with custom functions", func(t *testing.T) {
		templatePath := "testdata/func_template.txt"
		rawJSON := "{}"

		funcs := map[string]any{
			"echo": func(str string) string {
				return str
			},
		}

		got, err := RenderJSON(templatePath, rawJSON, funcs)
		assertNoError(t, err)

		if got != "Hello Custom" {
			t.Fatalf("Expected %s, got %s", "Hello Custom\n", got)
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
