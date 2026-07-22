package detect

import (
	"path/filepath"
	"strings"
)

// Directories that hold copies of published packages rather than source. Each
// copy carries the package.json from its own tarball, so without this every
// vendored library would be reported as a separate project.
//
// Keys with a slash only mean "vendored" under that parent. A bare lib, css or
// plugins folder is ordinary source in most projects; the ones inside wwwroot,
// assets or wp-content are where tooling installs third party code.
var vendorDirs = map[string]bool{
	"bower_components": true,
	"jspm_packages":    true,
	"web_modules":      true,
	"typings":          true,
	"vendors":          true,
	"third_party":      true,
	"thirdparty":       true,
	"external":         true,

	"wwwroot/lib":        true,
	"wwwroot/css":        true,
	"content/lib":        true,
	"assets/vendor":      true,
	"assets/plugins":     true,
	"assets/libs":        true,
	"static/vendor":      true,
	"public/vendor":      true,
	"wp-content/plugins": true,
	"wp-content/themes":  true,
}

func isVendorDir(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	if vendorDirs[name] {
		return true
	}

	parent := strings.ToLower(filepath.Base(filepath.Dir(path)))
	return vendorDirs[parent+"/"+name]
}
