package detect

type Technology struct {
	Name     string
	Version  string
	Category string
	Source   string
}

type Manifest struct {
	Ecosystem string
	Path      string
	FileName  string
	Content   string
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
