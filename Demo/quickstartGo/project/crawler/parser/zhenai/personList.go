package zhenai

import (
	"../../engine"
	"log"
	"regexp"
)

// http://album.zhenai.com/u/1416774094
var personUrlRe = regexp.MustCompile(`<a href="([0-9A-Za-z/:.]+)" target="_blank">([^<]+)</a>`)

func ParsePersonList(contents []byte) engine.ParseResult {
	match := personUrlRe.FindAllSubmatch(contents, -1)
	var result engine.ParseResult
	result.ItemType = "personList"

	for _, list := range match {
		//result.Items = append(result.Items, string(list[2]))
		tmp := string(list[2])
		log.Printf("ParsePersonList name: %s\n", tmp)
		result.Requests = append(result.Requests,
			engine.Request{
				Url: string(list[1]),
				ParseFunc: func(bytes []byte) engine.ParseResult {
					return ParseProfile(bytes, tmp)
				},
			},
		)

	}
	return result
}
