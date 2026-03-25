package tasks

import (
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

type delOperationDurationTask struct {
	dbm   *database.DBManager
	cache cache.Cache
}

func NewDelOperationDurationTask(dbm *database.DBManager, cache cache.Cache) *delOperationDurationTask {
	return &delOperationDurationTask{
		dbm:   dbm,
		cache: cache,
	}
}

// Execute 执行删除7天前的订单操作耗时记录
func (t *delOperationDurationTask) Execute() {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Error("删除订单操作耗时记录定时任务,发生panic", zap.Any("panic", r))
		}
	}()

	if err := repository.NewOrderOperationDurationRepo(t.dbm.GetDB(0)).Delete7DaysAgo(); err != nil {
		logger.Logger.Error("删除订单操作耗时记录失败", zap.Error(err))
	}
}
