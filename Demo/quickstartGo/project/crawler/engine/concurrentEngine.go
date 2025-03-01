package engine

import (
	"fmt"
	"time"
)

// 类似Spring Boot Run 修改Engine中的变量

func (e *ConcurrentEngine) Run(seeds ...Request) {

	// 创建worker 消费调度器
	// worker 在Run函数中创建 Run 方法只调用一次 in out 唯一

	tasks := make(chan Request)
	results := make(chan ParseResult)

	// tasks 在Scheduler 中 是私有变量
	e.Scheduler.ConfigureMasterWorkerChan(tasks)
	// 先创建消费者
	for i := 0; i < e.WorkerCount; i++ {
		e.CreateWorkerFunc(i, e.RequestSet, tasks, results)
	}
	// 将任务提交给调度器任务channel中
	//  !!!! 出现循环死锁现象， 第一次将seed提交给Scheduler后，分发给各个worker，各个worker将返回结果打印，并将新的url发送给Scheduler，此时是无阻塞的
	// 但是由于tasks channel是无缓冲channel， 在多个worker submit完之后，如果出现worker没有及时处理完新的任务（网络较慢），
	//则未处理的任务将阻塞tasks channel ，当worker完成任务后无法提交任务，也无法获取新的任务
	// 个人认为: 1. 给worker和 channel 一定的缓冲区 --》 但是有可能任务很多，很慢，依然是进入循环等待
	// 视频使用 每次submit 都开启一个协程进行submit -- 》 切换时间大于计算时间，每次都开启新的协程进行任务提交，没有直观的等待队列可以观察
	// 第二种请见 concurrentEngineWithQueue
	for _, r := range seeds {
		//fmt.Printf("submit task : %s",r.Url)
		err := e.Scheduler.Submit(r)
		if err != nil {
			fmt.Printf("request error :%s\n", err)
		}
	}

	// 如何准确的判断爬取完成
	for true {
		select {
		case result := <-results:
			for _, item := range result.Items {
				e.ItemSet.Put(item)
				//log.Printf("ParseResult Item :%s\n",item)
			}
			for _, request := range result.Requests {
				err := e.Scheduler.Submit(request)
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
