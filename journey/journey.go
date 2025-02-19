package journey

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mrichman/hargo"
)

type Journey struct {
	requests []requestConfig
}

type requestConfig struct {
	ID                   uint8
	Name                 string
	MimeType             string
	Method               string
	URL                  string
	Body                 []byte
	Headers              map[string]string
	Cookies              map[string]string
	Query                map[string]string
	ExpectedResponseCode int
}

// RequestDuration is a type that represents the duration of the request.
type RequestDuration struct {
	ID   uint8
	Name string
	time.Duration
}

type requestTimings []RequestDuration

// journeyTiming is a slice that represents the durations of each request in the journey.
type journeyTiming []requestTimings

func New(harFile string) (*Journey, error) {

	file, err := os.Open(harFile)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	harRequests, err := hargo.Decode(bufio.NewReader(file))
	if err != nil {
		return nil, fmt.Errorf("error parsing har file: %w", err)
	}

	var journey = new(Journey)
	journey.requests = make([]requestConfig, len(harRequests.Log.Entries))

	for i, entry := range harRequests.Log.Entries {
		journey.requests[i] = requestConfig{
			ID:                   uint8(i),
			Name:                 entry.Pageref,
			MimeType:             entry.Request.PostData.MimeType,
			Method:               entry.Request.Method,
			URL:                  entry.Request.URL,
			Body:                 []byte(entry.Request.PostData.Text),
			ExpectedResponseCode: entry.Response.Status,
		}

		if entry.Request.Headers != nil {
			journey.requests[i].Headers = make(map[string]string, len(entry.Request.Headers))
			for _, header := range entry.Request.Headers {
				if !strings.HasPrefix(header.Name, ":") {
					journey.requests[i].Headers[header.Name] = header.Value
				}
			}
		}

		if entry.Request.Cookies != nil {
			journey.requests[i].Cookies = make(map[string]string, len(entry.Request.Cookies))
			for _, cookie := range entry.Request.Cookies {
				journey.requests[i].Cookies[cookie.Name] = cookie.Value
			}
		}

		if entry.Request.QueryString != nil {
			journey.requests[i].Query = make(map[string]string, len(entry.Request.QueryString))
			for _, query := range entry.Request.QueryString {
				journey.requests[i].Query[query.Name] = query.Value
			}
		}
	}
	return journey, nil
}
