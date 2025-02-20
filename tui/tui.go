package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/kmulvey/replay/histogram"
	"github.com/navidys/tvxwidgets"
	"github.com/rivo/tview"
)

var barColors = []tcell.Color{
	tcell.ColorRed,
	tcell.ColorOrange,
	tcell.ColorYellow,
	tcell.ColorGreen,
	tcell.ColorBlue,
	tcell.ColorIndigo,
	tcell.ColorViolet,
}

func configureTUI(configs []histogram.HistogramData, buckets <-chan histogram.Bucket, redistributedBuckets <-chan histogram.HistogramData) *tview.Application {
	app := tview.NewApplication()
	numRows := (len(configs) + 2) / 3
	rows := make([]*tview.Flex, numRows)
	charts := make(map[string]*tvxwidgets.BarChart, len(configs))

	var rowNum int
	rows[rowNum] = tview.NewFlex().SetDirection(tview.FlexColumn)
	for i, config := range configs {
		if i%3 == 0 && i != 0 {
			rowNum++
			rows[rowNum] = tview.NewFlex().SetDirection(tview.FlexColumn)
		}
		var chart = createBarChart(config)
		charts[config[0].HistogramName] = chart
		rows[rowNum].AddItem(chart, 0, 1, false)
	}

	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	for _, row := range rows {
		layout.AddItem(row, 0, 1, false)
	}
	app.SetRoot(layout, true)

	go func() {
		// we need to index the bar names so we can remove them when the buckets are redistributed
		var barNames = make(map[string][]string, len(charts))
		for _, config := range configs {
			barNames[config[0].HistogramName] = make([]string, len(config))
			for i, bucket := range config {
				barNames[config[0].HistogramName][i] = bucket.Range
			}
		}
		var ticker = time.NewTicker(time.Millisecond * 500)
		for buckets != nil && redistributedBuckets != nil {
			select {
			case bucket, open := <-buckets:
				if !open {
					buckets = nil
					continue
				}
				if chart, found := charts[bucket.HistogramName]; found {
					chart.SetBarValue(bucket.Range, int(bucket.Count))
				} else {
					file, _ := os.OpenFile("errors.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					file.WriteString(fmt.Sprintf("chart not found: %s\n", bucket.HistogramName))
					panic(charts)
				}

			case newBuckets, open := <-redistributedBuckets:
				if !open {
					redistributedBuckets = nil
					continue
				}

				for _, name := range barNames[newBuckets[0].HistogramName] {
					charts[newBuckets[0].HistogramName].RemoveBar(name)
				}

				barNames[newBuckets[0].HistogramName] = make([]string, len(newBuckets))
				for i, bucket := range newBuckets {
					charts[newBuckets[0].HistogramName].AddBar(bucket.Range, int(bucket.Count), barColors[i])
					barNames[newBuckets[0].HistogramName][i] = bucket.Range
					if bucket.Count > 100 {
						charts[newBuckets[0].HistogramName].SetMaxValue(int(bucket.Count))
					}
				}

			case <-ticker.C:
				app.Draw()
			}
		}
	}()

	return app
}

func createBarChart(config histogram.HistogramData) *tvxwidgets.BarChart {
	barGraph := tvxwidgets.NewBarChart()
	barGraph.SetBorder(true)
	barGraph.SetTitle(config[0].HistogramName)

	for i, bucket := range config {
		barGraph.AddBar(bucket.Range, int(bucket.Count), barColors[i])
	}

	barGraph.SetMaxValue(100)
	barGraph.SetAxesColor(tcell.ColorAntiqueWhite)
	barGraph.SetAxesLabelColor(tcell.ColorAntiqueWhite)
	return barGraph
}
