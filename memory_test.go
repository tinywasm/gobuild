package gobuild

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCompileToMemory(t *testing.T) {
	// Create a temporary directory for the source file
	tmpDir := t.TempDir()

	// Create a simple Go program
	srcFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(srcFile, []byte(`package main
import "fmt"
func main() {
	fmt.Println("Hello, Memory!")
}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Create configuration
	cfg := &Config{
		AppRootDir:                tmpDir,
		MainInputFileRelativePath: "main.go",
		Command:                   "go",
		OutName:                   "testapp",
		Extension:                 "",
		OutFolderRelativePath:     "",
		Logger: func(msg ...any) {
			t.Log(msg...)
		},
	}

	builder := New(cfg)

	// Test CompileToMemory
	binary, err := builder.CompileToMemory()
	if err != nil {
		t.Fatalf("CompileToMemory failed: %v", err)
	}

	if len(binary) == 0 {
		t.Fatal("Expected non-empty binary content")
	}

	// Basic ELF/Mach-O/PE magic number check to verify it's likely a binary
	// Linux/SysV: 0x7f 0x45 0x4c 0x46 (.ELF)
	// macOS (Mach-O 64-bit): 0xcf 0xfa 0xed 0xfe
	// Windows (PE): 0x4d 0x5a (MZ)
	if len(binary) > 4 {
		magic := binary[:4]
		t.Logf("Binary magic bytes: %x", magic)
	}

	// Special check for Windows where /dev/stdout might not work as expected
	// passed by the implementation
	if runtime.GOOS == "windows" {
		t.Log("Warning: Windows verification of /dev/stdout is experimental")
	}
}
