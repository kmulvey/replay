package journey

import (
	"fmt"
	"os"
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

	fmt.Println(os.Getpid())
	var j, err = New("../localhost.har")
	assert.NoError(t, err)

	var req = j.Requests[0]
	var responses = make(chan RequestDuration)
	var done = make(chan struct{})
	go func() {
		var i int
		for resp := range responses {
			assert.NoError(t, resp.error)
			//fmt.Println(resp.Duration)
			if resp.error != nil {
				panic(i)
			}
			i++
			if i%1000 == 0 {
				fmt.Println(i)
			}
		}
		close(done)
	}()
	j.makeRequest(req, 65535, responses) //65535
	<-done
}
