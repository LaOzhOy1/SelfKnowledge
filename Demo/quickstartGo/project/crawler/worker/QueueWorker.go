package worker

import (
	"../../../datastruct/set"
	"../engine"
	"../fetcher"
	"../parser/zhenai"
	"fmt"
)

type QueueWorker struct {
	id             int
	parseFunc      engine.ParseFunc
	results        chan engine.ParseResult
	requestSet     *set.ReadWriteHashSet
	queueScheduler engine.QueueScheduler
	task           chan engine.Request
}

func (s *QueueWorker) DoWork() error {
	go func() {
		for true {
			// 阻塞 ，等待，并通过channel发送信号（channel）给主协程，通知channel准备就绪 比mutex 更加抽象，
			// 假如用mutex 实现：
			// lock锁住该协程，然后将mutex传到scheduler中，scheduler循环解锁， 如何实现信息传递呢？
			// 通过共享一段内存实现
			s.queueScheduler.WorkerReady(s.task)
			request := <-s.task
			if s.requestSet.Contains(request.Url) {
				continue
			}
			// 确保不会有重复url
			s.requestSet.Put(request.Url)
			//log.Printf("worker %d do worker: %s\n", s.id, request.Url)
			body, err := fetcher.Fetch(request.Url)
			if err != nil {
				fmt.Printf("fetch error : %s", err)
			}
			// 写死 灵活度不高
			//s.results <- s.parseFunc(body)
			s.results <- request.ParseFunc(body)
		}
	}()
	return nil
}

// 返回Worker 无需声明指针
func CreateQueueWorker(
	id int,
	requestSet *set.ReadWriteHashSet,
	queueScheduler engine.QueueScheduler, results chan engine.ParseResult) engine.Worker {

	// 每个worker都有自己的channel
	task := make(chan engine.Request)

	worker := &QueueWorker{
		id:             id,
		results:        results,
		parseFunc:      zhenai.ParseCityList,
		requestSet:     requestSet,
		queueScheduler: queueScheduler,
		task:           task,
	}
	err := worker.DoWork()
	if err != nil {
		fmt.Printf("%s", err)
	}
	return worker
}
