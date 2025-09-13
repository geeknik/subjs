# Changelog

## [1.2.1] - 2025-09-13
### Fixed
- Resolved "context canceled" errors during HTML parsing by buffering response body before timeout.

## [1.2.0] - 2024-09-06
### Added
- TLS certificate verification configurable via command-line flag.
- UserAgent rotation, random, and weighted selection strategies.
- Duplicate JavaScript file elimination.

### Fixed
- Compilation errors in `runner.go`.
