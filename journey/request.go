package journey

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/errgroup"
)

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
				for _, req := range j.requests {
					duration, err := j.makeRequest(req)
					if err != nil {
						return fmt.Errorf("error making request: %w", err)
					}
					responses <- RequestDuration{ID: req.ID, Name: req.Name, Duration: duration}
				}
			}
			return nil
		})
	}

	return errGroup.Wait()
}

func (j *Journey) makeRequest(requestConfig requestConfig) (time.Duration, error) {

	parsedURL, err := url.Parse(requestConfig.URL)
	if err != nil {
		return 0, fmt.Errorf("error parsing url: %s , error: %w", requestConfig.URL, err)
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
		return 0, fmt.Errorf("error building request: %s %s, error: %w", requestConfig.Method, requestConfig.URL, err)
	}

	if requestConfig.Headers != nil {
		for k, v := range requestConfig.Headers {
			req.Header.Set(k, v)
		}
	}

	if requestConfig.Cookies != nil {
		for k, v := range requestConfig.Cookies {
			req.AddCookie(&http.Cookie{Name: k, Value: v})
		}
	}

	if requestConfig.MimeType != "" {
		req.Header.Set("Content-Type", requestConfig.MimeType)
	}

	var start = time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error sending request: %s %s, error: %w", requestConfig.Method, requestConfig.URL, err)
	}
	var duration = time.Since(start)

	if resp.StatusCode != requestConfig.ExpectedResponseCode {
		return 0, fmt.Errorf("unexpected response code, wanted: %d, got :%d", requestConfig.ExpectedResponseCode, resp.StatusCode)
	}

	return duration, nil
}

func (j *Journey) collect(responses chan RequestDuration) journeyTiming {
	var timings = make(journeyTiming, len(j.requests))

	for response := range responses {
		timings[response.ID] = append(timings[response.ID], response)
	}

	return timings
}
