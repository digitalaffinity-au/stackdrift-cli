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
	result := &Result{}

	if err := scanTree(root, result); err != nil {
		return nil, err
	}

	scanHost(result)

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
