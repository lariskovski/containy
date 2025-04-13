package main

import (
	"fmt"
	"os"

	"github.com/lariskovski/containy/internal/overlay"
	"github.com/lariskovski/containy/internal/utils"
)

var rootfsUrl = "https://dl-cdn.alpinelinux.org/alpine/v3.21/releases/x86_64/alpine-minirootfs-3.21.3-x86_64.tar.gz"

func main() {
    err := utils.CreateDirectory("lower", "upper", "work", "merged")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating directories: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Directories created successfully.")

    // Download the root filesystem
    err = utils.DownloadRootFS(rootfsUrl, "lower")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error downloading root filesystem: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Root filesystem downloaded successfully.")

    // Create the overlay filesystem
    overlayFS := &overlay.OverlayFS{
        LowerDir:  "lower",
        UpperDir:  "upper",
        WorkDir:   "work",
        MergedDir: "merged",
    }
    err = overlayFS.Mount()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating overlay filesystem: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Overlay filesystem created successfully.")

}