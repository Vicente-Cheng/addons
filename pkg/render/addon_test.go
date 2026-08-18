package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Addon(t *testing.T) {
	assert := require.New(t)
	tmpDir := os.TempDir()
	tmpPath, err := os.MkdirTemp(tmpDir, "rendered")
	assert.NoError(err)
	defer os.RemoveAll(tmpPath)
	err = Addon(relativeAddonsPath, tmpPath, filepath.Join(relativeVersionFilePath, defaultVersionFile))
	assert.NoError(err)
}

func Test_generate_version_info_map(t *testing.T) {
	assert := require.New(t)
	result, err := generate_version_info_map(filepath.Join(relativeVersionFilePath, defaultVersionFile))
	assert.NoError(err, "expected no error while reading version info map")
	_, ok := result["VM_IMPORT_CONTROLLER_CHART_VERSION"]
	assert.True(ok, "expected to find key VM_IMPORT_CONTROLLER_CHART_VERSION")
}

func Test_Template(t *testing.T) {
	assert := require.New(t)
	tmpDir := os.TempDir()
	tmpPath, err := os.MkdirTemp(tmpDir, "template")
	assert.NoError(err, "expected no error during tmp dir creation")
	defer os.RemoveAll(tmpPath)
	err = Template(relativeAddonsPath, tmpPath, filepath.Join(relativeVersionFilePath, defaultVersionFile))
	assert.NoError(err)
}

func Test_LoadAddonDirs(t *testing.T) {
	assert := require.New(t)
	dirs, err := LoadAddonDirs(relativeAddonsPath)
	assert.NoError(err)
	assert.NotEmpty(dirs)
	for i, d := range dirs {
		assert.Equal(d.Meta.Name, filepath.Base(d.Path))
		if i > 0 {
			assert.Less(dirs[i-1].Meta.Name, d.Meta.Name, "expected addons sorted by name")
		}
	}
}

func Test_AssembleTemplate(t *testing.T) {
	assert := require.New(t)
	contents, err := AssembleTemplate(relativeAddonsPath)
	assert.NoError(err)
	assembled := string(contents)
	assert.True(strings.HasPrefix(assembled, "resources:\n"))

	dirs, err := LoadAddonDirs(relativeAddonsPath)
	assert.NoError(err)
	builtIn := 0
	for _, d := range dirs {
		if d.Meta.BuiltIn {
			builtIn++
		}
	}
	assert.Equal(builtIn, strings.Count(assembled, "kind: Addon"),
		"expected one resource per builtIn addon")
}

func Test_Validate(t *testing.T) {
	assert := require.New(t)
	err := Validate(relativeAddonsPath, filepath.Join(relativeVersionFilePath, defaultVersionFile))
	assert.NoError(err)
}
