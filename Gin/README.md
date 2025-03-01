## Gin 与 tomcat

### 路由设计

Tomcat
收到客户端请求后，容器根据请求 URL 的上下文名称匹配 Web 应用程序，然后根据去除上下文路径和路径参数的路径，按以下规则顺序匹配，并且只使用第一个匹配的 Servlet，后续不再尝试匹配：

精确匹配，查找一个与请求路径完全匹配的 Servlet（Map）
前缀路径匹配，递归的尝试匹配最长路径前缀的 Servlet，通过使用 "/" 作为路径分隔符，在路径树上一次一个目录的匹配，选择路径最长的
扩展名匹配，如果 URL 最后一部分包含扩展名，如 .jsp，则尝试匹配处理此扩展名请求的 Servlet
如果前三个规则没有匹配成功，那么容器要为请求提供一个默认 Servlet

Gin
由前缀树组成路由匹配规则，处理请求时仅需要最小的节点检索。

1.原理：通过公共前缀，来降低查询时间的开销，是一种空间换时间的做法。

2.比：hashmap快，是一种树形结构。常用于：统计、排序、保存大量的字符


### 减少GC
json 的反序列化在文本解析和网络通信过程中非常常见，当程序并发度非常高的情况下，短时间内需要创建大量的临时对象。而这些对象是都是分配在堆上的，会给 GC 造成很大压力，严重影响程序的性能

stu := studentPool.Get().(*Student)
json.Unmarshal(buf, stu)
studentPool.Put(stu)

$ go test -bench . -benchmem
goos: darwin
goarch: amd64
pkg: example/hpg-sync-pool
BenchmarkUnmarshal-8           1993   559768 ns/op   5096 B/op 7 allocs/op
BenchmarkUnmarshalWithPool-8   1976   550223 ns/op    234 B/op 6 allocs/op
PASS
ok      example/hpg-sync-pool   2.334s

var bufferPool = sync.Pool{
New: func() interface{} {
return &bytes.Buffer{}
},
}

var data = make([]byte, 10000)

func BenchmarkBufferWithPool(b *testing.B) {
for n := 0; n < b.N; n++ {
buf := bufferPool.Get().(*bytes.Buffer)
buf.Write(data)
buf.Reset()
bufferPool.Put(buf)
}
}

func BenchmarkBuffer(b *testing.B) {
for n := 0; n < b.N; n++ {
var buf bytes.Buffer
buf.Write(data)
}
}
