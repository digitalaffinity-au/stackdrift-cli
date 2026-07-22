package detect

import (
	"path/filepath"
	"strings"
)

// Directories that hold copies of published packages rather than source. Each
// copy carries the package.json from its own tarball, so without this every
// vendored library would be reported as a separate project.
//
// Keys with a slash only mean "vendored" under that parent. A bare lib folder
// is ordinary source in most projects, but the lib inside wwwroot is where
// LibMan installs client libraries.
var vendorDirs = map[string]bool{
	"bower_components": true,
	"jspm_packages":    true,
	"web_modules":      true,
	"typings":          true,
	"wwwroot/lib":      true,
	"content/lib":      true,
}

func isVendorDir(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	if vendorDirs[name] {
		return true
	}

	parent := strings.ToLower(filepath.Base(filepath.Dir(path)))
	return vendorDirs[parent+"/"+name]
}
