> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Part of the orchestrator: `../../docs/HOT_RELOAD_MASTER_PLAN.md` (Phase A). Runs in parallel with `gorun/docs/PLAN.md` and `depfind/docs/PLAN.md`; no dependencies.

# gobuild — extract a `Compiler` interface for fast unit testing

## Problem

`client` (WASM builder) and `server` (external server binary builder) both
depend directly on the concrete struct `*gobuild.GoBuild`. This is why every
hot-reload-adjacent test in those two libraries that needs a compile step
either shells out to a real `go build`/tinygo (slow, flaky in CI) or mocks
the compile at a much higher level (`client`'s `Storage` interface, which
doesn't cover `server` at all). There is no lightweight way to unit-test
"did the correct compile get triggered" without a real compiler.

## Required change

1. In `gobuild/gobuild.go` (or a new `gobuild/compiler.go` if cleaner),
   define a minimal interface capturing the public surface actually used by
   consumers today:

   ```go
   type Compiler interface {
       CompileProgram() error
       FinalOutputPath() string
       UnobservedFiles() []string
   }
   ```

   Confirm the exact method set by grepping `client` and `server` for
   `goCompiler.` / `.CompileProgram(` / `.FinalOutputPath(` /
   `.UnobservedFiles(` usages before finalizing the interface — do not add
   methods nobody calls, do not omit ones that are called.

2. Ensure `*gobuild.GoBuild` (the existing concrete type returned by
   `gobuild.New`) satisfies `Compiler` with zero changes to its method
   bodies — this is a pure interface extraction, not a behavior change.

3. Add `gobuild/mock/compiler_mock.go` (new package `gobuildmock` or a
   `_test` helper — pick whichever matches this repo's existing mock
   conventions; check `sqlite`/`postgres` packages in the ecosystem for the
   established mock-package pattern before deciding) exposing:

   ```go
   type FakeCompiler struct {
       CompileErr       error
       CompileCallCount int
       Output           string
       Unobserved       []string
   }

   func (f *FakeCompiler) CompileProgram() error {
       f.CompileCallCount++
       return f.CompileErr
   }
   func (f *FakeCompiler) FinalOutputPath() string { return f.Output }
   func (f *FakeCompiler) UnobservedFiles() []string { return f.Unobserved }
   ```

   Keep it dependency-free (no fsnotify, no real file I/O) so `client` and
   `server` can import it in unit tests without pulling real compiler
   invocation into their test binaries.

## Constraints (apply to all code in this plan)

- No hardcoded strings: if the mock needs a default output path or similar,
  make it a named constant, not an inline literal.
- Do not change `gobuild.New`'s signature or `Config` struct — only add the
  interface and the mock. This phase must not break `gobuild`'s own existing
  tests (`compiler_test.go`, `config_test.go`, `arguments_test.go`,
  `files_test.go`, `integration_test.go`, `race_test.go`).
- Run `gotest ./...` inside `gobuild/` after the change; all existing tests
  must still pass unmodified.

## Stages

| Stage | Description | Output |
|---|---|---|
| 1 | Grep actual usage of `GoBuild` methods in `client` and `server` to finalize interface method set | List of methods confirmed |
| 2 | Add `Compiler` interface to `gobuild` package, verify `*GoBuild` satisfies it (compile-time assertion `var _ Compiler = (*GoBuild)(nil)`) | `gobuild/compiler.go` |
| 3 | Add `FakeCompiler` mock in a new mock package/file | `gobuild/mock/compiler_mock.go` (or repo-conventional location) |
| 4 | Run existing `gobuild` test suite, confirm no regressions | Test output attached to PR |
