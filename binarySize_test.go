package gobuild

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBinarySizer_New tests the constructor
func TestBinarySizer_New(t *testing.T) {
	sizer := NewBinarySizer(func() []byte { return []byte("test") })
	if sizer == nil {
		t.Fatal("NewBinarySizer returned nil")
	}
	if sizer.getBinary == nil {
		t.Error("getBinary function should not be nil")
	}
	if sizer.log == nil {
		t.Error("log function should be initialized")
	}
}

// TestBinarySizer_SetLog tests the SetLog method
func TestBinarySizer_SetLog(t *testing.T) {
	var logOutput bytes.Buffer
	logFunc := func(msgs ...any) {
		for _, msg := range msgs {
			logOutput.WriteString(strings.TrimSpace(string(msg.(string))))
		}
	}

	sizer := NewBinarySizer(func() []byte { return []byte("test") })
	sizer.SetLog(logFunc)

	// Verify SetLog doesn't panic with nil
	sizer.SetLog(nil)
}

// TestBinarySizer_BinarySize_ZeroBytes tests 0 bytes scenario
func TestBinarySizer_BinarySize_ZeroBytes(t *testing.T) {
	sizer := NewBinarySizer(func() []byte { return []byte{} })
	result := sizer.BinarySize()
	if result != "0.0 KB" {
		t.Errorf("Expected '0.0 KB' for empty bytes, got '%s'", result)
	}
}

// TestBinarySizer_BinarySize_NilBytes tests nil bytes scenario
func TestBinarySizer_BinarySize_NilBytes(t *testing.T) {
	sizer := NewBinarySizer(func() []byte { return nil })
	result := sizer.BinarySize()
	if result != "0.0 KB" {
		t.Errorf("Expected '0.0 KB' for nil bytes, got '%s'", result)
	}
}

// TestBinarySizer_BinarySize_NilFunction tests nil function scenario
func TestBinarySizer_BinarySize_NilFunction(t *testing.T) {
	sizer := NewBinarySizer(nil)
	result := sizer.BinarySize()
	if result != "0.0 KB" {
		t.Errorf("Expected '0.0 KB' for nil function, got '%s'", result)
	}
}

// TestBinarySizer_BinarySize_SmallBytes tests sizes less than 1KB
func TestBinarySizer_BinarySize_SmallBytes(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected string
	}{
		{"100 bytes", 100, "0.1 KB"},
		{"500 bytes", 500, "0.5 KB"},
		{"512 bytes", 512, "0.5 KB"},
		{"1000 bytes", 1000, "1.0 KB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.size)
			sizer := NewBinarySizer(func() []byte { return data })
			result := sizer.BinarySize()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestBinarySizer_BinarySize_KB tests kilobyte range
func TestBinarySizer_BinarySize_KB(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected string
	}{
		{"1 KB", 1024, "1.0 KB"},
		{"10 KB", 10 * 1024, "10.0 KB"},
		{"10.4 KB", 10*1024 + 410, "10.4 KB"},
		{"100 KB", 100 * 1024, "100.0 KB"},
		{"999 KB", 999 * 1024, "999.0 KB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.size)
			sizer := NewBinarySizer(func() []byte { return data })
			result := sizer.BinarySize()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestBinarySizer_BinarySize_MB tests megabyte range
func TestBinarySizer_BinarySize_MB(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected string
	}{
		{"1 MB", 1024 * 1024, "1.0 MB"},
		{"2.3 MB", 2*1024*1024 + 307*1024, "2.3 MB"},
		{"10 MB", 10 * 1024 * 1024, "10.0 MB"},
		{"100 MB", 100 * 1024 * 1024, "100.0 MB"},
		{"999 MB", 999 * 1024 * 1024, "999.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.size)
			sizer := NewBinarySizer(func() []byte { return data })
			result := sizer.BinarySize()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestBinarySizer_BinarySize_GB tests gigabyte range
func TestBinarySizer_BinarySize_GB(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{"1 GB", 1024 * 1024 * 1024, "1.0 GB"},
		{"1.5 GB", int64(1.5 * 1024 * 1024 * 1024), "1.5 GB"},
		{"2 GB", 2 * 1024 * 1024 * 1024, "2.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.size)
			sizer := NewBinarySizer(func() []byte { return data })
			result := sizer.BinarySize()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestBinarySizer_BinarySize_ThresholdBoundaries tests exact threshold boundaries
func TestBinarySizer_BinarySize_ThresholdBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{"Just below 1MB", 1024*1024 - 1, "1024.0 KB"},
		{"Exactly 1MB", 1024 * 1024, "1.0 MB"},
		{"Just above 1MB", 1024*1024 + 1, "1.0 MB"},
		{"Just below 1GB", 1024*1024*1024 - 1, "1024.0 MB"},
		{"Exactly 1GB", 1024 * 1024 * 1024, "1.0 GB"},
		{"Just above 1GB", 1024*1024*1024 + 1, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.size)
			sizer := NewBinarySizer(func() []byte { return data })
			result := sizer.BinarySize()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestBinarySizer_BinarySize_DecimalFormatting tests decimal precision
func TestBinarySizer_BinarySize_DecimalFormatting(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected string
	}{
		{"0.1 KB", 102, "0.1 KB"},
		{"0.5 KB", 512, "0.5 KB"},
		{"1.0 KB", 1024, "1.0 KB"},
		{"1.1 KB", 1126, "1.1 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"1.9 KB", 1946, "1.9 KB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, tt.size)
			sizer := NewBinarySizer(func() []byte { return data })
			result := sizer.BinarySize()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestGoBuild_BinarySize_Integration tests integration with GoBuild
func TestGoBuild_BinarySize_Integration(t *testing.T) {
	config := &Config{
		Command:                   "go",
		MainInputFileRelativePath: "main.go",
		OutName:                   "test",
		Extension:                 ".exe",
		OutFolderRelativePath:     "build",
	}
	gb := New(config)

	// Test that BinarySize method exists and returns a string
	result := gb.BinarySize()
	if result == "" {
		t.Error("BinarySize() should return a non-empty string")
	}

	// Should return "0.0 KB" since file doesn't exist
	if result != "0.0 KB" {
		t.Errorf("Expected '0.0 KB' for non-existent file, got '%s'", result)
	}
}

// TestGoBuild_BinarySize_WithLogger tests logging integration
func TestGoBuild_BinarySize_WithLogger(t *testing.T) {
	var logOutput bytes.Buffer
	logFunc := func(msgs ...any) {
		for _, msg := range msgs {
			logOutput.WriteString(strings.TrimSpace(string(msg.(string))))
		}
	}

	config := &Config{
		Command:                   "go",
		MainInputFileRelativePath: "main.go",
		OutName:                   "test",
		Extension:                 ".exe",
		OutFolderRelativePath:     "build",
		Logger:                    logFunc,
	}
	gb := New(config)

	// Verify logger was set
	if gb.binarySizer.log == nil {
		t.Error("Logger should be set on binarySizer")
	}

	// Call BinarySize (should not panic)
	result := gb.BinarySize()
	if result != "0.0 KB" {
		t.Errorf("Expected '0.0 KB', got '%s'", result)
	}
}

// TestGoBuild_BinarySize_WithCompileToMemory tests BinarySize with in-memory compilation
func TestGoBuild_BinarySize_WithCompileToMemory(t *testing.T) {
	// Create a temporary directory for the source file
	tmpDir := t.TempDir()

	// Create a simple Go program
	srcFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(srcFile, []byte(`package main
import "fmt"
func main() {
	fmt.Println("Hello, BinarySize!")
}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Create configuration
	config := &Config{
		Command:                   "go",
		MainInputFileRelativePath: "main.go",
		OutName:                   "testapp",
		Extension:                 "",
		OutFolderRelativePath:     tmpDir,
		AppRootDir:                tmpDir,
		Logger: func(msg ...any) {
			t.Log(msg...)
		},
	}

	gb := New(config)

	// Compile to memory
	binary, err := gb.CompileToMemory()
	if err != nil {
		t.Fatalf("CompileToMemory failed: %v", err)
	}

	if len(binary) == 0 {
		t.Fatal("Expected non-empty binary")
	}

	// Now BinarySize() should return the in-memory binary size
	sizeStr := gb.BinarySize()
	if sizeStr == "0.0 KB" {
		t.Error("Expected BinarySize to return non-zero size after CompileToMemory")
	}

	// Verify the size makes sense (should be at least some KB for a compiled Go binary)
	t.Logf("Compiled binary size: %s (actual bytes: %d)", sizeStr, len(binary))

	// Check that size string has correct format
	if !strings.Contains(sizeStr, "KB") && !strings.Contains(sizeStr, "MB") && !strings.Contains(sizeStr, "GB") {
		t.Errorf("Expected size format with KB/MB/GB, got '%s'", sizeStr)
	}
}
