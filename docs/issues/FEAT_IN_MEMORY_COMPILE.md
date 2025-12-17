# Feature: In-Memory Compilation Support

**Objective**: Extend `gobuild` to support compiling Go programs directly to memory, avoiding temporary files on disk when possible, using `stdout` capture.

## Context
Currently, `gobuild` compiles to a temporary file on disk. For transient development environments or "In-Memory" modes in consuming libraries (like `server` or `client`), we want to compile directly to a byte buffer.

## Requirements

1.  **Stdout Capture**:
    *   Use `go build -o /dev/stdout ...` (or equivalent cross-platform approach) to direct binary output to `stdout`.
    *   Capture this output into a `bytes.Buffer`.
    *   **Compatibility**: Verify or fallback for Windows if `/dev/stdout` is not natively supported by `go build` in the same way (though Go 1.20+ often handles this). If not reliable, fallback to a temporary file + read + delete strategy transparently.

2.  **New API**:
    *   Add method `CompileToMemory() ([]byte, error)`.
    *   Ensure it respects existing `Config` (Env vars, Flags, Context/Timeout).

3.  **Refactoring**:
    *   Existing `CompileProgram()` writes to disk. Ensure `CompileToMemory` shares logic (args construction) but changes the output target.

4.  **Verification**:
    *   Unit test compiling a simple "Hello World" to memory.
    *   Verify the returned `[]byte` is a valid executable (magic headers?).

## Constraints
*   Must be thread-safe (respect `mu` lock).
*   Must handle cancellation/timeout correctly.
