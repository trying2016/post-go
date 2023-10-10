package utils

import (
	"fmt"
	"sync"
)

// StartThread 开启线程组
func StartThread(num int, fn func(i int) bool) {
	var job sync.WaitGroup
	job.Add(num)
	for i := 0; i < num; i++ {
		SafeGo(func() {
			for {
				if !fn(0) {
					break
				}
			}
			job.Done()
		}, func(err interface{}) {
			fmt.Println(err)
		})
	}
	job.Wait()
}
