# Changelog

## [1.2.0] - 2024-09-06
### Changed
- Incremented version number to 1.2.0.
### Added
- Made TLS certificate verification configurable via command-line flag.
- Eliminated duplicate JavaScript files in output by tracking seen URLs.

### Fixed
- Resolved compilation errors in `runner.go` by declaring the `seen` map inside the `fetch` function.
