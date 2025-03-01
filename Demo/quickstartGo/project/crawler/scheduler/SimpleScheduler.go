package scheduler

import (
	"../engine"
)

type SimpleScheduler struct {
	tasks chan engine.Request
}

// 如果方法中有指针调用者 初始化时得要传&
func (s *SimpleScheduler) Submit(request engine.Request) error {

	//s.tasks <- request
	//	fmt.Printf("Submit : %s\n",request.Url)
	// 每次提交都开启新的协程进行提交，不阻塞工作协程执行任务
	go func() {
		s.tasks <- request
		//log.Printf("Submit : %s\n",request.Url)
	}()

	return nil
}

func (s *SimpleScheduler) ConfigureMasterWorkerChan(workerChan chan engine.Request) {
	s.tasks = workerChan
}
