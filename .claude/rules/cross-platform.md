---
paths:
  - "**/*.go"
---

# Cross-Platform Go Code

Existing convention (established by commits `6c6a80d`, `1d27241`): OS-specific code is split into paired files with matching build tags. Violating this breaks the Windows CI job and leaks POSIX-only calls into the Windows build.

## Build tag rules

- Any file that uses OS-specific syscalls, file paths, or process APIs MUST carry a `//go:build` directive on line 1 (before `package`).
- Pair filenames use `_unix.go` / `_windows.go` suffixes. Example: `internal/jobs/pid_unix.go` + `internal/jobs/pid_windows.go`.
- POSIX-only code uses `//go:build !windows`. Windows-only code uses `//go:build windows`.
- Integration tests that depend on POSIX signals, symlinks, or process groups MUST use `//go:build !windows` (see `test/integration/dispatch_flow_test.go`).

## Forbidden

- NEVER gate code by `runtime.GOOS == "windows"` for behavior that differs structurally (syscall surface, file layout). Use build tags and paired files instead. Runtime checks are acceptable only for cosmetic differences (e.g. path display).
- NEVER add `_unix.go` / `_windows.go` files without the matching companion file. Missing companion = build failure on the other OS.
- NEVER call `syscall.SysProcAttr` fields, `unix.*`, or `golang.org/x/sys/windows` outside a `_unix.go` / `_windows.go` file.

## Test mirror rule

A test that compiles only on one OS MUST match the build tag of the file it tests. If both OSes need coverage for a feature, provide two test files with inverse build tags.
