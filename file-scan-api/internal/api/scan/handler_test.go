package scan

import (
	"os"
	"testing"
)

// TestValidateFileSize tests the validateFileSize function
func TestValidateFileSize(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write data to the file to make it within the size limit
	if _, err := tmpFile.Write(make([]byte, fileMaxSize)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	// Test case: file within size limit
	ok, err := validateFileSize(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !ok {
		t.Errorf("expected file to be within size limit")
	}

	// Write more data to exceed the size limit
	if _, err := tmpFile.Write(make([]byte, 1)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	// Test case: file exceeds size limit
	ok, err = validateFileSize(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ok {
		t.Errorf("expected file to exceed size limit")
	}
}
