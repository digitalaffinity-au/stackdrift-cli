package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Credential struct {
	BaseURL string `json:"baseUrl"`
	Token   string `json:"token"`
	Email   string `json:"email,omitempty"`
}

type credentialStore struct {
	Credentials []Credential `json:"credentials"`
}

func credentialPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "stackdrift", "credentials.json"), nil
}

func LoadCredential(baseURL string) (*Credential, error) {
	store, err := loadStore()
	if err != nil {
		return nil, err
	}
	for i := range store.Credentials {
		if store.Credentials[i].BaseURL == baseURL {
			return &store.Credentials[i], nil
		}
	}
	return nil, nil
}

func SaveCredential(cred Credential) error {
	store, err := loadStore()
	if err != nil {
		return err
	}

	replaced := false
	for i := range store.Credentials {
		if store.Credentials[i].BaseURL == cred.BaseURL {
			store.Credentials[i] = cred
			replaced = true
			break
		}
	}
	if !replaced {
		store.Credentials = append(store.Credentials, cred)
	}

	return saveStore(store)
}

func DeleteCredential(baseURL string) error {
	store, err := loadStore()
	if err != nil {
		return err
	}

	kept := store.Credentials[:0]
	for _, c := range store.Credentials {
		if c.BaseURL != baseURL {
			kept = append(kept, c)
		}
	}
	store.Credentials = kept
	return saveStore(store)
}

func loadStore() (*credentialStore, error) {
	path, err := credentialPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &credentialStore{}, nil
	}
	if err != nil {
		return nil, err
	}

	var store credentialStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	return &store, nil
}

func saveStore(store *credentialStore) error {
	path, err := credentialPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}
