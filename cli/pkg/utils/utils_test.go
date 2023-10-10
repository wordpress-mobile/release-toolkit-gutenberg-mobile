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

func TestValidVersion(t *testing.T) {
	t.Run("It returns true for a valid scheduled release", func(t *testing.T) {
		got := ValidateVersion("1.0.0")

		if !got {
			t.Fatalf("Expected %v to be true", got)
		}
	})

	t.Run("It returns true for a valid non-scheduled release", func(t *testing.T) {
		got := ValidateVersion("1.0.1")

		if !got {
			t.Fatalf("Expected %v to be true", got)
		}
	})

	t.Run("It returns false when the patch value is missing", func(t *testing.T) {
		got := ValidateVersion("1.0")

		if got {
			t.Fatalf("Expected %v to be false", got)
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

func TestNormalizeVersion(t *testing.T) {
	t.Run("It returns an error if the version is invalid", func(t *testing.T) {
		_, err := NormalizeVersion("1.0")
		if err == nil {
			t.Fatalf("Expected an error, got nil")
		}
	})

	t.Run("It returns the version without the 'v' prefix", func(t *testing.T) {
		got, err := NormalizeVersion("v1.0.0")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if got != "1.0.0" {
			t.Fatalf("Expected %s, got %s", "1.0.0", got)
		}
	})

	t.Run("It returns the version if it's valid", func(t *testing.T) {
		got, err := NormalizeVersion("1.0.0")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if got != "1.0.0" {
			t.Fatalf("Expected %s, got %s", "1.0.0", got)
		}
	})
}