# Bufio



## Scanner



```go
package main
import (
    "bufio"
    "fmt"
    "strings"
)
func main() {
    input := "foo   bar      baz"
    scanner := bufio.NewScanner(strings.NewReader(input))
    scanner.Split(bufio.ScanWords) // 定义空格的分割函数
    for scanner.Scan() {
        fmt.Println(scanner.Text())
    }
}

//ScanWords is a split function for a Scanner that returns each
//space-separated word of text, with surrounding spaces deleted. It will
//never return an empty string. The definition of space is set by
//unicode.IsSpace.
```



### 源码解析

```go

// 通过模板方法 split 对缓冲区里的数据进行抽取
// 支持自定义定义Split 函数或者使用原生的Split函数
// 当遇到IO.EOF或读取的文件过大时 Scanning 会返回失败（false），token返回任意一个很久之前的一个数据
// 对于大数据读取推荐使用reader
// token 是经过分割函数分割出来的某段字符串
advance, token, err := s.split(s.buf[s.start:s.end], s.err != nil)

// advance 通过校验 split 返回的advance数量来判定此次读取的正确性
func (s *Scanner) advance(n int) bool {
	if n < 0 {
		s.setErr(ErrNegativeAdvance)
		return false
	}
	if n > s.end-s.start {
		s.setErr(ErrAdvanceTooFar)
		return false
	}
	s.start += n
	return true
}
```

