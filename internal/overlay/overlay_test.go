package overlay

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lariskovski/containy/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewOverlayFS_BaseLayer(t *testing.T) {
	tmpDir := t.TempDir()
	config.BaseOverlayDir = tmpDir + "/"

	ofs, err := NewOverlayFS("", "test-id", true)
	assert.NoError(t, err)
	assert.NotNil(t, ofs)
	assert.DirExists(t, ofs.GetLowerDir())
	assert.DirExists(t, ofs.GetUpperDir())
	assert.DirExists(t, ofs.GetWorkDir())
	assert.DirExists(t, ofs.GetMergedDir())
}

func TestNewOverlayFS_DerivedLayer(t *testing.T) {
	tmpDir := t.TempDir()
	config.BaseOverlayDir = tmpDir + "/"

	lower := tmpDir + "/base"
	err := os.MkdirAll(lower, 0755)
	assert.NoError(t, err)

	ofs, err := NewOverlayFS(lower, "derived-id", false)
	assert.NoError(t, err)
	assert.Equal(t, lower, ofs.GetLowerDir())
	assert.DirExists(t, ofs.GetUpperDir())
	assert.DirExists(t, ofs.GetWorkDir())
	assert.DirExists(t, ofs.GetMergedDir())
}

func TestCheckIfLayerExists(t *testing.T) {
	tmpDir := t.TempDir()
	config.BaseOverlayDir = tmpDir + "/"

	id := "layer1"
	assert.False(t, CheckIfLayerExists(id))

	err := os.MkdirAll(filepath.Join(tmpDir, id), 0755)
	assert.NoError(t, err)
	assert.True(t, CheckIfLayerExists(id))
}

func TestCreateDirectory(t *testing.T) {
	dir := t.TempDir() + "/newdir"
	err := createDirectory(dir)
	assert.NoError(t, err)
	assert.DirExists(t, dir)

	// Should not error if already exists
	err = createDirectory(dir)
	assert.NoError(t, err)
}

func TestCreateAlias(t *testing.T) {
	tmpDir := t.TempDir()
	config.AliasDir = filepath.Join(tmpDir, "aliases")

	merged := filepath.Join(tmpDir, "merged")
	err := os.MkdirAll(merged, 0755)
	assert.NoError(t, err)

	ofs := &OverlayFS{
		ID:        "test",
		MergedDir: merged,
	}

	// First time should succeed
	err = ofs.CreateAlias("alias1")
	assert.NoError(t, err)

	// Second time should fail
	err = ofs.CreateAlias("alias1")
	assert.Error(t, err)

	// Symlink should exist
	link := filepath.Join(config.AliasDir, "alias1")
	_, err = os.Lstat(link)
	assert.NoError(t, err)
}
