package subjs

import (
	"bufio"
	"context"
	"crypto/tls"
	"time"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"math/rand"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

const version = `1.1.4`

type SubJS struct {
	client *http.Client
	opts   *Options
}

func New(opts *Options) *SubJS {
	c := &http.Client{
		Timeout:   time.Duration(opts.Timeout+10) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow following redirects
			return nil
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: opts.InsecureSkipVerify},
		},
	}
	opts.UserAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.3",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/56.0.2924.87 Safari/537.3",
	}
	rand.Seed(time.Now().UnixNano())
	return &SubJS{client: c, opts: opts}
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

	// Increase timeout to handle redirects and slow responses
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.opts.Timeout+10)*time.Second)
	defer cancel()

	// Initialize channels
	urls := make(chan string, s.opts.Workers)
	results := make(chan string, s.opts.Workers)


	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < s.opts.Workers; i++ {
		wg.Add(1)
		go func(ctx context.Context) {
			defer wg.Done()
			s.fetch(ctx, urls, results)
		}(ctx)
	}

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
	return nil
}
func (s *SubJS) fetch(ctx context.Context, urls <-chan string, results chan string) {
    seen := make(map[string]struct{})
    for u := range urls {
        if u == "" {
            log.Printf("Empty URL encountered")
            continue
        }

        var (
            resp *http.Response
            err  error
        )

        for retries := 0; retries < 3; retries++ {
            req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
            if err != nil {
                log.Printf("Error creating request for URL %s: %v", u, err)
                break
            }
            req.Header.Add("User-Agent", s.opts.RotateUserAgent())

            resp, err = s.client.Do(req)
            if err == nil && resp != nil && resp.StatusCode == http.StatusOK {
                break
            }
            if err != nil {
                log.Printf("Error fetching URL %s: %v", u, err)
            }
            if resp != nil {
                log.Printf("Retrying URL %s: attempt %d, Status Code: %d", u, retries+1, resp.StatusCode)
            } else {
                log.Printf("Retrying URL %s: attempt %d", u, retries+1)
            }
            time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
        }

        if err != nil || resp == nil {
            log.Printf("Failed to fetch URL %s after retries: %v", u, err)
            continue
        }

        defer resp.Body.Close()

        doc, err := goquery.NewDocumentFromReader(resp.Body)
        if err != nil {
            log.Printf("Error parsing document from URL %s: %v", u, err)
            continue
        }

        parsedURL, err := url.Parse(u)
        if err != nil || parsedURL == nil {
            log.Printf("Failed to parse URL %s: %v", u, err)
            continue
        }

        doc.Find("script").Each(func(index int, s *goquery.Selection) {
            js, _ := s.Attr("src")
            if js != "" {
                if strings.HasPrefix(js, "http://") || strings.HasPrefix(js, "https://") {
                    if _, exists := seen[js]; !exists {
                        seen[js] = struct{}{}
                        results <- js
                    }
                } else if strings.HasPrefix(js, "//") {
                    js = fmt.Sprintf("%s:%s", parsedURL.Scheme, js)
                    if _, exists := seen[js]; !exists {
                        seen[js] = struct{}{}
                        results <- js
                    }
                } else if strings.HasPrefix(js, "/") {
                    js = fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, js)
                    if _, exists := seen[js]; !exists {
                        seen[js] = struct{}{}
                        results <- js
                    }
                } else {
                    js = fmt.Sprintf("%s://%s/%s", parsedURL.Scheme, parsedURL.Host, js)
                    results <- js
                }
            }
        })
    }
}
