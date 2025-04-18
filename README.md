# Containy

Containy is a lightweight container runtime written in Go for study purposes. It allows you to build and run containerized environments using a simple overlay filesystem and custom instructions.

## Features
- Parse and execute custom container instructions from a file.
- Build container layers using an overlay filesystem.
- Run commands inside isolated container environments.

## Usage

### Build a Container
To build a container from a file (e.g., `TainyFile`):
```bash
$ sudo go run cmd/main.go build TainyFile
```

### Run a Container
To run a container from a specific layer:
```bash
$ sudo go run cmd/main.go run tmp/build/layers/<layer-id>/merged sh
```

Replace `<layer-id>` with the actual ID of the layer you want to run.

## Requirements
- Go 1.23.4 or higher.
- Root privileges to execute container operations.

## Cleanup
To unmount all overlay filesystems and clean up temporary files:
```bash
$ ./umount.sh
```
## Todo

- [ ] Create aliases for layers and use them on run command
- [ ] Add Cobra CLI
- [ ] Improve logging
- [ ] Better patterns and less verbose output
- [ ] Add networking ns and bridge setup