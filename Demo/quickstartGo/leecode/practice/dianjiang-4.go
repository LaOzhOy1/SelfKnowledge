package main

import (
	"fmt"
	"strconv"
	"strings"
)

/*
其中起始ip和结束ip定义了一个ip段，这个ip段中的ip都属于后面的国家名

IP库文件示例如下：
0.0.0.0,1.5.7.8,CN
1.5.7.9,2.255.1.39,US
2.255.1.40,5.2.255.255,CN
*/
func main() {
	fmt.Printf("%b", ipToInteger("1.5.7.8"))
}

// 为了方便处理，请实现一个函数，
// 将ip转化为整数(已知ip格式为.分隔的4个整数，每个整数的取值范围是0-255)，
// 并且需要保持ip之间的大小关系
func ipToInteger(ip string) int {
	arr := strings.Split(ip, `.`)
	step := 0
	result := 0
	for i := len(arr) - 1; i >= 0; i-- {
		number, _ := strconv.Atoi(arr[i])
		if number > 255 {
			panic("数据不合法")
		}
		result += number << step
		step += 8
	}
	return result
}
