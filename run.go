package replay

import (
	"sort"
	"time"
)

func sortAndAverageTimings(timings [][]time.Duration) ([]time.Duration, time.Duration) {

	var allTimings = make([]time.Duration, len(timings)*len(timings[0]))
	var totalDuration time.Duration
	var slowestRequest time.Duration
	var fastestRequest time.Duration
	var i int

	for _, timing := range timings {
		for _, duration := range timing {

			allTimings[i] = duration
			totalDuration += duration
			if duration > slowestRequest {
				slowestRequest = duration
			}
			if duration < fastestRequest {
				fastestRequest = duration
			}

			i++
		}
	}

	sort.Slice(allTimings, func(i, j int) bool {
		return allTimings[i] < allTimings[j]
	})

	return allTimings, totalDuration / time.Duration(i)
}
