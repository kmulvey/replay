package journey

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// BenchmarkXxx-32    	      13	  77846182 ns/op
// BenchmarkXxx-32    	     956	   1215107 ns/op

func BenchmarkXxx(b *testing.B) {

	var j, err = New("../localhost.har")
	assert.NoError(b, err)

	var client = makeClient()
	req, err := makeRequest(j.Requests[0])
	assert.NoError(b, err)

	for b.Loop() {
		j.runRequest(client, req, 200)
	}
}

func TestMakeRequest(t *testing.T) {

	fmt.Println(os.Getpid())
	var j, err = New("../localhost.har")
	assert.NoError(t, err)

	var reqConfig = j.Requests[0]
	var responses = make(chan RequestDuration)
	var done = make(chan struct{})
	go func() {
		var i int
		for resp := range responses {
			assert.NoError(t, resp.Error)
			//fmt.Println(resp.Duration)
			if resp.Error != nil {
				panic(i)
			}
			i++
			if i%1000 == 0 {
				fmt.Println(i)
			}
		}
		close(done)
	}()

	req, err := makeRequest(reqConfig)
	assert.NoError(t, err)

	j.runRequest(makeClient(), req, 200) //65535
	<-done
}
