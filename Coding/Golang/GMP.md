参考文献：

https://zhuanlan.zhihu.com/p/64672362

https://blog.haohtml.com/archives/23437

https://learnku.com/articles/41728

# Golang的协程调度



## 协程调度相关的数据结构

### 全局M队列

allm *m 定义在runtime2.go:1099



### 全局P队列

allp *p 定义在runtime2.go:1111



### 全局G队列

  allgs   []*g 定义在proc.go:515



### 调度器(Schedt)



#### midle 调度器的空闲M列表



#### pidle 调度器的空闲P列表



#### runq 调度器的可运行G列表



#### gFree 调度器的自由G列表





#### maximum 调度期间允许的最大线程创建数

```go
// A gQueue is a dequeue of Gs linked through g.schedlink. A G can only
// be on one gQueue or gList at a time.
type gQueue struct {
	head guintptr
	tail guintptr
}

type schedt struct {
	// accessed atomically. keep at top to ensure alignment on 32-bit systems.
	goidgen   uint64
	lastpoll  uint64 // time of last network poll, 0 if currently polling
	pollUntil uint64 // time to which current poll is sleeping

	lock mutex

	// When increasing nmidle, nmidlelocked, nmsys, or nmfreed, be
	// sure to call checkdead().

	midle        muintptr // idle m's waiting for work
	nmidle       int32    // number of idle m's waiting for work
	nmidlelocked int32    // number of locked m's waiting for work
	mnext        int64    // number of m's that have been created and next M ID
	maxmcount    int32    // maximum number of m's allowed (or die)
	nmsys        int32    // number of system m's not counted for deadlock
	nmfreed      int64    // cumulative number of freed m's

	ngsys uint32 // number of system goroutines; updated atomically

	pidle      puintptr // idle p's
	npidle     uint32
	nmspinning uint32 // See "Worker thread parking/unparking" comment in proc.go.

	// Global runnable queue.
	runq     gQueue
	runqsize int32

	// disable controls selective disabling of the scheduler.
	//
	// Use schedEnableUser to control this.
	//
	// disable is protected by sched.lock.
	disable struct {
		// user disables scheduling of user goroutines.
		user     bool
		runnable gQueue // pending runnable Gs
		n        int32  // length of runnable
	}

	// Global cache of dead G's.
    // _Gdead 
    // gfput 将退出的goroutine存放到本地p队列中，如果存放超过64个g时，会将一半g移动schedgFree中，如果当前的goroutine栈大小不是默认的2K（不同系统不一样），则会释放stack，并将goroutine放入noStack中
	gFree struct {
		lock    mutex
		stack   gList // Gs with stacks
		noStack gList // Gs without stacks 
		n       int32
	}

	// Central cache of sudog structs.
	sudoglock  mutex
	sudogcache *sudog

	// Central pool of available defer structs.
	deferlock mutex
	deferpool *_defer

	// freem is the list of m's waiting to be freed when their
	// m.exited is set. Linked through m.freelink.
	freem *m

	gcwaiting  uint32 // gc is waiting to run
	stopwait   int32
	stopnote   note
	sysmonwait uint32
	sysmonnote note

	// safepointFn should be called on each P at the next GC
	// safepoint if p.runSafePointFn is set.
	safePointFn   func(*p)
	safePointWait int32
	safePointNote note

	profilehz int32 // cpu profiling rate

	procresizetime int64 // nanotime() of last change to gomaxprocs
	totaltime      int64 // ∫gomaxprocs dt up to procresizetime

	// sysmonlock protects sysmon's actions on the runtime.
	//
	// Acquire and hold this mutex to block sysmon from interacting
	// with the rest of the runtime.
	sysmonlock mutex

	// timeToRun is a distribution of scheduling latencies, defined
	// as the sum of time a G spends in the _Grunnable state before
	// it transitions to _Grunning.
	//
	// timeToRun is protected by sched.lock.
	timeToRun timeHistogram
}
```



### Gobuf

gobuf结构体用于保存goroutine的调度信息，主要包括CPU的几个寄存器的值

```go
type gobuf struct {
    // The offsets of sp, pc, and g are known to (hard-coded in) libmach.
    //
    // ctxt is unusual with respect to GC: it may be a
    // heap-allocated funcval, so GC needs to track it, but it
    // needs to be set and cleared from assembly, where it's
    // difficult to have write barriers. However, ctxt is really a
    // saved, live register, and we only ever exchange it between
    // the real register and the gobuf. Hence, we treat it as a
    // root during stack scanning, which means assembly that saves
    // and restores it doesn't need write barriers. It's still
    // typed as a pointer so that any other writes from Go get
    // write barriers.
    sp   uintptr  // 保存CPU的rsp寄存器的值
    pc   uintptr  // 保存CPU的rip寄存器的值
    g    guintptr // 记录当前这个gobuf对象属于哪个goroutine
    ctxt unsafe.Pointer
 
    // 保存系统调用的返回值，因为从系统调用返回之后如果p被其它工作线程抢占，
    // 则这个goroutine会被放入全局运行队列被其它工作线程调度，其它线程需要知道系统调用的返回值。
    ret  sys.Uintreg  
    lr   uintptr
 
    // 保存CPU的rip寄存器的值
    bp   uintptr // for GOEXPERIMENT=framepointer
}
```



### M

概述： m 是machine 的缩写，代表一个用户空间的内核线程，和内核线程一一对应，但是各有各的调度算法。



#### 相关的数据结构



##### G0

- 存储m上进行协程调度代码的协程



##### allm

- 全局的M列表的头结点指针



##### p
- 与M结合的当前p



##### spinning
- m是否正在寻觅P



##### nextp
- 当m唤醒后可能执行的p

```GO
type m struct {
	g0      *g     // 存储m上进行协程调度代码的协程 goroutine with scheduling stack
	morebuf gobuf  // gobuf arg to morestack
	divmod  uint32 // div/mod denominator for arm - known to liblink
	_       uint32 // align next field to 8 bytes

	// Fields not known to debuggers.
	procid        uint64            // for debuggers, but offset not hard-coded
	gsignal       *g                // signal-handling g
	goSigStack    gsignalStack      // Go-allocated signal handling stack
	sigmask       sigset            // storage for saved signal mask
	tls           [tlsSlots]uintptr // thread-local storage (for x86 extern register)
	mstartfn      func()
	curg          *g       // current running goroutine
	caughtsig     guintptr // goroutine running during fatal signal
	p             puintptr // attached p for executing go code (nil if not executing go code)
	nextp         puintptr
	oldp          puintptr // the p that was attached before executing a syscall
	id            int64
	mallocing     int32
	throwing      throwType
	preemptoff    string // if != "", keep curg running on this m
	locks         int32
	dying         int32
	profilehz     int32
	spinning      bool // m is out of work and is actively looking for work
	blocked       bool // m is blocked on a note
	newSigstack   bool // minit on C thread called sigaltstack
	printlock     int8
	incgo         bool   // m is executing a cgo call
	freeWait      uint32 // if == 0, safe to free g0 and delete m (atomic)
	fastrand      uint64
	needextram    bool
	traceback     uint8
	ncgocall      uint64      // number of cgo calls in total
	ncgo          int32       // number of cgo calls currently in progress
	cgoCallersUse uint32      // if non-zero, cgoCallers in use temporarily
	cgoCallers    *cgoCallers // cgo traceback if crashing in cgo call
	park          note
	alllink       *m // on allm
	schedlink     muintptr
	lockedg       guintptr
	createstack   [32]uintptr // stack that created this thread.
	lockedExt     uint32      // tracking for external LockOSThread
	lockedInt     uint32      // tracking for internal lockOSThread
	nextwaitm     muintptr    // next m waiting for lock
	waitunlockf   func(*g, unsafe.Pointer) bool
	waitlock      unsafe.Pointer
	waittraceev   byte
	waittraceskip int
	startingtrace bool
	syscalltick   uint32
	freelink      *m // on sched.freem

	// these are here because they are too large to be on the stack
	// of low-level NOSPLIT functions.
	libcall   libcall
	libcallpc uintptr // for cpu profiler
	libcallsp uintptr
	libcallg  guintptr
	syscall   libcall // stores syscall parameters on windows

	vdsoSP uintptr // SP for traceback while in VDSO call (0 if not in call)
	vdsoPC uintptr // PC for traceback while in VDSO call

	// preemptGen counts the number of completed preemption
	// signals. This is used to detect when a preemption is
	// requested, but fails. Accessed atomically.
	preemptGen uint32

	// Whether this is a pending preemption signal on this M.
	// Accessed atomically.
	signalPending uint32

	dlogPerM

	mOS

	// Up to 10 locks held by this m, maintained by the lock ranking code.
	locksHeldLen int
	locksHeld    [10]heldLockInfo
}

```











### P

概述：包含一个自由G列表，存放着一些已经运行完成的G



#### 相关数据结构

##### gFree: P的自由G列表

##### runq: P的可运行G列表

##### m: p所在的m

##### preempt 是否允许被抢占

##### timers 封装了多个lib库获取时间的功能对象

##### schedtick

```go
type p struct {
	id          int32
	status      uint32 // one of pidle/prunning/...
	link        puintptr
	schedtick   uint32     // incremented on every scheduler call
	syscalltick uint32     // incremented on every system call
	sysmontick  sysmontick // last tick observed by sysmon
	m           muintptr   // back-link to associated m (nil if idle)
	mcache      *mcache
	pcache      pageCache
	raceprocctx uintptr

	deferpool    []*_defer // pool of available defer structs (see panic.go)
	deferpoolbuf [32]*_defer

	// Cache of goroutine ids, amortizes accesses to runtime·sched.goidgen.
	goidcache    uint64
	goidcacheend uint64

	// Queue of runnable goroutines. Accessed without lock.
	runqhead uint32
	runqtail uint32
	runq     [256]guintptr
	// runnext, if non-nil, is a runnable G that was ready'd by
	// the current G and should be run next instead of what's in
	// runq if there's time remaining in the running G's time
	// slice. It will inherit the time left in the current time
	// slice. If a set of goroutines is locked in a
	// communicate-and-wait pattern, this schedules that set as a
	// unit and eliminates the (potentially large) scheduling
	// latency that otherwise arises from adding the ready'd
	// goroutines to the end of the run queue.
	//
	// Note that while other P's may atomically CAS this to zero,
	// only the owner P can CAS it to a valid G.
	runnext guintptr

	// Available G's (status == Gdead)
	gFree struct {
		gList
		n int32
	}

	sudogcache []*sudog
	sudogbuf   [128]*sudog

	// Cache of mspan objects from the heap.
	mspancache struct {
		// We need an explicit length here because this field is used
		// in allocation codepaths where write barriers are not allowed,
		// and eliminating the write barrier/keeping it eliminated from
		// slice updates is tricky, moreso than just managing the length
		// ourselves.
		len int
		buf [128]*mspan
	}

	tracebuf traceBufPtr

	// traceSweep indicates the sweep events should be traced.
	// This is used to defer the sweep start event until a span
	// has actually been swept.
	traceSweep bool
	// traceSwept and traceReclaimed track the number of bytes
	// swept and reclaimed by sweeping in the current sweep loop.
	traceSwept, traceReclaimed uintptr

	palloc persistentAlloc // per-P to avoid mutex

	_ uint32 // Alignment for atomic fields below

	// The when field of the first entry on the timer heap.
	// This is updated using atomic functions.
	// This is 0 if the timer heap is empty.
	timer0When uint64

	// The earliest known nextwhen field of a timer with
	// timerModifiedEarlier status. Because the timer may have been
	// modified again, there need not be any timer with this value.
	// This is updated using atomic functions.
	// This is 0 if there are no timerModifiedEarlier timers.
	timerModifiedEarliest uint64

	// Per-P GC state
	gcAssistTime         int64 // Nanoseconds in assistAlloc
	gcFractionalMarkTime int64 // Nanoseconds in fractional mark worker (atomic)

	// limiterEvent tracks events for the GC CPU limiter.
	limiterEvent limiterEvent

	// gcMarkWorkerMode is the mode for the next mark worker to run in.
	// That is, this is used to communicate with the worker goroutine
	// selected for immediate execution by
	// gcController.findRunnableGCWorker. When scheduling other goroutines,
	// this field must be set to gcMarkWorkerNotWorker.
	gcMarkWorkerMode gcMarkWorkerMode
	// gcMarkWorkerStartTime is the nanotime() at which the most recent
	// mark worker started.
	gcMarkWorkerStartTime int64

	// gcw is this P's GC work buffer cache. The work buffer is
	// filled by write barriers, drained by mutator assists, and
	// disposed on certain GC state transitions.
	gcw gcWork

	// wbBuf is this P's GC write barrier buffer.
	//
	// TODO: Consider caching this in the running G.
	wbBuf wbBuf

	runSafePointFn uint32 // if 1, run sched.safePointFn at next safe point

	// statsSeq is a counter indicating whether this P is currently
	// writing any stats. Its value is even when not, odd when it is.
	statsSeq uint32

	// Lock for timers. We normally access the timers while running
	// on this P, but the scheduler can also do it from a different P.
	timersLock mutex

	// Actions to take at some time. This is used to implement the
	// standard library's time package.
	// Must hold timersLock to access.
	timers []*timer

	// Number of timers in P's heap.
	// Modified using atomic instructions.
	numTimers uint32

	// Number of timerDeleted timers in P's heap.
	// Modified using atomic instructions.
	deletedTimers uint32

	// Race context used while executing timer functions.
	timerRaceCtx uintptr

	// maxStackScanDelta accumulates the amount of stack space held by
	// live goroutines (i.e. those eligible for stack scanning).
	// Flushed to gcController.maxStackScan once maxStackScanSlack
	// or -maxStackScanSlack is reached.
	maxStackScanDelta int64

	// gc-time statistics about current goroutines
	// Note that this differs from maxStackScan in that this
	// accumulates the actual stack observed to be used at GC time (hi - sp),
	// not an instantaneous measure of the total stack size that might need
	// to be scanned (hi - lo).
	scannedStackSize uint64 // stack size of goroutines scanned by this P
	scannedStacks    uint64 // number of goroutines scanned by this P

	// preempt is set to indicate that this P should be enter the
	// scheduler ASAP (regardless of what G is running on it).
	preempt bool

	// Padding is no longer needed. False sharing is now not a worry because p is large enough
	// that its size class is an integer multiple of the cache line size (for any of our architectures).
}

```





##### 状态

```GO
const (
	// P status

	// _Pidle means a P is not being used to run user code or the
	// scheduler. Typically, it's on the idle P list and available
	// to the scheduler, but it may just be transitioning between
	// other states.
	//
	// The P is owned by the idle list or by whatever is
	// transitioning its state. Its run queue is empty.
	_Pidle = iota

	// _Prunning means a P is owned by an M and is being used to
	// run user code or the scheduler. Only the M that owns this P
	// is allowed to change the P's status from _Prunning. The M
	// may transition the P to _Pidle (if it has no more work to
	// do), _Psyscall (when entering a syscall), or _Pgcstop (to
	// halt for the GC). The M may also hand ownership of the P
	// off directly to another M (e.g., to schedule a locked G).
	_Prunning

	// _Psyscall means a P is not running user code. It has
	// affinity to an M in a syscall but is not owned by it and
	// may be stolen by another M. This is similar to _Pidle but
	// uses lightweight transitions and maintains M affinity.
	//
	// Leaving _Psyscall must be done with a CAS, either to steal
	// or retake the P. Note that there's an ABA hazard: even if
	// an M successfully CASes its original P back to _Prunning
	// after a syscall, it must understand the P may have been
	// used by another M in the interim.
	_Psyscall

	// _Pgcstop means a P is halted for STW and owned by the M
	// that stopped the world. The M that stopped the world
	// continues to use its P, even in _Pgcstop. Transitioning
	// from _Prunning to _Pgcstop causes an M to release its P and
	// park.
	//
	// The P retains its run queue and startTheWorld will restart
	// the scheduler on Ps with non-empty run queues.
	_Pgcstop

	// _Pdead means a P is no longer used (GOMAXPROCS shrank). We
	// reuse Ps if GOMAXPROCS increases. A dead P is mostly
	// stripped of its resources, though a few things remain
	// (e.g., trace buffers).
	_Pdead
)
```



### G

概述：goroutine的缩写，一个G代表了go的代码片段。



#### 相关数据结构

##### Stack

- 记录栈中的起始和结束位置

##### StackGuard0 

- 记录正常 goroutine 的运行时栈指针位置

##### stackguard1
- 记录g0或gsignal中cstack的栈指针位置

##### preempt
- 抢占标志

##### sched

- 调度过程信息

##### lockedm

- 当前g关联的m数量，当cgo创建一个新的M或者schedule时，该变量发生改变

```GO
type stack struct {
	lo uintptr
	hi uintptr
}


type g struct {
	// Stack parameters.
	// stack describes the actual stack memory: [stack.lo, stack.hi).
	// stackguard0 is the stack pointer compared in the Go stack growth prologue.
	// It is stack.lo+StackGuard normally, but can be StackPreempt to trigger a preemption.
	// stackguard1 is the stack pointer compared in the C stack growth prologue.
	// It is stack.lo+StackGuard on g0 and gsignal stacks.
	// It is ~0 on other goroutine stacks, to trigger a call to morestackc (and crash).
	stack       stack   // offset known to runtime/cgo
	stackguard0 uintptr // offset known to liblink
	stackguard1 uintptr // offset known to liblink

	_panic    *_panic // innermost panic - offset known to liblink
	_defer    *_defer // innermost defer
	m         *m      // current m; offset known to arm liblink
	sched     gobuf
	syscallsp uintptr // if status==Gsyscall, syscallsp = sched.sp to use during gc
	syscallpc uintptr // if status==Gsyscall, syscallpc = sched.pc to use during gc
	stktopsp  uintptr // expected sp at top of stack, to check in traceback
	// param is a generic pointer parameter field used to pass
	// values in particular contexts where other storage for the
	// parameter would be difficult to find. It is currently used
	// in three ways:
	// 1. When a channel operation wakes up a blocked goroutine, it sets param to
	//    point to the sudog of the completed blocking operation.
	// 2. By gcAssistAlloc1 to signal back to its caller that the goroutine completed
	//    the GC cycle. It is unsafe to do so in any other way, because the goroutine's
	//    stack may have moved in the meantime.
	// 3. By debugCallWrap to pass parameters to a new goroutine because allocating a
	//    closure in the runtime is forbidden.
	param        unsafe.Pointer
	atomicstatus uint32
	stackLock    uint32 // sigprof/scang lock; TODO: fold in to atomicstatus
	goid         int64
	schedlink    guintptr
	waitsince    int64      // approx time when the g become blocked
	waitreason   waitReason // if status==Gwaiting

	preempt       bool // preemption signal, duplicates stackguard0 = stackpreempt
	preemptStop   bool // transition to _Gpreempted on preemption; otherwise, just deschedule
	preemptShrink bool // shrink stack at synchronous safe point

	// asyncSafePoint is set if g is stopped at an asynchronous
	// safe point. This means there are frames on the stack
	// without precise pointer information.
	asyncSafePoint bool

	paniconfault bool // panic (instead of crash) on unexpected fault address
	gcscandone   bool // g has scanned stack; protected by _Gscan bit in status
	throwsplit   bool // must not split stack
	// activeStackChans indicates that there are unlocked channels
	// pointing into this goroutine's stack. If true, stack
	// copying needs to acquire channel locks to protect these
	// areas of the stack.
	activeStackChans bool
	// parkingOnChan indicates that the goroutine is about to
	// park on a chansend or chanrecv. Used to signal an unsafe point
	// for stack shrinking. It's a boolean value, but is updated atomically.
	parkingOnChan uint8

	raceignore     int8     // ignore race detection events
	sysblocktraced bool     // StartTrace has emitted EvGoInSyscall about this goroutine
	tracking       bool     // whether we're tracking this G for sched latency statistics
	trackingSeq    uint8    // used to decide whether to track this G
	runnableStamp  int64    // timestamp of when the G last became runnable, only used when tracking
	runnableTime   int64    // the amount of time spent runnable, cleared when running, only used when tracking
	sysexitticks   int64    // cputicks when syscall has returned (for tracing)
	traceseq       uint64   // trace event sequencer
	tracelastp     puintptr // last P emitted an event for this goroutine
	lockedm        muintptr
	sig            uint32
	writebuf       []byte
	sigcode0       uintptr
	sigcode1       uintptr
	sigpc          uintptr
	gopc           uintptr         // pc of go statement that created this goroutine
	ancestors      *[]ancestorInfo // ancestor information goroutine(s) that created this goroutine (only used if debug.tracebackancestors)
	startpc        uintptr         // pc of goroutine function
	racectx        uintptr
	waiting        *sudog         // sudog structures this g is waiting on (that have a valid elem ptr); in lock order
	cgoCtxt        []uintptr      // cgo traceback context
	labels         unsafe.Pointer // profiler labels
	timer          *timer         // cached timer for time.Sleep
	selectDone     uint32         // are we participating in a select and did someone win the race?

	// goroutineProfiled indicates the status of this goroutine's stack for the
	// current in-progress goroutine profile
	goroutineProfiled goroutineProfileStateHolder

	// Per-G GC state

	// gcAssistBytes is this G's GC assist credit in terms of
	// bytes allocated. If this is positive, then the G has credit
	// to allocate gcAssistBytes bytes without assisting. If this
	// is negative, then the G must correct this by performing
	// scan work. We track this in bytes to make it fast to update
	// and check for debt in the malloc hot path. The assist ratio
	// determines how this corresponds to scan work debt.
	gcAssistBytes int64
}






// sudog represents a g in a wait list, such as for sending/receiving
// on a channel.
//
// sudog is necessary because the g ↔ synchronization object relation
// is many-to-many. A g can be on many wait lists, so there may be
// many sudogs for one g; and many gs may be waiting on the same
// synchronization object, so there may be many sudogs for one object.
//
// sudogs are allocated from a special pool. Use acquireSudog and
// releaseSudog to allocate and free them.
type sudog struct {
	// The following fields are protected by the hchan.lock of the
	// channel this sudog is blocking on. shrinkstack depends on
	// this for sudogs involved in channel ops.

	g *g

	next *sudog
	prev *sudog
	elem unsafe.Pointer // data element (may point to stack)

	// The following fields are never accessed concurrently.
	// For channels, waitlink is only accessed by g.
	// For semaphores, all fields (including the ones above)
	// are only accessed when holding a semaRoot lock.

	acquiretime int64
	releasetime int64
	ticket      uint32

	// isSelect indicates g is participating in a select, so
	// g.selectDone must be CAS'd to win the wake-up race.
	isSelect bool

	// success indicates whether communication over channel c
	// succeeded. It is true if the goroutine was awoken because a
	// value was delivered over channel c, and false if awoken
	// because c was closed.
	success bool

	parent   *sudog // semaRoot binary tree
	waitlink *sudog // g.waiting list or semaRoot
	waittail *sudog // semaRoot
	c        *hchan // channel
}

```





##### 状态

```GO
// defined constants
const (
	// G status
	//
	// Beyond indicating the general state of a G, the G status
	// acts like a lock on the goroutine's stack (and hence its
	// ability to execute user code).
	//
	// If you add to this list, add to the list
	// of "okay during garbage collection" status
	// in mgcmark.go too.
	//
	// TODO(austin): The _Gscan bit could be much lighter-weight.
	// For example, we could choose not to run _Gscanrunnable
	// goroutines found in the run queue, rather than CAS-looping
	// until they become _Grunnable. And transitions like
	// _Gscanwaiting -> _Gscanrunnable are actually okay because
	// they don't affect stack ownership.

	// _Gidle means this goroutine was just allocated and has not
	// yet been initialized.
	_Gidle = iota // 0

	// _Grunnable means this goroutine is on a run queue. It is
	// not currently executing user code. The stack is not owned.
	_Grunnable // 1

	// _Grunning means this goroutine may execute user code. The
	// stack is owned by this goroutine. It is not on a run queue.
	// It is assigned an M and a P (g.m and g.m.p are valid).
	_Grunning // 2

	// _Gsyscall means this goroutine is executing a system call.
	// It is not executing user code. The stack is owned by this
	// goroutine. It is not on a run queue. It is assigned an M.
	_Gsyscall // 3

	// _Gwaiting means this goroutine is blocked in the runtime.
	// It is not executing user code. It is not on a run queue,
	// but should be recorded somewhere (e.g., a channel wait
	// queue) so it can be ready()d when necessary. The stack is
	// not owned *except* that a channel operation may read or
	// write parts of the stack under the appropriate channel
	// lock. Otherwise, it is not safe to access the stack after a
	// goroutine enters _Gwaiting (e.g., it may get moved).
	_Gwaiting // 4

	// _Gmoribund_unused is currently unused, but hardcoded in gdb
	// scripts.
	_Gmoribund_unused // 5

	// _Gdead means this goroutine is currently unused. It may be
	// just exited, on a free list, or just being initialized. It
	// is not executing user code. It may or may not have a stack
	// allocated. The G and its stack (if any) are owned by the M
	// that is exiting the G or that obtained the G from the free
	// list.
	_Gdead // 6

	// _Genqueue_unused is currently unused.
	_Genqueue_unused // 7

	// _Gcopystack means this goroutine's stack is being moved. It
	// is not executing user code and is not on a run queue. The
	// stack is owned by the goroutine that put it in _Gcopystack.
	_Gcopystack // 8

	// _Gpreempted means this goroutine stopped itself for a
	// suspendG preemption. It is like _Gwaiting, but nothing is
	// yet responsible for ready()ing it. Some suspendG must CAS
	// the status to _Gwaiting to take responsibility for
	// ready()ing this G.
	_Gpreempted // 9

	// _Gscan combined with one of the above states other than
	// _Grunning indicates that GC is scanning the stack. The
	// goroutine is not executing user code and the stack is owned
	// by the goroutine that set the _Gscan bit.
	//
	// _Gscanrunning is different: it is used to briefly block
	// state transitions while GC signals the G to scan its own
	// stack. This is otherwise like _Grunning.
	//
	// atomicstatus&~Gscan gives the state the goroutine will
	// return to when the scan completes.
	_Gscan          = 0x1000
	_Gscanrunnable  = _Gscan + _Grunnable  // 0x1001
	_Gscanrunning   = _Gscan + _Grunning   // 0x1002
	_Gscansyscall   = _Gscan + _Gsyscall   // 0x1003
	_Gscanwaiting   = _Gscan + _Gwaiting   // 0x1004
	_Gscanpreempted = _Gscan + _Gpreempted // 0x1009
)
```





## GMP调度流程

### 初始化阶段

#### 函数入口

- 使用 **info files** 命令找到程序入口（Entry point）地址为0x452270，并用gdb b断点进入函数入口

    ```shell
    bobo@ubuntu:~/study/go$ gobuild hello.go 
    bobo@ubuntu:~/study/go$ gdbhello
    GNU gdb (GDB) 8.0.1
    (gdb) info files
    Symbols from "/home/bobo/study/go/main".
    Local exec file:
    `/home/bobo/study/go/main', file type elf64-x86-64.
    Entry point: 0x452270
    0x0000000000401000 -0x0000000000486aac is .text
    0x0000000000487000 -0x00000000004d1a73 is .rodata
    0x00000000004d1c20 -0x00000000004d27f0 is .typelink
    0x00000000004d27f0 -0x00000000004d2838 is .itablink
    0x00000000004d2838 -0x00000000004d2838 is .gosymtab
    0x00000000004d2840 -0x00000000005426d9 is .gopclntab
    0x0000000000543000 -0x000000000054fa9c is .noptrdata
    0x000000000054faa0 -0x0000000000556790 is .data
    0x00000000005567a0 -0x0000000000571ef0 is .bss
    0x0000000000571f00 -0x0000000000574658 is .noptrbss
    0x0000000000400f9c -0x0000000000401000 is .note.go.buildid
    (gdb) b *0x452270
    Breakpoint 1at 0x452270: file /usr/local/go/src/runtime/rt0_linux_amd64.s, line 8.
    ```

#### 初始化g0

```shell
TEXT _rt0_amd64_linux(SB),NOSPLIT,$-8
    JMP _rt0_amd64(SB)

TEXT _rt0_amd64(SB),NOSPLIT,$-8
	// 将操作系统内核传递过来的参数argc和argv数组的地址分别放在DI和SI寄存器中
    MOVQ 0(SP), DI // argc 
    LEAQ 8(SP), SI // argv
    JMP runtime·rt0_go(SB)



TEXT runtime·rt0_go(SB),NOSPLIT|TOPFRAME,$0
	// copy arguments forward on an even stack
	MOVQ	DI, AX		// argc
	MOVQ	SI, BX		// argv
	SUBQ	$(5*8), SP		// 3args 2auto
	ANDQ	$~15, SP
	MOVQ	AX, 24(SP)
	MOVQ	BX, 32(SP)

	// create istack out of the given (operating system) stack.
	// _cgo_init may update stackguard.
	// 初始化g0在寄存器上的栈信息
	MOVQ	$runtime·g0(SB), DI
	LEAQ	(-64*1024+104)(SP), BX
	MOVQ	BX, g_stackguard0(DI)
	MOVQ	BX, g_stackguard1(DI)
	MOVQ	BX, (g_stack+stack_lo)(DI)
	MOVQ	SP, (g_stack+stack_hi)(DI)
```

![image-20220705055926937](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/img/202207050559001.png)



#### 将主线程与g0与m0绑定

只要每个工作线程拥有了各自私有的m结构体全局变量，我们就能在不同的工作线程中使用相同的全局变量名来访问不同的m结构体对象里的G0。



###### 通过tls初始化m0中的g0变量

```shell
runtime·rt0_go(SB)
··········
  	//下面开始初始化tls(thread local storage,线程本地存储)

	LEAQ	runtime·m0+m_tls(SB), DI // 取m0的tls成员的地址到DI寄存器
	CALL	runtime·settls(SB) // 调用settls设置线程本地存储，settls函数的参数在DI寄存器中

	// store through it, to make sure it works
    //验证settls是否可以正常工作，如果有问题则abort退出程序
    get_tls(BX) //获取fs段基地址并放入BX寄存器，其实就是m0.tls[1]的地址，get_tls的代码由编译器生成
    MOVQ $0x123, g(BX) //把整型常量0x123拷贝到fs段基地址偏移-8的内存位置，也就是m0.tls[0] =0x123
    MOVQ runtime·m0+m_tls(SB), AX//AX=m0.tls[0]
    CMPQ AX, $0x123 //检查m0.tls[0]的值是否是通过线程本地存储存入的0x123来验证tls功能是否正常
    JEQ 2(PC)
    CALL runtime·abort(SB) //如果线程本地存储不能正常工作，退出程序

	// Cgo版本
	MOVQ	$runtime·g0(SB), CX
	MOVQ	(g_stack+stack_lo)(CX), AX
	ADDQ	$const__StackGuard, AX
	MOVQ	AX, g_stackguard0(CX)
	MOVQ	AX, g_stackguard1(CX)

	LEAQ	runtime·m0+m_tls(SB), DI
	CALL	runtime·settls(SB)

	// store through it, to make sure it works
	get_tls(BX)
	MOVQ	$0x123, g(BX)
	MOVQ	runtime·m0+m_tls(SB), AX
	CMPQ	AX, $0x123
	JEQ 2(PC)
	CALL	runtime·abort(SB)
    
  	
ok:
	// set the per-goroutine and per-mach "registers"
	get_tls(BX) //获取fs段基址到BX寄存器
	LEAQ	runtime·g0(SB), CX  // CX=g0的地址
	MOVQ	CX, g(BX)  //把g0的地址保存在线程本地存储里面，也就是m0.tls[0]=&g0
	LEAQ	runtime·m0(SB), AX //AX=m0的地址
	
	//把m0和g0关联起来m0->g0 =g0，g0->m =m0
	// save m->g0 = g0 
	MOVQ	CX, m_g0(AX)
	// save m0 to g0->m
	MOVQ	AX, g_m(CX)  
	
```

```shell
   // 详细看一下settls函数是如何实现线程私有全局变量的。
   // 结合get_tls(SB) 可以看出，tls会将不同线程里的m.tls内存地址存放在寄存器DI中，通过arch_prctl系统调	 //	用把m0.tls[1]的地址设置成了线程栈里fs段的段基址，而每个线程都有自己的一组CPU寄存器值，操作系统在把线	// 程调离CPU运行时会帮我们把所有寄存器中的值保存在内存中，调度线程起来运行时又会从内存中把这些寄存器的值    // 恢复到CPU，在此之后，工作线程代码就可以通过fs寄存器来找到m.tls
   // set tls base to DI
   TEXT runtime·settls(SB),NOSPLIT,$32
   //......
   //DI寄存器中存放的是m.tls[0]的地址，m的tls成员是一个数组，读者如果忘记了可以回头看一下m结构体的定义
   //下面这一句代码把DI寄存器中的地址加8，为什么要+8呢，主要跟ELF可执行文件格式中的TLS实现的机制有关
   //执行下面这句指令之后DI寄存器中的存放的就是m.tls[1]的地址了
   ADDQ $8, DI// ELF wants to use -8(FS)

    //下面通过arch_prctl系统调用设置FS段基址
    MOVQ DI, SI//SI存放arch_prctl系统调用的第二个参数
    MOVQ $0x1002, DI// ARCH_SET_FS //arch_prctl的第一个参数
    MOVQ $SYS_arch_prctl, AX//系统调用编号
    SYS CALL
    CMP QAX, $0xfffffffffffff001
    JLS 2(PC)
    MOVL $0xf1, 0xf1 // crash //系统调用失败直接crash
    RET
```



此时，主线程，m0，g0以及g0的栈之间的关系如下图所示：

![image-20220705134506771](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/img/202207051345841.png)



#### 小结

在程序启动时，经过内核线程调度，主线程会调用Golang中不同系统的启动程序，将初始化g0，并通过线程本地共享机制，将g0放在m0的寄存器fs段，且互相独立，以便于golang的工作线程m可以方便快速的调用g0，最后把m和g0绑定在一起，这样，之后在主线程中通过get_tls可以获取到g0，通过g0的m成员又可以找到m0，于是这里就实现了m0和g0与主线程之间的关联。



### 启动阶段



#### 总览

在经过程序对用户输入的参数处理后，将以下面的顺序调用函数启动程序

```GO
// The bootstrap sequence is:
//
//	call osinit
//	call schedinit
//	make & queue new G
//	call runtime·mstart


runtime·rt0_go(SB)

	MOVW	8(RSP), R0	// copy argc
	MOVW	R0, -8(RSP)
	MOVD	16(RSP), R0		// copy argv
	MOVD	R0, 0(RSP)
	BL	runtime·args(SB)
	BL	runtime·osinit(SB)
	BL	runtime·schedinit(SB)

	// create a new goroutine to start program
	MOVD	$runtime·mainPC(SB), R0		// entry
	SUB	$16, RSP
	MOVD	R0, 8(RSP) // arg
	MOVD	$0, 0(RSP) // dummy LR
	BL	runtime·newproc(SB)
	ADD	$16, RSP

	// start this M
	BL	runtime·mstart(SB)

	// Prevent dead-code elimination of debugCallV2, which is
	// intended to be called by debuggers.
	MOVD	$runtime·debugCallV2<ABIInternal>(SB), R0

	MOVD	$0, R0
	MOVD	R0, (R0)	// boom
	UNDEF

```





#### OSInit--操作系统信息初始化

- 根据CPU的不同调用不同的初始化OS初始化函数

![image-20220705051028511](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/img/202207050510596.png)



- 通过读取设备信息，设置CPU的一些特性位, 调度器初始化时需要知道当前系统有多少个CPU核。

```go
//go:build arm64 && linux && !android

package cpu

func osInit() {
	hwcapInit("linux")
}


func hwcapInit(os string) {
	// HWCap was populated by the runtime from the auxiliary vector.
	// Use HWCap information since reading aarch64 system registers
	// is not supported in user space on older linux kernels.
	ARM64.HasAES = isSet(HWCap, hwcap_AES)
	ARM64.HasPMULL = isSet(HWCap, hwcap_PMULL)
	ARM64.HasSHA1 = isSet(HWCap, hwcap_SHA1)
	ARM64.HasSHA2 = isSet(HWCap, hwcap_SHA2)
	ARM64.HasCRC32 = isSet(HWCap, hwcap_CRC32)
	ARM64.HasCPUID = isSet(HWCap, hwcap_CPUID)

	// The Samsung S9+ kernel reports support for atomics, but not all cores
	// actually support them, resulting in SIGILL. See issue #28431.
	// TODO(elias.naur): Only disable the optimization on bad chipsets on android.
	ARM64.HasATOMICS = isSet(HWCap, hwcap_ATOMICS) && os != "android"

	// Check to see if executing on a NeoverseN1 and in order to do that,
	// check the AUXV for the CPUID bit. The getMIDR function executes an
	// instruction which would normally be an illegal instruction, but it's
	// trapped by the kernel, the value sanitized and then returned. Without
	// the CPUID bit the kernel will not trap the instruction and the process
	// will be terminated with SIGILL.
	if ARM64.HasCPUID {
		midr := getMIDR()
		part_num := uint16((midr >> 4) & 0xfff)
		implementor := byte((midr >> 24) & 0xff)

		if implementor == 'A' && part_num == 0xd0c {
			ARM64.IsNeoverseN1 = true
		}
		if implementor == 'A' && part_num == 0xd40 {
			ARM64.IsZeus = true
		}
	}
}

var ARM64 struct {
	_            CacheLinePad
	HasAES       bool
	HasPMULL     bool
	HasSHA1      bool
	HasSHA2      bool
	HasCRC32     bool
	HasATOMICS   bool
	HasCPUID     bool
	IsNeoverseN1 bool
	IsZeus       bool
	_            CacheLinePad
}
```





#### Schedinit -- 调度器初始化

##### 初始化调度器中的锁

```GO

// The new G calls runtime·main.
func schedinit() {
	lockInit(&sched.lock, lockRankSched)
	lockInit(&sched.sysmonlock, lockRankSysmon)
	lockInit(&sched.deferlock, lockRankDefer)
	lockInit(&sched.sudoglock, lockRankSudog)
	lockInit(&deadlock, lockRankDeadlock)
	lockInit(&paniclk, lockRankPanic)
	lockInit(&allglock, lockRankAllg)
	lockInit(&allpLock, lockRankAllp)
	lockInit(&reflectOffs.lock, lockRankReflectOffs)
	lockInit(&finlock, lockRankFin)
	lockInit(&trace.bufLock, lockRankTraceBuf)
	lockInit(&trace.stringsLock, lockRankTraceStrings)
	lockInit(&trace.lock, lockRankTrace)
	lockInit(&cpuprof.lock, lockRankCpuprof)
	lockInit(&trace.stackTab.lock, lockRankTraceStackTab)
	// Enforce that this lock is always a leaf lock.
	// All of this lock's critical sections should be
	// extremely short.
	lockInit(&memstats.heapStats.noPLock, lockRankLeafRank)
```



##### 获取主线程tls中的g0调度代码


```GO
	// raceinit must be the first call to race detector.
	// In particular, it must be done before mallocinit below calls racemapshadow.

    //getg函数在源代码中没有对应的定义，由编译器插入类似下面两行代码
    //get_tls(CX) 
    //MOVQ g(CX), BX; BX存器里面现在放的是当前g结构体对象的地址	
	_g_ := getg()
	if raceenabled {
		_g_.racectx, raceprocctx0 = raceinit()
	}

	//设置最多启动10000个操作系统线程，也是最多10000个M
	sched.maxmcount = 10000
```



##### m与p的初始化

- 通过mcommoninit初始化m0
- 通过procresize 初始化

```go
	// The world starts stopped.
	worldStopped()

	moduledataverify()
	stackinit()
	mallocinit()
	cpuinit()      // must run before alginit
	alginit()      // maps, hash, fastrand must not be used before this call
	fastrandinit() // must run before mcommoninit
	mcommoninit(_g_.m, -1)  //初始化m0，因为从前面的代码我们知道g0->m = &m0
	modulesinit()   // provides activeModules
	typelinksinit() // uses maps, activeModules
	itabsinit()     // uses activeModules
	stkobjinit()    // must run before GC starts

	sigsave(&_g_.m.sigmask)
	initSigmask = _g_.m.sigmask

	if offset := unsafe.Offsetof(sched.timeToRun); offset%8 != 0 {
		println(offset)
		throw("sched.timeToRun not aligned to 8 bytes")
	}

	goargs()
	goenvs()
	parsedebugvars()
	gcinit()

	lock(&sched.lock)
	sched.lastpoll = uint64(nanotime())

	// 根据cpu核心数，创建相应的P队列
	procs := ncpu
	if n, ok := atoi32(gogetenv("GOMAXPROCS")); ok && n > 0 {
		procs = n
	}
	if procresize(procs) != nil {
		throw("unknown runnable goroutine during bootstrap")
	}
	unlock(&sched.lock)

	// World is effectively started now, as P's can run.
	worldStarted()

	// For cgocheck > 1, we turn on the write barrier at all times
	// and check all pointer writes. We can't do this until after
	// procresize because the write barrier needs a P.
	if debug.cgocheck > 1 {
		writeBarrier.cgo = true
		writeBarrier.enabled = true
		for _, p := range allp {
			p.wbBuf.reset()
		}
	}

	if buildVersion == "" {
		// Condition should never trigger. This code just serves
		// to ensure runtime·buildVersion is kept in the resulting binary.
		buildVersion = "unknown"
	}
	if len(modinfo) == 1 {
		// Condition should never trigger. This code just serves
		// to ensure runtime·modinfo is kept in the resulting binary.
		modinfo = ""
	}
}
```



######  mcommoninit

- 为m0创建信号处理的gsignal
- 将m0挂在全局M队列上

```go
// Pre-allocated ID may be passed as 'id', or omitted by passing -1.
func mcommoninit(mp *m, id int64) {
	_g_ := getg()

	// g0 stack won't make sense for user (and is not necessary unwindable).
	if _g_ != _g_.m.g0 {
		callers(1, mp.createstack[:])
	}

	lock(&sched.lock)

	if id >= 0 {
		mp.id = id
	} else {
		mp.id = mReserveID()
	}

	lo := uint32(int64Hash(uint64(mp.id), fastrandseed))
	hi := uint32(int64Hash(uint64(cputicks()), ^fastrandseed))
	if lo|hi == 0 {
		hi = 1
	}
	// Same behavior as for 1.17.
	// TODO: Simplify ths.
	if goarch.BigEndian {
		mp.fastrand = uint64(lo)<<32 | uint64(hi)
	} else {
		mp.fastrand = uint64(hi)<<32 | uint64(lo)
	}

	//创建用于信号处理的gsignal，只是简单的从堆上分配一个g结构体对象,然后把栈设置好就返回了
	mpreinit(mp)
	if mp.gsignal != nil {
		mp.gsignal.stackguard1 = mp.gsignal.stack.lo + _StackGuard
	}

	// Add to allm so garbage collector doesn't free g->m
	// when it is just in a register or thread-local storage.
	mp.alllink = allm

	// NumCgoCall() iterates over allm w/o schedlock,
	// so we need to publish it safely.
	atomicstorep(unsafe.Pointer(&allm), unsafe.Pointer(mp))
	unlock(&sched.lock)

	// Allocate memory to hold a cgo traceback if the cgo call crashes.
	if iscgo || GOOS == "solaris" || GOOS == "illumos" || GOOS == "windows" {
		mp.cgoCallers = new(cgoCallers)
	}
}
```



###### procresize

- 初始化全局队列P
- 初始化MaxP数量的P
- 互挂P0与M0

```
// Change number of processors.
//
// sched.lock must be held, and the world must be stopped.
//
// gcworkbufs must not be being modified by either the GC or the write barrier
// code, so the GC must not be running if the number of Ps actually changes.
//
// Returns list of Ps with local work, they need to be scheduled by the caller.
func procresize(nprocs int32) *p {
	assertLockHeld(&sched.lock)
	assertWorldStopped()

	old := gomaxprocs
	if old < 0 || nprocs <= 0 {
		throw("procresize: invalid arg")
	}
	if trace.enabled {
		traceGomaxprocs(nprocs)
	}

	// update statistics
	now := nanotime()
	if sched.procresizetime != 0 {
		sched.totaltime += int64(old) * (now - sched.procresizetime)
	}
	sched.procresizetime = now

	maskWords := (nprocs + 31) / 32


	// Grow allp if necessary.
	//初始化时 len(allp) == 0
	if nprocs > int32(len(allp)) {
		// Synchronize with retake, which could be running
		// concurrently since it doesn't run on a P.
		lock(&allpLock)
		if nprocs <= int32(cap(allp)) {
			allp = allp[:nprocs]
		} else {
			nallp := make([]*p, nprocs)
			// 将底层数组的所有元素拷贝到新数组
			// Copy everything up to allp's cap so we
			// never lose old allocated Ps.
			copy(nallp, allp[:cap(allp)])
			allp = nallp
		}

		if maskWords <= int32(cap(idlepMask)) {
			idlepMask = idlepMask[:maskWords]
			timerpMask = timerpMask[:maskWords]
		} else {
			nidlepMask := make([]uint32, maskWords)
			// No need to copy beyond len, old Ps are irrelevant.
			copy(nidlepMask, idlepMask)
			idlepMask = nidlepMask

			ntimerpMask := make([]uint32, maskWords)
			copy(ntimerpMask, timerpMask)
			timerpMask = ntimerpMask
		}
		unlock(&allpLock)
	}

	// initialize new P's
	for i := old; i < nprocs; i++ {
		pp := allp[i]
		if pp == nil {
		    //调用内存分配器从堆上分配一个struct p
			pp = new(p)
		}
		pp.init(i)
		atomicstorep(unsafe.Pointer(&allp[i]), unsafe.Pointer(pp))
	}

	_g_ := getg()
	//初始化时m0->p还未初始化，所以不会执行这个分支
	if _g_.m.p != 0 && _g_.m.p.ptr().id < nprocs {
		// continue to use the current P
		_g_.m.p.ptr().status = _Prunning
		_g_.m.p.ptr().mcache.prepareForSweep()
	} else {
		// release the current P and acquire allp[0].
		//
		// We must do this before destroying our current P
		// because p.destroy itself has write barriers, so we
		// need to do that from a valid P.
		if _g_.m.p != 0 {
			if trace.enabled {
				// Pretend that we were descheduled
				// and then scheduled again to keep
				// the trace sane.
				traceGoSched()
				traceProcStop(_g_.m.p.ptr())
			}
			_g_.m.p.ptr().m = 0
		}
		//把p和m0关联起来，其实是这两个strct的成员相互赋值
		_g_.m.p = 0
		p := allp[0]
		p.m = 0
		p.status = _Pidle
		acquirep(p)
		if trace.enabled {
			traceGoStart()
		}
	}

	// g.m.p is now set, so we no longer need mcache0 for bootstrapping.
	mcache0 = nil

	// 销毁没有用的P
	// release resources from unused P's
	for i := nprocs; i < old; i++ {
		p := allp[i]
		p.destroy()
		// can't free P itself because it can be referenced by an M in syscall
	}

	// Trim allp.
	if int32(len(allp)) != nprocs {
		lock(&allpLock)
		allp = allp[:nprocs]
		idlepMask = idlepMask[:maskWords]
		timerpMask = timerpMask[:maskWords]
		unlock(&allpLock)
	}

	var runnablePs *p
	for i := nprocs - 1; i >= 0; i-- {
		p := allp[i]
		if _g_.m.p.ptr() == p {
			continue
		}
		p.status = _Pidle
		// 把除了allp[0]之外的所有p放入到全局变量sched的pidle空闲队列之中
		if runqempty(p) {
			pidleput(p, now)
		} else {
			p.m.set(mget())
			p.link.set(runnablePs)
			runnablePs = p
		}
	}
	stealOrder.reset(uint32(nprocs))
	var int32p *int32 = &gomaxprocs // make compiler check that gomaxprocs is an int32
	atomic.Store((*uint32)(unsafe.Pointer(int32p)), uint32(nprocs))
	if old != nprocs {
		// Notify the limiter that the amount of procs has changed.
		gcCPULimiter.resetCapacity(now, nprocs)
	}
	return runnablePs
}
```

![image-20220705145658564](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/img/202207051457583.png)



#### make & queue new G



##### mainPc

###### 将主函数入口和参数压入栈中

```shell
// create a new goroutine to start program
	MOVD	$runtime·mainPC(SB), R0		// entry 将main函数入口赋值给R0
	SUB	$16, RSP
	MOVD	R0, 8(RSP) // arg 参数1： newproc的第二个参数入栈，也就是新的goroutine需要执行的函数
	MOVD	$0, 0(RSP) // dummy LR 参数2：newproc的第一个参数入栈，该参数表示runtime.main函数需要的参数大小，因为runtime.main没有参数，所以这里是0
```





##### newProc

- 将rumtime.main函数打包为Goroutine，并加入运行队列（在g0 中 执行newProc时，不允许发生抢占）

```go
BL	runtime·newproc(SB) // 为main函数构建一个goroutine
ADD	$16, RSP // 调整pc到刚刚创建好的goroutine起始位置
	
// Create a new g running fn.
// Put it on the queue of g's waiting to run.
// The compiler turns a go statement into a call to this.
func newproc(fn *funcval) {
	gp := getg() // 获取g0
	pc := getcallerpc() // SUB	$16, RSP 获取该结果的PC位置
	
	//由于现在是在主线程的栈上进行计算的
    //systemstack的作用是切换到g0栈执行作为参数的函数
	systemstack(func() {
        
        // 创建或设置newg里面的调度信息
		newg := newproc1(fn, gp, pc)

		_p_ := getg().m.p.ptr()
        
        // 将可执行的G扔到本地P的可运行G里或者全局可运行G
		runqput(_p_, newg, true)

		if mainStarted {
			wakep()
		}
	})
}



// Create a new g in state _Grunnable, starting at fn. callerpc is the
// address of the go statement that created this. The caller is responsible
// for adding the new g to the scheduler.
func newproc1(fn *funcval, callergp *g, callerpc uintptr) *g {
	_g_ := getg()

	if fn == nil {
		fatal("go of nil func value")
	}
	// 不允许g（这里是g0）执行过程中 m发生抢占，即使g的preempt被设置为true
    // 返回当前g的m
	acquirem() // disable preemption because it can be holding p in a local var

    // 获取当前m结合的p， allp[0]
	_p_ := _g_.m.p.ptr()
    // //从p的本地缓冲里获取一个自由的g，初始化时没有，返回nil
	newg := gfget(_p_)
	if newg == nil {
		 //new一个g结构体对象，然后从堆上为其分配栈，并设置g的stack成员和两个stackgard成员
		newg = malg(_StackMin)
		// 初始化g的状态为_Gdead
		casgstatus(newg, _Gidle, _Gdead)
        // 将刚创建的G加入到全局G队列
		allgadd(newg) // publishes with a g->status of Gdead so GC scanner doesn't look at uninitialized stack.
	}
	if newg.stack.hi == 0 {
		throw("newproc1: newg missing stack")
	}

	if readgstatus(newg) != _Gdead {
		throw("newproc1: new g is not Gdead")
	}

	totalSize := uintptr(4*goarch.PtrSize + sys.MinFrameSize) // extra space in case of reads slightly beyond frame
	totalSize = alignUp(totalSize, sys.StackAlign)
	sp := newg.stack.hi - totalSize
	spArg := sp
	if usesLR {
		// caller's LR
		*(*uintptr)(unsafe.Pointer(sp)) = 0
		prepGoExitFrame(sp)
		spArg += sys.MinFrameSize
	}
	//把newg.sched结构体成员的所有成员设置为0
	memclrNoHeapPointers(unsafe.Pointer(&newg.sched), unsafe.Sizeof(newg.sched))
    
    // 记录g的调度信息
	newg.sched.sp = sp // newg栈顶
	newg.stktopsp = sp
    
    //newg.sched.pc表示当newg被调度起来运行时从这个地址开始执行指令
    //把pc设置成了goexit这个函数偏移1（sys.PCQuantum等于1）的位置，
    //至于为什么要这么做需要等到分析完gostartcallfn函数才知道 
	newg.sched.pc = abi.FuncPCABI0(goexit) + sys.PCQuantum // +PCQuantum so that previous instruction is in same function
	newg.sched.g = guintptr(unsafe.Pointer(newg))
    
    // 更新newg上的栈信息，再调用完fn后通过goexit进行goroutine善后处理
	gostartcallfn(&newg.sched, fn)
	newg.gopc = callerpc
	newg.ancestors = saveAncestors(callergp)
    //设置newg的startpc为fn.fn，该成员主要用于函数调用栈的traceback和栈收缩
    //newg真正从哪里开始执行并不依赖于这个成员，而是sched.pc
	newg.startpc = fn.fn
	if isSystemGoroutine(newg, false) {
		atomic.Xadd(&sched.ngsys, +1)
	} else {
		// Only user goroutines inherit pprof labels.
		if _g_.m.curg != nil {
			newg.labels = _g_.m.curg.labels
		}
		if goroutineProfile.active {
			// A concurrent goroutine profile is running. It should include
			// exactly the set of goroutines that were alive when the goroutine
			// profiler first stopped the world. That does not include newg, so
			// mark it as not needing a profile before transitioning it from
			// _Gdead.
			newg.goroutineProfiled.Store(goroutineProfileSatisfied)
		}
	}
	// Track initial transition?
	newg.trackingSeq = uint8(fastrand())
	if newg.trackingSeq%gTrackingPeriod == 0 {
		newg.tracking = true
	}
    
    // 设置g的状态为_Grunnable，表示这个g代表的goroutine可以运行了
	casgstatus(newg, _Gdead, _Grunnable)
	gcController.addScannableStack(_p_, int64(newg.stack.hi-newg.stack.lo))

	if _p_.goidcache == _p_.goidcacheend {
		// Sched.goidgen is the last allocated id,
		// this batch must be [sched.goidgen+1, sched.goidgen+GoidCacheBatch].
		// At startup sched.goidgen=0, so main goroutine receives goid=1.
		_p_.goidcache = atomic.Xadd64(&sched.goidgen, _GoidCacheBatch)
		_p_.goidcache -= _GoidCacheBatch - 1
		_p_.goidcacheend = _p_.goidcache + _GoidCacheBatch
	}
	newg.goid = int64(_p_.goidcache)
	_p_.goidcache++
	if raceenabled {
		newg.racectx = racegostart(callerpc)
		if newg.labels != nil {
			// See note in proflabel.go on labelSync's role in synchronizing
			// with the reads in the signal handler.
			racereleasemergeg(newg, unsafe.Pointer(&labelSync))
		}
	}
	if trace.enabled {
		traceGoCreate(newg, newg.startpc)
	}
	releasem(_g_.m)

	return newg
}

```

###### runqput

```go

// runqput tries to put g on the local runnable queue.
// If next is false, runqput adds g to the tail of the runnable queue.
// If next is true, runqput puts g in the _p_.runnext slot.
// If the run queue is full, runnext puts g on the global queue.
// Executed only by the owner P.
// 1. 可以选择预先检查runnext指针上是否有可运行的G，如果有将runnext 设置为newg，并将old放入本地队列
// 2. 本地队列是一个环形队列，如果本地队列没有满的情况下直接在队尾加入oldg
// 3. 如果队列为满的情况下，则将队列的head + len/2 的g批量加入到全局可运行g列表中
func runqput(_p_ *p, gp *g, next bool) {
	if randomizeScheduler && next && fastrandn(2) == 0 {
		next = false
	}

	if next {
	retryNext:
		oldnext := _p_.runnext
		if !_p_.runnext.cas(oldnext, guintptr(unsafe.Pointer(gp))) {
			goto retryNext
		}
		if oldnext == 0 {
			return
		}
		// Kick the old runnext out to the regular run queue.
		gp = oldnext.ptr()
	}

retry:
	h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with consumers
	t := _p_.runqtail
	if t-h < uint32(len(_p_.runq)) {
		_p_.runq[t%uint32(len(_p_.runq))].set(gp)
		atomic.StoreRel(&_p_.runqtail, t+1) // store-release, makes the item available for consumption
		return
	}
	if runqputslow(_p_, gp, h, t) {
		return
	}
	// the queue is not full, now the put above must succeed
	goto retry
}

// Put g and a batch of work from local runnable queue on global queue.
// Executed only by the owner P.
func runqputslow(_p_ *p, gp *g, h, t uint32) bool {
	var batch [len(_p_.runq)/2 + 1]*g

	// First, grab a batch from local queue.
	n := t - h
	n = n / 2
	if n != uint32(len(_p_.runq)/2) {
		throw("runqputslow: queue is not full")
	}
	for i := uint32(0); i < n; i++ {
		batch[i] = _p_.runq[(h+i)%uint32(len(_p_.runq))].ptr()
	}
	if !atomic.CasRel(&_p_.runqhead, h, h+n) { // cas-release, commits consume
		return false
	}
	batch[n] = gp

    // 重排序逻辑
	if randomizeScheduler {
		for i := uint32(1); i <= n; i++ {
			j := fastrandn(i + 1)
			batch[i], batch[j] = batch[j], batch[i]
		}
	}

	// Link the goroutines.
	for i := uint32(0); i < n; i++ {
		batch[i].schedlink.set(batch[i+1])
	}
	var q gQueue
	q.head.set(batch[0])
	q.tail.set(batch[n])

	// Now put the batch on global queue.
	lock(&sched.lock)
	globrunqputbatch(&q, int32(n+1))
	unlock(&sched.lock)
	return true
}
```

###### gostartcallfn

-  newg调度前的参数设置


```go



// adjust Gobuf as if it executed a call to fn
// and then stopped before the first instruction in fn.
func gostartcallfn(gobuf *gobuf, fv *funcval) {
	var fn unsafe.Pointer
	if fv != nil {
		fn  = unsafe.Pointer(fv.fn)
	} else {
		fn = unsafe.Pointer(abi.FuncPCABIInternal(nilfunc))
	}
	gostartcall(gobuf, fn, unsafe.Pointer(fv))
}




// sp: goexit + 1
// pc: runtime.main
// 等到newg被调度起来运行时，调度器会把buf.pc放入cpu的IP寄存器
// adjust Gobuf as if it executed a call to fn with context ctxt
// and then stopped before the first instruction in fn.
func gostartcall(buf *gobuf, fn, ctxt unsafe.Pointer) {
	sp := buf.sp
	sp -= goarch.PtrSize //为返回地址预留空间，向下代表增长栈空间
	*(*uintptr)(unsafe.Pointer(sp)) = buf.pc //在栈上放入goexit+1的地址
	buf.sp = sp // 重新设置newg的栈顶寄存器
	buf.pc = uintptr(fn) // 指向函数，告诉cpu下一步要执行的地址
	buf.ctxt = ctxt
}


```

###### acquirem &  releasem


```go


//go:nosplit
func acquirem() *m {
	_g_ := getg()
	_g_.m.locks++
	return _g_.m
}

//go:nosplit
func releasem(mp *m) {
	_g_ := getg()
	mp.locks--
	if mp.locks == 0 && _g_.preempt {
		// restore the preemption request in case we've cleared it in newstack
		_g_.stackguard0 = stackPreempt
	}
}


```

###### allgadd


```go


func allgadd(gp *g) {
    // _Gidle means this goroutine was just allocated and has not
	// yet been initialized.
    // 可以看出Gidle 不能进入到 allg中
	if readgstatus(gp) == _Gidle {
		throw("allgadd: bad status Gidle")
	}

	lock(&allglock)
	allgs = append(allgs, gp)
	if &allgs[0] != allgptr {
		atomicstorep(unsafe.Pointer(&allgptr), unsafe.Pointer(&allgs[0]))
	}
	atomic.Storeuintptr(&allglen, uintptr(len(allgs)))
	unlock(&allglock)
}

```

###### malg


```go


// Allocate a new g, with a stack big enough for stacksize bytes.
func malg(stacksize int32) *g {
	newg := new(g)
	if stacksize >= 0 {
		stacksize = round2(_StackSystem + stacksize)
		systemstack(func() {
			newg.stack = stackalloc(uint32(stacksize))
		})
		newg.stackguard0 = newg.stack.lo + _StackGuard
		newg.stackguard1 = ^uintptr(0)
		// Clear the bottom word of the stack. We record g
		// there on gsignal stack during VDSO on ARM and ARM64.
		*(*uintptr)(unsafe.Pointer(newg.stack.lo)) = 0
	}
	return newg
}

```

###### gfget


```go


// Get from gfree list.
// If local list is empty, grab a batch from global list.
func gfget(_p_ *p) *g {
retry:
    // 当本地P的自由G队列中没有G时，回去检测调度器中的自由G列表
	if _p_.gFree.empty() && (!sched.gFree.stack.empty() || !sched.gFree.noStack.empty()) {
        // 由于是在全局队列上进行数据操作，需要加锁
		lock(&sched.gFree.lock)
		// Move a batch of free Gs to the P.
        // 填充本地自由G队列直到数量达到32个
		for _p_.gFree.n < 32 {
			// Prefer Gs with stacks.
			gp := sched.gFree.stack.pop()
			if gp == nil {
				gp = sched.gFree.noStack.pop()
				if gp == nil {
					break
				}
			}
			sched.gFree.n--
			_p_.gFree.push(gp)
			_p_.gFree.n++
		}
		unlock(&sched.gFree.lock)
		goto retry
	}
    
    // 先入先出的弹出G
	gp := _p_.gFree.pop()
    // 初始化时获取的GFree是空的
	if gp == nil {
		return nil
	}
	_p_.gFree.n--
    
    // 对于之前有使用过的goRoutine，再次被调用，要清空之前的调用栈
	if gp.stack.lo != 0 && gp.stack.hi-gp.stack.lo != uintptr(startingStackSize) {
		// Deallocate old stack. We kept it in gfput because it was the
		// right size when the goroutine was put on the free list, but
		// the right size has changed since then.
		systemstack(func() {
			stackfree(gp.stack)
			gp.stack.lo = 0
			gp.stack.hi = 0
			gp.stackguard0 = 0
		})
	}
	if gp.stack.lo == 0 {
		// Stack was deallocated in gfput or just above. Allocate a new one.
		systemstack(func() {
			gp.stack = stackalloc(startingStackSize)
		})
		gp.stackguard0 = gp.stack.lo + _StackGuard
	} else {
		if raceenabled {
			racemalloc(unsafe.Pointer(gp.stack.lo), gp.stack.hi-gp.stack.lo)
		}
		if msanenabled {
			msanmalloc(unsafe.Pointer(gp.stack.lo), gp.stack.hi-gp.stack.lo)
		}
		if asanenabled {
			asanunpoison(unsafe.Pointer(gp.stack.lo), gp.stack.hi-gp.stack.lo)
		}
	}
	return gp
}

```

###### goexit

```go
// Finishes execution of the current goroutine.
func goexit1() {
    if raceenabled {  //与竞态检查有关，不关注
        racegoend()
    }
    if trace.enabled { //与backtrace有关，不关注
        traceGoEnd()
    }
    // 首先从当前运行的g切换到g0，这一步包括保存当前g的调度信息，把g0设置到tls中，修改CPU的rsp寄存器使其指向g0的栈；
	// 以当前g作为参数调用fn函数(此处为goexit0)。 
    mcall(goexit0)
}

func goexit0(gp *g) {
    // 这里的 _g_ 是g0
    _g_ := getg()

    // 状态切换
    casgstatus(gp, _Grunning, _Gdead)

    // 系统goroutines 数据减少1
    if isSystemGoroutine(gp, false) {
        atomic.Xadd(&sched.ngsys, -1)
    }

    // 释放相关资源，比较简单
    gp.m = nil
    locked := gp.lockedm != 0
    gp.lockedm = 0
    _g_.m.lockedg = 0
    gp.preemptStop = false
    gp.paniconfault = false
    gp._defer = nil // should be true already but just in case.
    gp._panic = nil // non-nil for Goexit during panic. points at stack-allocated data.
    gp.writebuf = nil
    gp.waitreason = 0
    gp.param = nil
    gp.labels = nil
    gp.timer = nil

    if gcBlackenEnabled != 0 && gp.gcAssistBytes > 0 {
        // Flush assist credit to the global pool. This gives
        // better information to pacing if the application is
        // rapidly creating an exiting goroutines.
        scanCredit := int64(gcController.assistWorkPerByte * float64(gp.gcAssistBytes))
        atomic.Xaddint64(&gcController.bgScanCredit, scanCredit)
        gp.gcAssistBytes = 0
    }

    // 删除 m 与当前gourtine m->curg 的关联
    dropg()

    if GOARCH == "wasm" { // no threads yet on wasm
        gfput(_g_.m.p.ptr(), gp)
        schedule() // never returns
    }

    if _g_.m.lockedInt != 0 {
        print("invalid m->lockedInt = ", _g_.m.lockedInt, "n")
        throw("internal lockOSThread error")
    }

    // 将g放入当前 P 的 gFree 列表，如果列表太长(>=64)，则转移一半的g到全局列表 sched.gFree
    // 把g放入p的freeg队列缓存起来供下次创建g时快速获取而不用从内存分配。freeg就是g的一个对象池；
    gfput(_g_.m.p.ptr(), gp)

    if locked {
        // The goroutine may have locked this thread because
        // it put it in an unusual kernel state. Kill it
        // rather than returning it to the thread pool.

        // Return to mstart, which will release the P and exit
        // the thread.
        if GOOS != "plan9" { // See golang.org/issue/22227.
            gogo(&_g_.m.g0.sched)
        } else {
            // Clear lockedExt on plan9 since we may end up re-using
            // this thread.
            _g_.m.lockedExt = 0
        }
    }

    // 调度
    schedule()
}

// Put on gfree list.
// If local list is too long, transfer a batch to the global list.
func gfput(_p_ *p, gp *g) {
    if readgstatus(gp) != _Gdead {
        throw("gfput: bad status (not Gdead)")
    }
    stksize := gp.stack.hi - gp.stack.lo
    if stksize != _FixedStack {
        // non-standard stack size - free it.
        // 非标准大小，就释放栈
        stackfree(gp.stack)
        gp.stack.lo = 0
        gp.stack.hi = 0
        gp.stackguard0 = 0
    }
    _p_.gFree.push(gp)
    _p_.gFree.n++
    if _p_.gFree.n >= 64 {
        lock(&sched.gFree.lock)
        for _p_.gFree.n >= 32 {
            _p_.gFree.n--
            gp = _p_.gFree.pop()
            if gp.stack.lo == 0 {
                sched.gFree.noStack.push(gp)
            } else {
                sched.gFree.stack.push(gp)
            }
            sched.gFree.n++
        }
        unlock(&sched.gFree.lock)
    }
}
```



![image-20220706045037448](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/img/202207060450069.png)

#### call runtime·mstart

##### rt0_go中的调用

```shell
// start this M
BL	runtime·mstart(SB) # 主线程进入调度循环，运行刚刚创建的goroutine

# 上面的mstart永远不应该返回的，如果返回了，一定是代码逻辑有问题，直接abort
CALL  runtime·abort(SB)// mstart should never return
RET

DATA  runtime·mainPC+0(SB)/8,$runtime·main(SB)
GLOB  Lruntime·mainPC(SB),RODATA,$8
```

##### mstart()

```go
func mstart() {
	_g_ := getg() //_g_ = g0

    //对于启动过程来说，g0的stack.lo早已完成初始化，所以onStack = false
	osStack := _g_.stack.lo == 0
	if osStack {
		// Initialize stack bounds from system stack.
		// Cgo may have left stack size in stack.hi.
		// minit may update the stack bounds.
		size := _g_.stack.hi
		if size == 0 {
			size = 8192 * sys.StackGuardMultiplier
		}
		_g_.stack.hi = uintptr(noescape(unsafe.Pointer(&size)))
		_g_.stack.lo = _g_.stack.hi - size + 1024
	}
	// Initialize stack guards so that we can start calling
	// both Go and C functions with stack growth prologues.
	_g_.stackguard0 = _g_.stack.lo + _StackGuard
	_g_.stackguard1 = _g_.stackguard0
    
	mstart1()

	// Exit this thread.
	if GOOS == "windows" || GOOS == "solaris" || GOOS == "plan9" || GOOS == "darwin" || GOOS == "aix" {
		// Window, Solaris, Darwin, AIX and Plan 9 always system-allocate
		// the stack, but put it in _g_.stack before mstart,
		// so the logic above hasn't set osStack yet.
		osStack = true
	}
	mexit(osStack)
}
```

##### mstart1()

- 通过getcallerpc、getcallersp 搭配save记录g0的函数栈调用信息，这个要和newg里的pc和sp区分开

```go
func mstart1() {
	_g_ := getg()  //启动过程时 _g_ = m0的g0

	if _g_ != _g_.m.g0 {
		throw("bad runtime·mstart")
	}

	// Record the caller for use as the top of stack in mcall and
	// for terminating the thread.
	// We're never coming back to mstart1 after we call schedule,
	// so other calls can reuse the current frame.
      //getcallerpc()g0.sched.pc指向的是mstart函数中调用mstart1函数之后的 if 语句，因为Schedule不会有返回值返回，将一直在进行调度循环,直到退出，此时会切回g0的堆栈，按g0.sched.pc执行
      //getcallersp()g0.sched.sp指向了mstart1函数执行完成后的返回地址，该地址保存在了mstart函数的栈帧之中
    // 这样设置之后，多个新建的M在生命结束时，都会走到相同的getcallerpc(), getcallersp()函数逻辑，执行mexit(osStack)，关闭OS线程
    // 但是runtime.main函数除外，当某个M执行到MainG时，如果MainG结束，则主线程也会随之结束
	save(getcallerpc(), getcallersp())
	asminit()  //在AMD64 Linux平台中，这个函数什么也没做，是个空函数
	minit()    //与信号相关的初始化，目前不需要关心

	// Install signal handlers; after minit so that minit can
	// prepare the thread to be able to handle the signals.
	if _g_.m == &m0 { //启动时_g_.m是m0，所以会执行下面的mstartm0函数
		mstartm0() //也是信号相关的初始化，现在我们不关注
	}

	if fn := _g_.m.mstartfn; fn != nil { //初始化过程中fn == nil
		fn()
	}

	if _g_.m != &m0 {// m0已经绑定了allp[0]，不是m0的话还没有p，所以需要获取一个p
		acquirep(_g_.m.nextp.ptr())
		_g_.m.nextp = 0
	}
    
        //schedule函数永远不会返回
	schedule()
}

// save updates getg().sched to refer to pc and sp so that a following
// gogo will restore pc and sp.
//
// save must not have write barriers because invoking a write barrier
// can clobber getg().sched.
//
//go:nosplit
//go:nowritebarrierrec
func save(pc, sp uintptr) {
	_g_ := getg()

	_g_.sched.pc = pc //再次运行时的指令地址
	_g_.sched.sp = sp //再次运行时到栈顶
	_g_.sched.lr = 0
	_g_.sched.ret = 0
	_g_.sched.g = guintptr(unsafe.Pointer(_g_))
	// We need to ensure ctxt is zero, but can't have a write
	// barrier here. However, it should always already be zero.
	// Assert that.
	if _g_.sched.ctxt != nil {
		badctxt()
	}
}
```

##### schedule()

- 检查m上的locks，检查是否在执行cgo

```go

// One round of scheduler: find a runnable goroutine and execute it.
// Never returns.
func schedule() {
	_g_ := getg()

	if _g_.m.locks != 0 {
		throw("schedule: holding locks")
	}

	if _g_.m.lockedg != 0 {
        // // Stops execution of the current m that is locked to a g until the g is runnable again.
		stoplockedm()
		execute(_g_.m.lockedg.ptr(), false) // Never returns.
	}

	// We should not schedule away from a g that is executing a cgo call,
	// since the cgo call is using the m's g0 stack.
	if _g_.m.incgo {
		throw("schedule: in cgo")
	}

top:
	pp := _g_.m.p.ptr()
	pp.preempt = false

	// Safety check: if we are spinning, the run queue should be empty.
	// Check this before calling checkTimers, as that might call
	// goready to put a ready goroutine on the local run queue.
	if _g_.m.spinning && (pp.runnext != 0 || pp.runqhead != pp.runqtail) {
		throw("schedule: spinning with local work")
	}

	gp, inheritTime, tryWakeP := findRunnable() // blocks until work is available

	// This thread is going to run a goroutine and is not spinning anymore,
	// so if it was marked as spinning we need to reset it now and potentially
	// start a new spinning M.
	if _g_.m.spinning {
		resetspinning()
	}

    // 如果当前的g被用户停止调度，并且不是runtime系统G，将会挂到disable.runnable下
	if sched.disable.user && !schedEnabled(gp) {
		// Scheduling of this goroutine is disabled. Put it on
		// the list of pending runnable goroutines for when we
		// re-enable user scheduling and look again.
		lock(&sched.lock)
		if schedEnabled(gp) {
			// Something re-enabled scheduling while we
			// were acquiring the lock.
			unlock(&sched.lock)
		} else {
			sched.disable.runnable.pushBack(gp)
			sched.disable.n++
			unlock(&sched.lock)
			goto top
		}
	}

	// If about to schedule a not-normal goroutine (a GCworker or tracereader),
	// wake a P if there is one.
	if tryWakeP {
		wakep()
	}
    // 防止多个m竞争处理一个g，上交该p给对应的lockedm
	if gp.lockedm != 0 {
		// Hands off own p to the locked m,
		// then blocks waiting for a new p.
		startlockedm(gp)
		goto top
	}
    //当前运行的是runtime的代码，函数调用栈使用的是g0的栈空间
    //调用execte切换到gp的代码和栈空间去运行
	execute(gp, inheritTime)
}



```

##### findrunable()

```GO
// Finds a runnable goroutine to execute.
// Tries to steal from other P's, get g from local or global queue, poll network.
// tryWakeP indicates that the returned goroutine is not normal (GC worker, trace
// reader) so the caller should try to wake a P.
func findRunnable() (gp *g, inheritTime, tryWakeP bool) {
	_g_ := getg()

	// The conditions here and in handoffp must agree: if
	// findrunnable would return a G to run, handoffp must start
	// an M.

top:
	_p_ := _g_.m.p.ptr()
	if sched.gcwaiting != 0 {
		gcstopm()
		goto top
	}
	if _p_.runSafePointFn != 0 {
		runSafePointFn()
	}

	// now and pollUntil are saved for work stealing later,
	// which may steal timers. It's important that between now
	// and then, nothing blocks, so these numbers remain mostly
	// relevant.
    // 返回当前时间， pollUntil 收包的最小结束时间
	now, pollUntil, _ := checkTimers(_p_, 0)

	// Try to schedule the trace reader.
	if trace.enabled || trace.shutdown {
		gp = traceReader()
		if gp != nil {
			casgstatus(gp, _Gwaiting, _Grunnable)
			traceGoUnpark(gp, 0)
			return gp, false, true
		}
	}

	// Try to schedule a GC worker.
	if gcBlackenEnabled != 0 {
		gp, now = gcController.findRunnableGCWorker(_p_, now)
		if gp != nil {
			return gp, false, true
		}
	}

    // 当本地队列处理了多次G后，偶尔去处理下全局队列上的G
	// Check the global runnable queue once in a while to ensure fairness.
	// Otherwise two goroutines can completely occupy the local runqueue
	// by constantly respawning each other.
	if _p_.schedtick%61 == 0 && sched.runqsize > 0 {
        lock(&sched.lock)
        gp = globrunqget(_p_, 1)
        unlock(&sched.lock)
        if gp != nil {
            return gp, false, false
        }
	}

	// Wake up the finalizer G.
	if fingwait && fingwake {
		if gp := wakefing(); gp != nil {
			ready(gp, 0, true)
		}
	}
	if *cgo_yield != nil {
		asmcgocall(*cgo_yield, nil)
	}

    // 从本地队列里面拿可运行的G
	// local runq
	if gp, inheritTime := runqget(_p_); gp != nil {
		return gp, inheritTime, false
	}
    // 从全局队列里面拿可运行的G
	// global runq
	if sched.runqsize != 0 {
		lock(&sched.lock)
		gp := globrunqget(_p_, 0)
		unlock(&sched.lock)
		if gp != nil {
			return gp, false, false
		}
	}

    // golang 高效使用Epoll的原因之一，当golang调度过程中如果发现网络Epoll上没有准备好，不会发生阻塞，这和操作系统上的Epoll是有区别的
    
	// Poll network.
	// This netpoll is only an optimization before we resort to stealing.
	// We can safely skip it if there are no waiters or a thread is blocked
	// in netpoll already. If there is any kind of logical race with that
	// blocked thread (e.g. it has already returned from netpoll, but does
	// not set lastpoll yet), this thread will do blocking netpoll below
	// anyway.
	if netpollinited() && atomic.Load(&netpollWaiters) > 0 && atomic.Load64(&sched.lastpoll) != 0 {
		if list := netpoll(0); !list.empty() { // non-blocking
			gp := list.pop()
			injectglist(&list)
			casgstatus(gp, _Gwaiting, _Grunnable)
			if trace.enabled {
				traceGoUnpark(gp, 0)
			}
			return gp, false, false
		}
	}

    // 偷窃过程的开始，每次偷窃Busy P上一般的Gs
	// Spinning Ms: steal work from other Ps.
	//
	// Limit the number of spinning Ms to half the number of busy Ps.
	// This is necessary to prevent excessive CPU consumption when
	// GOMAXPROCS>>1 but the program parallelism is low.
	procs := uint32(gomaxprocs)
    
    // 当m出现spinning状态或 调度器空闲P队列数量大于spinningm的两倍，则触发偷窃
    // 该偷窃是自旋的，第二个条件保证了偷窃行为不会消耗过多CPU
	if _g_.m.spinning || 2*atomic.Load(&sched.nmspinning) < procs-atomic.Load(&sched.npidle) {
		if !_g_.m.spinning {
			_g_.m.spinning = true
			atomic.Xadd(&sched.nmspinning, 1)
		}

		gp, inheritTime, tnow, w, newWork := stealWork(now)
		now = tnow
        // 偷到可执行的G
		if gp != nil {
			// Successfully stole.
			return gp, inheritTime, false
		}
        // 当前出现gc，进入下个循环然后等待
		if newWork {
			// There may be new timer or GC work; restart to
			// discover.
			goto top
		}
		if w != 0 && (pollUntil == 0 || w < pollUntil) {
			// Earlier timer to wait for.
			pollUntil = w
		}
	}

	// We have nothing to do.
	//
	// If we're in the GC mark phase, can safely scan and blacken objects,
	// and have work to do, run idle-time marking rather than give up the P.
	if gcBlackenEnabled != 0 && gcMarkWorkAvailable(_p_) && gcController.addIdleMarkWorker() {
		node := (*gcBgMarkWorkerNode)(gcBgMarkWorkerPool.pop())
		if node != nil {
			_p_.gcMarkWorkerMode = gcMarkWorkerIdleMode
			gp := node.gp.ptr()
			casgstatus(gp, _Gwaiting, _Grunnable)
			if trace.enabled {
				traceGoUnpark(gp, 0)
			}
			return gp, false, false
		}
		gcController.removeIdleMarkWorker()
	}

	// wasm only:
	// If a callback returned and no other goroutine is awake,
	// then wake event handler goroutine which pauses execution
	// until a callback was triggered.
	gp, otherReady := beforeIdle(now, pollUntil)
	if gp != nil {
		casgstatus(gp, _Gwaiting, _Grunnable)
		if trace.enabled {
			traceGoUnpark(gp, 0)
		}
		return gp, false, false
	}
	if otherReady {
		goto top
	}

    // 当前并没有窃取到任何的G，并且再次查询全局队列也没有发现可以执行的G，则会将当前P设置为IDLE并挂起
	// Before we drop our P, make a snapshot of the allp slice,
	// which can change underfoot once we no longer block
	// safe-points. We don't need to snapshot the contents because
	// everything up to cap(allp) is immutable.
	allpSnapshot := allp
	// Also snapshot masks. Value changes are OK, but we can't allow
	// len to change out from under us.
	idlepMaskSnapshot := idlepMask
	timerpMaskSnapshot := timerpMask

	// return P and block
	lock(&sched.lock)
	if sched.gcwaiting != 0 || _p_.runSafePointFn != 0 {
		unlock(&sched.lock)
		goto top
	}
	if sched.runqsize != 0 {
		gp := globrunqget(_p_, 0)
		unlock(&sched.lock)
		return gp, false, false
	}
	if releasep() != _p_ {
		throw("findrunnable: wrong p")
	}
	now = pidleput(_p_, now)
	unlock(&sched.lock)

	// Delicate dance: thread transitions from spinning to non-spinning
	// state, potentially concurrently with submission of new work. We must
	// drop nmspinning first and then check all sources again (with
	// #StoreLoad memory barrier in between). If we do it the other way
	// around, another thread can submit work after we've checked all
	// sources but before we drop nmspinning; as a result nobody will
	// unpark a thread to run the work.
	//
	// This applies to the following sources of work:
	//
	// * Goroutines added to a per-P run queue.
	// * New/modified-earlier timers on a per-P timer heap.
	// * Idle-priority GC work (barring golang.org/issue/19112).
	//
	// If we discover new work below, we need to restore m.spinning as a signal
	// for resetspinning to unpark a new worker thread (because there can be more
	// than one starving goroutine). However, if after discovering new work
	// we also observe no idle Ps it is OK to skip unparking a new worker
	// thread: the system is fully loaded so no spinning threads are required.
	// Also see "Worker thread parking/unparking" comment at the top of the file.
	// 执行到这里，需要确认全局P队列的快照中没有需要执行的任务，还有不存在有IDLEP的情况下，需要进行垃圾回收的任务，则可以将m设置为non-spinning 并在之后stop掉
    wasSpinning := _g_.m.spinning
	if _g_.m.spinning {
		_g_.m.spinning = false
		if int32(atomic.Xadd(&sched.nmspinning, -1)) < 0 {
			throw("findrunnable: negative nmspinning")
		}

		// Note the for correctness, only the last M transitioning from
		// spinning to non-spinning must perform these rechecks to
		// ensure no missed work. We are performing it on every M that
		// transitions as a conservative change to monitor effects on
		// latency. See golang.org/issue/43997.

		// Check all runqueues once again.
		_p_ = checkRunqsNoP(allpSnapshot, idlepMaskSnapshot)
		if _p_ != nil {
			acquirep(_p_)
			_g_.m.spinning = true
			atomic.Xadd(&sched.nmspinning, 1)
			goto top
		}

		// Check for idle-priority GC work again.
		_p_, gp = checkIdleGCNoP()
		if _p_ != nil {
			acquirep(_p_)
			_g_.m.spinning = true
			atomic.Xadd(&sched.nmspinning, 1)

			// Run the idle worker.
			_p_.gcMarkWorkerMode = gcMarkWorkerIdleMode
			casgstatus(gp, _Gwaiting, _Grunnable)
			if trace.enabled {
				traceGoUnpark(gp, 0)
			}
			return gp, false, false
		}

		// Finally, check for timer creation or expiry concurrently with
		// transitioning from spinning to non-spinning.
		//
		// Note that we cannot use checkTimers here because it calls
		// adjusttimers which may need to allocate memory, and that isn't
		// allowed when we don't have an active P.
		pollUntil = checkTimersNoP(allpSnapshot, timerpMaskSnapshot, pollUntil)
	}

	// Poll network until next timer.
	if netpollinited() && (atomic.Load(&netpollWaiters) > 0 || pollUntil != 0) && atomic.Xchg64(&sched.lastpoll, 0) != 0 {
		atomic.Store64(&sched.pollUntil, uint64(pollUntil))
		if _g_.m.p != 0 {
			throw("findrunnable: netpoll with p")
		}
		if _g_.m.spinning {
			throw("findrunnable: netpoll with spinning")
		}
		// Refresh now.
		now = nanotime()
		delay := int64(-1)
		if pollUntil != 0 {
			delay = pollUntil - now
			if delay < 0 {
				delay = 0
			}
		}
		if faketime != 0 {
			// When using fake time, just poll.
			delay = 0
		}
        // 在这里的epoll还是会阻塞的，在这里
		list := netpoll(delay) // block until new work is available
		atomic.Store64(&sched.pollUntil, 0)
		atomic.Store64(&sched.lastpoll, uint64(now))
		if faketime != 0 && list.empty() {
			// Using fake time and nothing is ready; stop M.
			// When all M's stop, checkdead will call timejump.
			stopm()
			goto top
		}
		lock(&sched.lock)
		_p_, _ = pidleget(now)
		unlock(&sched.lock)
		if _p_ == nil {
            // 会将采集的到Gs，如果当前P已经被挂起，通过检查当前的Idle全局P，将所有Gs放入全局可运行G列表中，去启动相应数量的M，返回。，如果当前P没被挂起，则将idle数量的G传到全局可运行G列表，其他的在当前P上允许
			injectglist(&list)
		} else {
            // 
			acquirep(_p_)
			if !list.empty() {
				gp := list.pop()
				injectglist(&list)
				casgstatus(gp, _Gwaiting, _Grunnable)
				if trace.enabled {
					traceGoUnpark(gp, 0)
				}
				return gp, false, false
			}
			if wasSpinning {
				_g_.m.spinning = true
				atomic.Xadd(&sched.nmspinning, 1)
			}
			goto top
		}
	} else if pollUntil != 0 && netpollinited() {
		pollerPollUntil := int64(atomic.Load64(&sched.pollUntil))
		if pollerPollUntil == 0 || pollerPollUntil > pollUntil {
			netpollBreak()
		}
	}
    // stopm的核心是调用mput把m结构体对象放入sched的midle空闲队列，然后通过notesleep(&m.park)函数让自己进入睡眠状态。
	stopm()
	goto top
}


// Get g from local runnable queue.
// If inheritTime is true, gp should inherit the remaining time in the
// current time slice. Otherwise, it should start a new time slice.
// Executed only by the owner P.
func runqget(_p_ *p) (gp *g, inheritTime bool) {
	// If there's a runnext, it's the next G to run.
	next := _p_.runnext
	// If the runnext is non-0 and the CAS fails, it could only have been stolen by another P,
	// because other Ps can race to set runnext to 0, but only the current P can set it to non-0.
	// Hence, there's no need to retry this CAS if it falls.
	if next != 0 && _p_.runnext.cas(next, 0) {
		return next.ptr(), true
	}

	for {
		h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with other consumers
		t := _p_.runqtail
		if t == h {
			return nil, false
		}
		gp := _p_.runq[h%uint32(len(_p_.runq))].ptr()
		if atomic.CasRel(&_p_.runqhead, h, h+1) { // cas-release, commits consume
			return gp, false
		}
	}
}
```

##### stealwork()

```GO
// stealWork attempts to steal a runnable goroutine or timer from any P.
//
// If newWork is true, new work may have been readied.
//
// If now is not 0 it is the current time. stealWork returns the passed time or
// the current time if now was passed as 0.
func stealWork(now int64) (gp *g, inheritTime bool, rnow, pollUntil int64, newWork bool) {
	pp := getg().m.p.ptr()

	ranTimer := false

	const stealTries = 4
	for i := 0; i < stealTries; i++ {
        // 根据重试1次数决定要不要偷nextp
		stealTimersOrRunNextG := i == stealTries-1

		for enum := stealOrder.start(fastrand()); !enum.done(); enum.next() {
			if sched.gcwaiting != 0 {
				// GC work may be available.
				return nil, false, now, pollUntil, true
			}
            // 随机偷窃某个P队列
			p2 := allp[enum.position()]
			if pp == p2 {
				continue
			}

			// Steal timers from p2. This call to checkTimers is the only place
			// where we might hold a lock on a different P's timers. We do this
			// once on the last pass before checking runnext because stealing
			// from the other P's runnext should be the last resort, so if there
			// are timers to steal do that first.
			//
			// We only check timers on one of the stealing iterations because
			// the time stored in now doesn't change in this loop and checking
			// the timers for each P more than once with the same value of now
			// is probably a waste of time.
			//
			// timerpMask tells us whether the P may have timers at all. If it
			// can't, no need to check at all.
			if stealTimersOrRunNextG && timerpMask.read(enum.position()) {
				tnow, w, ran := checkTimers(p2, now)
				now = tnow
				if w != 0 && (pollUntil == 0 || w < pollUntil) {
					pollUntil = w
				}
				if ran {
					// Running the timers may have
					// made an arbitrary number of G's
					// ready and added them to this P's
					// local run queue. That invalidates
					// the assumption of runqsteal
					// that it always has room to add
					// stolen G's. So check now if there
					// is a local G to run.
					if gp, inheritTime := runqget(pp); gp != nil {
						return gp, inheritTime, now, pollUntil, ranTimer
					}
					ranTimer = true
				}
			}

			// Don't bother to attempt to steal if p2 is idle.
			if !idlepMask.read(enum.position()) {
				if gp := runqsteal(pp, p2, stealTimersOrRunNextG); gp != nil {
					return gp, false, now, pollUntil, ranTimer
				}
			}
		}
	}

	// No goroutines found to steal. Regardless, running a timer may have
	// made some goroutine ready that we missed. Indicate the next timer to
	// wait for.
	return nil, false, now, pollUntil, ranTimer
}

// Steal half of elements from local runnable queue of p2
// and put onto local runnable queue of p.
// Returns one of the stolen elements (or nil if failed).
func runqsteal(_p_, p2 *p, stealRunNextG bool) *g {
	t := _p_.runqtail
	n := runqgrab(p2, &_p_.runq, t, stealRunNextG)
	if n == 0 {
		return nil
	}
	n--
	gp := _p_.runq[(t+n)%uint32(len(_p_.runq))].ptr()
	if n == 0 {
		return gp
	}
	h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with consumers
	if t-h+n >= uint32(len(_p_.runq)) {
		throw("runqsteal: runq overflow")
	}
	atomic.StoreRel(&_p_.runqtail, t+n) // store-release, makes the item available for consumption
	return gp
}


// Grabs a batch of goroutines from _p_'s runnable queue into batch.
// Batch is a ring buffer starting at batchHead.
// Returns number of grabbed goroutines.
// Can be executed by any P.
func runqgrab(_p_ *p, batch *[256]guintptr, batchHead uint32, stealRunNextG bool) uint32 {
	for {
		h := atomic.LoadAcq(&_p_.runqhead) // load-acquire, synchronize with other consumers
		t := atomic.LoadAcq(&_p_.runqtail) // load-acquire, synchronize with the producer
		n := t - h
		n = n - n/2
		if n == 0 {
            // 当P队列上没得偷时，考虑runnext，非常粗暴的让他休息，然后偷走
			if stealRunNextG {
				// Try to steal from _p_.runnext.
				if next := _p_.runnext; next != 0 {
					if _p_.status == _Prunning {
						// Sleep to ensure that _p_ isn't about to run the g
						// we are about to steal.
						// The important use case here is when the g running
						// on _p_ ready()s another g and then almost
						// immediately blocks. Instead of stealing runnext
						// in this window, back off to give _p_ a chance to
						// schedule runnext. This will avoid thrashing gs
						// between different Ps.
						// A sync chan send/recv takes ~50ns as of time of
						// writing, so 3us gives ~50x overshoot.
						if GOOS != "windows" && GOOS != "openbsd" && GOOS != "netbsd" {
							usleep(3)
						} else {
							// On some platforms system timer granularity is
							// 1-15ms, which is way too much for this
							// optimization. So just yield.
							osyield()
						}
					}
					if !_p_.runnext.cas(next, 0) {
						continue
					}
					batch[batchHead%uint32(len(batch))] = next
					return 1
				}
			}
			return 0
		}
        // 当我们读取runqhead之后但还未读取runqtail之前，如果有其它线程快速的在增加（这是完全有可能的，其它偷取者从队列中偷取goroutine会增加runqhead，而队列的所有者往队列中添加goroutine会增加runqtail）这两个值，则会导致我们读取出来的runqtail已经远远大于我们之前读取出来放在局部变量h里面的runqhead了
		if n > uint32(len(_p_.runq)/2) { // read inconsistent h and t
			continue
		}
		for i := uint32(0); i < n; i++ {
			g := _p_.runq[(h+i)%uint32(len(_p_.runq))]
			batch[(batchHead+i)%uint32(len(batch))] = g
		}
        // 偷窃当前队列中前半部分Gs
		if atomic.CasRel(&_p_.runqhead, h, h+n) { // cas-release, commits consume
			return n
		}
	}
}
```

##### execute()

```go
// Schedules gp to run on the current M.
// If inheritTime is true, gp inherits the remaining time in the
// current time slice. Otherwise, it starts a new time slice.
// Never returns.
//
// Write barriers are allowed because this is called immediately after
// acquiring a P in several places.
//
//go:yeswritebarrierrec
func execute(gp *g, inheritTime bool) {
	_g_ := getg() //g0

	if goroutineProfile.active {
		// Make sure that gp has had its stack written out to the goroutine
		// profile, exactly as it was when the goroutine profiler first stopped
		// the world.
		tryRecordGoroutineProfile(gp, osyield)
	}

	// Assign gp.m before entering _Grunning so running Gs have an
	// M.
     //把g和m关联起来，这样通过m就可以找到当前工作线程正在执行哪个goroutine
	_g_.m.curg = gp
	gp.m = _g_.m 
	casgstatus(gp, _Grunnable, _Grunning)
	gp.waitsince = 0
	gp.preempt = false
	gp.stackguard0 = gp.stack.lo + _StackGuard
	if !inheritTime {
		_g_.m.p.ptr().schedtick++
	}

	// Check whether the profiler needs to be turned on or off.
	hz := sched.profilehz
	if _g_.m.profilehz != hz {
		setThreadCPUProfiler(hz)
	}

	if trace.enabled {
		// GoSysExit has to happen when we have a P, but before GoStart.
		// So we emit it here.
		if gp.syscallsp != 0 && gp.sysblocktraced {
			traceGoSysExit(gp.sysexitticks)
		}
		traceGoStart()
	}
	//gogo完成从g0到gp真正的切换
	gogo(&gp.sched)
}
```

##### gogo

```shell
# func gogo(buf *gobuf)
# restore state from Gobuf; longjmp
TEXT runtime·gogo(SB), NOSPLIT, $16-8
	#buf = &gp.sched
	// 使用BX寄存器存储调度信息
	MOVQ	buf+0(FP), BX		# BX = buf
	
	#gobuf->g --> dx register
	// 把gp.sched.g读取到DX寄存器，注意这条指令的源操作数是间接寻址
	MOVQ	gobuf_g(BX), DX  # DX = gp.sched.g
	
	#下面这行代码没有实质作用，检查gp.sched.g是否是nil，如果是nil进程会crash死掉
	MOVQ	0(DX), CX		# make sure g != nil
	
	get_tls(CX) 
	
	#把要运行的g的指针放入线程本地存储，这样后面的代码就可以通过线程本地存储
	#获取到当前正在执行的goroutine的g结构体对象，从而找到与之关联的m和p
	MOVQ	DX, g(CX)
	
	#把CPU的SP寄存器设置为sched.sp，完成了栈的切换
	MOVQ	gobuf_sp(BX), SP	# restore SP
	
	#下面三条同样是恢复调度上下文到CPU相关寄存器
	MOVQ	gobuf_ret(BX), AX  #系统调用的返回值放入AX寄存器
	MOVQ	gobuf_ctxt(BX), DX
	MOVQ	gobuf_bp(BX), BP
	
	#清空sched的值，因为我们已把相关值放入CPU对应的寄存器了，不再需要，这样做可以少gc的工作量
	MOVQ	$0, gobuf_sp(BX)	# clear to help garbage collector
	MOVQ	$0, gobuf_ret(BX)
	MOVQ	$0, gobuf_ctxt(BX)
	MOVQ	$0, gobuf_bp(BX)
	
	#把sched.pc值放入BX寄存器,因为上面已经把调度信息从内存转移到寄存器了
	// gobuf_pc(BX) 是gp 的函数入口 runtime.main
	MOVQ	gobuf_pc(BX), BX
	
	#JMP把BX寄存器的包含的地址值放入CPU的IP寄存器，于是，CPU跳转到该地址继续执行指令，
	JMP	BX
```

##### runtime.main

```go
// The main goroutine.
func main() {
	g := getg() // 此时获取的是GP

	// Racectx of m0->g0 is used only as the parent of the main goroutine.
	// It must not be used for anything else.
	g.m.g0.racectx = 0

	// Max stack size is 1 GB on 64-bit, 250 MB on 32-bit.
	// Using decimal instead of binary GB and MB because
	// they look nicer in the stack overflow failure message.
	if goarch.PtrSize == 8 {
		maxstacksize = 1000000000
	} else {
		maxstacksize = 250000000
	}

	// An upper limit for max stack size. Used to avoid random crashes
	// after calling SetMaxStack and trying to allocate a stack that is too big,
	// since stackalloc works with 32-bit sizes.
	maxstackceiling = 2 * maxstacksize

	// Allow newproc to start new Ms.
	mainStarted = true

	if GOARCH != "wasm" { // no threads on wasm yet, so no sysmon
		systemstack(func() {
			newm(sysmon, nil, -1)
		})
	}

	// Lock the main goroutine onto this, the main OS thread,
	// during initialization. Most programs won't care, but a few
	// do require certain calls to be made by the main thread.
	// Those can arrange for main.main to run in the main thread
	// by calling runtime.LockOSThread during initialization
	// to preserve the lock.
	lockOSThread()

	if g.m != &m0 {
		throw("runtime.main not on m0")
	}

	// Record when the world started.
	// Must be before doInit for tracing init.
	runtimeInitTime = nanotime()
	if runtimeInitTime == 0 {
		throw("nanotime returning zero")
	}

	if debug.inittrace != 0 {
		inittrace.id = getg().goid
		inittrace.active = true
	}
    //调用runtime包的初始化函数，由编译器实现
    // 启动一个sysmon系统监控线程，该线程负责整个程序的gc、抢占调度以及netpoll等功能的监控，在调度阶段我们再继续分析sysmon是如何协助完成goroutine的抢占调度的；
    //执行runtime包的初始化；
	doInit(&runtime_inittask) // Must be before defer.

	// Defer unlock so that runtime.Goexit during init does the unlock too.
	needUnlock := true
	defer func() {
		if needUnlock {
			unlockOSThread()
		}
	}()

	gcenable()

	main_init_done = make(chan bool)
	if iscgo {
		if _cgo_thread_start == nil {
			throw("_cgo_thread_start missing")
		}
		if GOOS != "windows" {
			if _cgo_setenv == nil {
				throw("_cgo_setenv missing")
			}
			if _cgo_unsetenv == nil {
				throw("_cgo_unsetenv missing")
			}
		}
		if _cgo_notify_runtime_init_done == nil {
			throw("_cgo_notify_runtime_init_done missing")
		}
		// Start the template thread in case we enter Go from
		// a C-created thread and need to create a new thread.
		startTemplateThread()
		cgocall(_cgo_notify_runtime_init_done, nil)
	}
	 //main 包的初始化函数，也是由编译器实现，会递归的调用我们import进来的包的初始化函数
	doInit(&main_inittask)

	// Disable init tracing after main init done to avoid overhead
	// of collecting statistics in malloc and newproc
	inittrace.active = false

	close(main_init_done)

	needUnlock = false
	unlockOSThread()

	if isarchive || islibrary {
		// A program compiled with -buildmode=c-archive or c-shared
		// has a main, but it is not executed.
		return
	}
	//调用main.main函数
	fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
	fn()
	if raceenabled {
		racefini()
	}

	// Make racy client program work: if panicking on
	// another goroutine at the same time as main returns,
	// let the other goroutine finish printing the panic trace.
	// Once it does, it will exit. See issues 3934 and 20018.
	if atomic.Load(&runningPanicDefers) != 0 {
		// Running deferred functions should not take long.
		for c := 0; c < 1000; c++ {
			if atomic.Load(&runningPanicDefers) == 0 {
				break
			}
			Gosched()
		}
	}
	if atomic.Load(&panicking) != 0 {
		gopark(nil, nil, waitReasonPanicWait, traceEvGoStop, 1)
	}
	//进入系统调用，退出进程，可以看出main goroutine并未返回，而是直接进入系统调用退出进程了
    //可是从前面的分析中我们得知，在创建goroutine的时候已经在其栈上放好了一个返回地址，伪造成goexit函数调用了goroutine的入口函数，这里怎么没有用到这个返回地址啊？其实那是为非main goroutine准备的，非main goroutine执行完成后就会返回到goexit继续执行，而main goroutine执行完成后整个进程就结束了，这是main goroutine与其它goroutine的一个区别。
	exit(0)
	for {
		var x *int32
		*x = 0
	}
}
```

##### sysmon.retake （抢占式调度）

```GO
//sysmon会根据系统当前的繁忙程度睡一小段时间，然后每隔10ms至少进行一次epoll并唤醒相应的goroutine。同时，它还会检测是否有P长时间处于Psyscall状态或Prunning状态，并进行抢占式调度。
for(;;) {
    runtime.usleep(delay);
    if(lastpoll != 0 && lastpoll + 10*1000*1000 > now) {
        runtime.netpoll();
    }
    retake(now);    // 根据每个P的状态和运行时间决定是否要进行抢占
}

```



#### 小结

一次循环调度

```
schedule()->execute()->gogo()->g2()->goexit()->goexit1()->mcall()->goexit0()->schedule()  
```



分为四个阶段

1. 第一个阶段进行操作系统的信息初始化，调用OsInit，如读取CPU的核心数
2. 第二个阶段是调度相关的初始化,调用SchedInit，一方面是调度器schedt中lock的初始化，另一方面是将m0挂载到runtime中的allm列表中，还有根据cpu核心数创建相应的p，将其分别放入allp和idlep中，并将p[0] 与 m0 进行关联
3. 第三个阶段，通过汇编函数MainPc将runtime.main的函数入口地址传给newProc，然后调用NewProc将runtime.main和参数一起包装为main-goroutine（函数指针地址为run_time.main），设置完CPU调度信息以及G的状态信息后，丢到可执行G队列里等待调度。
4. 第四个阶段，通过mstart正式启动m0，通过schedule函数进入循环调度流程，此时会读取到我们之前的main函数Goroutine，在gogo函数中从g0切换到main-goroutine的上下文，在runtime.main中初始化sysmon与runtime与import包相关的函数，最终执行我们main.main的函数入口





### 调度阶段





##### 正常运行时：

- 执行findrunable函数找到可执行的G执行

##### 被动调度：

###### goparkunlock->gopark->mcall

函数直接调用gopark函数，gopark则调用mcall从当前main goroutine切换到g0去执行park_m函数（mcall前面我们分析过，其主要作用就是保存当前goroutine的现场，然后切换到g0栈去调用作为参数传递给它的函数）

```go
// Puts the current goroutine into a waiting state and unlocks the lock.
// The goroutine can be made runnable again by calling goready(gp).
func goparkunlock(lock*mutex, reasonwaitReason, traceEvbyte, traceskipint) {
    gopark(parkunlock_c, unsafe.Pointer(lock), reason, traceEv, traceskip)
}
 
// runtime/proc.go : 276
// Puts the current goroutine into a waiting state and calls unlockf.
// If unlockf returns false, the goroutine is resumed.
// unlockf must not access this G's stack, as it may be moved between
// the call to gopark and the call to unlockf.
// Reason explains why the goroutine has been parked.
// It is displayed in stack traces and heap dumps.
// Reasons should be unique and descriptive.
// Do not re-use reasons, add new ones.
func gopark(unlockffunc(*g, unsafe.Pointer) bool, lockunsafe.Pointer, reason    waitReason, traceEvbyte, traceskipint) {
    ......
    // can't do anything that might move the G between Ms here.
    mcall(park_m) //切换到g0栈执行park_m函数
}
```

###### park_m & ready （g0上运行）

```go
// park continuation on g0.
func park_m(gp*g) {
    _g_ := getg()
 
    if trace.enabled {
        traceGoPark(_g_.m.waittraceev, _g_.m.waittraceskip)
    }
	// park_m首先把当前goroutine的状态设置为_Gwaiting
    casgstatus(gp, _Grunning, _Gwaiting)
    dropg()  //解除g和m之间的关系
 
    ......
   
    schedule()
}

func dropg() {
	_g_ := getg()

	setMNoWB(&_g_.m.curg.m, nil)
	setGNoWB(&_g_.m.curg, nil)
}


// Mark gp ready to run.
func ready(gp *g, traceskip int, next bool) {
    ......
    // Mark runnable.
    _g_ := getg()
    ......
    // status is Gwaiting or Gscanwaiting, make Grunnable and put on runq
    casgstatus(gp, _Gwaiting, _Grunnable)
    runqput(_g_.m.p.ptr(), gp, next) //放入运行队列
    if atomic.Load(&sched.npidle) != 0 && atomic.Load(&sched.nmspinning) == 0 {
        //有空闲的p而且没有正在偷取goroutine的工作线程，则需要唤醒p出来工作
        wakep()
    }
    ......
}

// Tries to add one more P to execute G's.
// Called when a G is made runnable (newproc, ready).
// wakep首先通过cas操作再次确认是否有其它工作线程正处于spinning状态，这里之所以需要使用cas操作再次进行确认，原因在于，在当前工作线程通过如下条件
func wakep() {
    // be conservative about spinning threads
    if !atomic.Cas(&sched.nmspinning, 0, 1)  {
        return
    }
    // 调用startm创建一个新的或唤醒一个处于睡眠状态的工作线程出来工作。
    startm(nil, true)
}

// Schedules some M to run the p (creates an M if necessary).
// If p==nil, tries to get an idle P, if no idle P's does nothing.
// May run with m.p==nil, so write barriers are not allowed.
// If spinning is set, the caller has incremented nmspinning and startm will
// either decrement nmspinning or set m.spinning in the newly started M.
//go:nowritebarrierrec
func startm(_p_ *p, spinning bool)  {
    lock(&sched.lock)
    if _p_ == nil { //没有指定p的话需要从p的空闲队列中获取一个p
        _p_ = pidleget() //从p的空闲队列中获取空闲p
        if _p_ == nil {
            unlock(&sched.lock)
            if spinning {
                // The caller incremented nmspinning, but there are no idle Ps,
                // so it's okay to just undo the increment and give up.
                //spinning为true表示进入这个函数之前已经对sched.nmspinning加了1，需要还原
                if int32(atomic.Xadd(&sched.nmspinning, -1)) < 0 {
                    throw("startm: negative nmspinning")
                }
            }
            return//没有空闲的p，直接返回
        }
    }
    mp := mget() //从m空闲队列中获取正处于睡眠之中的工作线程，所有处于睡眠状态的m都在此队列中
    unlock(&sched.lock)
    if mp == nil {
       //没有处于睡眠状态的工作线程
        var fn func()
        if spinning {
            // The caller incremented nmspinning, so set m.spinning in the new M.
            fn = mspinning
        }
        newm(fn, _p_) //创建新的工作线程
        return
    }
    if mp.spinning {
        throw("startm: m is spinning")
    }
    if mp.nextp != 0 {
        throw("startm: m has p")
    }
    if spinning && !runqempty(_p_)  {
        throw("startm: p has runnable gs")
    }
    // The caller incremented nmspinning, so set m.spinning in the new M.
    mp.spinning = spinning
    mp.nextp.set(_p_)
   
    //唤醒处于休眠状态的工作线程
    notewakeup(&mp.park)
}
```

例子： 

##### 主动调度：

###### Gosched 

**Goroutine的主动调度是指当前正在运行的goroutine通过直接调用runtime.Gosched()函数暂时放弃运行而发生的调度**

```go
// Gosched yields the processor, allowing other goroutines to run. It does not
// suspend the current goroutine, so execution resumes automatically.
func Gosched() {
    checkTimeouts() //amd64 linux平台空函数
   
    // 切换到当前m的g0栈执行gosched_m函数
    // void mcall(fn func(*g))
    // Switch to m->g0's stack, call fn(g).
    // Fn must never return. It should gogo(&g->sched)
    // to keep running g.
    mcall(gosched_m)
    //再次被调度起来则从这里开始继续运行
}
// gosched_m函数只是简单的在调用goschedImpl：
func goschedImpl(gp *g) {
    ......
    casgstatus(gp, _Grunning, _Grunnable)
    dropg() //设置当前m.curg = nil, gp.m = nil
    lock(&sched.lock)
    globrunqput(gp) //把gp放入sched的全局运行队列runq
    unlock(&sched.lock)

    schedule() //进入新一轮调度
}
```





## 问题

### 在一个G中go func 新增的G 是否会在当前M的P队列上

是，会以一定几率替换next的值，并将替换前的g加入到队列中



### 一个G在执行完后会发生什么

设置状态为GDead 加入到的GIdle队列后，调用mcall（Exit） 函数进入g0后，调用Schedule（），进行下一次调度



### 当M的P中没有G可运行，会发生什么？

进入自旋状态，从local、global、net、其他P 获取可运行的G，如果没有则P进入Idle 挂起





### 当M发生阻塞系统调用会发生什么（chan -> park）？

 假定当前除了 M3 和 M4 为自旋线程，还有 M5 和 M6 为空闲的线程 (没有得到 P 的绑定，注意我们这里最多就只能够存在 4 个 P，所以 P 的数量应该永远是 M>=P, 大部分都是 M 在抢占需要运行的 P)，G8 创建了 G9，G8 进行了阻塞的系统调用，M2 和 P2 立即解绑，P2 会执行以下判断：如果 P2 本地队列有 G、全局队列有 G 或有空闲的 M，P2 都会立马唤醒 1 个 M 和它绑定，否则 P2 则会加入到空闲 P 列表，等待 M 来获取可用的 p。本场景中，P2 本地队列有 G9，可以和其他空闲的线程 M5 绑定。

![image-20220708014329821](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/img/202207080143866.png)



当M发生非阻塞系统调用会发生什么？



M2 和 P2 会解绑，但 M2 会记住 P2，然后 G8 和 M2 进入系统调用状态。当 G8 和 M2 退出系统调用时，会尝试获取 P2，如果无法获取，则获取空闲的 P，如果依然没有，G8 会被记为可运行状态，并加入到全局队列，M2 因为没有 P 的绑定而变成休眠状态 (长时间休眠等待 GC 回收销毁)。





### 为什么mcall和systemstack这种操作要切换到g0进行操作

mcall 和 systemstack 的调用的方法，都是涉及







#### MCALL

- 将当前G的sched信息存到寄存器中

- 将内存中的g0栈通过tls共享到寄存器的FS段中

- 只允许调用不会返回的函数，如Schedule() 进行下一次调度、或者发生阻塞，被挂在某个g的结构体上（chan），等待被ready唤醒。

- Mcall的意义在于将goroutine的栈和m的栈隔离，goroutine中的栈只做和业务有关的东西，m中的栈只做和调度有关的东西。

  ![image-20220708003919025](C:/Users/92175/AppData/Roaming/Typora/typora-user-images/image-20220708003919025.png)

```
TEXT runtime·mcall<ABIInternal>(SB), NOSPLIT, $0-8
	MOVQ	AX, DX	// DX = fn

	// save state in g->sched
	MOVQ	0(SP), BX	// caller's PC
	MOVQ	BX, (g_sched+gobuf_pc)(R14)
	LEAQ	fn+0(FP), BX	// caller's SP
	MOVQ	BX, (g_sched+gobuf_sp)(R14)
	MOVQ	BP, (g_sched+gobuf_bp)(R14)

	// switch to m->g0 & its stack, call fn
	MOVQ	g_m(R14), BX
	MOVQ	m_g0(BX), SI	// SI = g.m.g0
	CMPQ	SI, R14	// if g == m->g0 call badmcall
	JNE	goodm
	JMP	runtime·badmcall(SB)
goodm:
	MOVQ	R14, AX		// AX (and arg 0) = g
	MOVQ	SI, R14		// g = g.m.g0
	get_tls(CX)		// Set G in TLS
	MOVQ	R14, g(CX)
	MOVQ	(g_sched+gobuf_sp)(R14), SP	// sp = g0.sched.sp
	PUSHQ	AX	// open up space for fn's arg spill slot
	MOVQ	0(DX), R12
	CALL	R12		// fn(g)
	POPQ	AX
	JMP	runtime·badmcall2(SB)
	RET
```

#### SYSTEMSTACK

- 对于G0和Gsignal中的Systemstack调用可以直接执行对应函数，并返回原先的调用地址

```
// func systemstack(fn func())
TEXT runtime·systemstack(SB), NOSPLIT, $0-8
	MOVQ	fn+0(FP), DI	// DI = fn
	get_tls(CX)
	MOVQ	g(CX), AX	// AX = g
	MOVQ	g_m(AX), BX	// BX = m

	CMPQ	AX, m_gsignal(BX)
	JEQ	noswitch

	MOVQ	m_g0(BX), DX	// DX = g0
	CMPQ	AX, DX
	JEQ	noswitch

	CMPQ	AX, m_curg(BX)
	JNE	bad

	// switch stacks
	// save our state in g->sched. Pretend to
	// be systemstack_switch if the G stack is scanned.
	CALL	gosave_systemstack_switch<>(SB)

	// switch to g0
	MOVQ	DX, g(CX)
	MOVQ	DX, R14 // set the g register
	MOVQ	(g_sched+gobuf_sp)(DX), BX
	MOVQ	BX, SP

	// call target function
	MOVQ	DI, DX
	MOVQ	0(DI), DI
	CALL	DI

	// switch back to g
	get_tls(CX)
	MOVQ	g(CX), AX
	MOVQ	g_m(AX), BX
	MOVQ	m_curg(BX), AX
	MOVQ	AX, g(CX)
	MOVQ	(g_sched+gobuf_sp)(AX), SP
	MOVQ	$0, (g_sched+gobuf_sp)(AX)
	RET

noswitch:
	// already on m stack; tail call the function
	// Using a tail call here cleans up tracebacks since we won't stop
	// at an intermediate systemstack.
	MOVQ	DI, DX
	MOVQ	0(DI), DI
	JMP	DI

bad:
	// Bad: g is not gsignal, not g0, not curg. What is it?
	MOVQ	$runtime·badsystemstack(SB), AX
	CALL	AX
	INT	$3
```



### 新建m是个什么流程

#### newm的调用

```GO
systemstack(func() {
			sysmon, nil, -1)newm(
		})

newm(templateThread, nil, -1)

// Start M to run P.  Do not start another M below.
newm(nil, p, -1)
```

newm

```

//go:nosplit
func acquirem() *m {
	_g_ := getg()
	_g_.m.locks++
	return _g_.m
}

// Create a new m. It will start off with a call to fn, or else the scheduler.
// fn needs to be static and not a heap allocated closure.
// May run with m.p==nil, so write barriers are not allowed.
//
// id is optional pre-allocated m ID. Omit by passing -1.
//
//go:nowritebarrierrec
// 创建一个m 它将以对fn的调用开始
func newm(fn func(), _p_ *p, id int64) {
	// allocm adds a new M to allm, but they do not start until created by
	// the OS in newm1 or the template thread.
	//
	// doAllThreadsSyscall requires that every M in allm will eventually
	// start and be signal-able, even with a STW.
	//
	// Disable preemption here until we start the thread to ensure that
	// newm is not preempted between allocm and starting the new thread,
	// ensuring that anything added to allm is guaranteed to eventually
	// start.
	acquirem()

	// 创建一个没有和任何线程有关联的M
	// 其中会初始化M的G0变量，也是通过 malg -> mcommoninit 创建的和普通的g一样
	// 对比g0，g0初始化后会放在寄存器的FS段上，在初始化期间高校读取
	mp := allocm(_p_, fn, id)
	mp.nextp.set(_p_)
	mp.sigmask = initSigmask
	if gp := getg(); gp != nil && gp.m != nil && (gp.m.lockedExt != 0 || gp.m.incgo) && GOOS != "plan9" {
		// We're on a locked M or a thread that may have been
		// started by C. The kernel state of this thread may
		// be strange (the user may have locked it for that
		// purpose). We don't want to clone that into another
		// thread. Instead, ask a known-good thread to create
		// the thread for us.
		//
		// This is disabled on Plan 9. See golang.org/issue/22227.
		//
		// TODO: This may be unnecessary on Windows, which
		// doesn't model thread creation off fork.
		lock(&newmHandoff.lock)
		if newmHandoff.haveTemplateThread == 0 {
			throw("on a locked thread with no template thread")
		}
		mp.schedlink = newmHandoff.newm
		newmHandoff.newm.set(mp)
		if newmHandoff.waiting {
			newmHandoff.waiting = false
			notewakeup(&newmHandoff.wake)
		}
		unlock(&newmHandoff.lock)
		// The M has not started yet, but the template thread does not
		// participate in STW, so it will always process queued Ms and
		// it is safe to releasem.
		releasem(getg().m)
		return
	}
	newm1(mp)
	releasem(getg().m)
}

```

#### newm1

```go

func newm1(mp *m) {
	if iscgo {
		var ts cgothreadstart
		if _cgo_thread_start == nil {
			throw("_cgo_thread_start missing")
		}
		ts.g.set(mp.g0)
		ts.tls = (*uint64)(unsafe.Pointer(&mp.tls[0]))
		ts.fn = unsafe.Pointer(abi.FuncPCABI0(mstart))
		if msanenabled {
			msanwrite(unsafe.Pointer(&ts), unsafe.Sizeof(ts))
		}
		if asanenabled {
			asanwrite(unsafe.Pointer(&ts), unsafe.Sizeof(ts))
		}
		execLock.rlock() // Prevent process clone.
		asmcgocall(_cgo_thread_start, unsafe.Pointer(&ts))
		execLock.runlock()
		return
	}
	execLock.rlock() // Prevent process clone.
	newosproc(mp)
	execLock.runlock()
}
```

#### newosproc

```go
//go:nowritebarrier
func newosproc(mp *m) {
	stk := unsafe.Pointer(mp.g0.stack.hi)
	/*
	 * note: strace gets confused if we use CLONE_PTRACE here.
	 */
	if false {
		print("newosproc stk=", stk, " m=", mp, " g=", mp.g0, " clone=", abi.FuncPCABI0(clone), " id=", mp.id, " ostk=", &mp, "\n")
	}

	// Disable signals during clone, so that the new thread starts
	// with signals disabled. It will enable them in minit.
	var oset sigset
	sigprocmask(_SIG_SETMASK, &sigset_all, &oset)
    // 通过系统克隆创建线程，线程的任务函数为mstart,并将初始化好的m和g作为参数传递
	ret := clone(cloneFlags, stk, unsafe.Pointer(mp), unsafe.Pointer(mp.g0), unsafe.Pointer(abi.FuncPCABI0(mstart)))
	sigprocmask(_SIG_SETMASK, &oset, nil)

	if ret < 0 {
		print("runtime: failed to create new OS thread (have ", mcount(), " already; errno=", -ret, ")\n")
		if ret == -_EAGAIN {
			println("runtime: may need to increase max user processes (ulimit -u)")
		}
		throw("newosproc")
	}
}
```





### G0是分配在主线程栈上还是堆上的？

G0是分配在段寄存器FS位上，通过g0可以找到m0进而找到p0，当栈的发生上下文切换后，G0上的寄存器内容又会被放入内存中



### 为什么队列要分运行时G和自由G

把g放入p的freeg队列缓存起来供下次创建g时快速获取而不用从内存分配。freeg就是g的一个对象池；




### Golang 栈是从高往低增长的吗？

函数执行时候，需要有足够的内存空间，供他存放局部变量，参数等数据，这段空间对应到虚拟地址空间的栈。

分配给函数的栈空间被称为**函数栈帧**，Go 语言中函数栈帧布局是如下的，先是调用者栈基地址，然后是函数的局部变量，最后是被调用函数的返回值和参数。

![image-20220706024225764](https://typroa-pic-sh-1258186845.cos.ap-shanghai.myqcloud.com/img/202207060242935.png)

举个粒子：

```GO
func A() {
    var a1, a2, r1, r2 int64
    a1, a2 = 1, 2
    r1, r2 = B(a1, a2)
    r1 = C(a1)
    println(r1, r2)
}
func B(p1, p2 int64) (int64, int64) {
    return p2, p1
}
func C(p1 int64) int64 {
    return p1
}
```



#### 函数内部

函数 A 的栈帧分布就由上至下，分别是局部变量 `a1, a2, r1, r2` ，被调函数 B 的返回值 `r2,r1` ，被调用函数 B 的参数 `a2,a1`。

**注意观察参数的顺序，返回值和参数都是先入栈的第二个，然后再入栈的第一个，相当于是从右至左逐一入栈的。**

被调用函数是通过栈指针加上偏移量这样相对寻址的方式来定位自己的参数和返回值，刚好由下至上先找到第一个参数在找到第二个参数（通过**增加偏移量**的方式）。所以说参数和返回值采用**由右至左的入栈顺序**比较合适。


<!--通常，我们认为返回值是通过寄存器传递的，但是 Go 语言支持多返回值，所以在栈上分配返回值空间更合适-->

#### 函数相互调用

对于函数 B 的调用会被编译器编译成 `Call` 指令。`Call` 指令只做两件事情。

1. 将下一条指令的地址入栈，被调用函数结束后，跳回到该地址继续执行，这就是返回地址
2. 跳转到被调用函数的指令入口处执行，所以返回地址下面就是函数 B 的栈帧。

其余的部分会按照 A 函数布局一样布局。

当函数 B 执行结束后会释放栈帧，然后就到 `Ret` 指令了，`Ret` 指令也会干两件事。

1. 弹出`Call` 指令压栈的返回地址
2. 跳转到返回地址











  

