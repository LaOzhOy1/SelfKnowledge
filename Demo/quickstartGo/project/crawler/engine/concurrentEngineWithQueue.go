package engine

import (
	"fmt"
	"log"
	"time"
)

func (m *ConcurrentEngineWithQueue) Run(seeds ...Request) {
	results := make(chan ParseResult)
	saver := m.CreateSaverFunc()
	// 开启调度器，负责开启两个通道，接受request 与 workerChannel 放到队列中等待派遣
	m.QueueScheduler.Run()

	// 创建多个worker 每个worker都有一条独自的channel
	for i := 0; i < m.WorkerCount; i++ {
		m.CreateWorkerFunc(i, m.RequestSet, m.QueueScheduler, results)
	}
	for _, seed := range seeds {
		err := m.QueueScheduler.Submit(seed)
		if err != nil {
			fmt.Printf("submit error :%s", err)
		}
	}

	// 如何准确的判断爬取完成
	for true {
		select {
		case result := <-results:
			for _, item := range result.Items {
				if result.ItemType == "city" {
					m.ItemSet.Put(item)
				} else if result.ItemType == "personList" {

				} else if result.ItemType == "personProfile" {

					// 存储到数据库 --》 子携程
					saver <- item
				}
				log.Printf("ParseResult Item :%s\n", item)
			}
			for _, request := range result.Requests {
				err := m.QueueScheduler.Submit(request)
				if err != nil {
					fmt.Printf("request error :%s\n", err)
				}
			}
		case <-time.After(time.Second * 5):
			fmt.Printf("no task in tasks to run!")
			return
		}
	}

}
