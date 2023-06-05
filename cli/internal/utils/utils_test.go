package utils

import (
	"strings"
	"testing"
)

func TestNextReleaseDate(t *testing.T) {

	t.Run("It returns the next Thursday", func(t *testing.T) {
		got := NextReleaseDate()

		if !strings.Contains(got, "Thursday") {
			t.Fatalf("Expected %s to contain %s", got, "Thursday")
		}
	})
}

func TestIsScheduledRelease(t *testing.T) {

	t.Run("It returns true if the release is scheduled", func(t *testing.T) {
		got := IsScheduledRelease("v1.0.0")

		if !got {
			t.Fatalf("Expected %v to be true", got)
		}
	})

	t.Run("It returns false if the release is not scheduled", func(t *testing.T) {
		got := IsScheduledRelease("v1.0.1")

		if got {
			t.Fatalf("Expected %v to be false", got)
		}
	})

	t.Run("It ignores the 'v' prefix", func(t *testing.T) {
		got := IsScheduledRelease("1.0.0")
		if !got {
			t.Fatalf("Expected %v to be true", got)
		}
	})
}
