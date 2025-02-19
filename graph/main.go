package main

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/navidys/tvxwidgets"
	"github.com/rivo/tview"
)

func main() {

	app := tview.NewApplication()

	sinData := func() [][]float64 {
		n := 220
		data := make([][]float64, 2)
		data[0] = make([]float64, n)
		data[1] = make([]float64, n)
		for i := 0; i < n; i++ {
			data[0][i] = 1 + math.Sin(float64(i+1)/5)
			// Avoid taking Cos(0) because it creates a high point of 2 that
			// will never be hit again and makes the graph look a little funny
			data[1][i] = 1 + math.Cos(float64(i+1)/5)
		}
		return data
	}()

	bmLineChart := tvxwidgets.NewPlot()
	bmLineChart.SetBorder(true)
	bmLineChart.SetTitle("line chart (braille mode)")
	bmLineChart.SetLineColor([]tcell.Color{
		tcell.ColorSteelBlue,
		tcell.ColorGreen,
	})
	bmLineChart.SetMarker(tvxwidgets.PlotMarkerBraille)
	bmLineChart.SetData(sinData)
	bmLineChart.SetDrawXAxisLabel(false)

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
