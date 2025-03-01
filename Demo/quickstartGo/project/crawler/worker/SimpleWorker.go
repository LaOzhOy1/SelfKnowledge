package worker

import (
	"../../../datastruct/set"
	"../engine"
	"../fetcher"
	"../parser/zhenai"
	"fmt"
	"log"
)

type SimpleWorker struct {
	id         int
	tasks      chan engine.Request
	results    chan engine.ParseResult
	parseFunc  engine.ParseFunc
	requestSet *set.ReadWriteHashSet
}

func (s *SimpleWorker) DoWork() error {
	go func() {
		for true {
			//
			request := <-s.tasks

			if s.requestSet.Contains(request.Url) {
				continue
			}
			s.requestSet.Put(request.Url)
			log.Printf("worker %d do worker: %s\n", s.id, request.Url)
			body, err := fetcher.Fetch(request.Url)
			if err != nil {
				fmt.Printf("fetch error : %s", err)
			}
			s.results <- s.parseFunc(body)
		}
	}()
	return nil
}

// 返回Worker 无需声明指针
func CreateSimpleWorker(id int, requestSet *set.ReadWriteHashSet,
	tasks chan engine.Request, results chan engine.ParseResult) engine.Worker {
	// 由于调用方法中有指针调用者，需返回引用
	//return &SimpleWorker{
	//	tasks:   tasks,
	//	results: results,
	//}
	// 无需返回直接调用
	worker := &SimpleWorker{
		id:         id,
		tasks:      tasks,
		results:    results,
		parseFunc:  zhenai.ParseCityList,
		requestSet: requestSet,
	}
	err := worker.DoWork()
	if err != nil {
		fmt.Printf("%s", err)
	}
	return worker
}
