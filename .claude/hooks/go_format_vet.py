#!/usr/bin/env python3
"""PostToolUse hook: format and vet Go files after Write or Edit.

Behavior:
  - Silently skips when the edited path is not a *.go file.
  - Runs `gofmt -w` and (if installed) `goimports -w` to auto-fix style.
  - Runs `go vet` scoped to the edited file's package.
  - Exit 2 with vet output on failure; exit 0 otherwise.
  - Total budget: <5s for a small package, capped at 25s hard timeout.

Input: JSON on stdin with tool_input.file_path.
Env:   CLAUDE_PROJECT_DIR (set by Claude Code).
"""
from __future__ import annotations

import json
import os
import shutil
import subprocess
import sys
from pathlib import Path


def resolve_project_dir(fallback: Path) -> Path:
    env = os.environ.get("CLAUDE_PROJECT_DIR")
    if env:
        return Path(env).resolve()
    return fallback.resolve()


def main() -> int:
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0

    tool_input = payload.get("tool_input") or {}
    file_path_str = tool_input.get("file_path")
    if not file_path_str:
        return 0

    file_path = Path(file_path_str).resolve()
    if file_path.suffix != ".go":
        return 0
    if not file_path.exists():
        return 0

    project_dir = resolve_project_dir(file_path.parent)

    go = shutil.which("go")
    if not go:
        print("go not on PATH; skipping Go format/vet check", file=sys.stderr)
        return 0

    gofmt = shutil.which("gofmt")
    if gofmt:
        subprocess.run([gofmt, "-w", str(file_path)], check=False, timeout=10)

    goimports = shutil.which("goimports")
    if goimports:
        subprocess.run([goimports, "-w", str(file_path)], check=False, timeout=10)

    try:
        rel = file_path.parent.relative_to(project_dir)
        package_spec = f"./{rel.as_posix()}" if rel.parts else "./..."
    except ValueError:
        package_spec = "./..."

    try:
        result = subprocess.run(
            [go, "vet", package_spec],
            cwd=str(project_dir),
            capture_output=True,
            text=True,
            timeout=25,
        )
    except subprocess.TimeoutExpired:
        print("go vet timed out after 25s", file=sys.stderr)
        return 0

    if result.returncode != 0:
        message = result.stderr.strip() or result.stdout.strip() or "go vet failed"
        print(message, file=sys.stderr)
        return 2

    return 0


if __name__ == "__main__":
    sys.exit(main())
