# subjs
[![License](https://img.shields.io/badge/license-MIT-_red.svg)](https://opensource.org/licenses/MIT)
[![Go ReportCard](https://goreportcard.com/badge/github.com/geeknik/subjs)](https://goreportcard.com/report/github.com/geeknik/subjs)

subjs fetches javascript files from a list of URLS or subdomains. Analyzing javascript files can help you find undocumented endpoints, secrets, and more.

It's recommended to pair this with [gau](https://github.com/lc/gau) and then [linkfinder](https://github.com/GerbenJavado/LinkFinder). Or even [gofuzz](https://github.com/nullenc0de/gofuzz).

# Resources
- [Usage](#usage)
- [Installation](#installation)

## Usage:
Examples:
```bash
$ cat urls.txt | subjs 
$ subjs -i urls.txt
$ cat hosts.txt | gau | subjs
```

To display the help for the tool use the `-h` flag:

```bash
$ subjs -h
Usage of subjs:
  -c int
        Number of concurrent workers (default 10)
  -i string
        Input file containing URLS
  -insecure
        Skip TLS certificate verification
  -t int
        Timeout (in seconds) for http client (default 15)
  -ua string
        User-Agent to send in requests
  -ua-strategy string
        UserAgent selection strategy (rotation/random/weighted) (default "rotation")
  -version
```

### UserAgent Selection Strategies

subjs now supports three UserAgent selection strategies:

- **`rotation`** (default): Round-robin rotation through UserAgent pool - maintains consistency and is memory efficient
- **`random`**: Pure random selection from the comprehensive UserAgent pool - maximum diversity
- **`weighted`**: Weighted random selection based on real-world browser usage statistics - most realistic distribution

Examples:
```bash
$ subjs -i urls.txt -ua-strategy random    # Random UserAgent for each request
$ subjs -i urls.txt -ua-strategy weighted  # Weighted random based on browser popularity
$ subjs -i urls.txt                        # Default rotation strategy
```

### Advanced Features

**Enhanced Error Handling:**
- Comprehensive error types and structured error reporting
- Exponential backoff retry logic with jitter to prevent thundering herd
- Robust URL validation and normalization
- Per-request timeout context
- Proper resource cleanup and graceful shutdown

**Improved UserAgent Pool:**
- 20+ modern UserAgent strings covering Chrome, Firefox, Safari, Edge
- Support for Windows, macOS, Linux, and mobile platforms
- Includes legacy UserAgents for compatibility testing

## Installation
### From Source:

```
git clone https://github.com/geeknik/subjs
cd subjs
go build .
go install
subjs -version
```

Original author: [lc](https://github.com/lc/)
