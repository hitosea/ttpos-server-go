package tasks

import (
	"ttpos-server-go/app/service"
	"ttpos-server-go/app/service/rpc/message"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
)

// TranslateWhenSyncTask 翻译同步时产生的多语言
type CheckMaterialStockTask struct {
	dbm   *database.DBManager
	cache cache.Cache
}

// NewTranslateWhenSyncTask 翻译同步时产生的多语言
func NewCheckMaterialStockTask(dbm *database.DBManager, cache cache.Cache) *CheckMaterialStockTask {
	return &CheckMaterialStockTask{
		dbm:   dbm,
		cache: cache,
	}
}

// Execute 执行定时任务
func (t *CheckMaterialStockTask) Execute() {
	defer func() {
		if err := recover(); err != nil {
			logger.Logger.Error("检查物品库存发生panic", zap.Any("error", err))
		}
	}()
	logger.Logger.Info("开始执行检查物品库存任务")
	settingSrv := setting.NewSrv(t.dbm, t.cache)
	translateSrv := service.NewTranslateSrv(t.dbm, t.cache)
	messageSrv := message.NewIMessageSrv(t.dbm)
	localeSrv := service.NewLocaleSrv()
	ctx := context.NewDefaultContext()
	service.NewMaterialSrv(t.dbm, localeSrv, settingSrv, translateSrv, messageSrv).CheckMaterialSafetyStock(ctx, 0)
}
