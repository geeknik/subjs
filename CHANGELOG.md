# Changelog

## [Unreleased]
### Added
- Made TLS certificate verification configurable via command-line flag.
- Eliminated duplicate JavaScript files in output by tracking seen URLs.

### Fixed
- Resolved compilation errors in `runner.go` by declaring the `seen` map inside the `fetch` function.
