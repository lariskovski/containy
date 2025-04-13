package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)


func createOverlayFS(lower, upper, work, merged string) error {
	// Build overlay mount options
	data := "lowerdir=" + lower + ",upperdir=" + upper + ",workdir=" + work

	// Call mount syscall directly
	err := unix.Mount("overlay", merged, "overlay", 0, data)
	if err != nil {
		return fmt.Errorf("failed to mount overlay filesystem: %v", err)
	}

	return nil
}
