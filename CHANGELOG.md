# Changelog

All notable changes to cc-dispatch will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.1.3] - 2026-04-24

### Fixed

- `ccd show`, `tail`, `cancel`, and `resume-cmd` now accept the 8-character session id prefix printed by `ccd list`.
  Previously the daemon required the full UUID, so ids copy-pasted from `list` never resolved.

### Changed

- Daemon RPC endpoints `dispatch_status`, `dispatch_tail`, and `dispatch_cancel` resolve id prefixes server-side.
  Ambiguous prefixes return HTTP 409 `ambiguous_id`; full ids continue to work unchanged.
