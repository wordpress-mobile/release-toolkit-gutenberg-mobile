package yq

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/diff"
)

func TestYqEval(t *testing.T) {

	t.Run("it adds a key", func(t *testing.T) {

		test := read(t, "./testdata/ref_tag.yaml")
		want := read(t, "./testdata/ref_tag_commit.yaml")

		got, err := YqEval(".ref.commit = \"1.2.3\"", test)
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.CreateTemp("", "result.yaml")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(f.Name())

		_, err = f.WriteString(got)

		if err != nil {
			t.Fatal(err)
		}
		assertEqual(t, got, want)
	})

	t.Run("it removes a key", func(t *testing.T) {
		test := read(t, "./testdata/ref_tag_commit.yaml")
		want := read(t, "./testdata/ref_commit.yaml")

		got, err := YqEval("del(.ref.tag)", test)

		if err != nil {
			t.Fatal(err)
		}
		assertEqual(t, got, want)
	})

}
func assertEqual(t testing.TB, got, want string) {
	t.Helper()

	// Trim spaces to avoid false negatives
	got = strings.Trim(got, " ")
	want = strings.Trim(want, " ")

	if !reflect.DeepEqual(got, want) {
		t.Fatal("Found a difference", diff.Diff(got, want))
	}
}
func read(t testing.TB, file string) string {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
