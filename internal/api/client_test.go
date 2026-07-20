package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPollDeviceToken_PendingThenApproved_ReturnsStatuses(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"error":"authorization_pending"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"accessToken":"sdp_abc","tokenType":"Bearer"}`))
	}))
	defer server.Close()

	client := New(server.URL, "")

	_, status1, err := client.PollDeviceToken("dc")
	if err != nil || status1 != http.StatusAccepted {
		t.Fatalf("expected 202, got %d (err %v)", status1, err)
	}

	token, status2, err := client.PollDeviceToken("dc")
	if err != nil || status2 != http.StatusOK {
		t.Fatalf("expected 200, got %d (err %v)", status2, err)
	}
	if token.AccessToken != "sdp_abc" {
		t.Fatalf("expected token, got %q", token.AccessToken)
	}
}

func TestPollDeviceToken_Denied_ReturnsGoneWithoutToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"access_denied"}`))
	}))
	defer server.Close()

	token, status, err := New(server.URL, "").PollDeviceToken("dc")
	if err != nil {
		t.Fatal(err)
	}
	if status != http.StatusForbidden || token != nil {
		t.Fatalf("expected 403 nil token, got %d %+v", status, token)
	}
}

func TestAddTechnology_SendsBearerAndBody(t *testing.T) {
	var gotAuth string
	var gotBody AddTechnologyRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		_, _ = w.Write([]byte(`{"id":1,"name":"Demo","technologies":[]}`))
	}))
	defer server.Close()

	client := New(server.URL, "sdp_token")
	_, err := client.AddTechnology(3, AddTechnologyRequest{Name: "Ubuntu", Version: "24.04", Category: "OperatingSystem"})
	if err != nil {
		t.Fatal(err)
	}

	if gotAuth != "Bearer sdp_token" {
		t.Fatalf("expected bearer header, got %q", gotAuth)
	}
	if gotBody.Name != "Ubuntu" || gotBody.Category != "OperatingSystem" {
		t.Fatalf("unexpected body: %+v", gotBody)
	}
}

func TestUploadManifests_SendsFilesAndEcosystem(t *testing.T) {
	var got UploadManifestsRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		_, _ = w.Write([]byte(`{"summary":{"totalCount":2,"groups":[]},"unsupportedFiles":[]}`))
	}))
	defer server.Close()

	resp, err := New(server.URL, "t").UploadManifests(3, UploadManifestsRequest{
		Ecosystem: "Npm",
		GroupName: "web npm",
		Files:     []ManifestFile{{FileName: "package.json", Content: "{}"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Ecosystem != "Npm" || len(got.Files) != 1 || got.Files[0].FileName != "package.json" {
		t.Fatalf("unexpected upload: %+v", got)
	}
	if resp.Summary.TotalCount != 2 {
		t.Fatalf("expected 2 deps, got %d", resp.Summary.TotalCount)
	}
}

func TestDo_ErrorBody_SurfacesMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"Name is required."}`))
	}))
	defer server.Close()

	_, err := New(server.URL, "t").CreateProject("", "")
	if err == nil {
		t.Fatal("expected an error")
	}
	apiErr, ok := err.(*Error)
	if !ok || apiErr.Message != "Name is required." {
		t.Fatalf("expected surfaced message, got %v", err)
	}
}

func TestDo_Unauthorized_GivesLoginHint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	_, err := New(server.URL, "").ListProjects()
	apiErr, ok := err.(*Error)
	if !ok || apiErr.Status != http.StatusUnauthorized {
		t.Fatalf("expected 401 error, got %v", err)
	}
}
