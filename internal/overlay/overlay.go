package overlay

import (
	"fmt"

	"golang.org/x/sys/unix"
)

type OverlayFS struct {
	LowerDir string
	UpperDir string
	WorkDir  string
	MergedDir string
}

func (o *OverlayFS) Mount() error {
	// Build overlay mount options
	data := "lowerdir=" + o.LowerDir + ",upperdir=" + o.UpperDir + ",workdir=" + o.WorkDir

	// Call mount syscall directly
	err := unix.Mount("overlay", o.MergedDir, "overlay", 0, data)
	if err != nil {
		return fmt.Errorf("failed to mount overlay filesystem: %v", err)
	}

	return nil
}
