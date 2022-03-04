package channels

import (
	"fmt"
	"sync"
)

// closes the channel when the wait group is done waiting
func CloseWhenDone(wg *sync.WaitGroup, sigChan chan<- interface{}) {
	go func() {
		wg.Wait()
		close(sigChan)
	}()
}

type MergedChannel struct {
	OnClosed      func()
	ReceivingChan <-chan interface{}
}

// returns a wait group that is done when all the passed channels have been closed
func Merge(cs ...MergedChannel) *sync.WaitGroup {
	var wg sync.WaitGroup

	doneOnClose := func(c MergedChannel) {
		for v := range c.ReceivingChan {
			if v != nil {
				fmt.Printf("%v\n", v)
			}
		}
		// invoke on closed function
		c.OnClosed()
		wg.Done()
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go doneOnClose(c)
	}

	return &wg
}
