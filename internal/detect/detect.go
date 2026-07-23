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

// Dedupe puts two detections of the same thing onto one row. It has to run again
// after versions are resolved to catalog lines, because resolution is what makes
// two differently worded detections identical: two composer constraints under one
// major, or the same site found both in the scanned directory and under a web
// root. Two rows sharing a name and version also share a tracking key, so
// unticking one of them would delete the technology the other still represents.
func Dedupe(result *Result) {
	result.Technologies = dedupeTechnologies(result.Technologies)
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

// dedupeTechnologies keeps the first entry for a name and version, because the
// earlier one describes the scanned directory and the later one only the machine.
// The exact build is the exception: whichever detection knows it wins, since a
// Dockerfile naming the host's own release would otherwise shadow the
// /etc/os-release entry and silently drop the running kernel, and the same
// applies to a WordPress install reached both ways.
func dedupeTechnologies(techs []Technology) []Technology {
	at := make(map[string]int)
	out := make([]Technology, 0, len(techs))
	for _, t := range techs {
		key := t.Name + "|" + t.Version
		if i, seen := at[key]; seen {
			if out[i].Kernel == "" {
				out[i].Kernel = t.Kernel
			}
			continue
		}
		at[key] = len(out)
		out = append(out, t)
	}
	return out
}
