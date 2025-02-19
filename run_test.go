package replay

import (
	"testing"

	"github.com/kmulvey/replay/histogram"
	"github.com/kmulvey/replay/journey"
	"github.com/stretchr/testify/assert"
)

func TestRunBenchmark(t *testing.T) {
	t.Parallel()

	var j, err = journey.New("localhost.har")
	assert.NoError(t, err)

	var responses = make(chan journey.RequestDuration)
	err = j.Stream(100, 2, responses)
	assert.NoError(t, err)

	var graphs = make([]chan journey.RequestDuration, 2)

	var oneBuckets = make(chan histogram.Bucket)
	one := histogram.New(5, 10, graphs[0], oneBuckets)

	// // Print the histogram
	// histogram.Print()
}

func splitter(samples <-chan journey.RequestDuration, graphs ...chan<- journey.RequestDuration) {
	for sample := range samples {
		graphs[sample.ID] <- sample
	}
}
