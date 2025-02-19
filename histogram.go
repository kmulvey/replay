package replay

import (
	"fmt"
	"time"
)

// Define a histogram bucket structure
type Histogram struct {
	buckets    []int
	bucketSize time.Duration
	min        time.Duration
	max        time.Duration
}

// Create a new histogram with a predefined number of buckets
func NewHistogram(numBuckets int, durations []time.Duration) *Histogram {
	// Calculate the min and max durations from the dataset
	var min, max time.Duration
	for _, d := range durations {
		if min == 0 || d < min {
			min = d
		}
		if max == 0 || d > max {
			max = d
		}
	}

	// Calculate the bucket size based on the range of the data
	rangeDuration := max - min
	bucketSize := rangeDuration / time.Duration(numBuckets)

	// If the bucket size is too small (0), set it to a default size
	if bucketSize == 0 {
		bucketSize = 1 * time.Millisecond
	}

	// Create and return the histogram
	return &Histogram{
		buckets:    make([]int, numBuckets),
		bucketSize: bucketSize,
		min:        min,
		max:        max,
	}
}

// Function to add a duration to the histogram
func (h *Histogram) Add(duration time.Duration) {
	// Determine the bucket index for the duration
	if duration < h.min {
		duration = h.min
	}
	if duration > h.max {
		duration = h.max
	}
	bucketIndex := int((duration - h.min) / h.bucketSize)

	// Ensure the bucket index is within the valid range
	if bucketIndex >= len(h.buckets) {
		bucketIndex = len(h.buckets) - 1
	}

	// Increment the appropriate bucket
	h.buckets[bucketIndex]++
}

// Function to print the histogram
func (h *Histogram) Print() {
	fmt.Println("Request Duration Histogram:")
	for i, count := range h.buckets {
		// Calculate the range for each bucket
		start := h.min + time.Duration(i)*h.bucketSize
		end := start + h.bucketSize
		fmt.Printf("[%v - %v): %d requests\n", start, end, count)
	}
}
