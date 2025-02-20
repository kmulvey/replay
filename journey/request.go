package journey

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
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

func (j *Journey) Stream(numRequests uint16, concurrency uint8, responses chan<- RequestDuration) error {

	var errGroup = new(errgroup.Group)
	defer close(responses)

	for range concurrency {
		errGroup.Go(func() error {
			for range numRequests / uint16(concurrency) {
				for _, req := range j.Requests {
					j.makeRequest(req, numRequests, responses)
				}
			}
			return nil
		})
	}

	return errGroup.Wait()
}
func (j *Journey) makeRequest(requestConfig requestConfig, numRequests uint16, responses chan<- RequestDuration) {
	defer close(responses)

	parsedURL, err := url.Parse(requestConfig.URL)
	if err != nil {
		responses <- RequestDuration{
			ID:       requestConfig.ID,
			Name:     requestConfig.Name,
			Duration: 0,
			error:    fmt.Errorf("error parsing url: %s , error: %w", requestConfig.URL, err),
		}
		return
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
		responses <- RequestDuration{
			ID:       requestConfig.ID,
			Name:     requestConfig.Name,
			Duration: 0,
			error:    fmt.Errorf("error building request: %s %s, error: %w", requestConfig.Method, requestConfig.URL, err),
		}
		return
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

	var cache = dnscache.New(5 * time.Minute)
	//addrValue := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 45678}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				separator := strings.LastIndex(address, ":")
				ips, err := cache.Fetch(address[:separator])
				if err != nil {
					return nil, fmt.Errorf("failed to fetch IPs: %w", err)
				}
				var lastErr error
				for _, ip := range ips {
					conn, err := (&net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 30 * time.Second,
						//LocalAddr: &addrValue,
					}).DialContext(ctx, network, net.JoinHostPort(ip.String(), address[separator+1:]))
					if err == nil {
						return conn, nil
					}
					lastErr = err
				}
				return nil, fmt.Errorf("failed to connect to any IP: %w", lastErr)
			},
			DialTLSContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				separator := strings.LastIndex(address, ":")
				ips, err := cache.Fetch(address[:separator])
				if err != nil {
					return nil, fmt.Errorf("failed to fetch IPs: %w", err)
				}
				var lastErr error
				for _, ip := range ips {
					conn, err := tls.DialWithDialer(&net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 30 * time.Second,
					}, network, net.JoinHostPort(ip.String(), address[separator+1:]), &tls.Config{
						InsecureSkipVerify: true,
					})
					if err == nil {
						return conn, nil
					}
					lastErr = err
				}
				return nil, fmt.Errorf("failed to connect to any IP: %w", lastErr)
			},
		},
	}

	for range numRequests {
		start := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(start)

		if err != nil {
			responses <- RequestDuration{
				ID:       requestConfig.ID,
				Name:     requestConfig.Name,
				Duration: duration,
				error:    fmt.Errorf("error sending request: %s %s, error: %w", requestConfig.Method, requestConfig.URL, err),
			}
			continue
		}

		if resp.StatusCode != requestConfig.ExpectedResponseCode {
			responses <- RequestDuration{
				ID:       requestConfig.ID,
				Name:     requestConfig.Name,
				Duration: duration,
				error:    fmt.Errorf("unexpected response code, wanted: %d, got :%d", requestConfig.ExpectedResponseCode, resp.StatusCode),
			}
			continue
		}

		responses <- RequestDuration{
			ID:       requestConfig.ID,
			Name:     requestConfig.Name,
			Duration: duration,
		}
	}
}

func (j *Journey) collect(responses chan RequestDuration) journeyTiming {
	var timings = make(journeyTiming, len(j.Requests))

	for response := range responses {
		timings[response.ID] = append(timings[response.ID], response)
	}

	return timings
}
