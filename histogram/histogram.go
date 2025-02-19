package histogram

import (
	"fmt"
	"sort"
	"time"

	"github.com/kmulvey/replay/journey"
)

// Define a histogram bucket structure
type Histogram struct {
	buckets             []Bucket
	durations           []time.Duration
	bucketSize          time.Duration
	min                 time.Duration
	max                 time.Duration
	redistributeInerval uint32
}

type Bucket struct {
	Range string
	Count uint64
}

// New sreates a new histogram with a predefined number of buckets. The bucket ranges
// are 0-1s initially but will be recalculated as the data comes in.
func New(numBuckets uint8, redistributeInerval uint32, samples <-chan journey.RequestDuration, buckets chan<- Bucket) *Histogram {
	if numBuckets == 0 {
		numBuckets = 5
	}

	// Create and return the histogram
	var h = &Histogram{
		buckets:             make([]Bucket, numBuckets),
		durations:           make([]time.Duration, 0),
		bucketSize:          time.Second / time.Duration(numBuckets),
		min:                 0,
		max:                 time.Second,
		redistributeInerval: redistributeInerval,
	}

	// Initialize the buckets
	for i := range h.buckets {
		start := h.min + time.Duration(i)*h.bucketSize
		end := start + h.bucketSize
		h.buckets[i] = Bucket{
			Range: fmt.Sprintf("%v - %v", start, end),
			Count: 0,
		}
	}

	// Send the initial buckets
	for _, bucket := range h.buckets {
		buckets <- bucket
	}

	go h.collect(samples, buckets)
	return h
}

func (h *Histogram) collect(samples <-chan journey.RequestDuration, buckets chan<- Bucket) {
	var seen uint32
	for sample := range samples {
		h.insertDuration(sample.Duration)
		buckets <- h.bucketDuration(sample.Duration)

		seen++
		if seen == h.redistributeInerval {
			h.redistributeBuckets()
			for _, bucket := range h.buckets {
				buckets <- bucket
			}
			seen = 0
		}
	}
	close(buckets)
}

func (h *Histogram) redistributeBuckets() {
	bucketCount := len(h.buckets)
	if bucketCount == 0 {
		return
	}

	// Clear the current buckets
	for i := range h.buckets {
		h.buckets[i].Count = 0
	}

	// Calculate the new bucket size based on the min and max durations
	h.bucketSize = (h.max - h.min) / time.Duration(bucketCount)

	// Redistribute the durations into the new buckets
	for _, duration := range h.durations {
		bucketIndex := int((duration - h.min) / h.bucketSize)
		if bucketIndex >= bucketCount {
			bucketIndex = bucketCount - 1
		}
		h.buckets[bucketIndex].Count++
	}

	// Update the range for each bucket
	for i := range h.buckets {
		start := h.min + time.Duration(i)*h.bucketSize
		end := start + h.bucketSize
		h.buckets[i].Range = fmt.Sprintf("%v - %v", start, end)
	}
}

// insertDuration inserts a time.Duration into a sorted slice while maintaining order.
func (h *Histogram) insertDuration(value time.Duration) {
	// Find the insertion index using binary search
	index := sort.Search(len(h.durations), func(i int) bool {
		return h.durations[i] >= value
	})

	// Insert the value at the correct position
	h.durations = append(h.durations, 0)             // Add a dummy element to extend the slice
	copy(h.durations[index+1:], h.durations[index:]) // Shift elements to the right
	h.durations[index] = value                       // Insert the new value
	h.min = h.durations[0]
	h.max = h.durations[len(h.durations)-1]
}

// bucketDuration adds the duration to the correct bucket.
func (h *Histogram) bucketDuration(sample time.Duration) Bucket {
	// Determine the bucket index for the duration
	if sample < h.min {
		sample = h.min
	}
	if sample > h.max {
		sample = h.max
	}
	bucketIndex := int((sample - h.min) / h.bucketSize)

	// Ensure the bucket index is within the valid range
	if bucketIndex >= len(h.buckets) {
		bucketIndex = len(h.buckets) - 1
	}

	// Increment the appropriate bucket
	h.buckets[bucketIndex].Count++
	return h.buckets[bucketIndex]
}

// Function to print the histogram
func (h *Histogram) Print() {
	fmt.Println("Request Duration Histogram:")
	for i, count := range h.buckets {
		// Calculate the range for each bucket
		start := h.min + time.Duration(i)*h.bucketSize
		end := start + h.bucketSize
		fmt.Printf("[%v - %v): %d requests\n", start, end, count.Count)
	}
}
