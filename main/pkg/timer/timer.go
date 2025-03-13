package timer

import (
	"sync"
	"time"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// 定时任务接口
type Task interface {
	// 执行任务
	Execute()
	// 获取任务名称
	GetName() string
}

// 定时任务管理器
type TimerManager struct {
	tasks     map[string]Task
	taskMutex sync.RWMutex
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

var (
	instance *TimerManager
	once     sync.Once
)

// 获取定时任务管理器实例
func GetTimerManager() *TimerManager {
	once.Do(func() {
		instance = &TimerManager{
			tasks:    make(map[string]Task),
			stopChan: make(chan struct{}),
		}
	})
	return instance
}

// 注册定时任务
func (tm *TimerManager) RegisterTask(task Task) {
	tm.taskMutex.Lock()
	defer tm.taskMutex.Unlock()
	
	taskName := task.GetName()
	tm.tasks[taskName] = task
	logger.Logger.Info("注册定时任务", zap.String("任务名称", taskName))
}

// 启动秒级定时器
func (tm *TimerManager) Start() {
	tm.wg.Add(1)
	go func() {
		defer tm.wg.Done()
		
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				tm.executeAllTasks()
			case <-tm.stopChan:
				logger.Logger.Info("定时任务管理器已停止")
				return
			}
		}
	}()
	logger.Logger.Info("定时任务管理器已启动")
}

// 执行所有任务
func (tm *TimerManager) executeAllTasks() {
	tm.taskMutex.RLock()
	defer tm.taskMutex.RUnlock()
	
	for _, task := range tm.tasks {
		go func(t Task) {
			defer func() {
				if r := recover(); r != nil {
					logger.Logger.Error("定时任务执行异常", 
						zap.String("任务名称", t.GetName()),
						zap.Any("错误", r))
				}
			}()
			
			t.Execute()
		}(task)
	}
}

// 停止定时器
func (tm *TimerManager) Stop() {
	close(tm.stopChan)
	tm.wg.Wait()
}
