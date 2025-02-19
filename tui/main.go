package main

import (
	"flag"
	"os"
	"os/signal"

	"github.com/kmulvey/goutils"
	"github.com/kmulvey/replay/histogram"
	"github.com/kmulvey/replay/journey"
	"github.com/rivo/tview"
)

func main() {

	var totalNumberOfRequests int
	var concurrentRequests int
	var harFile string
	flag.IntVar(&totalNumberOfRequests, "total", 100, "total number of requests")
	flag.IntVar(&concurrentRequests, "concurrent", 10, "number of concurrent requests")
	flag.StringVar(&harFile, "har", "../localhost.har", "har file to replay")
	flag.Parse()

	var tui, err = journeyUI(harFile, totalNumberOfRequests, concurrentRequests)
	if err != nil {
		panic(err)
	}

	go func() {
		if err := tui.Run(); err != nil {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	tui.Stop()
	os.Exit(0)
}

func journeyUI(harFile string, totalNumberOfRequests, concurrentRequests int) (*tview.Application, error) {

	var j, err = journey.New(harFile)
	if err != nil {
		return nil, err
	}

	// journeyResponses are the timings of each request
	var journeyResponses = make(chan journey.RequestDuration)

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

	var initialBuckets = make([]BarChartConfig, len(j.Requests))
	for i := range buckets {
		var chartConfig = BarChartConfig{
			title: j.Requests[i].Name,
		}
		_, bucket := histogram.New(j.Requests[i].Name, 5, 10, graphs[i], buckets[i])
		chartConfig.buckets = bucket
		initialBuckets[i] = chartConfig
	}

	var tui = configureTUI(initialBuckets, goutils.MergeChannels(buckets...))

	// stream makes the requests and sends the timings to journeyResponses
	go func() {
		err = j.Stream(uint16(totalNumberOfRequests), uint8(concurrentRequests), journeyResponses)
		if err != nil {
			panic(err) // TODO: i dont love this
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
