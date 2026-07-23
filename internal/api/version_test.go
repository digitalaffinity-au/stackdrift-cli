package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// The server refuses anything that does not name a version, so sending it is
// not optional decoration.
func TestDo_SendsTheVersionHeaderOnEveryRequest(t *testing.T) {
	previous := Version
	Version = "0.1.26"
	t.Cleanup(func() { Version = previous })

	got := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get(VersionHeader)
		_, _ = w.Write([]byte(`{"authenticated":true}`))
	}))
	t.Cleanup(server.Close)

	if _, err := New(server.URL, "token").Me(); err != nil {
		t.Fatal(err)
	}

	if got != "0.1.26" {
		t.Fatalf("expected the version to be sent, got %q", got)
	}
}

func TestIsUpgradeRequired_UpgradeRequiredStatus_IsRecognised(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUpgradeRequired)
	}))
	t.Cleanup(server.Close)

	_, err := New(server.URL, "token").ListProjects()

	if !IsUpgradeRequired(err) {
		t.Fatalf("expected an upgrade required verdict, got %v", err)
	}
}

// Updating and re-running is a heavy response, so it must not fire for a
// failure that has nothing to do with the build.
func TestIsUpgradeRequired_OtherFailures_AreNot(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))

		_, err := New(server.URL, "token").ListProjects()
		server.Close()

		if IsUpgradeRequired(err) {
			t.Fatalf("status %d must not be read as needing an upgrade", status)
		}
	}
}

func TestIsUpgradeRequired_NoError_IsFalse(t *testing.T) {
	if IsUpgradeRequired(nil) {
		t.Fatal("nil must not be read as needing an upgrade")
	}
}
