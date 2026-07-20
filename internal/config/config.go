package config

import (
	"os"
	"strings"
)

const DefaultBaseURL = "https://stackdrift.net"

func BaseURL() string {
	if fromEnv := strings.TrimSpace(os.Getenv("STACKDRIFT_URL")); fromEnv != "" {
		return normalizeURL(fromEnv)
	}
	return normalizeURL(DefaultBaseURL)
}

func normalizeURL(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}
