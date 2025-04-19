package overlay

import (
	"fmt"

	"github.com/lariskovski/containy/internal/config"
	"github.com/lariskovski/containy/internal/utils"
	"golang.org/x/sys/unix"
)

type OverlayFS struct {
	ID        string
	LowerDir  string
	UpperDir  string
	WorkDir   string
	MergedDir string
}

// NewOverlayFS creates and sets up a new OverlayFS instance
func NewOverlayFS(lowerDir, id string, isBaseLayer bool) (*OverlayFS, error) {
	config.Log.Debugf("Creating new overlay filesystem with ID: %s", id)

	baseDir := config.BaseOverlayDir + id + "/"
	upperDir := baseDir + "upper"
	workDir := baseDir + "work"
	mergedDir := baseDir + "merged"

	// For FROM instructions, create a lowerDir
	if isBaseLayer {
		lowerDir = baseDir + "lower"
		if err := utils.CreateDirectory(lowerDir); err != nil {
			config.Log.Errorf("Failed to create lowerDir overlay directory: %v", err)
			return nil, fmt.Errorf("failed to create lowerDir overlay directory: %v", err)
		}
	}

	overlay := &OverlayFS{
		ID:        id,
		LowerDir:  lowerDir,
		UpperDir:  upperDir,
		WorkDir:   workDir,
		MergedDir: mergedDir,
	}

	if err := utils.CreateDirectory(upperDir, workDir, mergedDir); err != nil {
		config.Log.Errorf("Failed to create overlay directories: %v", err)
		return nil, fmt.Errorf("failed to create overlay directories: %v", err)
	}

	return overlay, nil
}

func (o *OverlayFS) Mount() error {
	config.Log.Debugf("Mounting overlay filesystem at %s", o.MergedDir)
	// Build overlay mount options
	data := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", o.LowerDir, o.UpperDir, o.WorkDir)
	config.Log.Debugf("Mount options: %s", data)
	// Call mount syscall directly
	err := unix.Mount("overlay", o.MergedDir, "overlay", 0, data)
	if err != nil {
		config.Log.Errorf("Failed to mount overlay filesystem: %v", err)
		return fmt.Errorf("failed to mount overlay filesystem: %v", err)
	}

	return nil
}

// Implement the instructions.Layer interface
func (o *OverlayFS) GetID() string        { return o.ID }
func (o *OverlayFS) GetLowerDir() string  { return o.LowerDir }
func (o *OverlayFS) GetUpperDir() string  { return o.UpperDir }
func (o *OverlayFS) GetWorkDir() string   { return o.WorkDir }
func (o *OverlayFS) GetMergedDir() string { return o.MergedDir }
