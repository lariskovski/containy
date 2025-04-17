package overlay

import (
	"fmt"
	// "path/filepath"
	// "strings"

	"github.com/lariskovski/containy/internal/utils"
	"golang.org/x/sys/unix"
)

var baseOverlayDir = "tmp/build/layers/"
var IDLength = 10

type OverlayFS struct {
	Instruction string
	ID string
	LowerDir string
	UpperDir string
	WorkDir  string
	MergedDir string
}

func (o *OverlayFS) Setup() (*OverlayFS, error) {
	baseDir := baseOverlayDir + o.ID + "/"
	// Create lowerDir only if the instruction is FROM
	if o.Instruction == "FROM" {
		o.LowerDir = baseDir + o.LowerDir
		if err := utils.CreateDirectory(o.LowerDir); err != nil {
			return nil, fmt.Errorf("failed to create lowerDir overlay directory: %v", err)
		}
	}
	o.UpperDir = baseDir + o.UpperDir
	o.WorkDir = baseDir + o.WorkDir
	o.MergedDir = baseDir + o.MergedDir

	// Create directories if they don't exist
	if err := utils.CreateDirectory(o.UpperDir, o.WorkDir, o.MergedDir); err != nil {
		return nil, fmt.Errorf("failed to create overlay directories: %v", err)
	}
	return o, nil

}

func (o *OverlayFS) Mount() error {
	// // Ensure all paths are absolute
	// lowerDirs := strings.Split(o.LowerDir, ":")
	// for i, dir := range lowerDirs {
	// 	absDir, err := filepath.Abs(dir)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to get absolute path for lowerdir: %v", err)
	// 	}
	// 	lowerDirs[i] = absDir
	// }
	// o.LowerDir = strings.Join(lowerDirs, ":")

	// Build overlay mount options
	data := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", o.LowerDir, o.UpperDir, o.WorkDir)

	// Call mount syscall directly
	err := unix.Mount("overlay", o.MergedDir, "overlay", 0, data)
	if err != nil {
		return fmt.Errorf("failed to mount overlay filesystem: %v - data: %s", err, data)
	}

	return nil
}
