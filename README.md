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
$ sudo go run main.go build examples/TainyFile --alias test
```

### Run a Container
To run a container from an alias:
```bash
$ sudo go run main.go run test sh
```

## Requirements
- Go 1.23.4 or higher.
- Root privileges to execute container operations.

## Cleanup
To unmount all overlay filesystems and clean up temporary files:
```bash
$ ./umount.sh
```
