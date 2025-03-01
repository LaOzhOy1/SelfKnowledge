package main

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

/*
通过 channel实现routine 同步

通过 mutex 实现routine 同步
*/

func main() {
	// 不要执行 会导致 cpu 和 内存占用标高
	//number := math.MaxInt64
	//for i := 0; i < number; i++ {
	//	go func() {
	//		fmt.Print("hello world")
	//		time.Sleep(time.Second)
	//	}()
	//}

	//chanel 实现协程同步

	// mutex 实现协程同步

	// waitgroup
	for i := 0; i < 2; i++ {
		go incrShareCnt()
	}

	time.Sleep(1000 * time.Second)
}

var share_cnt uint64 = 0

var lck sync.Mutex

func showMessageByChannel() {
	//
	msg_chan := make(chan string)
	done := make(chan bool)

	i := 0
	// 生产者
	go func() {
		for {
			i++
			time.Sleep(1 * time.Second)
			msg_chan <- "on message"
			// 阻塞
			<-done
		}
	}()
	// 消费者
	go func() {
		for {
			select {
			// 其他情况阻塞
			case msg := <-msg_chan:
				i++
				fmt.Println(msg + " " + strconv.Itoa(i))
				time.Sleep(2 * time.Second)
				done <- true
			}
		}

	}()
	time.Sleep(20 * time.Second)
}

func incrShareCnt() {
	for i := 0; i < 1000000; i++ {
		lck.Lock()
		share_cnt++
		lck.Unlock()
	}

	fmt.Println(share_cnt)

}

func showUrlsByWaitGroup() {
	for _, url := range urls {
		// Increment the WaitGroup counter.
		wg.Add(1)
		// Launch a goroutine to fetch the URL.
		go func(url string) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// Fetch the URL.
			http.Get(url)
		}(url)
	}
}

var wg sync.WaitGroup
var urls = []string{
	"http://www.baidu.com/",
	"http://www.taobao.com/",
	"http://www.tianmao.com/",
}
