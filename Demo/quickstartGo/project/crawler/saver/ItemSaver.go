package saver

import "log"

func ItemSaver() chan interface{} {
	saveChannel := make(chan interface{})
	go func() {
		itemCount := 0
		for true {
			item := <-saveChannel
			log.Printf("saver receive number :%d,%v", itemCount, item)
			itemCount++
		}
	}()
	return saveChannel
}
