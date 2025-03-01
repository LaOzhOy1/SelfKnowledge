package main

import (
	"fmt"
	"regexp"
)

const string_url = "aabbcc@qq.com"

func main() {
	re, _ := regexp.Compile(`aa`)
	match := re.FindString(string_url)
	fmt.Println(match)
}
