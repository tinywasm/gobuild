package gobuild

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// CompileToMemory compiles the Go program returning the binary as a byte slice.
// It avoids writing to physical disk by using stdout.
func (h *GoBuild) CompileToMemory() ([]byte, error) {
	h.mu.Lock()

	// Cancel any active compilation
	if h.active != nil {
		h.active.cancel()
		h.active = nil
	}

	// Create new compilation context
	ctx, cancel := context.WithTimeout(context.Background(), h.config.Timeout)

	// In-memory compilation doesn't use temp files on disk, but we need a "compilation" struct
	// to track state/cancellation.
	comp := &compilation{
		cancel:    cancel,
		done:      make(chan error, 1),
		tempFile:  "memory", // Virtual placeholder
		startTime: time.Now(),
	}

	h.active = comp
	h.mu.Unlock()

	// Build arguments: -o /dev/stdout ...
	// Note: We use h.buildArguments but we need to override the output file.
	// h.buildArguments appends -o <dest> at the beginning.
	// We'll construct args manually here reusing logic or refactor buildArguments later if needed.
	// For minimal invasion, we construct base args and prepend our special output.

	// Use "/dev/stdout" for output. This works on Linux/Mac.
	// TODO: Windows verification. Go 1.20+ might support it natively or we need fallback.
	// User requested to try this approach first.
	outputDest := "/dev/stdout"

	// Construct arguments using the shared logic which handles ldflags, input paths, etc.
	// Because outputDest starts with /dev/, buildArguments will treat it as absolute/special.
	args := h.buildArguments(outputDest)

	cmd := exec.CommandContext(ctx, h.config.Command, args...)
	cmd.Dir = h.config.AppRootDir

	// Environment variables
	cmd.Env = os.Environ() // Inherit current env
	// Add/Override env vars from config if any
	// (gobuild.go doesn't seem to have explicit Env map in Config visible in previous view,
	// assuming standard behavior or none for now based on viewed files)

	// Capture Stdout
	var wasmBuffer bytes.Buffer
	cmd.Stdout = &wasmBuffer

	// Capture Stderr for logs (and pass to logger if needed)
	// We can use a buffer for stderr too to log on error
	var stderrBuffer bytes.Buffer
	cmd.Stderr = &stderrBuffer

	if h.config.Logger != nil {
		h.config.Logger("Compiling to memory...")
	}

	err := cmd.Run()

	// Clean up active state
	h.mu.Lock()
	if h.active == comp {
		h.active = nil
	}
	h.mu.Unlock()

	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("compilation timed out after %v", h.config.Timeout)
	}

	if err != nil {
		return nil, fmt.Errorf("compilation failed: %w\nOutput: %s", err, stderrBuffer.String())
	}

	if h.config.Logger != nil {
		h.config.Logger("Compilation to memory success. Size:", wasmBuffer.Len(), "bytes")
	}

	return wasmBuffer.Bytes(), nil
}
