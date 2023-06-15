package repo

import (
	"reflect"
	"testing"
)

func TestInitOrgs(t *testing.T) {

	t.Run("It sets up the default orgs", func(t *testing.T) {
		InitOrgs()
		assertEqual(t, WordPressOrg, "WordPress")
		assertEqual(t, AutomatticOrg, "Automattic")
		assertEqual(t, WpMobileOrg, "wordpress-mobile")
	})

	t.Run("It returns the orgs from the environment variables", func(t *testing.T) {
		t.Setenv("GBM_WPMOBILE_ORG", "my-wordpress-mobile")
		t.Setenv("GBM_WORDPRESS_ORG", "my-wordpress")
		t.Setenv("GBM_AUTOMATTIC_ORG", "my-automattic")
		defer clearEnv(t)

		InitOrgs()

		assertEqual(t, WordPressOrg, "my-wordpress")
		assertEqual(t, AutomatticOrg, "my-automattic")
		assertEqual(t, WpMobileOrg, "my-wordpress-mobile")
	})
}

func assertNoError(t testing.TB, got error) {
	t.Helper()
	if got != nil {
		t.Fatalf("got an error but didn't expect one: %v", got)
	}
}

func assertError(t testing.TB, got error) {
	t.Helper()
	if got == nil {
		t.Fatal("expected an error but didn't get one")
	}
}

func assertEqual(t testing.TB, got, want interface{}) {
	t.Helper()
	eq := reflect.DeepEqual(got, want)
	if !eq {
		t.Fatalf("got %v want %v", got, want)
	}
}

func assertNotEqual(t testing.TB, got, want interface{}) {
	t.Helper()
	eq := reflect.DeepEqual(got, want)
	if eq {
		t.Fatalf("got %v want %v", got, want)
	}
}

func clearEnv(t testing.TB) {
	t.Helper()
	t.Setenv("GBM_WPMOBILE_ORG", "")
	t.Setenv("GBM_WORDPRESS_ORG", "")
	t.Setenv("GBM_AUTOMATTIC_ORG", "")
	InitOrgs()
}
