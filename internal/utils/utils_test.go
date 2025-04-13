package utils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	dirPath := filepath.Join(tempDir, "testdir")

	err := CreateDirectory(dirPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Fatalf("expected directory to exist, but it does not")
	}
}

func TestDownloadRootFS(t *testing.T) {
	tempDir := t.TempDir()
	url := "https://dl-cdn.alpinelinux.org/alpine/v3.21/releases/x86_64/alpine-minirootfs-3.21.3-x86_64.tar.gz"

	err := DownloadRootFS(url, tempDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check if files were extracted
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("expected files to be extracted, but directory is empty")
	}
}

func TestDownloadFile(t *testing.T) {
	tempDir := t.TempDir()
	dest := filepath.Join(tempDir, "testfile")
	url := "https://www.google.com/robots.txt"

	err := downloadFile(url, dest)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Fatalf("expected file to exist, but it does not")
	}

	// Check if file is not empty
	fileInfo, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if fileInfo.Size() == 0 {
		t.Fatalf("expected file to have content, but it is empty")
	}
}

func TestExtractTarGz(t *testing.T) {
	tempDir := t.TempDir()
	tarGzPath := filepath.Join(tempDir, "test.tar.gz")
	extractDir := filepath.Join(tempDir, "extracted")

	// Create a sample tar.gz file
	err := createSampleTarGz(tarGzPath)
	if err != nil {
		t.Fatalf("failed to create sample tar.gz: %v", err)
	}

	// Extract the tar.gz file
	err = extractTarGz(tarGzPath, extractDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check if files were extracted
	files, err := os.ReadDir(extractDir)
	if err != nil {
		t.Fatalf("failed to read directory: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("expected files to be extracted, but directory is empty")
	}
}

// Helper function to create a sample tar.gz file
func createSampleTarGz(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add a sample file to the tarball
	content := "hello world"
	header := &tar.Header{
		Name: "sample.txt",
		Mode: 0600,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	if _, err := io.WriteString(tw, content); err != nil {
		return err
	}

	return nil
}
