// Demo code for the bar chart primitive.
package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/navidys/tvxwidgets"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	barGraph := tvxwidgets.NewBarChart()
	barGraph.SetRect(4, 2, 50, 20)
	barGraph.SetBorder(true)
	barGraph.SetTitle("System Resource Usage")
	// display system metric usage
	barGraph.AddBar("cpu", 100, tcell.ColorBlue)
	barGraph.AddBar("mem", 20, tcell.ColorRed)
	barGraph.AddBar("swap", 40, tcell.ColorGreen)
	barGraph.AddBar("disk", 40, tcell.ColorOrange)
	barGraph.SetMaxValue(100)
	barGraph.SetAxesColor(tcell.ColorAntiqueWhite)
	barGraph.SetAxesLabelColor(tcell.ColorAntiqueWhite)

	update := func() {
		for pct := 100; pct >= 0; pct-- {
			barGraph.SetBarValue("cpu", pct)
			app.Draw()
			time.Sleep(100 * time.Millisecond)
		}
	}
	go update()

	if err := app.SetRoot(barGraph, false).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
