package install

import (
	"os"
	"path/filepath"
	"testing"

	config "github.com/Sabique-Islam/catalyst/internal/config"
)

func TestDownloadResource(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test", "file.txt")

	// Test downloading a simple text file (using a reliable public URL)
	url := "https://httpbin.org/uuid"

	err := DownloadResource(url, testFile)
	if err != nil {
		t.Fatalf("Failed to download resource: %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatalf("Downloaded file does not exist: %s", testFile)
	}

	// Check if file has content
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat downloaded file: %v", err)
	}

	if info.Size() == 0 {
		t.Fatalf("Downloaded file is empty")
	}

	t.Logf("Successfully downloaded file: %s (size: %d bytes)", testFile, info.Size())
}

func TestDownloadResourceAlreadyExists(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "existing.txt")

	// Create an existing file
	err := os.MkdirAll(filepath.Dir(testFile), 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	existingContent := "existing content"
	err = os.WriteFile(testFile, []byte(existingContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Try to download to the same location
	url := "https://httpbin.org/uuid"
	err = DownloadResource(url, testFile)
	if err != nil {
		t.Fatalf("Failed to handle existing file: %v", err)
	}

	// Check that the existing file content is preserved
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file after download attempt: %v", err)
	}

	if string(content) != existingContent {
		t.Fatalf("Existing file content was modified. Expected: %s, Got: %s", existingContent, string(content))
	}

	t.Log("Correctly skipped downloading to existing file")
}

func TestInstallResources(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test config with resources
	cfg := &config.Config{
		Resources: []config.Resource{
			{
				URL:  "https://httpbin.org/uuid",
				Path: filepath.Join(tempDir, "resource1.json"),
			},
			{
				URL:  "https://httpbin.org/base64/SFRUUEJJTiBpcyBhd2Vzb21l",
				Path: filepath.Join(tempDir, "data", "resource2.txt"),
			},
		},
	}

	err := InstallResources(cfg)
	if err != nil {
		t.Fatalf("Failed to install resources: %v", err)
	}

	// Check if both files were created
	for _, resource := range cfg.Resources {
		if _, err := os.Stat(resource.Path); os.IsNotExist(err) {
			t.Fatalf("Resource file does not exist: %s", resource.Path)
		}
		t.Logf("Successfully created resource: %s", resource.Path)
	}
}
