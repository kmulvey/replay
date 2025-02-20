package journey

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// BenchmarkXxx-32    	      13	  77846182 ns/op

// func BenchmarkXxx(b *testing.B) {

// 	var j, err = New("../localhost.har")
// 	assert.NoError(b, err)

// 	for b.Loop() {
// 		j.makeRequest(j.Requests[0])
// 	}
// }

func TestMakeRequest(t *testing.T) {

	var j, err = New("../localhost.har")
	assert.NoError(t, err)

	var req = j.Requests[0]
	var responses = make(chan Response)
	var done = make(chan struct{})
	go func() {
		for resp := range responses {
			assert.NoError(t, resp.error)
			//fmt.Println(resp.Duration)
		}
		close(done)
	}()
	j.makeRequest(req, 50, responses)
	<-done
}
