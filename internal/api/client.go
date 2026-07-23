package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// VersionHeader tells the server which build is calling. The server turns away
// anything behind the current release, so a build that does not send this is
// read as predating the check and refused with it.
const VersionHeader = "X-StackDrift-CLI-Version"

// Set from main at startup, where the release version is stamped in.
var Version = "dev"

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("request failed with status %d", e.Status)
}

// IsUpgradeRequired reports whether the server refused the build itself rather
// than anything about the request, which no retry of the same binary can fix.
func IsUpgradeRequired(err error) bool {
	var apiErr *Error
	return errors.As(err, &apiErr) && apiErr.Status == http.StatusUpgradeRequired
}

// IsUnauthorized reports whether the server rejected the token outright. Note
// that /api/auth/me is anonymous and answers 200 with Authenticated false
// instead, so a session check has to read that flag rather than call this.
func IsUnauthorized(err error) bool {
	var apiErr *Error
	return errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *Client) do(method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set(VersionHeader, Version)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &Error{Status: resp.StatusCode, Message: extractMessage(data, resp.StatusCode)}
	}

	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) raw(method, path string, body []byte) (int, []byte, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return 0, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, data, nil
}

func marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func unmarshal(data []byte, out any) error {
	return json.Unmarshal(data, out)
}

func extractMessage(data []byte, status int) string {
	var problem struct {
		Message string   `json:"message"`
		Detail  string   `json:"detail"`
		Title   string   `json:"title"`
		Errors  []string `json:"errors"`
	}
	if err := json.Unmarshal(data, &problem); err == nil {
		switch {
		case problem.Message != "":
			return problem.Message
		case problem.Detail != "":
			return problem.Detail
		case problem.Title != "":
			return problem.Title
		case len(problem.Errors) > 0:
			return strings.Join(problem.Errors, " ")
		}
	}
	if status == http.StatusUnauthorized {
		return "not signed in, run: stackdrift login"
	}
	return fmt.Sprintf("request failed with status %d", status)
}
