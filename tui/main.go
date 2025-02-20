package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kmulvey/goutils"
	"github.com/kmulvey/replay/histogram"
	"github.com/kmulvey/replay/journey"
	"github.com/rivo/tview"
)

var errLog *os.File

func main() {
	errLog, _ = os.OpenFile("errors.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	var totalNumberOfRequests int
	var concurrentRequests int
	var harFile string
	flag.IntVar(&totalNumberOfRequests, "total", 1000, "total number of requests")
	flag.IntVar(&concurrentRequests, "concurrent", 10, "number of concurrent requests")
	flag.StringVar(&harFile, "har", "../localhost.har", "har file to replay")
	flag.Parse()

	var tui, err = journeyUI(harFile, totalNumberOfRequests, concurrentRequests)
	if err != nil {
		panic(err)
	}

	if err := tui.Run(); err != nil {
		errLog.WriteString(fmt.Sprintf("error running tui: %v\n", err))
	}
}

func journeyUI(harFile string, totalNumberOfRequests, concurrentRequests int) (*tview.Application, error) {

	var j, err = journey.New(harFile)
	if err != nil {
		return nil, err
	}

	// journeyResponses are the timings of each request
	var journeyResponses = make(chan journey.RequestDuration, 1000)

	// we need to fan out the journeyResponses to each graph, one graph per request
	var graphs = make([]chan journey.RequestDuration, len(j.Requests))
	for i := range graphs {
		graphs[i] = make(chan journey.RequestDuration)
	}
	go fanOut(journeyResponses, graphs...)

	// buckets are the histogram bucket values for each request
	var buckets = make([]chan histogram.Bucket, len(j.Requests))
	for i := range buckets {
		buckets[i] = make(chan histogram.Bucket)
	}

	// redistributedBuckets are the histogram bucket values after the buckets have been redistributed
	// based on the new min and max values
	var redistributedBuckets = make([]chan histogram.HistogramData, len(j.Requests))
	for i := range buckets {
		redistributedBuckets[i] = make(chan histogram.HistogramData)
	}

	var initialBuckets = make([]histogram.HistogramData, len(j.Requests))
	for i := range buckets {
		_, initialBuckets[i] = histogram.New(j.Requests[i].Name, 5, 100, graphs[i], buckets[i], redistributedBuckets[i])
	}

	var tui = configureTUI(initialBuckets, goutils.MergeChannels(buckets...), goutils.MergeChannels(redistributedBuckets...))

	// stream makes the requests and sends the timings to journeyResponses
	go func() {
		err = j.Stream(uint16(totalNumberOfRequests), uint8(concurrentRequests), journeyResponses)
		if err != nil {
			errLog.WriteString(fmt.Sprintf("error running Stream(): %v\n", err)) // TODO: i dont love this
		}
	}()

	return tui, nil
}

func fanOut(samples <-chan journey.RequestDuration, graphs ...chan journey.RequestDuration) {
	for sample := range samples {
		graphs[sample.ID] <- sample
	}
	for _, graph := range graphs {
		close(graph)
	}
}
