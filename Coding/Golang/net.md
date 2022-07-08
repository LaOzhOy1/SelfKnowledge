# Epoll 机制

https://zhuanlan.zhihu.com/p/367591714



## 内核数据接收过程

计算机系统的外部数据从网卡进入，到最后应用程序接收，这个是一个链路相对较长的过程。大概流程如下：

- 网卡接收数据，通过DMA控制器将数据拷贝到内核空间，同时往CPU发起硬中断。
- CPU接收到硬中断后简单处理下，然后交给ksoftiqd进程，这个过程称为软中断。
- ksoftirqd进程用于和IO复用，比如select、epoll等交互，它将触发poll()。
- 拷贝完成后通过IO模型(epoll为例)，会将数据添加到就绪列表的某个fd中。
- 通过回调(epoll)为例）唤醒相应进程，访问数据。

关于ksoftirqd进程，它用于处理软中断，调用注册的poll方法开始收包。

```shell
➜  ps -aux | grep ksoft
root          12  0.0  0.0      0     0 ?        S    9月6   0:01 [ksoftirqd/0]
root          20  0.0  0.0      0     0 ?        S    9月6   0:00 [ksoftirqd/1]
root          26  0.0  0.0      0     0 ?        S    9月6   0:00 [ksoftirqd/2]
root          32  0.0  0.0      0     0 ?        S    9月6   0:02 [ksoftirqd/3]
```

## epoll的整体软硬交互过程

内核与外部设备以及应用程序之间的交互逻辑大致如下：

- 操作系统启动时会创建调用`epoll_create`，创建epoll管理对象。
- 应用程序在启动时会通过http/rpc等方式绑定和监听七层数据。
- 往内核注册监听事件，然后调用`epoll_wait`阻塞等待数据就绪。
- 数据就绪后会将数据从内核拷贝到用户空间并唤醒执行。

## 初始化epfd

```go
// runtime/netpoll.go

var netpollInited   uint32

func netpollGenericInit() {
	if atomic.Load(&netpollInited) == 0 {
		lockInit(&netpollInitLock, lockRankNetpollInit)
		lock(&netpollInitLock)
		if netpollInited == 0 {
			netpollinit()
			atomic.Store(&netpollInited, 1)
		}
		unlock(&netpollInitLock)
	}
}
```

首先用`atomic`原语来实现单例模式，防止被多次初始化。具体实例化在`netpollinit()`

```go
// runtime/netpoll.go

func netpollinit() {
    // 尝试epollcreate1
	epfd = epollcreate1(_EPOLL_CLOEXEC)
	if epfd < 0 {
        // 尝试epollcreate
		epfd = epollcreate(1024)
		if epfd < 0 {
			println("runtime: epollcreate failed with", -epfd)
			throw("runtime: netpollinit failed")
		}
		closeonexec(epfd)
	}
	r, w, errno := nonblockingPipe()
	if errno != 0 {
		println("runtime: pipe failed with", -errno)
		throw("runtime: pipe failed")
	}
	ev := epollevent{
		events: _EPOLLIN,
	}
	*(**uintptr)(unsafe.Pointer(&ev.data)) = &netpollBreakRd
	errno = epollctl(epfd, _EPOLL_CTL_ADD, r, &ev)
	if errno != 0 {
		println("runtime: epollctl failed with", -errno)
		throw("runtime: epollctl failed")
	}
	netpollBreakRd = uintptr(r)
	netpollBreakWr = uintptr(w)
}
```

- 关于epollcreate

  由于Cow机制，在子进程执行之前，它拥有和其父进程一样的数据空间、堆栈和fd列表，当子进程执行新代码时，父进程的fd也会随之被覆盖，此时会导致无法维护父进程的fd，所以需要子进程先关掉无用的fd，再执行相应代码。

> func epollcreate(size int32) int32
>
> func epollcreate1(flags int32) int32
>
> 当flags=_EPOLL_CLOEXEC时，它会自动在netpoll关闭时调用closeonexec()

- 关于`nonblockingPipe`

  所谓IO多路复用，是指所有线程复用一个连接。`netpoll`则起到一个承上启下的作用。

  > r, w, errno := nonblockingPipe()
  >
  > netpollBreakRd = uintptr(r)
  >
  > netpollBreakWr = uintptr(w)
  >
  > r和w是最底层的读写fd，也是要进行复用的fd，之后在netpoll注册监听的所有fd都是复用这个`netpollBreakRd`和`netpollBreakWr`

- 关于`epollctl`

  在`netpoll`模型中，所有对`netpoll`对象的修改都是通过`epollctl`实现的。`netpoll`对象本身也是一个文件描述符。也就是代码中的`epfd`。

  > errno = epollctl(epfd, _EPOLL_CTL_ADD, r, &ev)
  >
  > 这里给底层的netpollBreakRd绑定了可读事件，即外部网络连接过来的数据都会通过netpollBreakRd写入到内核中，然后再分发到不同的fd上。

## 给用户fd绑定事件

```go
// runtime/netpoll_epoll.go

func netpollopen(fd uintptr, pd *pollDesc) int32 {
	var ev epollevent
	ev.events = _EPOLLIN | _EPOLLOUT | _EPOLLRDHUP | _EPOLLET
	*(**pollDesc)(unsafe.Pointer(&ev.data)) = pd
	return -epollctl(epfd, _EPOLL_CTL_ADD, int32(fd), &ev)
}
```

这里比较简单，就是调用`epollctl`绑定可读、可写、挂起和边缘触发事件。这样当就绪时就能马上唤起对应的fd读写数据。

## 给用户fd解绑事件

```go
// runtime/netpoll_epoll.go

func netpollclose(fd uintptr) int32 {
	var ev epollevent
	return -epollctl(epfd, _EPOLL_CTL_DEL, int32(fd), &ev)
}
```

同样的，调用`epollctl`，指定操作为`_EPOLL_CTL_DEL`。

## 监听事件

```go
// runtime/netpoll_epoll.go

// `delay`表示阻塞多久，然后返回需要通知的协程列表`gList`
func netpoll(delay int64) gList {
    // 如果`epfd`关闭了，那就直接返回，因为没有fd需要监听。
    if epfd == -1 {
		return gList{}
	}
    ...
    	var events [128]epollevent
retry:
    n := epollwait(epfd, &events[0], int32(len(events)), waitms)
    if n < 0 {
        ...
        // 超时
        if waitms > 0 {
			return gList{}
	    }
        goto retry
    }
    var toRun gList
    for i := int32(0); i < n; i++ {
        ev := &events[i]
        // 没有绑定的监听事件
        if ev.events == 0 {
			continue
	    }
        // 底层可读事件:netpollBreakRd,接收数据，写入内核数据空间
        if *(**uintptr)(unsafe.Pointer(&ev.data)) == &netpollBreakRd {
            if ev.events != _EPOLLIN {
				println("runtime: netpoll: break fd ready for", ev.events)
				throw("runtime: netpoll: break fd ready for something unexpected")
		    }
            if delay != 0 {
				// 写入
				var tmp [16]byte                
				read(int32(netpollBreakRd), noescape(unsafe.Pointer(&tmp[0])), int32(len(tmp)))
				atomic.Store(&netpollWakeSig, 0)
			}
			continue
        }
        var mode int32
         // 只要注册了_EPOLLIN可读事件，就是读模式
		if ev.events&(_EPOLLIN|_EPOLLRDHUP|_EPOLLHUP|_EPOLLERR) != 0 {
			mode += 'r'
		}
         // 只要注册了_EPOLLOUT可写事件，就是写模式
		if ev.events&(_EPOLLOUT|_EPOLLHUP|_EPOLLERR) != 0 {
			mode += 'w'
		}
		if mode != 0 {
			pd := *(**pollDesc)(unsafe.Pointer(&ev.data))
			pd.everr = false
			if ev.events == _EPOLLERR {
				pd.everr = true
			}
             // 通知就绪的fd所对应的gorotine
			netpollready(&toRun, pd, mode)
		}
    }
}

func netpollready(toRun *gList, pd *pollDesc, mode int32) {
	var rg, wg *g
	if mode == 'r' || mode == 'r'+'w' {
		rg = netpollunblock(pd, 'r', true)
	}
	if mode == 'w' || mode == 'r'+'w' {
		wg = netpollunblock(pd, 'w', true)
	}
	if rg != nil {
        // 唤醒 可读 协程
		toRun.push(rg)
	}
	if wg != nil {
		// 唤醒 可写 协程
		toRun.push(wg)
	}
}

func netpollunblock(pd *pollDesc, mode int32, ioready bool) *g {
	gpp := &pd.rg
	if mode == 'w' {
		gpp = &pd.wg
	}

	for {
		old := *gpp
		if old == pdReady {
			return nil
		}
		if old == 0 && !ioready {
			// Only set pdReady for ioready. runtime_pollWait
			// will check for timeout/cancel before waiting.
			return nil
		}
		var new uintptr
		if ioready {
			new = pdReady
		}
		if atomic.Casuintptr(gpp, old, new) {
			if old == pdWait {
				old = 0
			}
			return (*g)(unsafe.Pointer(old))
		}
	}
}
```

监听部分是对通过`netpoll_ctx`注册的事件不断循环监听，当复用连接(`netpollBreakRd`和`netpollBreakWd`)读写就绪时，就唤醒相应事件绑定的协程。

主要流程：

1. `初始化` 即获取底层连接`netpollBreakRd`和`netpollBreakWr`。这两个是对外的读写连接流，它们本身也是一个socket对象。`epoll`对这两个fd进行复用。
2. `初始化` 初始化`eventpollfd`，通过它来管理其他socket的读写事件，`epfd`是`epoll`模型的核心结构。
3. `注册事件` 某个协程通过系统调用发起一个IO请求，内核将该`fd`注册到`epfd`上，调用`epollctl`绑定需要的事件，如读事件、写事件、挂起事件等。然后协程进入阻塞状态，等待被唤醒。
4. `监听事件` 内核会不断监听底层连接`netpollBreadRd`和`netpollBreakWr`。当底层连接可读时，会读`netpollBreadRd`，然后缓存到内核数据空间上。
5. `监听事件` 内核根据`epfd`注册的各种事件不断循环，当注册的事件出现时，如可读、可写等。就唤醒对应绑定事件的协程，即上述的`toRun`列表。
6. `协程就绪` 协程收到通知后，开始向`fd`读、写数据。完成之后关闭`fd`。内核会将其从`epfd`上删除绑定的事件。



