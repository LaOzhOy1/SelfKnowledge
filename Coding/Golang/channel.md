# Channel 源码解读

## 1. 找到Channel 源码

-  通过 go tool compile -N -l -S  xxxx.go 命令调试任意含有channel 关键字的文件， 可以看到channel的相关代码。
  ![image-20220623131510022](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206231315058.png)

- 在go 源码中搜索该代码，位于 runtime.chan 文件中.



## 代码分析


### 数据结构明细
```Golang
const (
	maxAlign  = 8
	hchanSize = unsafe.Sizeof(hchan{}) + uintptr(-int(unsafe.Sizeof(hchan{}))&(maxAlign-1))
	debugChan = false
)

type hchan struct {
	qcount   uint           // 当前循环队列的大小
	dataqsiz uint           // 数据循环队列容量大小，一旦初始化后就不会更改，大小等于取决于用户的传值
	buf      unsafe.Pointer // points to an array of dataqsiz elements
	elemsize uint16
	closed   uint32
	elemtype *_type // element type
	sendx    uint   // send index
	recvx    uint   // receive index
	recvq    waitq  // 正在等待中的收取消息协程队列， FIFO
	sendq    waitq  // 正在等待中的发送消息协程队列， FIFO

	// lock protects all fields in hchan, as well as several
	// fields in sudogs blocked on this channel.
	//
	// Do not change another G's status while holding this lock
	// (in particular, do not ready a G), as this can deadlock
	// with stack shrinking.
	lock mutex
}

type waitq struct {
	first *sudog
	last  *sudog
}
```


### 初始化函数
```golang
var c *hchan
	switch {
	case mem == 0:
		c = (*hchan)(mallocgc(hchanSize, nil, true))
		c.buf = c.raceaddr()
    // var ch = make(chan int) || 	var ch = make(chan int, 0)
	case elem.ptrdata == 0:
		c = (*hchan)(mallocgc(hchanSize+mem, nil, true))
		c.buf = add(unsafe.Pointer(c), hchanSize)
	default:
		c = new(hchan)
		c.buf = mallocgc(mem, elem, true)
	}

	c.elemsize = uint16(elem.size)
	c.elemtype = elem
	c.dataqsiz = uint(size)
	// 管理整个Channel的锁
	lockInit(&c.lock, lockRankHchan)
```





### 通过队列发送消息

1. 语法糖接入口：
   ```golang
   // entry point for c <- x from compiled code
   //
   //go:nosplit
   func chansend1(c *hchan, elem unsafe.Pointer) {
   	chansend(c, elem, true, getcallerpc())
   }
   
   // compiler implements
   //
   //	select {
   //	case c <- v:
   //		... foo
   //	default:
   //		... bar
   //	}
   //
   // as
   //
   //	if selectnbsend(c, v) {
   //		... foo
   //	} else {
   //		... bar
   //	}
   func selectnbsend(c *hchan, elem unsafe.Pointer) (selected bool) {
   	return chansend(c, elem, false, getcallerpc())
   }
   ```

   

2. 发送逻辑实现：

- 阻塞实现：

  - 当发送不成功时都会抛出异常

  ```mermaid
  graph TD;
  
  SendChannel[SendChannel] --> IsInit{初始Channel};
  
  IsInit --> Lock{加锁}
  IsInit -->|未初始化| HandleNotInitByBlock{挂起协程}
  HandleNotInitByBlock -->|唤醒|Panic
  
  Lock --> closeChannelCheck{检查Channel是否已经关闭}
  closeChannelCheck --> |close|Panic
  closeChannelCheck --> CheckRecvQ{检查是否有挂起的Recv协程}
   
  CheckRecvQ -->|Exist|SendDirectly{SendDirectly给待接收的协程}
  
  
  SendDirectly --> UnLock{释放锁}
  
  UnLock --> True[发送成功]
  
  CheckRecvQ --> CheckQueueCount{检查循环队列中的数量}
  
  
  CheckQueueCount --> |队满|Unlock2[释放锁]
  Unlock2 -->Parking[挂起协程SendQ]  
  
  
  Parking -->|被Close唤醒|Panic[异常Panic]
  Parking -->|被Recv唤醒|True[发送成功]
  
  CheckQueueCount --> AddData{往循环队列里添加数据}
  AddData --> Unlock3[释放锁]
  
  Unlock3 --> True
  
  
  
  
  ```

  

  

- 非阻塞实现（不需要挂起当前协程的操作）：

  ```mermaid
  graph TD;
  
  SendChannel[SendChannel] --> IsInit{初始Channel};
  
  IsInit --> FullChannelCheck[检查Channel是否队满]
  IsInit -->|未初始化| False[发送失败]
  FullChannelCheck --> Lock[加锁]
  FullChannelCheck --> |Channel已队满|False
  Lock -->CloseChannelCheck{通道是否已经关闭}
  
  CloseChannelCheck -->|Channel已经关闭|UnlockByPanic[释放锁]
  UnlockByPanic -->Panic
  
  CloseChannelCheck--> CheckRecvQ{检查是否有挂起的Recv协程}
   
  CheckRecvQ -->|Exist|SendDirectly{SendDirectly给待接收的协程}
  
  SendDirectly --> UnLock[释放锁]
  
  UnLock --> True[发送成功]
  
  CheckRecvQ --> CheckQueueCount{检查循环队列中的数量}
  
  
  CheckQueueCount --> |队满|Unlock2[释放锁]
  Unlock2 -->False  
  
  CheckQueueCount --> AddData{往循环队列里添加数据}
  AddData --> Unlock3[释放锁]
  
  Unlock3 --> True
  
  ```

  

```golang
相关代码逻辑：

func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
	if c == nil {
		if !block {
			return false
		}
		gopark(nil, nil, waitReasonChanSendNilChan, traceEvGoStop, 2)
		throw("unreachable")
	}

    // 疑问1：为什么这里不加锁进行队满判断
	if !block && c.closed == 0 && full(c) {
		return false
	}

	var t0 int64
	if blockprofilerate > 0 {
		t0 = cputicks()
	}

	lock(&c.lock)

	if c.closed != 0 {
		unlock(&c.lock)
		panic(plainError("send on closed channel"))
	}

	if sg := c.recvq.dequeue(); sg != nil {
        // 疑问2：消息是如何进行Goroutine之间转发的
		send(c, sg, ep, func() { unlock(&c.lock) }, 3)
		return true
	}

	if c.qcount < c.dataqsiz {
		qp := chanbuf(c, c.sendx)
		if raceenabled {
			racenotify(c, c.sendx, nil)
		}
         // 疑问3：typedmemmove在这里的作用
		typedmemmove(c.elemtype, qp, ep)
		c.sendx++
		if c.sendx == c.dataqsiz {
			c.sendx = 0
		}
		c.qcount++
		unlock(&c.lock)
		return true
	}

	if !block {
		unlock(&c.lock)
		return false
	}


	gp := getg()
	mysg := acquireSudog()
	mysg.releasetime = 0
	if t0 != 0 {
		mysg.releasetime = -1
	}
	mysg.elem = ep
	mysg.waitlink = nil
	mysg.g = gp
	mysg.isSelect = false
	mysg.c = c
	gp.waiting = mysg
	gp.param = nil
	c.sendq.enqueue(mysg)
	// 保证当前的堆栈在parking 前不会发生变化
	atomic.Store8(&gp.parkingOnChan, 1)
    // 疑问四： gopark做了什么事情
	gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanSend, traceEvGoBlockSend, 2)
	// 疑问五： KeepAlive做了什么事情
	KeepAlive(ep)
	// 疑问六： 什么会改变gp的waiting
	if mysg != gp.waiting {
		throw("G waiting list is corrupted")
	}
	gp.waiting = nil
	gp.activeStackChans = false
    // 疑问七： 什么会改变mysg的success
	closed := !mysg.success
	gp.param = nil
	if mysg.releasetime > 0 {
		blockevent(mysg.releasetime-t0, 2)
	}
	mysg.c = nil
    // 疑问八 releaseSudog做了啥
	releaseSudog(mysg)
	if closed {
		if c.closed == 0 {
			throw("chansend: spurious wakeup")
		}
		panic(plainError("send on closed channel"))
	}
	return true
}
```





### Q&A

1. **为什么发送Channel 消息时，非阻塞状态下判断是否队满不用在同步资源区中**

- 判断队满需要Full 方法和isClose两变量判断。没有加锁时，读取这两个变量是不确定的
  - 当读取IsClose 为0 当实际已经被close掉： 不一致，但不影响
    - 当读取Full为True 实际被更改为False， 
      - 源代码的效果：直接返回False
      - 实际的效果：命中Panic
    - 当读取Full为True 和实际值一致： 一致
      - 源代码的效果：直接返回False
      - 实际的效果：Return False
  - 当读取IsClose 为0 与实际一致： 导致本应该发送的数据但是没有发送，但其实不影响，因为当从Full 改为Not Full的这种情况下，会有另外的协程被Recv唤醒进行Send操作，可能是官方所说的We rely on the side effects of lock release in chanrecv() and closechan() to update this thread's view of c.closed and full()这个意思把。
    - 当读取Full为True 实际被更改为False， 
      - 源代码的效果：直接返回False
      - 实际的效果：发送一条数据
    - 当读取Full为True 和实际值一致： 一致
      - 源代码的效果：直接返回False
      - 实际的效果：Return False
        

2. **消息是如何进行Goroutine之间转发的**

- 转发到环形队列

  - ```go
    typedmemmove(c.elemtype, qp, ep)
    c.sendx++
    if c.sendx == c.dataqsiz {
        c.sendx = 0
    }
    c.qcount++
    
    func typedmemmove(typ *_type, dst, src unsafe.Pointer) {
    	if dst == src {
    		return
    	}
        // 通过混合屏障保证在赋值过程中，不会出现回收不到的垃圾
    	if writeBarrier.needed && typ.ptrdata != 0 {
    		bulkBarrierPreWrite(uintptr(dst), uintptr(src), typ.ptrdata)
    	}
    	// 将类型大小为typ.size 的src 赋值到 dst
    	memmove(dst, src, typ.size)
    	if writeBarrier.cgo {
    		cgoCheckMemmove(typ, dst, src, 0, typ.size)
    	}
    }
    ```

    

- 转发到对应的Goroutine

  - ```go
    func send(){
    	if sg.elem != nil {
            // 发送消息
            sendDirect(c.elemtype, sg, ep)
            sg.elem = nil
        }
        gp := sg.g
        unlockf()
        gp.param = unsafe.Pointer(sg)
        sg.success = true
        if sg.releasetime != 0 {
            sg.releasetime = cputicks()
        }
        goready(gp, skip+1)
    }
    func sendDirect(t *_type, sg *sudog, src unsafe.Pointer) {
    	// src is on our stack, dst is a slot on another stack.
    	dst := sg.elem
    	typeBitsBulkBarrier(t, uintptr(dst), uintptr(src), t.size).
    	memmove(dst, src, t.size)
    }
    ```

3. **关键函数memmove的作用**

- 见疑问2

4. **Goparking做了什么**

- ```go
  // 获取当前执行这个G的M
  mp := acquirem()
  gp := mp.curg
  // 获取当前G的状态
  status := readgstatus(gp)
  // 仅对Running中的G进行park
  if status != _Grunning && status != _Gscanrunning {
  throw("gopark: bad g status")
  }
  mp.waitlock = lock
  mp.waitunlockf = unlockf
  gp.waitreason = reason
  mp.waittraceev = traceEv
  mp.waittraceskip = traceskip
  // 释放M与P之间的关系
  releasem(mp)
  // 将M的Stack 切回G0，并将G的现场保存在g->scheduled中，m可以继续和其他goroutine交互
  mcall(park_m)
  ```

5. KeepAlive 做了什么？

- 保证在读取数据时，不会析构

6. 什么方法会修改gp.waiting？
7. 什么方法会修改mysdg.success
   close 方法会将succes 改为false，消息发送成功后设置为true
8. releaseSudog做了啥 ?



### 通过队列接受消息

1. 语法糖接入口

   ```golang
   // entry points for <- c from compiled code
   //
   //go:nosplit
   func chanrecv1(c *hchan, elem unsafe.Pointer) {
   	chanrecv(c, elem, true)
   }
   // for C range <-ch
   //go:nosplit
   func chanrecv2(c *hchan, elem unsafe.Pointer) (received bool) {
   	_, received = chanrecv(c, elem, true)
   	return
   }
   
   // compiler implements
   //
   //	select {
   //	case v, ok = <-c:
   //		... foo
   //	default:
   //		... bar
   //	}
   //
   // as
   //
   //	if selected, ok = selectnbrecv(&v, c); selected {
   //		... foo
   //	} else {
   //		... bar
   //	}
   func selectnbrecv(elem unsafe.Pointer, c *hchan) (selected, received bool) {
   	return chanrecv(c, elem, false)
   }
   ```

2. 实现逻辑

``` go

	if c == nil {
		if !block {
			return
		}
		gopark(nil, nil, waitReasonChanReceiveNilChan, traceEvGoStop, 2)
		throw("unreachable")
	}

	if !block && empty(c) {
		
        
  
// 为什么这里返回false，false      
		if atomic.Load(&c.closed) == 0 {
			return
		}
        
       
		if empty(c) {
			if ep != nil {
				typedmemclr(c.elemtype, ep)
			}
			return true, false
		}
	}

	lock(&c.lock)

	if c.closed != 0 && c.qcount == 0 {
		unlock(&c.lock)
		if ep != nil {
			typedmemclr(c.elemtype, ep)
		}
		return true, false
	}

	if sg := c.sendq.dequeue(); sg != nil {
		recv(c, sg, ep, func() { unlock(&c.lock) }, 3)
		return true, true
	}

	if c.qcount > 0 {
		qp := chanbuf(c, c.recvx)
		if ep != nil {
			typedmemmove(c.elemtype, ep, qp)
		}
		typedmemclr(c.elemtype, qp)
		c.recvx++
		if c.recvx == c.dataqsiz {
			c.recvx = 0
		}
		c.qcount--
		unlock(&c.lock)
		return true, true
	}

	if !block {
		unlock(&c.lock)
		return false, false
	}

	// no sender available: block on this channel.
	gp := getg()
	mysg := acquireSudog()

	mysg.elem = ep
	mysg.waitlink = nil
	gp.waiting = mysg
	mysg.g = gp
	mysg.isSelect = false
	mysg.c = c
	gp.param = nil
	c.recvq.enqueue(mysg)

	atomic.Store8(&gp.parkingOnChan, 1)
	gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanReceive, traceEvGoBlockRecv, 2)

	if mysg != gp.waiting {
		throw("G waiting list is corrupted")
	}
	gp.waiting = nil
	gp.activeStackChans = false
	if mysg.releasetime > 0 {
		blockevent(mysg.releasetime-t0, 2)
	}
	success := mysg.success
	gp.param = nil
	mysg.c = nil
	releaseSudog(mysg)
	return true, success



===================================================


    if c.dataqsiz == 0 {
        if ep != nil {
            recvDirect(c.elemtype, sg, ep)
        }
    } else {
        qp := chanbuf(c, c.recvx)
        if ep != nil {
            typedmemmove(c.elemtype, ep, qp)
        }
        // copy data from sender to queue
        typedmemmove(c.elemtype, qp, sg.elem)
        c.recvx++
        if c.recvx == c.dataqsiz {
            c.recvx = 0
        }
        c.sendx = c.recvx // c.sendx = (c.sendx+1) % c.dataqsiz
    }
    sg.elem = nil
    gp := sg.g
    unlockf()
    gp.param = unsafe.Pointer(sg)
    sg.success = true
    if sg.releasetime != 0 {
        sg.releasetime = cputicks()
    }
    goready(gp, skip+1)
```

### Q&A

1. **为什么这里返回false，false**

- 第一个返回值控制Select语句，当通道关闭或有数据时，会触发Select语句抽取Channel中的数据。

  会一直触发获取false和默认值

  ![image-20220624062653656](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206240626763.png)

- 第二个返回值表示是否接受到值

  - 当挂起的协程被close 唤醒时，会收到false和默认值
    ![image-20220624062534360](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206240625471.png)
  - 





### 关闭Channel

1. 获取锁资源，锁住Channel所有数据结构

2. 更新channel close状态，并设置为消息推送失败

3. 释放锁

4. 将两个等待队列中的Goroutine 全面唤醒。
   ```golang
   func closechan(c *hchan) {
   	if c == nil {
   		panic(plainError("close of nil channel"))
   	}
   
   	lock(&c.lock)
   	if c.closed != 0 {
   		unlock(&c.lock)
   		panic(plainError("close of closed channel"))
   	}
   
   	if raceenabled {
   		callerpc := getcallerpc()
   		racewritepc(c.raceaddr(), callerpc, abi.FuncPCABIInternal(closechan))
   		racerelease(c.raceaddr())
   	}
   
   	c.closed = 1
   
   	var glist gList
   
   	for {
   		sg := c.recvq.dequeue()
   		if sg == nil {
   			break
   		}
   		if sg.elem != nil {
   			typedmemclr(c.elemtype, sg.elem)
   			sg.elem = nil
   		}
   		if sg.releasetime != 0 {
   			sg.releasetime = cputicks()
   		}
   		gp := sg.g
   		gp.param = unsafe.Pointer(sg)
   		sg.success = false
   		if raceenabled {
   			raceacquireg(gp, c.raceaddr())
   		}
   		glist.push(gp)
   	}
   
   	for {
   		sg := c.sendq.dequeue()
   		if sg == nil {
   			break
   		}
   		sg.elem = nil
   		if sg.releasetime != 0 {
   			sg.releasetime = cputicks()
   		}
   		gp := sg.g
   		gp.param = unsafe.Pointer(sg)
   		sg.success = false
   		if raceenabled {
   			raceacquireg(gp, c.raceaddr())
   		}
   		glist.push(gp)
   	}
   	unlock(&c.lock)
   
   	for !glist.empty() {
   		gp := glist.pop()
   		gp.schedlink = 0
   		goready(gp, 3)
   	}
   }
   ```





总结：

1. **向未初始化的Channel发送信息，分两种情况，**

-  当通过select 发送时，发送失败进入下一个select循环； 当通过阻塞方式发送时，则会进入协程等待。

2. **向已经关闭的Channel发送消息**

-  无论什么方式，立即Panic。	

3. **如何优雅的关闭Channel**【不报Panic】

   - 当有多个生产者时和一个消费者

     - 如果提前close 会导致消费者一直接受false信息

     通过设置WaitGroup 统一等待所有生产者发送完成，通知消费者可以关闭channel，消费者可以在通道消费完后进行close。
     ![image-20220624064200624](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/202206240642731.png)

   - 当有多个生产者时和多个消费者

​				同上
