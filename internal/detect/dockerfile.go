package detect

import (
	"bufio"
	"regexp"
	"strings"
)

var fromRe = regexp.MustCompile(`(?i)^\s*FROM\s+(\S+)`)
var leadingIntRe = regexp.MustCompile(`^\d+`)

var dockerImageDistros = map[string]string{
	"ubuntu": "Ubuntu",
	"debian": "Debian",
	"fedora": "Fedora",
	"alpine": "Alpine Linux",
}

func detectDockerfile(result *Result, path string) {
	content, ok := readCapped(path)
	if !ok {
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		match := fromRe.FindStringSubmatch(scanner.Text())
		if match == nil {
			continue
		}
		detectDockerImage(result, match[1])
	}
}

func detectDockerImage(result *Result, image string) {
	image = strings.TrimPrefix(image, "docker.io/library/")

	repo, tag, _ := strings.Cut(image, ":")

	if strings.Contains(repo, "dotnet/") {
		if version := imageVersion(tag); version != "" {
			result.Technologies = append(result.Technologies, Technology{
				Name:     ".NET Core Runtime",
				Version:  version,
				Category: "Framework",
				Source:   "Dockerfile",
			})
		}
		return
	}

	base := repo
	if slash := strings.LastIndex(repo, "/"); slash >= 0 {
		base = repo[slash+1:]
	}

	name, known := dockerImageDistros[base]
	if !known {
		return
	}

	result.Technologies = append(result.Technologies, Technology{
		Name:     name,
		Version:  imageVersion(tag),
		Category: "OperatingSystem",
		Source:   "Dockerfile",
	})
}

func imageVersion(tag string) string {
	if match := kernelVersionRe.FindString(tag); match != "" {
		return match
	}
	if match := leadingIntRe.FindString(tag); match != "" {
		return match
	}
	return ""
}
