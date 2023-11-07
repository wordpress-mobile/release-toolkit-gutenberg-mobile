package integrate

import "testing"

func TestRun(t *testing.T) {
	t.Run("It returns an error if no platform is specified", func(t *testing.T) {
		t.Skip()
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
