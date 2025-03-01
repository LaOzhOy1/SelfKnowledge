package scheduler

import "../engine"

// SimpleScheduler 在run方法（在主协程）中 通过 Submit （开启一个新的协程） 提交 到tasks channel中 多个worker 接受 task 返回 result 到 results中， 在run方法（在主协程）for循环监听返回来的结果
// 要点 ： 多个协程共用 results ， tasks channels, 为了防止阻塞，每次提交任务都会新开启一个协程传输，由于子协程只有传输任务，无需等待新任务，所以不会有死锁条件产生、
// 缺点： 协程开太多，切换协程获取的总体效率低
// 改进
type QueueScheduler struct {
	requestChan chan engine.Request
	// 每个worker有自己的channel
	workerChan chan chan engine.Request
}

func (q *QueueScheduler) Submit(r engine.Request) error {
	q.requestChan <- r
	return nil
}

// 返回worker Ready 的 Channel
func (q *QueueScheduler) WorkerReady(w chan engine.Request) {
	q.workerChan <- w
}

func (q *QueueScheduler) ConfigureMasterWorkerChan(requestChan chan engine.Request, workerChan chan chan engine.Request) {
	q.requestChan = requestChan
	q.workerChan = workerChan
}

// 程序执行入口
func (q *QueueScheduler) Run() {

	q.workerChan = make(chan chan engine.Request)
	q.requestChan = make(chan engine.Request)

	go func() {
		var requestQ []engine.Request
		var workerQ []chan engine.Request
		for true {
			var activeRequest engine.Request
			var activeWorker chan engine.Request
			if len(requestQ) > 0 && len(workerQ) > 0 {
				activeRequest = requestQ[0]
				activeWorker = workerQ[0]
			}
			select {
			case r := <-q.requestChan:
				requestQ = append(requestQ, r)
			case w := <-q.workerChan:
				workerQ = append(workerQ, w)
			case activeWorker <- activeRequest:
				workerQ = workerQ[1:]
				requestQ = requestQ[1:]
				//	不需要重置为nil吗？ 每次var定义时重置为nil
			}
		}
	}()
}
