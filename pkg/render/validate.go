package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

var validStages = map[string]bool{
	StageExperimental: true,
	StagePreview:      true,
	StageGA:           true,
}

// Validate enforces the addon layout invariants:
//   - metadata.yaml is well-formed, name matches the directory, stage is one of
//     experimental/preview/ga, namespace is set, names are unique
//   - deprecated: true requires builtIn: false
//   - every addon directory has a README.md
//   - builtIn addons have addon-template.yaml (and no addon.yaml); the
//     assembled template renders with no missing variables, and the rendered
//     resources match metadata (name, namespace, stage labels)
//   - non-builtIn addons have a static, fully-rendered addon.yaml (and no
//     addon-template.yaml) whose Addon resource matches metadata
func Validate(root, versionFilePath string) error {
	dirs, err := LoadAddonDirs(root)
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		return fmt.Errorf("no addon directory found under %s", root)
	}

	seen := map[string]bool{}
	hasBuiltIn := false
	for _, d := range dirs {
		if err := validateDir(d); err != nil {
			return err
		}
		if seen[d.Meta.Name] {
			return fmt.Errorf("duplicate addon name %s", d.Meta.Name)
		}
		seen[d.Meta.Name] = true
		if d.Meta.BuiltIn {
			hasBuiltIn = true
		} else if err := validateStaticManifest(d); err != nil {
			return err
		}
	}

	if hasBuiltIn {
		if err := validateRenderedResources(root, versionFilePath, dirs); err != nil {
			return err
		}
	}
	return nil
}

func validateDir(d AddonDir) error {
	m := d.Meta
	if m.Name == "" || m.Namespace == "" {
		return fmt.Errorf("%s: metadata name and namespace must be set", d.Path)
	}
	if m.Name != filepath.Base(d.Path) {
		return fmt.Errorf("%s: metadata name %s does not match directory name", d.Path, m.Name)
	}
	if !validStages[m.Stage] {
		return fmt.Errorf("%s: invalid stage %q (want experimental, preview or ga)", d.Path, m.Stage)
	}
	if m.Deprecated && m.BuiltIn {
		return fmt.Errorf("%s: a deprecated addon cannot be builtIn", d.Path)
	}
	if _, err := os.Stat(filepath.Join(d.Path, readmeFileName)); err != nil {
		return fmt.Errorf("%s: missing %s", d.Path, readmeFileName)
	}

	fragmentExists := fileExists(filepath.Join(d.Path, templateFragmentFileName))
	staticExists := fileExists(filepath.Join(d.Path, staticManifestFileName))
	if m.BuiltIn && (!fragmentExists || staticExists) {
		return fmt.Errorf("%s: a builtIn addon must have %s and no %s", d.Path, templateFragmentFileName, staticManifestFileName)
	}
	if !m.BuiltIn && (fragmentExists || !staticExists) {
		return fmt.Errorf("%s: a non-builtIn addon must have %s and no %s", d.Path, staticManifestFileName, templateFragmentFileName)
	}
	return nil
}

// validateStaticManifest checks that a non-builtIn addon.yaml is valid,
// fully-rendered YAML containing exactly one Addon resource that matches the
// addon metadata.
func validateStaticManifest(d AddonDir) error {
	manifest := filepath.Join(d.Path, staticManifestFileName)
	contents, err := os.ReadFile(manifest)
	if err != nil {
		return fmt.Errorf("error reading %s: %v", manifest, err)
	}

	var addonDocs []map[string]interface{}
	for _, doc := range strings.Split(string(contents), "\n---") {
		if strings.TrimSpace(doc) == "" {
			continue
		}
		parsed := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(doc), &parsed); err != nil {
			return fmt.Errorf("%s: invalid YAML: %v", manifest, err)
		}
		if parsed["kind"] == "Addon" {
			addonDocs = append(addonDocs, parsed)
		}
	}
	if len(addonDocs) != 1 {
		return fmt.Errorf("%s: expected exactly one Addon resource, found %d", manifest, len(addonDocs))
	}
	return matchResource(d.Meta, addonDocs[0], manifest)
}

// validateRenderedResources fully renders the builtIn addons and checks each
// rendered resource against its metadata.
func validateRenderedResources(root, versionFilePath string, dirs []AddonDir) error {
	resources, err := renderResources(root, versionFilePath)
	if err != nil {
		return err
	}

	byName := map[string]map[string]interface{}{}
	for _, r := range resources.Resources {
		metadata, _ := r["metadata"].(map[string]interface{})
		name, _ := metadata["name"].(string)
		byName[name] = r
	}

	for _, d := range dirs {
		if !d.Meta.BuiltIn {
			continue
		}
		r, ok := byName[d.Meta.Name]
		if !ok {
			return fmt.Errorf("%s: builtIn addon missing from rendered template", d.Path)
		}
		if err := matchResource(d.Meta, r, d.Path); err != nil {
			return err
		}
	}
	return nil
}

// matchResource checks an Addon resource against the addon metadata: name,
// namespace, and the stage/deprecated labels derived from metadata.
func matchResource(m Metadata, resource map[string]interface{}, source string) error {
	metadata, _ := resource["metadata"].(map[string]interface{})
	if metadata["name"] != m.Name {
		return fmt.Errorf("%s: Addon resource name %v does not match metadata name %s", source, metadata["name"], m.Name)
	}
	if metadata["namespace"] != m.Namespace {
		return fmt.Errorf("%s: Addon resource namespace %v does not match metadata namespace %s", source, metadata["namespace"], m.Namespace)
	}

	labels := map[string]interface{}{}
	if l, ok := metadata["labels"].(map[string]interface{}); ok {
		labels = l
	}
	expected := map[string]bool{
		labelExperimental: m.Stage == StageExperimental,
		labelPreview:      m.Stage == StagePreview,
		labelDeprecated:   m.Deprecated,
	}
	for label, want := range expected {
		_, present := labels[label]
		if want && labels[label] != "true" {
			return fmt.Errorf("%s: label %s: \"true\" is required by metadata", source, label)
		}
		if !want && present {
			return fmt.Errorf("%s: label %s must not be set according to metadata", source, label)
		}
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
