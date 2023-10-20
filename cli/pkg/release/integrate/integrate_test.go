package integrate

import "testing"

func TestGetRepo(t *testing.T) {
	t.Run("It returns the repo based on the platform", func(t *testing.T) {
		androidRi := ReleaseIntegration{Type: androidIntegration{}}
		got := androidRi.getRepo()
		assertEqual(t, got, "WordPress-Android")

		iosRi := ReleaseIntegration{Type: iosIntegration{}}
		got = iosRi.getRepo()
		assertEqual(t, got, "WordPress-iOS")
	})
}

func TestRun(t *testing.T) {
	t.Run("It returns an error if no platform is specified", func(t *testing.T) {
		ri := ReleaseIntegration{}
		_, err := ri.Run("")
		assertError(t, err)
	})
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("Expected %s, got %s", want, got)
	}
}
