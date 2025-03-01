package main

import "fmt"
import "./engine"
import "./scheduler"
import "./worker"
import "./parser/zhenai"
import "../../datastruct/set"
import "./saver"

/**
单任务模型
爬取到的城市数量为 :5169
开始时间  2021/04/30 15:31:22 Fetcher  : https://www.zhenai.com/zhenghun
结束时间  2021/04/30 15:44:04 Fetcher  : http://www.zhenai.com/zhenghun/zunyixian
拉取城市链接过慢
并发模型 100 channel
开始时间  2021/05/02 16:47:43 worker 4 do worker: https://www.zhenai.com/zhenghun
结束时间  2021/05/02 16:47:59 worker 57 do worker: http://www.zhenai.com/zhenghun/linying
总计 16 s
no task in tasks to run!requestSet size are :2625, itemSet size are :5169

并发模型2 100 channel
开始时间  2021/05/02 16:48:04 worker 3 do worker: https://www.zhenai.com/zhenghun
结束时间  2021/05/02 16:48:20 worker 2 do worker: http://www.zhenai.com/zhenghun/luxi1
总计 16 s
no task in tasks to run!requestSet size are :2625, itemSet size are :5169

*/

const url = "https://www.zhenai.com/zhenghun"

func main() {
	//  定义了解析的url 和解析的方法
	//package command-line-arguments
	//	imports ./engine
	//	imports ../zhenai/parser
	//	imports ../../engine: import cycle not allowed

	// 单任务版
	// 对方法中直接调用其他包里的方法改为参数传递 原因是engine包中声明了接口，但在使用过程中直接调用其他包中接口实现类使用，而接口实现类引用了engine包
	//engine.Run(zhenai.ParseCityList, engine.Request{
	//	Url: url,
	//})

	// 并发版本1
	//concurrentEngine := &engine.ConcurrentEngine{
	//	// 为什么要传&？ 方法调用为*
	//	Scheduler:        &scheduler.SimpleScheduler{},
	//	WorkerCount:      100,
	//	CreateWorkerFunc: worker.CreateSimpleWorker,
	//	RequestSet:       set.NewReadWriteHashSet(),
	//	ItemSet:          set.NewReadWriteHashSet(),
	//}
	//concurrentEngine.Run(engine.Request{Url: url})
	//fmt.Printf("requestSet size are :%d, itemSet size are :%d", concurrentEngine.RequestSet.Size(), concurrentEngine.ItemSet.Size())

	concurrentEngineWithQueue := &engine.ConcurrentEngineWithQueue{
		QueueScheduler:   &scheduler.QueueScheduler{},
		WorkerCount:      100,
		CreateWorkerFunc: worker.CreateQueueWorker,
		CreateSaverFunc:  saver.ItemSaver,
		ItemSet:          set.NewReadWriteHashSet(),
		RequestSet:       set.NewReadWriteHashSet(),
	}
	concurrentEngineWithQueue.Run(engine.Request{
		Url:       url,
		ParseFunc: zhenai.ParseCityList,
	})
	fmt.Printf("requestSet size are :%d, itemSet size are :%d", concurrentEngineWithQueue.RequestSet.Size(), concurrentEngineWithQueue.ItemSet.Size())

}

//
