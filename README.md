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
  -version
```

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
