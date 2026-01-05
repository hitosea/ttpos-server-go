package controller

import (
	goCtx "context"

	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ICacheObjectController 缓存对象控制器接口
// 统一管理对象的缓存查询和更新，支持观察者模式更新缓存
// 可用于 Desk、ProductPackage 等各种对象类型
type ICacheObjectController interface {
	// GetByUuid 根据 UUID 获取对象（带缓存）
	// 参数：
	//   - ctx: 上下文（用于提取 companyUuid）
	//   - db: 数据库连接
	//   - uuid: 对象 UUID
	// 返回：
	//   - *model.Desk: 对象指针
	//   - error: 错误信息
	GetByUuid(ctx goCtx.Context, db *gorm.DB, uuid uint64) (*model.Desk, error)

	// BatchGetByUuids 批量根据 UUID 列表获取对象（带缓存）
	// 参数：
	//   - ctx: 上下文（用于提取 companyUuid）
	//   - db: 数据库连接
	//   - uuids: 对象 UUID 列表
	//   - opts: 选项函数（可选），如 WithSkipCache() 跳过缓存直接查询
	// 返回：
	//   - map[uint64]*model.Desk: UUID 到对象的映射
	//   - error: 错误信息
	BatchGetByUuids(ctx goCtx.Context, db *gorm.DB, uuids []uint64, opts ...func(*BatchGetByUuidsOption)) (map[uint64]*model.Desk, error)

	// Update 更新对象的缓存（用于观察者模式）
	// 当对象发生变化时，调用此方法更新缓存
	// 通过重新调用 BatchGetByUuids 并跳过缓存，从数据库重新查询并更新缓存
	// 参数：
	//   - ctx: 上下文（用于提取 companyUuid）
	//   - db: 数据库连接
	//   - uuids: 对象 UUID 列表
	// 返回：
	//   - error: 错误信息
	Update(ctx goCtx.Context, db *gorm.DB, uuids []uint64) error
}
