package overlay

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lariskovski/containy/internal/config"
	"golang.org/x/sys/unix"
)

// OverlayFS implements a Linux overlay filesystem abstraction for container layers.
// It manages the directories required for overlay mounting (lower, upper, work, merged)
// and provides operations to create and mount overlay filesystems.
//
// Overlay filesystems are used to implement container image layering, where each
// instruction in a build file creates a new filesystem layer that can see the content
// of all previous layers but only modifies its own "upper" directory.
type OverlayFS struct {
	// ID is the unique identifier for this layer
	ID string

	// LowerDir contains the read-only base files (may be a colon-separated list for multiple lower dirs)
	LowerDir string

	// UpperDir stores modifications made in this layer
	UpperDir string

	// WorkDir is used by overlayfs for internal operations
	WorkDir string

	// MergedDir is where the combined view of lower and upper is mounted
	MergedDir string
}

// NewOverlayFS creates and sets up a new OverlayFS instance.
// It prepares the directory structure needed for an overlay filesystem
// but does not mount it (Mount() must be called separately).
//
// For base layers (created by FROM instructions), it also creates a new
// lower directory. For derived layers, it uses the provided lower directory.
//
// Parameters:
//   - lowerDir: Path to the read-only base directory/directories
//     (for derived layers) or empty (for base layers)
//   - id: Unique identifier for this layer
//   - isBaseLayer: true if this is a base layer (FROM instruction),
//     false for derived layers (RUN, COPY, etc.)
//
// Returns:
//   - *OverlayFS: The created overlay filesystem instance
//   - error: Any error encountered during setup
func NewOverlayFS(lowerDir, id string) (*OverlayFS, error) {
	config.Log.Debugf("Creating new overlay filesystem with ID: %s", id)

	baseDir := config.BaseOverlayDir + id + "/"
	upperDir := baseDir + "upper"
	workDir := baseDir + "work"
	mergedDir := baseDir + "merged"

	// For FROM instructions, create a lowerDir
	if lowerDir == "" {
		lowerDir = baseDir + "lower"
		if err := createDirectory(lowerDir); err != nil {
			return nil, fmt.Errorf("failed to create lowerDir overlay directory: %w", err)
		}
	}

	overlay := &OverlayFS{
		ID:        id,
		LowerDir:  lowerDir,
		UpperDir:  upperDir,
		WorkDir:   workDir,
		MergedDir: mergedDir,
	}

	if err := createDirectory(upperDir, workDir, mergedDir); err != nil {
		return nil, fmt.Errorf("failed to create overlay directories: %w", err)
	}

	return overlay, nil
}

// Mount creates an overlay mount that combines the lower and upper directories
// into a unified view at the merged directory location.
//
// The overlay mount uses:
// - Lower directory: Read-only base layer(s)
// - Upper directory: Read-write layer to capture changes
// - Work directory: Used internally by overlayfs
// - Merged directory: The mount point where the unified view is presented
//
// Returns:
//   - error: Any error encountered during mounting
func (o *OverlayFS) Mount() error {
	config.Log.Debugf("Mounting overlay filesystem at %s", o.MergedDir)
	// Build overlay mount options
	data := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", o.LowerDir, o.UpperDir, o.WorkDir)
	config.Log.Debugf("Mount options: %s", data)
	// Call mount syscall directly
	err := unix.Mount("overlay", o.MergedDir, "overlay", 0, data)
	if err != nil {
		return fmt.Errorf("failed to mount overlay filesystem: %w", err)
	}

	return nil
}

func (o *OverlayFS) CreateAlias(alias string) error {
	// Check if the alias already exists
	if _, err := os.Stat(filepath.Join(config.AliasDir, alias)); !os.IsNotExist(err) {
		return fmt.Errorf("alias %s already exists", alias)
	}
	// Create the alias directory if it doesn't exist
	if err := os.MkdirAll(config.AliasDir, 0755); err != nil {
		return fmt.Errorf("failed to create alias directory: %v", err)
	}

	// Create a symbolic link to the merged directory
	err := os.Symlink(o.MergedDir, filepath.Join(config.AliasDir, alias))
	if err != nil {
		return fmt.Errorf("failed to create alias %s: %v", alias, err)
	}

	return nil
}

// GetID returns the unique identifier for this layer.
func (o *OverlayFS) GetID() string { return o.ID }

// GetLowerDir returns the path to the read-only base directory/directories.
func (o *OverlayFS) GetLowerDir() string { return o.LowerDir }

// GetUpperDir returns the path to the read-write directory that stores changes.
func (o *OverlayFS) GetUpperDir() string { return o.UpperDir }

// GetWorkDir returns the path to the directory used by overlayfs for internal operations.
func (o *OverlayFS) GetWorkDir() string { return o.WorkDir }

// GetMergedDir returns the path to the directory where the unified filesystem view is mounted.
func (o *OverlayFS) GetMergedDir() string { return o.MergedDir }
