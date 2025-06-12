package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileStorage_ReadLastIP(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		fileExists  bool
		expected    string
		expectError bool
	}{
		{
			name:        "reads existing IP",
			fileContent: "192.168.1.1",
			fileExists:  true,
			expected:    "192.168.1.1",
			expectError: false,
		},
		{
			name:        "reads IP with whitespace",
			fileContent: "  192.168.1.1\n  ",
			fileExists:  true,
			expected:    "192.168.1.1",
			expectError: false,
		},
		{
			name:        "returns empty string for non-existent file",
			fileExists:  false,
			expected:    "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test_ip.txt")

			if tt.fileExists {
				err := os.WriteFile(filePath, []byte(tt.fileContent), 0600)
				if err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			fs := NewFileStorage(filePath)
			result, err := fs.ReadLastIP()

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFileStorage_WriteIP(t *testing.T) {
	tests := []struct {
		name        string
		ip          string
		expectError bool
	}{
		{
			name:        "writes IP successfully",
			ip:          "192.168.1.1",
			expectError: false,
		},
		{
			name:        "writes empty IP",
			ip:          "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test_ip.txt")

			fs := NewFileStorage(filePath)
			err := fs.WriteIP(tt.ip)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError {
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("failed to read written file: %v", err)
				}
				if string(content) != tt.ip {
					t.Errorf("expected file content %q, got %q", tt.ip, string(content))
				}
			}
		})
	}
}

func TestFileStorage_ReadWriteRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_ip.txt")
	fs := NewFileStorage(filePath)

	testIP := "203.0.113.1"

	err := fs.WriteIP(testIP)
	if err != nil {
		t.Fatalf("failed to write IP: %v", err)
	}

	readIP, err := fs.ReadLastIP()
	if err != nil {
		t.Fatalf("failed to read IP: %v", err)
	}

	if readIP != testIP {
		t.Errorf("round trip failed: wrote %q, read %q", testIP, readIP)
	}
}