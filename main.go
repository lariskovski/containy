package main

import (
    "fmt"
    "os"
)

var rootfsUrl = "https://dl-cdn.alpinelinux.org/alpine/v3.21/releases/x86_64/alpine-minirootfs-3.21.3-x86_64.tar.gz"
func main() {
    err := CreateDirectory("lower", "upper", "work", "merged")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating directories: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Directories created successfully.")

    err = DownloadRootFS(rootfsUrl, "lower")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error downloading root filesystem: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Root filesystem downloaded successfully.")

    err = createOverlayFS("lower", "upper", "work", "merged")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating overlay filesystem: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("Overlay filesystem created successfully.")

}