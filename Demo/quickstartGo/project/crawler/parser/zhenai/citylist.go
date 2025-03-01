package zhenai

import (
	"../../engine"
	"regexp"
)

// 解析珍爱网城市数据
func ParseCityList(content []byte) engine.ParseResult {
	urls := getUrl(content)
	return getMetaData(urls)
}

// 同一包级别可以使用   获取所有符合条件的url
func getUrl(content []byte) [][]byte {
	//<a href="http://www.zhenai.com/zhenghun/aba" data-v-1573aa7c>阿坝</a>
	//re := regexp.MustCompile(`<a target="_blank" href="http://www.zhenai.com/zhenghun/[0-9a-z]+"[^>]*>[^<]+</a> `)
	re := regexp.MustCompile(`<a href="http://www.zhenai.com/zhenghun/[0-9a-zA-Z]+"[^>]*>[^<]+</a>`)
	matches := re.FindAll(content, -1)
	return matches
}

// 筛选url种有用的信息
func getMetaData(bytes [][]byte) engine.ParseResult {

	//re := regexp.MustCompile(`<a target="_blank" href="(http://www.zhenai.com/zhenghun/[0-9a-z]+)"[^>]*>([^<]+)</a> `)
	re := regexp.MustCompile(`<a href="(http://www.zhenai.com/zhenghun/[0-9a-zA-Z]+)"[^>]*>([^<]+)</a>`)

	var result engine.ParseResult
	result.ItemType = "city"
	// 将bytes 看作string   data为string[][]
	for _, data := range bytes {
		// 将bytes 看作string   data为string[][]
		matches := re.FindAllSubmatch(data, -1)
		//value 为 string[] 代表每一个网址 与其拆分结果
		for _, values := range matches {

			result.Items = append(result.Items, string(values[2]))
			result.Requests = append(result.Requests, engine.Request{
				Url: string(values[1]),
				// 加入解析城市种人物的逻辑
				ParseFunc: ParsePersonList,
			})
		}
	}
	return result
}
