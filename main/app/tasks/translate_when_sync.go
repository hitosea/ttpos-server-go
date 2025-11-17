package tasks

import (
	"ttpos-server-go/app/service"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// TranslateWhenSyncTask 翻译同步时产生的多语言
type TranslateWhenSyncTask struct {
	dbm   *database.DBManager
	cache cache.Cache
}

// NewTranslateWhenSyncTask 翻译同步时产生的多语言
func NewTranslateWhenSyncTask(dbm *database.DBManager, cache cache.Cache) *TranslateWhenSyncTask {
	return &TranslateWhenSyncTask{
		dbm:   dbm,
		cache: cache,
	}
}

// Execute 执行定时任务
func (t *TranslateWhenSyncTask) Execute() {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error("翻译任务发生panic", zap.Any("error", err))
		}
	}()
	logger.Logger.Info("开始执行翻译任务")
	service.NewTranslateSrv(t.dbm, t.cache).TranslateAll()
}
