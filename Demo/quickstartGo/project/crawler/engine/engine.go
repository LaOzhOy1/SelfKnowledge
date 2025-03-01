package engine

import (
	"../../../datastruct/set"
	"../fetcher"
	"fmt"
	"log"
)

// 维护一个队列 并将队列的数据进行重复爬取

func Run(parseFunc ParseFunc, seeds ...Request) {
	var requests []Request
	for _, seed := range seeds {
		requests = append(requests, seed)
	}

	requestSet := set.NewHashSet()

	citySet := set.NewHashSet()

	for len(requests) > 0 {
		//hash of unhashable type engine.Request
		r := requests[0]
		if requestSet.Contains(r) {
			requests = requests[1:]
			continue
		}
		requestSet.Put(r)
		requests = requests[1:]
		// 对url内容进行抓取  invalid memory address or nil pointer dereference
		body, err := fetcher.Fetch(r.Url)

		if err != nil {
			log.Printf("Fetcher : error : %s ,%v", r.Url, err)
			continue
		} else {
			log.Printf("Fetcher  : %s", r.Url)
		}
		// 根据request中的parseFunc 解析返回 解析后的数据
		parseResult := parseFunc(body)
		// 将解析结果中新的解析request 放入队列中
		requests = append(requests, parseResult.Requests...)
		for _, item := range parseResult.Items {
			//log.Printf("got Item : %s", item)
			citySet.Put(item)
		}
	}
	fmt.Printf("爬取到的城市数量为 :%d", citySet.Size())
	citySet.Print()
}
