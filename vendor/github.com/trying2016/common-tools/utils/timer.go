package utils

import (
	"time"
)

func StartTimeWithChan(callback func(), interval int, exitSignal chan struct{}) {
	go func() {
		for {
			select {
			case <-time.After(time.Duration(interval) * time.Millisecond):
				callback()
			case <-exitSignal:
				return
			}
		}
	}()
}

func StartTime(callback func(), interval int) (exitSignal chan struct{}) {
	exitSignal = make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(time.Duration(interval) * time.Millisecond):
				callback()
			case <-exitSignal:
				return
			}
		}
	}()
	return exitSignal
}

func StopTime(exitSignal chan struct{}) {
	exitSignal <- struct{}{}
}
