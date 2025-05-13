package main

import (
	"os"
	"testing"
)

func TestDetectDefaultShell(t *testing.T) {
	passwdContent := `testuser:x:1000:1000:Test User:/home/testuser:/bin/sh`
	tempFile, err := os.CreateTemp("", "passwd")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(passwdContent); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	tempFile.Close()

	tests := []struct {
		name     string
		user     string
		expected string
	}{
		{"Valid user", "testuser", "/bin/sh"},
		{"Non-existent user", "nonexistent", "/bin/bash"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectDefaultShell(tt.user, tempFile.Name())
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
