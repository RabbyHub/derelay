package relay

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSendChanWithNoReceiver(t *testing.T) {

	var wg sync.WaitGroup
	send := make(chan int, 3)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case i := <-send:
				fmt.Printf("received: %v\n", i)
				if i == 3 {
					fmt.Printf("receiving routine exit\n")
					return
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		i := 0
		for {
			select {
			case send <- i:
				fmt.Printf("send: %v\n", i)
				i++
				time.Sleep(1 * time.Second)
			default:
				fmt.Printf("send buffer is full\n")
				return
			}
		}
	}()

	wg.Wait()
}
