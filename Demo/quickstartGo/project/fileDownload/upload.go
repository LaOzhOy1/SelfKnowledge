package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/*
https://www.cnblogs.com/Vikyanite/p/17141666.html
前端切割代码 分片上传
var reader = new FileReader();
reader.readAsArrayBuffer(file);
reader.addEventListener("load", function(e) {
    //每10M切割一段,这里只做一个切割演示，实际切割需要循环切割，
    var slice = e.target.result.slice(0, 10*1024*1024);
});
后端将切片进行合并
后端支持小文件断点续传
https://blog.csdn.net/a1309525802/article/details/131766466
https://blog.csdn.net/JineD/article/details/125872088
*/

/*
后端分片1
https://learnku.com/articles/64380

分布式
https://github.com/cvenwu/DistributedFileServer

range
https://jasonkayzk.github.io/2020/09/28/Go%E5%AE%9E%E7%8E%B0HTTP%E6%96%AD%E7%82%B9%E7%BB%AD%E4%BC%A0%E5%A4%9A%E7%BA%BF%E7%A8%8B%E4%B8%8B%E8%BD%BD/

https://learn.lianglianglee.com/%E4%B8%93%E6%A0%8F/%E9%80%8F%E8%A7%86HTTP%E5%8D%8F%E8%AE%AE/16%20%20%E6%8A%8A%E5%A4%A7%E8%B1%A1%E8%A3%85%E8%BF%9B%E5%86%B0%E7%AE%B1%EF%BC%9AHTTP%E4%BC%A0%E8%BE%93%E5%A4%A7%E6%96%87%E4%BB%B6%E7%9A%84%E6%96%B9%E6%B3%95.md
*/

/*
有了分块传输编码，服务器就可以轻松地收发大文件了，但对于上 G 的超大文件，还有一些问题需要考虑。

比如，你在看当下正热播的某穿越剧，想跳过片头，直接看正片，或者有段剧情很无聊，想拖动进度条快进几分钟，这实际上是想获取一个大文件其中的片段数据，而分块传输并没有这个能力。

HTTP 协议为了满足这样的需求，提出了“范围请求”（range requests）的概念，允许客户端在请求头里使用专用字段来表示只获取文件的一部分，相当于是客户端的“化整为零”。

范围请求不是 Web 服务器必备的功能，可以实现也可以不实现，所以服务器必须在响应头里使用字段“Accept-Ranges: bytes”明确告知客户端：“我是支持范围请求的”。

如果不支持的话该怎么办呢？服务器可以发送“Accept-Ranges: none”，或者干脆不发送“Accept-Ranges”字段，这样客户端就认为服务器没有实现范围请求功能，只能老老实实地收发整块文件了。

请求头Range是 HTTP 范围请求的专用字段，格式是“bytes=x-y”，其中的 x 和 y 是以字节为单位的数据范围。

要注意 x、y 表示的是“偏移量”，范围必须从 0 计数，例如前 10 个字节表示为“0-9”，第二个 10 字节表示为“10-19”，而“0-10”实际上是前 11 个字节。

Range 的格式也很灵活，起点 x 和终点 y 可以省略，能够很方便地表示正数或者倒数的范围。假设文件是 100 个字节，那么：

“0-”表示从文档起点到文档终点，相当于“0-99”，即整个文件；
“10-”是从第 10 个字节开始到文档末尾，相当于“10-99”；
“-1”是文档的最后一个字节，相当于“99-99”；
“-10”是从文档末尾倒数 10 个字节，相当于“90-99”。
服务器收到 Range 字段后，需要做四件事。

第一，它必须检查范围是否合法，比如文件只有 100 个字节，但请求“200-300”，这就是范围越界了。服务器就会返回状态码416，意思是“你的范围请求有误，我无法处理，请再检查一下”。

第二，如果范围正确，服务器就可以根据 Range 头计算偏移量，读取文件的片段了，返回状态码“206 Partial Content”，和 200 的意思差不多，但表示 body 只是原数据的一部分。

第三，服务器要添加一个响应头字段Content-Range，告诉片段的实际偏移量和资源的总大小，格式是“bytes x-y/length”，与 Range 头区别在没有“=”，范围后多了总长度。例如，对于“0-10”的范围请求，值就是“bytes 0-10/100”。

最后剩下的就是发送数据了，直接把片段用 TCP 发给客户端，一个范围请求就算是处理完了。

断点续传原理
HTTP1.1 协议（RFC2616）开始支持获取文件的部分内容，这为并行下载以及断点续传提供了技术支持：通过在 Header里两个参数Range和Content-Range实现：

客户端发请求时对应的是 Range，服务器端响应时对应的是 Content-Range。

https://blog.csdn.net/luckytanggu/article/details/79830493

writer.Header().Add("Accept-ranges", "bytes")   //告诉客户端支持Range，并且是bytes级的断点续传
writer.Header().Add("Content-Length", 文件大小)     //文件的总体大小
writer.Header().Add("Content-Disposition", "attachment; filename=文件名称“)
https://rehtt.com/index.php/archives/220/

https://github.com/zhuchangwu/large-file-upload


https://github.com/zhuchangwu/large-file-upload/blob/master/beego-file-uploader/controllers/file_upload_controller.go
*/

const MaxAllowMb = 1
const MB = 1024 * 1024
const AllowFileSuffix = "."
const uploadPath = "./"

/*
先说一下思路：想实现断点续传，主要就是记住上一次已经传递了多少数据，那我们可以创建一个临时文件，
记录已经传递的数据量，当恢复传递的时候，先从临时文件中读取上次已经传递的数据量，然后通过Seek()方法，设置到该读和该写的位置，再继续传递数据。
*/
func main() {
	r := gin.Default()
	r.POST("/upload", func(context *gin.Context) {
		src, fileHeader, err := context.Request.FormFile("file")
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer func() {
			err = src.Close()
			if err != nil {
				return
			}
		}()

		filename := fileHeader.Filename
		//if !strings.HasSuffix(filename, AllowFileSuffix) {
		//	context.JSON(http.StatusBadRequest, gin.H{"error": "only allow xlsx file"})
		//	return
		//}

		dst, err := os.OpenFile(uploadPath+filename, os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "open file failed"})
			return
		}
		defer dst.Close()

		size := fileHeader.Size
		if size/MB <= MaxAllowMb {
			smallFileUpload(src, dst)
			return
		}
		bigFileUpload(fileHeader, src, dst)
		return
	})
	server := &http.Server{
		Addr:           ":9999",
		Handler:        r,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	//启动HTTP服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Log(nil, slog.LevelError, "run gin web failed")
		}
	}()

	gracefulExit(server)
}

func gracefulExit(server *http.Server) {
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Log(nil, slog.LevelInfo, "Shutdown Server ...")

	//创建超时上下文，Shutdown可以让未处理的连接在这个时间内关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//停止HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		slog.Log(nil, slog.LevelInfo, "Shutdown Server ...", err)
	}

	slog.Log(nil, slog.LevelInfo, "Server exiting")
}

func smallFileUpload(src multipart.File, dst *os.File) {
	// io CopyBuffer 默认以32KB进行读写
	buff := make([]byte, 64*1024)
	io.CopyBuffer(dst, src, buff)
}

func bigFileUpload(srcHdr *multipart.FileHeader, src multipart.File, dst *os.File) {
	//  获取本地文件大小
	info, err := dst.Stat()
	if err != nil {
		slog.Log(nil, slog.LevelError, "open dst file stat failed")
		return
	}

	// 获取目标文件大小
	size := srcHdr.Size
	// TODO 通过文件大小判断是否已经上传成功

	// 将写指针指向临时文件末尾
	_, err = src.Seek(info.Size(), io.SeekStart)
	if err != nil {
		slog.Log(nil, slog.LevelError, "seek dst file stat failed")
		return
	}

	fmt.Println(size)
	fmt.Println(info.Size())

	// 64KB
	buff := make([]byte, 64*1024)
	var bar Bar
	bar.NewOption(info.Size(), size)
	ch := make(chan struct{})
	go func() {
		for true {
			select {
			case <-ch:
				slog.Log(nil, slog.LevelInfo, "receive file write finish")
				return
			default:
				stat, _ := dst.Stat()
				bar.Play(stat.Size())
			}
		}
	}()

	io.CopyBuffer(dst, src, buff)
	ch <- struct{}{}
	// io copy https://blog.csdn.net/chenjiazhanxiao/article/details/131630626
}

type Bar struct {
	percent int64  //百分比
	cur     int64  //当前进度位置
	total   int64  //总进度
	rate    string //进度条
	graph   string //显示符号
}

func (bar *Bar) NewOption(start, total int64) {
	bar.cur = start
	bar.total = total
	if bar.graph == "" {
		bar.graph = "█"
	}
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += bar.graph //初始化进度条位置
	}
}

func (bar *Bar) getPercent() int64 {
	return int64(float32(bar.cur) / float32(bar.total) * 100)
}

func (bar *Bar) NewOptionWithGraph(start, total int64, graph string) {
	bar.graph = graph
	bar.NewOption(start, total)
}

func (bar *Bar) Play(cur int64) {
	bar.cur = cur
	last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != last && bar.percent%2 == 0 {
		bar.rate += bar.graph
	}
	fmt.Printf("\r[%-50s]%3d%%  %8d/%d", bar.rate, bar.percent, bar.cur, bar.total)
}
