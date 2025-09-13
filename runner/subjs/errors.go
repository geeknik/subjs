package subjs

import (
	"fmt"
	"time"
)

// Error types for better error handling and testability
type SubJSError interface {
	error
	ErrorType() string
}

// URLValidationError indicates invalid URL format or constraints
type URLValidationError struct {
	URL    string
	Reason string
}

func (e URLValidationError) Error() string {
	return fmt.Sprintf("URL validation error for '%s': %s", e.URL, e.Reason)
}

func (e URLValidationError) ErrorType() string {
	return "URL_VALIDATION"
}

// HTTPRequestError indicates HTTP request failures
type HTTPRequestError struct {
	URL         string
	StatusCode  int
	Reason      string
	IsRetryable bool
}

func (e HTTPRequestError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("HTTP request error for '%s': status %d - %s", e.URL, e.StatusCode, e.Reason)
	}
	return fmt.Sprintf("HTTP request error for '%s': %s", e.URL, e.Reason)
}

func (e HTTPRequestError) ErrorType() string {
	return "HTTP_REQUEST"
}

// ParseError indicates document parsing failures
type ParseError struct {
	URL     string
	Reason  string
	Content string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("parse error for '%s': %s", e.URL, e.Reason)
}

func (e ParseError) ErrorType() string {
	return "PARSE"
}

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries    int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	JitterFactor  float64
}

// NewRetryConfig creates default retry configuration
func NewRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    5,
		BaseDelay:     500 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  1.1,
	}
}

// GetDelay calculates delay for given retry attempt with exponential backoff and jitter
func (rc *RetryConfig) GetDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return 0
	}

	// Exponential backoff with integer calculation
	baseDelayFloat := float64(rc.BaseDelay)
	expFactor := rc.BackoffFactor
	if attempt > 1 {
		// Calculate exponential factor safely
		factor := 1.0
		for i := 1; i < attempt; i++ {
			factor *= rc.BackoffFactor
		}
		expFactor = factor
	}
	delay := baseDelayFloat * expFactor

	// Apply maximum delay
	if delay > float64(rc.MaxDelay) {
		delay = float64(rc.MaxDelay)
	}

	// Add jitter to prevent thundering herd (0.1 to 2.0 range)
	jitter := rc.JitterFactor + float64(time.Now().UnixNano()%1000)/1000.0
	if jitter > 2.0 {
		jitter = 2.0
	}
	delay *= jitter

	return time.Duration(delay)
}

// ShouldRetry determines if an error is retryable
func ShouldRetry(err error, statusCode int, attempt int, maxRetries int) bool {
	if attempt >= maxRetries {
		return false
	}

	// Don't retry HTTP client errors (4xx)
	if statusCode >= 400 && statusCode < 500 {
		return false
	}

	// Retry temporary errors
	if reqErr, ok := err.(*HTTPRequestError); ok {
		return reqErr.IsRetryable
	}

	return true
}
