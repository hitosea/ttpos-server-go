package task_center

import (
	"context"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/utils"
)

type TaskCenter struct {
	task     []*model.Task           // 任务列表
	handlers map[HandlerType]Handler // 任务处理器列表
}

var TaskCenterInstance *TaskCenter

func init() {
	TaskCenterInstance = NewTaskCenter()
}

func InitTaskCenter(ctx context.Context) {
	TaskCenterInstance = NewTaskCenter()
}

func NewTaskCenter() *TaskCenter {
	return &TaskCenter{}
}

func (t *TaskCenter) Run() {
	utils.Go(func() {
		for {

		}
	})
}
