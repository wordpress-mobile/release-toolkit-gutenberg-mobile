package gbm

import (
	"testing"

	"github.com/andreyvit/diff"
)

func assertNoError(t testing.TB, got error) {
	t.Helper()
	if got != nil {
		t.Fatalf("got an error but didn't expect one: %v", got)
	}
}

func assertStringEqual(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("strings are not equal: %s", diff.LineDiff(got, want))
	}
}
