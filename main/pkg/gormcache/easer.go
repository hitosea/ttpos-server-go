package gormcache

import (
	"sync"
)

// task 任务接口
type task interface {
	GetId() string
	Run()
}

// eased 包装正在执行的任务
type eased struct {
	wg   sync.WaitGroup
	task task
}

// ease 执行任务，确保相同 ID 的任务只执行一次
// 如果任务已在执行中，后续请求会等待其完成并共享结果
//
// 原理：
//  1. 尝试将任务放入队列（sync.Map）
//  2. 如果队列中已存在相同 ID 的任务，等待其完成
//  3. 如果是首个请求，执行任务并通知所有等待者
func ease(t task, queue *sync.Map) {
	id := t.GetId()

	// 创建 eased 包装器
	e := &eased{task: t}
	e.wg.Add(1)

	// 尝试放入队列（原子操作）
	actual, loaded := queue.LoadOrStore(id, e)

	if loaded {
		// 已存在相同任务 → 等待其完成
		existing := actual.(*eased)
		existing.wg.Wait()
		return
	}

	// 首个请求 → 执行任务
	defer func() {
		e.wg.Done()      // 通知等待者
		queue.Delete(id) // 清理队列
	}()

	t.Run() // 实际执行查询
}

// queryTask 查询任务实现
type queryTask struct {
	id      string              // 任务唯一标识（= SQL + 参数）
	queryCb func()              // 原始查询回调
}

func (t *queryTask) GetId() string {
	return t.id
}

func (t *queryTask) Run() {
	t.queryCb()
}
