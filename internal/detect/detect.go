package detect

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
)

func IsHostSource(source string) bool {
	switch source {
	case SourceOsRelease, SourceHostKern, SourceHost:
		return true
	default:
		return false
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

// ScanWithProgress reads the host first because that part is instant, then calls
// onHostDone before walking the tree, which is the slow part and the one that goes
// looking for WordPress and the like. Without that hook the caller has nothing to
// say while a large directory is being walked and the scan looks hung.
//
// Host detection runs first but its technologies are appended LAST, so a tree
// detection still wins the dedupe exactly as it did when the walk ran first. That
// matters because the source decides whether an entry defaults to selected: a
// Dockerfile-declared Ubuntu must not turn into a host detection just because the
// order of work changed.
func ScanWithProgress(root string, onHostDone func()) (*Result, error) {
	host := &Result{}
	scanHost(host)

	if onHostDone != nil {
		onHostDone()
	}

	tree := &Result{}
	if err := scanTree(root, tree); err != nil {
		return nil, err
	}

	result := &Result{
		Technologies: append(tree.Technologies, host.Technologies...),
		Manifests:    tree.Manifests,
	}

	result.Technologies = dedupeTechnologies(result.Technologies)
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
