package test

import (
	"../../zhenai"
	"fmt"
	"io/ioutil"
	"testing"
)

const person_profile_filename = "parser_person_profile_test_data.txt"
const person_profile_url = "https://album.zhenai.com/u/1740528662"

func TestParsePersonProfile(t *testing.T) {

	content, err := ioutil.ReadFile(person_profile_filename)
	if err != nil {
		panic(err)
	}
	result := zhenai.ParseProfile(content, "小海豹")

	fmt.Print(result)

	//因为测试可能进行在内部环境，所以要将网上的数据进行拷贝存放在本地 ，而且已经知道返回的地点数量为固定24个
	//contents,err := fetcher.Fetch(person_profile_url)
	//for err!=nil {
	//	//panic(err)
	//	contents ,err= fetcher.Fetch(person_profile_url)
	//	time.Sleep(time.Second)
	//}
	////fmt.Printf("%s\n",contents)
	//permissions := 0664
	//err2 := ioutil.WriteFile(person_profile_filename, contents, os.FileMode(permissions))
	//if err2 != nil {
	//	// handle error
	//}
}
