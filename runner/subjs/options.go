package subjs

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
)

// SelectionStrategy defines how UserAgents are selected
type SelectionStrategy int

const (
	StrategyRotation SelectionStrategy = iota // Round-robin rotation (default)
	StrategyRandom                            // Pure random selection
	StrategyWeighted                          // Weighted random selection
)

// String returns string representation of SelectionStrategy
func (s SelectionStrategy) String() string {
	switch s {
	case StrategyRotation:
		return "rotation"
	case StrategyRandom:
		return "random"
	case StrategyWeighted:
		return "weighted"
	default:
		return "rotation"
	}
}

// ParseSelectionStrategy parses string to SelectionStrategy
func ParseSelectionStrategy(s string) SelectionStrategy {
	switch strings.ToLower(s) {
	case "random":
		return StrategyRandom
	case "weighted":
		return StrategyWeighted
	default:
		return StrategyRotation
	}
}

type Options struct {
	InputFile          string
	Workers            int
	Timeout            int
	UserAgent          string
	UserAgents         []string
	InsecureSkipVerify bool
	UserAgentStrategy  string              // CLI specified strategy
	userAgentMutex     sync.Mutex          // Mutex for thread-safe UserAgent operations
	currentIndex       int                 // For rotation strategy
	weightedUserAgents []WeightedUserAgent // For weighted strategy
}

// WeightedUserAgent represents a UserAgent with selection weight
type WeightedUserAgent struct {
	UserAgent string
	Weight    int
}

// GetExpandedUserAgents returns the comprehensive default UserAgent pool
func GetExpandedUserAgents() []string {
	return []string{
		// Chrome Windows (most common)
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",

		// Chrome macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",

		// Chrome Linux
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",

		// Firefox Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:119.0) Gecko/20100101 Firefox/119.0",

		// Firefox macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0",

		// Safari macOS
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.15",

		// Edge Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0",

		// Mobile Chrome Android
		"Mozilla/5.0 (Linux; Android 10; SM-G975F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",

		// Safari iPhone
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1",

		// Opera Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 OPR/106.0.0.0",

		// Brave Windows
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",

		// Old Chrome for compatibility testing
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.3",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.3",
	}
}

// GetWeightedUserAgents returns UserAgents with weights (higher weight = more common)
func GetWeightedUserAgents() []WeightedUserAgent {
	return []WeightedUserAgent{
		// Most common modern browsers (higher weight)
		{UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36", Weight: 20},
		{UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36", Weight: 15},
		{UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0", Weight: 12},
		{UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15", Weight: 10},
		{UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36", Weight: 8},

		// Moderately common browsers
		{UserAgent: "Mozilla/5.0 (Linux; Android 10; SM-G975F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36", Weight: 6},
		{UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36", Weight: 6},
		{UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0", Weight: 5},

		// Less common but still valid browsers
		{UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1", Weight: 3},
		{UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3", Weight: 2},
		{UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:120.0) Gecko/20100101 Firefox/120.0", Weight: 2},
	}
}

// SelectUserAgent intelligently selects a UserAgent based on strategy
func (opts *Options) SelectUserAgent() string {
	if opts.UserAgent != "" {
		return opts.UserAgent // Override with user-specified UA if provided
	}

	strategy := ParseSelectionStrategy(opts.UserAgentStrategy)

	opts.userAgentMutex.Lock()
	defer opts.userAgentMutex.Unlock()

	if len(opts.UserAgents) == 0 {
		// Fallback to first expanded UA if somehow empty
		uas := GetExpandedUserAgents()
		if len(uas) > 0 {
			return uas[0]
		}
		return ""
	}

	switch strategy {
	case StrategyRandom:
		return opts.selectRandom()
	case StrategyWeighted:
		return opts.selectWeighted()
	default: // StrategyRotation
		return opts.selectRotation()
	}
}

// selectRandom performs pure random selection
func (opts *Options) selectRandom() string {
	return opts.UserAgents[rand.Intn(len(opts.UserAgents))]
}

// selectWeighted performs weighted random selection
func (opts *Options) selectWeighted() string {
	if len(opts.weightedUserAgents) == 0 {
		// Fall back to pure random if no weighted data
		return opts.selectRandom()
	}

	// Calculate total weight
	totalWeight := 0
	for _, ua := range opts.weightedUserAgents {
		totalWeight += ua.Weight
	}

	// Select random point in weight range
	randomPoint := rand.Intn(totalWeight)

	// Find the UserAgent that corresponds to this point
	cumulativeWeight := 0
	for _, ua := range opts.weightedUserAgents {
		cumulativeWeight += ua.Weight
		if randomPoint < cumulativeWeight {
			return ua.UserAgent
		}
	}

	// Fallback to first (shouldn't reach here)
	return opts.weightedUserAgents[0].UserAgent
}

// selectRotation performs round-robin rotation (original behavior)
func (opts *Options) selectRotation() string {
	ua := opts.UserAgents[opts.currentIndex]
	opts.currentIndex = (opts.currentIndex + 1) % len(opts.UserAgents)
	return ua
}

// RotateUserAgent provides legacy compatibility method
// TODO: Consider deprecating in favor of SelectUserAgent
func (opts *Options) RotateUserAgent() string {
	return opts.SelectUserAgent()
}

// InitializeUserAgents sets up the UserAgent pool and strategy
func (opts *Options) InitializeUserAgents() {
	if len(opts.UserAgents) == 0 {
		// Use expanded pool instead of hardcoded 3
		opts.UserAgents = GetExpandedUserAgents()

		// Initialize weighted pool for weighted strategy
		opts.weightedUserAgents = GetWeightedUserAgents()
	}
}

func ParseOptions() *Options {
	opts := &Options{}
	flag.StringVar(&opts.InputFile, "i", "", "Input file containing URLS")
	flag.StringVar(&opts.UserAgent, "ua", "", "User-Agent to send in requests")
	flag.StringVar(&opts.UserAgentStrategy, "ua-strategy", "rotation", "UserAgent selection strategy (rotation/random/weighted)")
	flag.IntVar(&opts.Workers, "c", 10, "Number of concurrent workers")
	flag.IntVar(&opts.Timeout, "t", 15, "Timeout (in seconds) for http client")
	showVersion := flag.Bool("version", false, "Show version number")
	flag.BoolVar(&opts.InsecureSkipVerify, "insecure", false, "Skip TLS certificate verification")
	flag.Parse()

	// Initialize UserAgents after parsing options
	opts.InitializeUserAgents()

	if *showVersion {
		fmt.Printf("subjs version: %s\n", version)
		os.Exit(0)
	}
	return opts
}
