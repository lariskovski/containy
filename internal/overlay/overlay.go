package overlay

import (
	"fmt"

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

func (o *OverlayFS) Setup(instruction string) (*OverlayFS, error) {
	baseDir := baseOverlayDir + utils.GenerateHexID(instruction, IDLength) + "/"
	// Create lowerDir only if the instruction is FROM
	if o.Instruction == "FROM" {
		o.LowerDir = baseDir + o.LowerDir
	}
	o.UpperDir = baseDir + o.UpperDir
	o.WorkDir = baseDir + o.WorkDir
	o.MergedDir = baseDir + o.MergedDir

	return o, nil

}

func (o *OverlayFS) Mount() error {
	// Create directories if they don't exist
	if err := utils.CreateDirectory(o.LowerDir, o.UpperDir, o.WorkDir, o.MergedDir); err != nil {
		return fmt.Errorf("failed to create overlay directories: %v", err)
	}
	// Build overlay mount options
	// data := "lowerdir=" + o.LowerDir + ",upperdir=" + o.UpperDir + ",workdir=" + o.WorkDir
	data := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", o.LowerDir, o.UpperDir, o.WorkDir)

	// Call mount syscall directly
	err := unix.Mount("overlay", o.MergedDir, "overlay", 0, data)
	if err != nil {
		return fmt.Errorf("failed to mount overlay filesystem: %v", err)
	}

	return nil
}
