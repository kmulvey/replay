package main

import (
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

type BarChartConfig struct {
	title   string
	buckets []histogram.Bucket
}

func configureTUI(configs []BarChartConfig, buckets <-chan histogram.Bucket) *tview.Application {
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
		charts[config.title] = chart
		rows[rowNum].AddItem(chart, 0, 1, false)
	}

	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	for _, row := range rows {
		layout.AddItem(row, 0, 1, false)
	}
	app.SetRoot(layout, true).EnableMouse(true)

	go func() {
		for bucket := range buckets {
			charts[bucket.HistogramName].SetBarValue(bucket.Range, int(bucket.Count))
			app.Draw()
		}
	}()

	return app
}

func createBarChart(config BarChartConfig) *tvxwidgets.BarChart {
	barGraph := tvxwidgets.NewBarChart()
	barGraph.SetBorder(true)
	barGraph.SetTitle(config.title)

	for i, bucket := range config.buckets {
		barGraph.AddBar(bucket.Range, int(bucket.Count), barColors[i])
	}

	barGraph.SetMaxValue(int(time.Second))
	barGraph.SetAxesColor(tcell.ColorAntiqueWhite)
	barGraph.SetAxesLabelColor(tcell.ColorAntiqueWhite)
	return barGraph
}
