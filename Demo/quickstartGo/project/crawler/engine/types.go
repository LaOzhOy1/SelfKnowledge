package engine

import "../../../datastruct/set"

// 整个工程中需要的数据结构
// 封装爬取的url 和 爬取后解析的方法
type ConcurrentEngine struct {
	Scheduler        Scheduler
	WorkerCount      int
	CreateWorkerFunc func(int, *set.ReadWriteHashSet, chan Request, chan ParseResult) Worker
	RequestSet       *set.ReadWriteHashSet
	ItemSet          *set.ReadWriteHashSet
}

type ConcurrentEngineWithQueue struct {
	QueueScheduler   QueueScheduler
	WorkerCount      int
	CreateWorkerFunc func(int, *set.ReadWriteHashSet, QueueScheduler, chan ParseResult) Worker
	CreateSaverFunc  func() chan interface{}
	ItemSet          *set.ReadWriteHashSet
	RequestSet       *set.ReadWriteHashSet
}

type Scheduler interface {
	Submit(request Request) error
	ConfigureMasterWorkerChan(chan Request)
}

type QueueScheduler interface {
	Submit(request Request) error
	ConfigureMasterWorkerChan(chan Request, chan chan Request)
	Run()
	WorkerReady(chan Request)
}

type Worker interface {
	DoWork() error
}

type ParseFunc func([]byte) ParseResult

type Request struct {
	Url       string
	ParseFunc func([]byte) ParseResult
}

// 封装爬取结果 返回的字符串数组和新的请求
type ParseResult struct {
	Items    []interface{}
	ItemType string
	Requests []Request
}

// 模拟一个空解析结果返回
func NilParser([]byte) ParseResult {
	return ParseResult{}
}
