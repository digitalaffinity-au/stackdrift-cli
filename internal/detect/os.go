package detect

import (
	"bufio"
	"os"
	"regexp"
	"runtime"
	"strings"
)

var distroNames = map[string]string{
	"ubuntu":        "Ubuntu",
	"debian":        "Debian",
	"fedora":        "Fedora",
	"rhel":          "Red Hat Enterprise Linux",
	"opensuse-leap": "openSUSE Leap",
	"alpine":        "Alpine Linux",
	"linuxmint":     "Linux Mint",
}

var kernelVersionRe = regexp.MustCompile(`^(\d+\.\d+)`)

func scanHost(result *Result) {
	if runtime.GOOS != "linux" {
		if runtime.GOOS == "windows" {
			result.Technologies = append(result.Technologies, Technology{
				Name:     "Windows",
				Category: "OperatingSystem",
				Source:   SourceHost,
			})
		}
		return
	}

	detectOsRelease(result, "/etc/os-release")
	detectKernel(result)
}

func detectOsRelease(result *Result, path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	fields := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, found := strings.Cut(scanner.Text(), "=")
		if !found {
			continue
		}
		fields[key] = strings.Trim(value, `"`)
	}

	name, known := distroNames[fields["ID"]]
	if !known {
		return
	}

	result.Technologies = append(result.Technologies, Technology{
		Name:     name,
		Version:  fields["VERSION_ID"],
		Category: "OperatingSystem",
		Source:   SourceOsRelease,
	})
}

func detectKernel(result *Result) {
	data, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return
	}
	match := kernelVersionRe.FindString(strings.TrimSpace(string(data)))
	if match == "" {
		return
	}
	result.Technologies = append(result.Technologies, Technology{
		Name:     "Linux Kernel",
		Version:  match,
		Category: "OperatingSystem",
		Source:   SourceHostKern,
	})
}
