package detect

import "strings"

type Technology struct {
	Name     string
	Version  string
	Kernel   string
	Category string
	Source   string
}

const (
	SourceOsRelease = "/etc/os-release"
	SourceHostKern  = "host kernel"
	SourceHost      = "host"
	// Prefixes a path found on the machine rather than inside the scanned
	// directory, so the hint still names the install while the entry is still
	// treated as a host detection.
	SourceHostPrefix = "host: "
)

func IsHostSource(source string) bool {
	switch source {
	case SourceOsRelease, SourceHostKern, SourceHost:
		return true
	default:
		return strings.HasPrefix(source, SourceHostPrefix)
	}
}

type Manifest struct {
	Ecosystem string
	Path      string
	FileName  string
	Content   string
	Primary   bool
}

type Result struct {
	Technologies []Technology
	Manifests    []Manifest
}

var skipDirs = map[string]bool{
	"node_modules": true,
	"bin":          true,
	"obj":          true,
	".git":         true,
	"vendor":       true,
	"dist":         true,
	"build":        true,
	".vs":          true,
	".idea":        true,
	"packages":     true,
	"publish":      true,
	"docs":         true,
}

func Scan(root string) (*Result, error) {
	return ScanWithProgress(root, nil)
}

// ScanWithProgress reads the OS and kernel first because that part is instant,
// then calls onHostDone before the two slow parts: searching the machine's web
// roots and walking the scanned directory. Without that hook the caller has
// nothing to say while they run and the scan looks hung.
//
// Technologies are appended tree first, then web roots, then host, so a detection
// made inside the scanned directory keeps winning the dedupe. That matters
// because the source decides whether an entry defaults to selected: a WordPress
// install in the directory being scanned is the project, whereas the same
// install found under a web root describes the machine.
func ScanWithProgress(root string, onHostDone func()) (*Result, error) {
	host := &Result{}
	scanHost(host)

	if onHostDone != nil {
		onHostDone()
	}

	web := &Result{}
	scanWebRoots(web)

	tree := &Result{}
	if err := scanTree(root, tree); err != nil {
		return nil, err
	}

	technologies := append(tree.Technologies, web.Technologies...)
	technologies = append(technologies, host.Technologies...)

	result := &Result{
		Technologies: dedupeTechnologies(technologies),
		Manifests:    tree.Manifests,
	}

	return result, nil
}

func dedupeTechnologies(techs []Technology) []Technology {
	seen := make(map[string]bool)
	out := make([]Technology, 0, len(techs))
	for _, t := range techs {
		key := t.Name + "|" + t.Version
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, t)
	}
	return out
}
