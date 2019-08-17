package routines

import (
	"annotator/server/lib"
	"fmt"
	"time"
)

// Delete removes annotations older than retention period
func Delete(ch chan<- string) {
	i := 0
	for {
		lib.Mu.Lock()
		fmt.Printf("Delete: %d\n", i)
		i++
		lib.Mu.Unlock()
		if i == 10 {
			ch <- fmt.Sprintf("%s", "Delete")
			return
		}
		time.Sleep(10 * time.Second)
	}
}
