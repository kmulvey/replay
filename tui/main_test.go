package main

import "testing"

func TestXXX(t *testing.T) {
	t.Parallel()

	tui, err := journeyUI("../localhost.har", 200, 2)
	if err != nil {
		panic(err)
	}
	tui.Run()
}
