package speedTest

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func stringPlus(b *testing.B) {
	for i := 0; i < b.N; i++ {
		//StringPlus()
		//StringFmt()
		//StringJoin()
		//StringBuffer()
		StringBuilder()
	}
}

// 0.263s 它是非常灵活的一个结构体，不止可以拼接字符串，还是可以byte,rune等，并且实现了io.Writer接口
func StringBuffer() string {
	var b bytes.Buffer
	b.WriteString("昵称")
	b.WriteString(":")
	b.WriteString("飞雪无情")
	b.WriteString("\n")
	b.WriteString("博客")
	b.WriteString(":")
	b.WriteString("http://www.flysnow.org/")
	b.WriteString("\n")
	b.WriteString("微信公众号")
	b.WriteString(":")
	b.WriteString("flysnow_org")
	return b.String()
}

// 0.256s 和buffer 一样用于提升字符串拼接的性能
func StringBuilder() string {
	var b strings.Builder
	b.WriteString("昵称")
	b.WriteString(":")
	b.WriteString("飞雪无情")
	b.WriteString("\n")
	b.WriteString("博客")
	b.WriteString(":")
	b.WriteString("http://www.flysnow.org/")
	b.WriteString("\n")
	b.WriteString("微信公众号")
	b.WriteString(":")
	b.WriteString("flysnow_org")
	return b.String()
}

// string +  0.239s

func StringPlus() string {
	var s string
	s += "昵称" + ":" + "飞雪无情" + "\n"
	s += "博客" + ":" + "http://www.flysnow.org/" + "\n"
	s += "微信公众号" + ":" + "flysnow_org"
	return s
}

// 0.304s
func StringFmt() string {
	return fmt.Sprint("昵称", ":", "飞雪无情", "\n", "博客", ":", "http://www.flysnow.org/", "\n", "微信公众号", ":", "flysnow_org")
}

// 0.261s
func StringJoin() string {
	s := []string{"昵称", ":", "飞雪无情", "\n", "博客", ":", "http://www.flysnow.org/", "\n", "微信公众号", ":", "flysnow_org"}
	return strings.Join(s, "")
}
