package utils 

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"archive/tar"
	"compress/gzip"
)

func CreateDirectory(paths ...string) error {
	for _, path := range paths {
		// Check if the directory already exists
		if _, err := os.Stat(path); err == nil {
			// Directory exists, no need to create it
			continue
		} else if !os.IsNotExist(err) {
			// An error occurred while checking the directory
			return fmt.Errorf("failed to check directory %s: %v", path, err)
		}
		// Directory does not exist, create it
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", path, err)
		}
	}
	return nil
}

// DownloadRootFS downloads the Alpine root filesystem from the given URL and extracts it to the specified destination directory.
// download alpine root fs  https://dl-cdn.alpinelinux.org/alpine/v3.21/releases/x86_64/alpine-minirootfs-3.21.3-x86_64.tar.gz
func DownloadRootFS(url string, dest string) error{
	outputTarName := filepath.Join(dest, "alpine-minirootfs.tar.gz")
	// Check if the destination directory exists
	if _, err := os.Stat(dest); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check destination directory %s: %v", dest, err)
		}
	}	

	// Check if lower directory has no files
	if _, err := os.Stat(dest); err == nil {
		// Directory exists, check if it's empty
		files, err := os.ReadDir(dest)
		if err != nil {
			return fmt.Errorf("failed to read directory %s: %v", dest, err)
		}
		if len(files) > 0 {
			return fmt.Errorf("destination directory %s is not empty", dest)
		}
	}

	// Check if the destination directory is a valid directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dest, err)
	}
	// Download the root filesystem tarball
	if err := downloadFile(url, outputTarName); err != nil {
		return fmt.Errorf("failed to download root filesystem: %v", err)
	}
	// Check if the downloaded file is a valid tarball
	file, err := os.Open(outputTarName)
	if err != nil {
		return fmt.Errorf("failed to open downloaded file: %v", err)
	}
	defer file.Close()
	// Extract the tarball
	if err := extractTarGz(outputTarName, dest); err != nil {
		return fmt.Errorf("failed to extract root filesystem: %v", err)
	}
	// Remove the tarball after extraction
	if err := os.Remove(outputTarName); err != nil {
		return fmt.Errorf("failed to remove tarball: %v", err)
	}
	return nil
}

func downloadFile(url, dest string) error {
	// Create the file
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file: %s", resp.Status)
	}

	// Write the body to file
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	return nil
}

func extractTarGz(gzipPath, dest string) error {
	// Open the .tar.gz file
	file, err := os.Open(gzipPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
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
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create containing directory if necessary
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			// Create file
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		case tar.TypeSymlink:
			// Create symbolic link
			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}
		default:
			// Skip other file types
			fmt.Printf("Skipping unknown type: %v in %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}