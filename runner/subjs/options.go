package subjs

import (
	"flag"
	"fmt"
	"os"
)

type Options struct {
	InputFile string
	Workers          int
	Timeout          int
	UserAgent        string
	UserAgents       []string
	InsecureSkipVerify bool
}

func (opts *Options) RotateUserAgent() string {
	if len(opts.UserAgents) == 0 {
		return ""
	}
	ua := opts.UserAgents[0]
	opts.UserAgents = append(opts.UserAgents[1:], ua)
	return ua
}

func ParseOptions() *Options {
	opts := &Options{}
	flag.StringVar(&opts.InputFile, "i", "", "Input file containing URLS")
	flag.StringVar(&opts.UserAgent, "ua", "", "User-Agent to send in requests")
	flag.IntVar(&opts.Workers, "c", 10, "Number of concurrent workers")
	flag.IntVar(&opts.Timeout, "t", 15, "Timeout (in seconds) for http client")
	showVersion := flag.Bool("version", false, "Show version number")
	flag.BoolVar(&opts.InsecureSkipVerify, "insecure", false, "Skip TLS certificate verification")
	flag.Parse()
	if *showVersion {
		fmt.Printf("subjs version: %s\n", version)
		os.Exit(0)
	}
	return opts
}
