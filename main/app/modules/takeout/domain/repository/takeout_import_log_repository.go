package repository

import (
	"context"
	"ttpos-server-go/app/modules/takeout/domain/model"
)

// TakeoutImportLogRepository 外卖导入日志仓储接口
type TakeoutImportLogRepository interface {
	// Create 创建导入日志
	Create(ctx context.Context, log *model.TakeoutImportLog) error

	// Update 更新导入日志
	Update(ctx context.Context, log *model.TakeoutImportLog) error

	// FindByUUID 根据 UUID 查询导入日志
	FindByUUID(ctx context.Context, uuid uint64) (*model.TakeoutImportLog, error)

	// FindByID 根据 ID 查询导入日志
	FindByID(ctx context.Context, id uint64) (*model.TakeoutImportLog, error)

	// FindLatestByPlatform 查询指定平台最新的一条导入日志
	FindLatestByPlatform(ctx context.Context, platform string) (*model.TakeoutImportLog, error)

	// FindInProgressByPlatform 查询指定平台进行中的导入日志
	FindInProgressByPlatform(ctx context.Context, platform string) (*model.TakeoutImportLog, error)

	// List 查询导入日志列表
	// platform: 平台筛选，空字符串表示不筛选
	// importType: 导入类型筛选，0 表示不筛选
	// status: 状态筛选，-1 表示不筛选
	// page: 页码（从 1 开始）
	// pageSize: 每页数量
	List(ctx context.Context, platform string, importType int8, status int8, page, pageSize int) ([]*model.TakeoutImportLog, int64, error)

	// Delete 软删除导入日志
	Delete(ctx context.Context, uuid uint64) error
}
