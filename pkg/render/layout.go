package render

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sigs.k8s.io/yaml"
)

const (
	metadataFileName         = "metadata.yaml"
	templateFragmentFileName = "addon-template.yaml"
	staticManifestFileName   = "addon.yaml"
	readmeFileName           = "README.md"

	StageExperimental = "experimental"
	StagePreview      = "preview"
	StageGA           = "ga"

	labelExperimental = "addon.harvesterhci.io/experimental"
	labelPreview      = "addon.harvesterhci.io/preview"
	labelDeprecated   = "addon.harvesterhci.io/deprecated"
)

// Metadata describes one addon directory. Stage (maturity/support) and BuiltIn
// (ISO packaging) are orthogonal: BuiltIn is a usage-driven packaging decision.
type Metadata struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Stage      string `json:"stage"`
	BuiltIn    bool   `json:"builtIn"`
	Deprecated bool   `json:"deprecated,omitempty"`
}

type AddonDir struct {
	Path string
	Meta Metadata
}

// LoadAddonDirs reads every direct child directory of root that contains a
// metadata.yaml and returns them sorted by addon name.
func LoadAddonDirs(root string) ([]AddonDir, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("error reading addons root %s: %v", root, err)
	}

	var dirs []AddonDir
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(root, e.Name())
		metaFile := filepath.Join(path, metadataFileName)
		contents, err := os.ReadFile(metaFile)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("addon directory %s has no %s", path, metadataFileName)
			}
			return nil, fmt.Errorf("error reading %s: %v", metaFile, err)
		}
		meta := Metadata{}
		if err := yaml.UnmarshalStrict(contents, &meta); err != nil {
			return nil, fmt.Errorf("error parsing %s: %v", metaFile, err)
		}
		dirs = append(dirs, AddonDir{Path: path, Meta: meta})
	}

	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Meta.Name < dirs[j].Meta.Name })
	return dirs, nil
}

// AssembleTemplate concatenates the addon-template.yaml fragments of all
// builtIn addons (sorted by name) into the single rancherd bootstrap template.
// Fragments are stored without list indentation; each is re-indented as one
// item of the top-level `resources:` list.
func AssembleTemplate(root string) ([]byte, error) {
	dirs, err := LoadAddonDirs(root)
	if err != nil {
		return nil, err
	}

	var b strings.Builder
	b.WriteString("resources:\n")
	count := 0
	for _, d := range dirs {
		if !d.Meta.BuiltIn {
			continue
		}
		fragFile := filepath.Join(d.Path, templateFragmentFileName)
		contents, err := os.ReadFile(fragFile)
		if err != nil {
			return nil, fmt.Errorf("error reading %s: %v", fragFile, err)
		}
		for i, line := range strings.Split(strings.TrimRight(string(contents), "\n"), "\n") {
			switch {
			case i == 0:
				b.WriteString("  - " + line + "\n")
			case line == "":
				b.WriteString("\n")
			default:
				b.WriteString("    " + line + "\n")
			}
		}
		count++
	}
	if count == 0 {
		return nil, fmt.Errorf("no builtIn addon found under %s", root)
	}
	return []byte(b.String()), nil
}
