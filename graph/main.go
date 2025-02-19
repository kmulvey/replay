package main

import (
	"flag"
	"math/rand"

	"github.com/gdamore/tcell"
	"github.com/kmulvey/replay/histogram"
	"github.com/navidys/tvxwidgets"
	"github.com/rivo/tview"
)

func main() {

	var totalNumberOfRequests int
	var concurrentRequests int
	var harFile string
	flag.IntVar(&totalNumberOfRequests, "total", 100, "total number of requests")
	flag.IntVar(&concurrentRequests, "concurrent", 10, "number of concurrent requests")
	flag.StringVar(&harFile, "har", "localhost.har", "har file to replay")
	flag.Parse()

}

func configureTUI(journeyLength uint8) *tview.Application {
	app := tview.NewApplication()
	rows := make([]*tview.Flex, journeyLength/3)

	for i := range journeyLength {
	}
	firstRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	firstRow.AddItem(bmLineChart, 0, 1, false)
	firstRow.SetRect(0, 0, 100, 15)

	secondRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	secondRow.SetRect(0, 0, 100, 15)

	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	layout.AddItem(firstRow, 0, 1, false)
	layout.AddItem(secondRow, 0, 1, false)
	layout.SetRect(0, 0, 100, 30)

	if err := app.SetRoot(layout, false).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func createBarChart(title string, buckets []histogram.Bucket) *tvxwidgets.BarChart {
	barGraph := tvxwidgets.NewBarChart()
	barGraph.SetRect(4, 2, 50, 20)
	barGraph.SetBorder(true)
	barGraph.SetTitle(title)
	// display system metric usage
	for _, bucket := range buckets {
		barGraph.AddBar(bucket.Range, int(bucket.Count), tcell.NewHexColor(rand.Intn(0xFFFFFF)))
	}
	barGraph.SetMaxValue(100)
	barGraph.SetAxesColor(tcell.ColorAntiqueWhite)
	barGraph.SetAxesLabelColor(tcell.ColorAntiqueWhite)
	return barGraph
}
