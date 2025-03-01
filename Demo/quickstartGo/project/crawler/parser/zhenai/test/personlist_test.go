package test

import (
	"../../../fetcher"
	"io/ioutil"
	"os"
	"testing"
)

const personlist_url = "https://www.zhenai.com/zhenghun/beijing"

const personlist_filename = "parser_person_list_test_data.txt"

func TestParsePersonList(t *testing.T) {

	/*	content, err := ioutil.ReadFile(personlist_filename)
		if err != nil {
			panic(err)
		}
		 zhenai.ParsePersonList(content)

		if len(result.Requests) != 24 {
			t.Errorf("result should have %d , but get %d ", stdParseDataSize, len(result.Requests))
		}
	*/
	//因为测试可能进行在内部环境，所以要将网上的数据进行拷贝存放在本地 ，而且已经知道返回的地点数量为固定24个
	contents, err := fetcher.Fetch(personlist_url)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%s\n",contents)
	permissions := 0664
	err2 := ioutil.WriteFile(personlist_filename, contents, os.FileMode(permissions))
	if err2 != nil {
		// handle error
	}
}
