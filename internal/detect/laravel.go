package detect

import (
	"encoding/json"
	"regexp"
)

var versionLeadRe = regexp.MustCompile(`(\d+(?:\.\d+)*)`)

func detectLaravel(result *Result, path string) {
	content, ok := readCapped(path)
	if !ok {
		return
	}

	var composer struct {
		Require map[string]string `json:"require"`
	}
	if err := json.Unmarshal([]byte(content), &composer); err != nil {
		return
	}

	constraint, present := composer.Require["laravel/framework"]
	if !present {
		return
	}

	result.Technologies = append(result.Technologies, Technology{
		Name:     "Laravel",
		Version:  cleanVersion(constraint),
		Category: "Framework",
		Source:   "composer.json",
	})
}

func cleanVersion(constraint string) string {
	if match := versionLeadRe.FindString(constraint); match != "" {
		return match
	}
	return ""
}
