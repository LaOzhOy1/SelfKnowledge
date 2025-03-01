package pkg

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	cmdOnce      sync.Once      // 防止多次命令调用
	shutdownOnce sync.Once      // 防止多次调用，关停服务
	barrier      chan os.Signal // 等待命令到来
	serverStopFn []func()       // 优雅关机数组
	m            sync.Mutex     // 安全加入数组
)

func GracefulShutdown() {
	// 确保 并发调用后，只有一次shutdown请求成功。
	shutdownOnce.Do(
		func() {
			barrier = make(chan os.Signal, 1)
			signal.Notify(barrier, syscall.SIGTERM, syscall.SIGINT)
		})

	<-barrier
	//确保多次ctrl+D or C情况下，只进行一次 （去重）
	cmdOnce.Do(CleanUp)

	// 退出程序
	os.Exit(0)
}

func Add(f func()) {
	m.Lock()
	defer m.Unlock()
	serverStopFn = append(serverStopFn, f)
}

func CleanUp() {
	for _, i2 := range serverStopFn {
		i2()
	}
}
