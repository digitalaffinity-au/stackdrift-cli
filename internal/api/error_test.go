package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExtractMessage_PrefersMessageOverEverythingElse(t *testing.T) {
	body := []byte(`{"message":"first","detail":"second","title":"third","errors":["fourth"]}`)
	if got := extractMessage(body, http.StatusBadRequest); got != "first" {
		t.Fatalf("expected the message field to win, got %q", got)
	}
}

func TestExtractMessage_ProblemDetails_UsesDetail(t *testing.T) {
	// The server returns RFC 9457 ProblemDetails for a rejected technology, so
	// detail carries the reason the user needs to read.
	body := []byte(`{"title":"Unknown technology","detail":"StackDrift does not track 'Wordpres'.","status":400}`)
	if got := extractMessage(body, http.StatusBadRequest); got != "StackDrift does not track 'Wordpres'." {
		t.Fatalf("expected the detail, got %q", got)
	}
}

func TestExtractMessage_TitleOnly_FallsBackToTitle(t *testing.T) {
	body := []byte(`{"title":"Unknown technology","status":400}`)
	if got := extractMessage(body, http.StatusBadRequest); got != "Unknown technology" {
		t.Fatalf("expected the title, got %q", got)
	}
}

func TestExtractMessage_ValidationErrors_AreJoined(t *testing.T) {
	body := []byte(`{"errors":["Name is required.","Version is invalid."]}`)
	got := extractMessage(body, http.StatusBadRequest)
	if !strings.Contains(got, "Name is required.") || !strings.Contains(got, "Version is invalid.") {
		t.Fatalf("expected both errors, got %q", got)
	}
}

func TestExtractMessage_UnauthorizedWithNoBody_GivesLoginHint(t *testing.T) {
	got := extractMessage(nil, http.StatusUnauthorized)
	if !strings.Contains(got, "stackdrift login") {
		t.Fatalf("expected a login hint, got %q", got)
	}
}

func TestExtractMessage_NonJsonBody_FallsBackToStatus(t *testing.T) {
	got := extractMessage([]byte("<html>502 Bad Gateway</html>"), http.StatusBadGateway)
	if !strings.Contains(got, "502") {
		t.Fatalf("expected the status in the fallback, got %q", got)
	}
}

func TestExtractMessage_EmptyJsonObject_FallsBackToStatus(t *testing.T) {
	got := extractMessage([]byte(`{}`), http.StatusInternalServerError)
	if !strings.Contains(got, "500") {
		t.Fatalf("expected the status in the fallback, got %q", got)
	}
}

func TestError_WithoutMessage_DescribesTheStatus(t *testing.T) {
	err := &Error{Status: 503}
	if !strings.Contains(err.Error(), "503") {
		t.Fatalf("expected the status in the error text, got %q", err.Error())
	}
}

func TestDo_WithoutToken_SendsNoAuthorizationHeader(t *testing.T) {
	var hadAuth bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hadAuth = r.Header["Authorization"]
		_, _ = w.Write([]byte(`{"email":"a@b.c"}`))
	}))
	defer server.Close()

	if _, err := New(server.URL, "").Me(); err != nil {
		t.Fatal(err)
	}
	if hadAuth {
		t.Fatal("an anonymous client must not send an Authorization header")
	}
}

func TestDo_TrailingSlashInBaseURL_DoesNotDoubleUp(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"email":"a@b.c"}`))
	}))
	defer server.Close()

	if _, err := New(server.URL+"/", "t").Me(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(gotPath, "//") {
		t.Fatalf("expected a single slash in the path, got %q", gotPath)
	}
}

func TestDo_EmptySuccessBody_IsNotAnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	if err := New(server.URL, "t").DeleteTechnology(1); err != nil {
		t.Fatalf("a 204 with no body is a success, got %v", err)
	}
}
