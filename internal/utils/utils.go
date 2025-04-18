package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"

	"github.com/lariskovski/containy/internal/config"
)

func CreateDirectory(paths ...string) error {
	for _, path := range paths {
		config.Log.Debugf("Creating directory: %s", path)
		// Check if the directory already exists
		if _, err := os.Stat(path); err == nil {
			// Directory exists, no need to create it
			continue
		} else if !os.IsNotExist(err) {
			// An error occurred while checking the directory
			config.Log.Errorf("Failed to check directory %s: %v", path, err)
			return fmt.Errorf("failed to check directory %s: %v", path, err)
		}
		// Directory does not exist, create it
		if err := os.MkdirAll(path, 0755); err != nil {
			config.Log.Errorf("Failed to create directory %s: %v", path, err)
			return fmt.Errorf("failed to create directory %s: %v", path, err)
		}
	}
	return nil
}

// DownloadRootFS downloads the Alpine root filesystem from the given URL and extracts it to the specified destination directory.
// download alpine root fs  https://dl-cdn.alpinelinux.org/alpine/v3.21/releases/x86_64/alpine-minirootfs-3.21.3-x86_64.tar.gz
func DownloadRootFS(url string, dest string) error {
	config.Log.Debugf("Downloading root filesystem from %s to %s", url, dest)
	outputTarName := filepath.Join(dest, "alpine-minirootfs.tar.gz")
	// Check if the destination directory exists
	if _, err := os.Stat(dest); err != nil {
		if !os.IsNotExist(err) {
			config.Log.Errorf("Failed to check destination directory %s: %v", dest, err)
			return fmt.Errorf("failed to check destination directory %s: %v", dest, err)
		}
	}	

	// Check if lower directory has no files
	if _, err := os.Stat(dest); err == nil {
		// Directory exists, check if it's empty
		files, err := os.ReadDir(dest)
		if err != nil {
			config.Log.Errorf("Failed to read directory %s: %v", dest, err)
			return fmt.Errorf("failed to read directory %s: %v", dest, err)
		}
		if len(files) > 0 {
			config.Log.Errorf("Destination directory %s is not empty", dest)
			return fmt.Errorf("destination directory %s is not empty", dest)
		}
	}

	// Check if the destination directory is a valid directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		config.Log.Errorf("Failed to create directory %s: %v", dest, err)
		return fmt.Errorf("failed to create directory %s: %v", dest, err)
	}
	// Download the root filesystem tarball
	if err := downloadFile(url, outputTarName); err != nil {
		config.Log.Errorf("Failed to download root filesystem: %v", err)
		return fmt.Errorf("failed to download root filesystem: %v", err)
	}
	// Check if the downloaded file is a valid tarball
	file, err := os.Open(outputTarName)
	if err != nil {
		config.Log.Errorf("Failed to open downloaded file: %v", err)
		return fmt.Errorf("failed to open downloaded file: %v", err)
	}
	defer file.Close()
	// Extract the tarball
	if err := extractTarGz(outputTarName, dest); err != nil {
		config.Log.Errorf("Failed to extract root filesystem: %v", err)
		return fmt.Errorf("failed to extract root filesystem: %v", err)
	}
	// Remove the tarball after extraction
	if err := os.Remove(outputTarName); err != nil {
		config.Log.Errorf("Failed to remove tarball: %v", err)
		return fmt.Errorf("failed to remove tarball: %v", err)
	}
	return nil
}

func downloadFile(url, dest string) error {
	config.Log.Debugf("Downloading file from %s to %s", url, dest)
	// Create the file
	out, err := os.Create(dest)
	if err != nil {
		config.Log.Errorf("Failed to create file %s: %v", dest, err)
		return fmt.Errorf("failed to create file %s: %v", dest, err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		config.Log.Errorf("Failed to get data from %s: %v", url, err)
		return fmt.Errorf("failed to get data from %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		config.Log.Errorf("Failed to download file: %s", resp.Status)
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	// Write the body to file
	if _, err := io.Copy(out, resp.Body); err != nil {
		config.Log.Errorf("Failed to write data to file %s: %v", dest, err)
		return fmt.Errorf("failed to write data to file %s: %v", dest, err)
	}
	return nil
}

func extractTarGz(gzipPath, dest string) error {
	config.Log.Debugf("Extracting tar.gz file %s to %s", gzipPath, dest)
	// Open the .tar.gz file
	file, err := os.Open(gzipPath)
	if err != nil {
		config.Log.Errorf("Failed to open tar.gz file %s: %v", gzipPath, err)
		return fmt.Errorf("failed to open tar.gz file %s: %v", gzipPath, err)
	}
	defer file.Close()

	// Create a gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		config.Log.Errorf("Failed to create gzip reader for file %s: %v", gzipPath, err)
		return fmt.Errorf("failed to create gzip reader for file %s: %v", gzipPath, err)
	}
	defer gzr.Close()

	// Create a tar reader
	tr := tar.NewReader(gzr)

	// Iterate through the files in the archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			config.Log.Errorf("Failed to read tar header: %v", err)
			return fmt.Errorf("failed to read tar header: %v", err)
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				config.Log.Errorf("Failed to create directory %s: %v", target, err)
				return fmt.Errorf("failed to create directory %s: %v", target, err)
			}
		case tar.TypeReg:
			// Create containing directory if necessary
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				config.Log.Errorf("Failed to create directory %s: %v", filepath.Dir(target), err)
				return fmt.Errorf("failed to create directory %s: %v", filepath.Dir(target), err)
			}
			// Create file
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				config.Log.Errorf("Failed to create file %s: %v", target, err)
				return fmt.Errorf("failed to create file %s: %v", target, err)
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				config.Log.Errorf("Failed to write to file %s: %v", target, err)
				return fmt.Errorf("failed to write to file %s: %v", target, err)
			}
			outFile.Close()
		case tar.TypeSymlink:
			// Create symbolic link
			if err := os.Symlink(header.Linkname, target); err != nil {
				config.Log.Errorf("Failed to create symlink %s -> %s: %v", target, header.Linkname, err)
				return fmt.Errorf("failed to create symlink %s -> %s: %v", target, header.Linkname, err)
			}
		default:
			// Skip other file types
			config.Log.Warnf("Skipping unknown type: %v in %s", header.Typeflag, header.Name)
		}
	}

	return nil
}

func GenerateHexID(input string) string {
	length := config.IDLength
	config.Log.Debugf("Generating hex ID for input: %s", input)
	hash := sha256.Sum256([]byte(input))
	hexString := hex.EncodeToString(hash[:])
	if length > len(hexString) {
		length = len(hexString)
	}
	return hexString[:length]
}

func CheckIfLayerExists(path string) bool {
	config.Log.Debugf("Checking if layer exists at path: %s", path)
	// Check if the directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false // Directory does not exist
	}
	return true // Directory exists
}