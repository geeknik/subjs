package subjs

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const version = `1.2.0`

type SubJS struct {
	client      *http.Client
	opts        *Options
	seen        *sync.Map // Shared map for duplicate detection across all workers
	retryConfig *RetryConfig
}

func New(opts *Options) *SubJS {
	c := &http.Client{
		Timeout: time.Duration(opts.Timeout+10) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow following redirects
			return nil
		},
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: opts.InsecureSkipVerify},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
	opts.UserAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.3",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.3",
	}
	rand.Seed(time.Now().UnixNano())
	return &SubJS{
		client:      c,
		opts:        opts,
		seen:        &sync.Map{},
		retryConfig: NewRetryConfig(),
	}
}
func (s *SubJS) Run() error {
	// Setup input
	var input *os.File
	var err error
	// if input file not specified then read from stdin
	if s.opts.InputFile == "" {
		input = os.Stdin
	} else {
		// otherwise read from file
		input, err = os.Open(s.opts.InputFile)
		if err != nil {
			log.Printf("Error fetching URL: %v", err)
			log.Printf("Could not open input file: %s", err)
			return err
		}
		defer input.Close()
	}

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Increase timeout to handle redirects and slow responses
	shutdownTimer := time.AfterFunc(time.Duration(s.opts.Timeout+10)*time.Second, func() {
		log.Println("Global timeout reached, initiating shutdown...")
		cancel()
	})

	// Monitor for shutdown signals
	go func() {
		select {
		case sig := <-sigChan:
			log.Printf("Received signal: %v. Initiating graceful shutdown...", sig)
			cancel()
		case <-ctx.Done():
			// Context was already cancelled
		}
	}()

	// Initialize channels
	urls := make(chan string, s.opts.Workers)
	results := make(chan string, s.opts.Workers)
	errors := make(chan fetchError, s.opts.Workers)

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < s.opts.Workers; i++ {
		wg.Add(1)
		go func(ctx context.Context) {
			defer wg.Done()
			s.fetch(ctx, urls, results, errors)
		}(ctx)
	}

	// Handle errors from workers
	go func() {
		for err := range errors {
			log.Printf("Error processing URL '%s': %v", err.URL, err.Error)
		}
	}()

	// Setup output
	go func() {
		for result := range results {
			fmt.Println(result)
		}
	}()

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		u := scanner.Text()
		if u != "" {
			urls <- u
		}
	}
	close(urls)
	wg.Wait()
	close(results)

	// Stop the automatic shutdown timer since we completed normally
	shutdownTimer.Stop()

	return nil
}

// fetchError represents the result of processing a URL with error information
type fetchError struct {
	URL     string
	Error   error
	IsRetry bool
}

func (s *SubJS) fetch(ctx context.Context, urls <-chan string, results chan string, errors chan<- fetchError) {
	for u := range urls {
		if err := s.processURL(ctx, u, results); err != nil {
			errors <- fetchError{URL: u, Error: err, IsRetry: false}
		}
	}
}

// processURL handles the complete workflow for a single URL with improved error handling
func (s *SubJS) processURL(ctx context.Context, inputURL string, results chan string) error {
	// Validate and normalize URL
	validatedURL, err := ValidateAndNormalizeURL(inputURL)
	if err != nil {
		return err
	}

	// Convert back to string for consistency
	urlStr := validatedURL.String()

	// Fetch the document with retry logic
	resp, err := s.fetchWithRetry(ctx, urlStr)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Parse the document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ParseError{URL: urlStr, Reason: fmt.Sprintf("failed to parse HTML: %v", err)}
	}

	// Extract JavaScript files
	doc.Find("script").Each(func(index int, sel *goquery.Selection) {
		jsSrc, exists := sel.Attr("src")
		if exists && jsSrc != "" {
			s.processJavaScriptURL(jsSrc, validatedURL, results)
		}
	})

	return nil
}

// processJavaScriptURL formats and deduplicates JavaScript URLs
func (s *SubJS) processJavaScriptURL(jsSrc string, baseURL *url.URL, results chan string) {
	// Format the JavaScript URL
	formattedURL, err := FormatJavaScriptURL(jsSrc, baseURL)
	if err != nil {
		log.Printf("Warning: failed to format JavaScript URL '%s': %v", jsSrc, err)
		return
	}

	if formattedURL != "" {
		// Check for duplicates using the shared map
		if _, exists := s.seen.Load(formattedURL); !exists {
			s.seen.Store(formattedURL, struct{}{})
			results <- formattedURL
		}
	}
}

// fetchWithRetry performs HTTP request with exponential backoff retry logic
func (s *SubJS) fetchWithRetry(ctx context.Context, url string) (*http.Response, error) {
	var lastErr error
	var lastResponse *http.Response

	for attempt := 1; attempt <= s.retryConfig.MaxRetries; attempt++ {
		// Create request with timeout context
		reqCtx, cancel := context.WithTimeout(ctx, time.Duration(s.opts.Timeout)*time.Second)

		req, err := http.NewRequestWithContext(reqCtx, "GET", url, nil)
		if err != nil {
			cancel()
			return nil, HTTPRequestError{URL: url, Reason: fmt.Sprintf("failed to create request: %v", err), IsRetryable: false}
		}

		// Add headers
		req.Header.Add("User-Agent", s.opts.RotateUserAgent())

		resp, err := s.client.Do(req)
		cancel() // Always cancel, regardless of success/failure

		if err != nil {
			lastErr = HTTPRequestError{URL: url, Reason: fmt.Sprintf("request failed: %v", err), IsRetryable: true}
		} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Success (2xx status codes)
			return resp, nil
		} else {
			// Handle non-success status codes
			lastResponse = resp
			lastErr = HTTPRequestError{URL: url, StatusCode: resp.StatusCode, Reason: fmt.Sprintf("HTTP %d", resp.StatusCode), IsRetryable: ShouldRetry(nil, resp.StatusCode, attempt, s.retryConfig.MaxRetries)}
		}

		// Check if we should retry
		if !ShouldRetry(lastErr, resp.StatusCode, attempt, s.retryConfig.MaxRetries) {
			break
		}

		// Clean up response body if we have one
		if resp != nil {
			resp.Body.Close()
		}

		// Wait before retrying (only if not the last attempt)
		if attempt < s.retryConfig.MaxRetries {
			delay := s.retryConfig.GetDelay(attempt)
			log.Printf("Retrying URL %s in %v (attempt %d/%d)", url, delay, attempt, s.retryConfig.MaxRetries)
			time.Sleep(delay)
		}
	}

	// Clean up final response if unsuccessful
	if lastResponse != nil {
		lastResponse.Body.Close()
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, HTTPRequestError{URL: url, Reason: "max retries exceeded", IsRetryable: false}
}
