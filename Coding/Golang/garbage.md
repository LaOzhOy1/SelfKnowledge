参考资料：
 https://golang.design/under-the-hood/zh-cn/part2runtime/ch08gc/barrier/

https://blog.csdn.net/CSDN_bang/article/details/107572440

# 垃圾回收

概述： Golang 采用基于三色原理的标记清除算法处理内存里面的对象，并使用混合写屏障保护在并发标记发生修改的标记对象的引用关系；在Go中栈上内存仍由编译器负责管理回收，而堆上的内存由编译器和垃圾收集器负责管理回收，给编程人员带来了极大的便利性。

### 触发时机

触发GC有俩个条件，一是堆内存的分配达到控制器计算的触发堆大小，初始大小环境变量GOGC，之后堆内存达到上一次垃圾收集的 2 倍时才会触发GC。二是如果一定时间内没有触发，就会触发新的循环，该触发条件由`runtime.forcegcperiod`变量控制，默认为 2 分钟



### 执行阶段

垃圾回收器从根路径出发，根据引用关系将白色对象标记为灰色对象，放入灰色对象集合中，之后与用户线程并发循环遍历灰色对象集合，将灰色对象标记为黑色对象，白色对象标记为灰色对象，直到灰色集合对象为空，回收所有白色对象；但是标记阶段并不是阻塞进行的，而是和用户线程一起执行，会导致标记遗漏的情况。



### 垃圾回收的阈值

- 垃圾回收期间会保证回收占用的CPU不超过设置好的阈值
  

### 强三色不变性

- 保证条件: 不允许黑色对象直接链接白色对象  

### 弱三色不变性

- 保证条件：允许黑色对象直接链接白色对象，但必须保证有一条未扫描过的灰色路径链接到该白色对象

 

为了实现这俩种不变式的设计思想，从而引出了屏障机制，即在程序的执行过程中加一个判断机制，满足判断机制则执行回调函数。



## 写屏障

概述： 保障代码在内存操作的顺序，由golang编译器在编译期注入代码，保证代码不会在编译期被打乱或者在运行期间被CPU重排序



### 触发时机

1. 写屏障的代码在编译期间生成好，之后不会再变化；
2. 堆上对象赋值才会生成写屏障；
3. 哪些对象分配在栈上，哪些分配在堆上？也是编译期间由编译器决定，这个过程叫做“逃逸分析”；



### 插入屏障

插入写屏障将赋值器对已存活的对象的插入行为通知给回收器，并将新插入的对象置灰

优点：

- 在GC开始时无需STW，可以进行并发标记垃圾对象。


缺点：

- 确保不会有黑-白的情况产生，但会为每次赋值操作引入写屏障
- 无法保证黑-灰的引用被删除的情况，此时灰色的对象只能在下一次垃圾回收扫描周期被删除
- 如果在栈上使用插入屏障，则会引入性能问题；如果不使用插入屏障的情况下，默认将栈上对象分配堆上的对象设置为灰，在标记结束阶段需要对栈进行STW重新扫描一遍灰色对象

### 删除屏障

删除写屏障将赋值器灰白路径进行删除时，将白色对象判定为灰色对象，从而保证其后如果有黑色直接链接该对象时，不会出现黑-白路径
优点：

- 完成GC时不需要再次进行STW扫描栈

缺点：

- 必须在开始GC时对所有的Goroutine 栈进行快照，从而使用删除屏障保证弱三色不变性，不适用于栈比较大的服务器上，适用于嵌入式或者物联网这些栈小的设备。
- 对于没有引用的灰色对象，需要等下一次扫描才能彻底删除（波面后退）
- 不需要在标记结束阶段重新扫描，



### 混合写屏障

Go 使用了 Dijkstra 与 Yuasa 屏障的结合， 即**混合写屏障（Hybrid write barrier）技术**。



#### 具体实现：

1. 编译时期加入写屏障
   ![image-20220705015910647](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/img/202207050159706.png)

2. 对赋值对象和操作对象都进行置灰操作(批量置灰)

	```shell
	TEXT runtime·gcWriteBarrier(SB),NOSPLIT,$120
	 
	  get_tls(R13)
	  MOVQ  g(R13), R13
	  MOVQ  g_m(R13), R13
	  MOVQ  m_p(R13), R13
	  MOVQ  (p_wbBuf+wbBuf_next)(R13), R14
	 
	  LEAQ  16(R14), R14
	  MOVQ  R14, (p_wbBuf+wbBuf_next)(R13)
	    // 检查 buffer 队列是否满？
	  CMPQ  R14, (p_wbBuf+wbBuf_end)(R13)
	 
	    // 赋值的前后两个值都会被入队
	 
	  // 把 value 存到指定 buffer 位置
	  MOVQ  AX, -16(R14)  // Record value
	    // 把 *slot 存到指定 buffer 位置
	  MOVQ  (DI), R13
	  MOVQ  R13, -8(R14)
	 
	    // 如果 wbBuffer 队列满了，那么就下刷处理，比如置灰，置黑等操作
	  JEQ  flush
	ret:
	    // 赋值：*slot = val
	  MOVQ  104(SP), R14
	  MOVQ  112(SP), R13
	  MOVQ  AX, (DI)
	  RET
	flush:
	    。。。
	 
	  //  队列满了，统一处理，这个其实是一个批量优化手段
	  CALL  runtime·wbBufFlush(SB)
	    。。。
	 
	  JMP  ret
	```

3. 代码运行期间，各个goroutine栈中分配到堆上的对象通过写屏障被置灰，当GC发生时，他会并发的扫描各个goroutine的栈，期间会暂停goroutine分配器或发起扫描抢占正在运行的Goroutine，直到最终扫描完栈（仅仅是栈）

   ```go
   // 切系统调度栈
   systemstack(func() {
     userG := getg().m.curg
       // 如果是在自己的 goroutine 运行的时候去协助处理 gc 任务，恰好处理到自己的时候，需要做些处理；
     selfScan := gp == userG && readgstatus(userG) == _Grunning
     if selfScan {
       casgstatus(userG, _Grunning, _Gwaiting)
       userG.waitreason = waitReasonGarbageCollectionScan
     }
    
       // 扫描 goroutine 栈
     scang(gp, gcw)
    
     if selfScan {
       casgstatus(userG, _Gwaiting, _Grunning)
     }
   })
   
   
   func scang(gp *g, gcw *gcWork) {
       // 栈扫描是否完成的标识
       gp.gcscandone = false
    
       // 轮询的时长间隔
       const yieldDelay = 10 * 1000
       var nextYield int64
    
   // 循环
   loop:
       for i := 0; !gp.gcscandone; i++ {
           // 读取 goroutine 的状态标识；
           switch s := readgstatus(gp); s {
           default:
               dumpgstatus(gp)
               throw("stopg: invalid status")
    
           // 如果是已经释放的 goroutine，那么跳出；
           case _Gdead:
               // No stack.
               gp.gcscandone = true
               break loop
    
           // 拷贝栈的过程，等一下，稍后需要重试；
           case _Gcopystack:
           // Stack being switched. Go around again.
    
           // 如果该是 goroutine 是已经挂起的状态（非运行状态）
           case _Grunnable, _Gsyscall, _Gwaiting:
               // 重要：设置扫描标识（GScan）这个标识会阻塞该 goroutine 的运行，直到栈扫描完成；
               if castogscanstatus(gp, s, s|_Gscan) {
                   if !gp.gcscandone {
                   // 调用 scanstack 扫描栈
                       scanstack(gp, gcw)
                       gp.gcscandone = true
                   }
                   // 重启 goroutine ，这个 goroutine 又可以继续跑业务代码了；
                   restartg(gp)
                   break loop
               }
    
           // 如果已经是扫描状态了，那么说明别的地方已经在扫描这个g栈了，等别人完成就好了；
           case _Gscanwaiting:
           // newstack is doing a scan for us right now. Wait.
    
           // 如果这个 goroutine 是一个 runing 状态，那么需要抢占调度，然后让它自己去扫描 g 栈，现场就等他自己扫描完之后就好了；
           // 这里只需要设置抢占标识和扫描标识就可以了，真正的扫描现场在这个 gp 自己运行现场；
           // stackguard0 是一个抢占信号（go 1.13）
           case _Grunning:
               if gp.preemptscan && gp.preempt && gp.stackguard0 == stackPreempt {
                   break
               }
    
               // 打上 Scan 标识，通知 gp 抢占调度，让它自己扫描栈；
               if castogscanstatus(gp, _Grunning, _Gscanrunning) {
                   if !gp.gcscandone {
                   // 打上抢占标识
                       gp.preemptscan = true
                       gp.preempt = true
                       // 设置魔数标识
                       gp.stackguard0 = stackPreempt
                   }
                   // 设置完抢占标识，就可以把 goroutine 的 Scan 去掉了，下面就是循环等待它自己处理完了；
                   casfrom_Gscanstatus(gp, _Gscanrunning, _Grunning)
               }
           }
    
           if i == 0 {
               nextYield = nanotime() + yieldDelay
           }
           if nanotime() < nextYield {
               procyield(10)
           } else {
               osyield()
               nextYield = nanotime() + yieldDelay/2
           }
       }
    
       gp.preemptscan = false // cancel scan request if no longer needed
   }	
   ```

4. 当栈扫描完后，垃圾回收遵从插入屏障和删除屏障进行并发标记，保证对象在新建引用和删除引用的三色不变性



## 总结

插入屏障和删除屏障各有优劣， Golang1.18采取混合屏障机制解决了以下问题

- 插入屏障对于栈上的分配默认置灰问题导致在GC结束时需要STW重新扫描栈上的灰色对象

- 删除屏障需要在一开始STW将栈上的对象进行置灰，但是这样不利于栈比较大的服务器程序性能

  

### 备注

 golang 内部对象并没有保存颜色的属性，三色只是对他们的状态的描述，是通过一个队列 + 掩码位图 来实现的：

- 白色对象：对象所在 span 的 gcmarkBits 中对应的 bit 为 0，不在队列；
- 灰色对象：对象所在 span 的 gcmarkBits 中对应的 bit 为 1，且对象在扫描队列中；
- 黑色对象：对象所在 span 的 gcmarkBits 中对应的 bit 为 1，且对象已经从扫描队列中处理并摘除掉；
- 
