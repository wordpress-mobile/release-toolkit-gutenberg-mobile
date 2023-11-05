package semver

import "testing"

func TestSemVer(t *testing.T) {

	t.Run("It returns an error if the version is invalid", func(t *testing.T) {
		_, err := NewSemVer("1.0")
		assertError(t, err)
	})

	t.Run("It returns the version without the 'v' prefix", func(t *testing.T) {
		semver, err := NewSemVer("v1.0.0")
		assertNotError(t, err)
		assertEqual(t, semver.String(), "1.0.0")
	})

	t.Run("It returns the version with the 'v' prefix", func(t *testing.T) {
		semver, err := NewSemVer("1.0.0")
		assertNotError(t, err)
		assertEqual(t, semver.Vstring(), "v1.0.0")
	})

	t.Run("It returns the prior version of patch release", func(t *testing.T) {
		semver, err := NewSemVer("1.0.1")
		assertNotError(t, err)
		assertEqual(t, semver.PriorVersion().String(), "1.0.0")
	})

	t.Run("It can determine a scheduled release", func(t *testing.T) {
		semver, err := NewSemVer("1.0.0")
		assertNotError(t, err)
		if !semver.IsScheduledRelease() {
			t.Fatal("Expected 1.0.0 to be a scheduled release")
		}

		semver, err = NewSemVer("1.0.1")
		assertNotError(t, err)
		if semver.IsScheduledRelease() {
			t.Fatal("Expected 1.0.1 to not be a scheduled release")
		}
	})

	t.Run("It can determine a patch release", func(t *testing.T) {
		semver, err := NewSemVer("1.0.1")
		assertNotError(t, err)
		if !semver.IsPatchRelease() {
			t.Fatal("Expected 1.0.1 to be a patch release")
		}

		semver, err = NewSemVer("1.0.0")
		assertNotError(t, err)
		if semver.IsPatchRelease() {
			t.Fatal("Expected 1.0.0 to not be a patch release")
		}
	})

}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
}

func assertNotError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("Expected %s, got %s", want, got)
	}
}
