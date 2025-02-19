package replay

import (
	"testing"
	"time"

	"github.com/kmulvey/goutils"
	"github.com/kmulvey/replay/histogram"
	"github.com/kmulvey/replay/journey"
	"github.com/stretchr/testify/assert"
)

func TestRunBenchmark(t *testing.T) {
	t.Parallel()

	var j, err = journey.New("localhost.har")
	assert.NoError(t, err)

	var responses = make(chan journey.RequestDuration)
	go func() {
		err = j.Stream(200, 2, responses)
		assert.NoError(t, err)
	}()

	var graphs = make([]chan journey.RequestDuration, 3)
	for i := range graphs {
		graphs[i] = make(chan journey.RequestDuration)
	}
	go fanOut(responses, graphs...)

	var buckets = make([]chan histogram.Bucket, 3)
	for i := range buckets {
		buckets[i] = make(chan histogram.Bucket)
	}
	var done = make(chan struct{})

	go func() {
		for bucket := range goutils.MergeChannels(buckets...) {
			t.Logf("One: %+v", bucket)
		}
		close(done)
	}()

	time.Sleep(time.Second)

	one := histogram.New(5, 10, graphs[0], buckets[0])
	two := histogram.New(5, 10, graphs[1], buckets[1])
	three := histogram.New(5, 10, graphs[2], buckets[2])

	for {
		select {
		case <-done:
			return
		case <-time.After(time.Second * 10):
			one.Print()
			two.Print()
			three.Print()
		}
	}
}

func fanOut(samples <-chan journey.RequestDuration, graphs ...chan journey.RequestDuration) {
	for sample := range samples {
		graphs[sample.ID] <- sample
	}
	for _, graph := range graphs {
		close(graph)
	}
}
