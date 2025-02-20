package journey

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/viki-org/dnscache"
	"golang.org/x/sync/errgroup"
)

type Response struct {
	time.Duration
	error
}

func (j *Journey) Replay(numRequests uint16, concurrency uint8) (journeyTiming, error) {
	var timing journeyTiming
	var responses = make(chan RequestDuration)
	var done = make(chan struct{})

	go func() {
		timing = j.collect(responses)
		close(done)
	}()

	j.Stream(numRequests, concurrency, responses)

	<-done
	return timing, nil
}

func (j *Journey) Stream(totalNumRequests uint16, concurrency uint8, responses chan<- RequestDuration) error {

	var errGroup = new(errgroup.Group)
	defer close(responses)

	for range concurrency {
		errGroup.Go(func() error {
			var client = makeClient()
			var err error
			var requests = make([]*http.Request, len(j.Requests))
			for i, req := range j.Requests {
				requests[i], err = makeRequest(req)
				if err != nil {
					return err
				}
			}

			for range totalNumRequests / uint16(concurrency) {
				for i, req := range requests {
					duration, err := j.runRequest(client, req, j.Requests[i].ExpectedResponseCode)
					if err != nil {
						return err
					}
					responses <- RequestDuration{ID: j.Requests[i].ID, Name: j.Requests[i].Name, Duration: duration, Error: err}
				}
			}
			return nil
		})
	}

	return errGroup.Wait()
}

func (j *Journey) runRequest(client *http.Client, req *http.Request, expectedResponseCode int) (time.Duration, error) {

	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)
	if err != nil {
		return 0, fmt.Errorf("error sending request: %s %s, error: %w", req.Method, req.URL, err)
	}

	// Read the response body to ensure the connection is reused
	io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != expectedResponseCode {
		return 0, fmt.Errorf("unexpected response code, wanted: %d, got :%d", expectedResponseCode, resp.StatusCode)
	}

	return duration, nil
}

func (j *Journey) collect(responses chan RequestDuration) journeyTiming {
	var timings = make(journeyTiming, len(j.Requests))

	for response := range responses {
		timings[response.ID] = append(timings[response.ID], response)
	}

	return timings
}

func makeRequest(requestConfig requestConfig) (*http.Request, error) {
	parsedURL, err := url.Parse(requestConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("error parsing url: %s , error: %w", requestConfig.URL, err)
	}

	if requestConfig.Query != nil {
		q := parsedURL.Query()
		for k, v := range requestConfig.Query {
			q.Add(k, v)
		}
		parsedURL.RawQuery = q.Encode()
	}

	req, err := http.NewRequest(requestConfig.Method, parsedURL.String(), bytes.NewReader(requestConfig.Body))
	if err != nil {
		return nil, fmt.Errorf("error building request: %s %s, error: %w", requestConfig.Method, requestConfig.URL, err)
	}

	for k, v := range requestConfig.Headers {
		req.Header.Set(k, v)
	}

	for k, v := range requestConfig.Cookies {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}

	if requestConfig.MimeType != "" {
		req.Header.Set("Content-Type", requestConfig.MimeType)
	}

	return req, nil
}

func makeClient() *http.Client {
	var cache = dnscache.New(5 * time.Minute)
	return &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				// Split the address to get the host and port
				separator := strings.LastIndex(address, ":")
				if separator == -1 {
					return nil, fmt.Errorf("invalid address format: %s", address)
				}
				host := address[:separator]
				port := address[separator+1:]

				// Check if the IP is cached, otherwise resolve it
				ips, err := cache.Fetch(host)
				if err != nil {
					return nil, fmt.Errorf("failed to fetch IPs: %w", err)
				}

				// Try to connect to each IP address
				var lastErr error
				for _, ip := range ips {
					conn, err := (&net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 30 * time.Second,
					}).DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
					if err == nil {
						return conn, nil
					}
					lastErr = err
				}
				return nil, fmt.Errorf("failed to connect to any IP: %w", lastErr)
			},
		},
	}
}
