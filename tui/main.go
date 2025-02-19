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

	var tui = journeyUI(harFile, totalNumberOfRequests, concurrentRequests)
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

func journeyUI(harFile string, totalNumberOfRequests, concurrentRequests int) *tview.Application {

	var j, err = journey.New(harFile)
	if err != nil {
		return nil
	}

	var responses = make(chan journey.RequestDuration)
	var graphs = make([]chan journey.RequestDuration, len(j.Requests))
	for i := range graphs {
		graphs[i] = make(chan journey.RequestDuration)
	}
	go fanOut(responses, graphs...)

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

	go func() {
		err = j.Stream(uint16(totalNumberOfRequests), uint8(concurrentRequests), responses)
		if err != nil {
			//return err
			// ???
		}
	}()

	return configureTUI(initialBuckets, goutils.MergeChannels(buckets...))
}

func fanOut(samples <-chan journey.RequestDuration, graphs ...chan journey.RequestDuration) {
	for sample := range samples {
		graphs[sample.ID] <- sample
	}
	for _, graph := range graphs {
		close(graph)
	}
}
