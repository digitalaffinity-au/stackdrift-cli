package commands

import (
	"errors"
	"net/http"

	"github.com/digitalaffinity-au/stackdrift-cli/internal/api"
	"github.com/digitalaffinity-au/stackdrift-cli/internal/config"
)

var errNotSignedIn = errors.New("not signed in, run: stackdrift login")

func isNotFound(err error) bool {
	var apiErr *api.Error
	return errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound
}

func authenticatedClient() (*api.Client, string, error) {
	baseURL := config.BaseURL()
	cred, err := config.LoadCredential(baseURL)
	if err != nil {
		return nil, baseURL, err
	}
	if cred == nil || cred.Token == "" {
		return nil, baseURL, errNotSignedIn
	}
	return api.New(baseURL, cred.Token), baseURL, nil
}
