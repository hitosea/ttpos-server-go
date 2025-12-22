package persistence

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ITakeoutOrderRepo 外卖订单仓储接口
type ITakeoutOrderRepo interface {
	Create(order *model.TakeoutOrder) error
	Update(order *model.TakeoutOrder, options ...DBOption) error
	UpdateByMap(uuid uint64, data map[string]interface{}) error
	GetByUuid(uuid uint64, options ...DBOption) (*model.TakeoutOrder, error)
	GetByPlatformOrderId(platform, platformOrderId string, options ...DBOption) (*model.TakeoutOrder, error)
	GetList(options ...DBOption) ([]*model.TakeoutOrder, int64, error)
	Delete(uuid uint64) error

	// 选项方法
	WhereUuid(uuid uint64) DBOption
	WherePlatform(platform string) DBOption
	WhereOrderState(orderState int) DBOption
	WhereTimeRange(startTime, endTime int64) DBOption
	WhereSearch(search string) DBOption
	Limit(limit int) DBOption
	Offset(offset int) DBOption
}

// NewTakeoutOrderRepo 创建外卖订单仓储
func NewTakeoutOrderRepo(db *gorm.DB) ITakeoutOrderRepo {
	return &TakeoutOrderRepoImpl{db: db}
}

// TakeoutOrderRepoImpl 外卖订单仓储实现
type TakeoutOrderRepoImpl struct {
	db *gorm.DB
}

// Create 创建外卖订单（不保存关联的商品和修饰符数据）
func (r *TakeoutOrderRepoImpl) Create(order *model.TakeoutOrder) error {
	return errors.WithMessage(r.db.Model(&model.TakeoutOrder{}).Omit(clause.Associations).Create(&order).Error)
}

// Update 更新外卖订单
func (r *TakeoutOrderRepoImpl) Update(order *model.TakeoutOrder, options ...DBOption) error {
	db := r.db.Model(&model.TakeoutOrder{})
	for _, option := range options {
		db = option(db)
	}
	return errors.WithMessage(db.Where("uuid = ?", order.Uuid).Updates(order).Error)
}

// UpdateByMap 使用 map 更新外卖订单
func (r *TakeoutOrderRepoImpl) UpdateByMap(uuid uint64, data map[string]interface{}) error {
	return errors.WithMessage(
		r.db.Model(&model.TakeoutOrder{}).
			Where("uuid = ?", uuid).
			Updates(data).Error,
	)
}

// GetByUuid 根据UUID获取外卖订单
func (r *TakeoutOrderRepoImpl) GetByUuid(uuid uint64, options ...DBOption) (*model.TakeoutOrder, error) {
	var order model.TakeoutOrder
	db := r.db.Model(&model.TakeoutOrder{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	err := db.Where("uuid = ?", uuid).First(&order).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err)
	}

	return &order, nil
}

// GetByPlatformOrderId 根据平台订单ID获取外卖订单
func (r *TakeoutOrderRepoImpl) GetByPlatformOrderId(platform, platformOrderId string, options ...DBOption) (*model.TakeoutOrder, error) {
	var order model.TakeoutOrder
	db := r.db.Model(&model.TakeoutOrder{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	err := db.Where("platform = ? AND platform_order_id = ?", platform, platformOrderId).First(&order).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err)
	}

	return &order, nil
}

// GetList 获取外卖订单列表
func (r *TakeoutOrderRepoImpl) GetList(options ...DBOption) ([]*model.TakeoutOrder, int64, error) {
	var orders []*model.TakeoutOrder
	var total int64

	db := r.db.Model(&model.TakeoutOrder{}).Where("delete_time = ?", constant.NotDeleted)

	// 应用选项构建查询条件
	countDB := db
	for _, option := range options {
		countDB = option(countDB)
	}

	// 获取总数
	if err := countDB.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	// 应用选项并获取列表
	for _, option := range options {
		db = option(db)
	}

	err := db.Order("order_time DESC").Find(&orders).Error

	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	return orders, total, nil
}

// Delete 软删除外卖订单
func (r *TakeoutOrderRepoImpl) Delete(uuid uint64) error {
	return errors.WithMessage(
		r.db.Model(&model.TakeoutOrder{}).
			Where("uuid = ? AND delete_time = ?", uuid, constant.NotDeleted).
			Update("delete_time", time.Now().Unix()).Error,
	)
}

// WhereUuid 根据UUID查询选项
func (r *TakeoutOrderRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WherePlatform 根据平台筛选
func (r *TakeoutOrderRepoImpl) WherePlatform(platform string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if platform != "" {
			return db.Where("platform = ?", platform)
		}
		return db
	}
}

// WhereOrderState 根据订单状态筛选
func (r *TakeoutOrderRepoImpl) WhereOrderState(orderState int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if orderState > 0 {
			return db.Where("order_state = ?", orderState)
		}
		return db
	}
}

// WhereTimeRange 根据时间范围筛选
func (r *TakeoutOrderRepoImpl) WhereTimeRange(startTime, endTime int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if startTime > 0 && endTime > 0 {
			return db.Where("order_time >= ? AND order_time <= ?", startTime, endTime)
		} else if startTime > 0 {
			return db.Where("order_time >= ?", startTime)
		} else if endTime > 0 {
			return db.Where("order_time <= ?", endTime)
		}
		return db
	}
}

// WhereSearch 根据关键词搜索
func (r *TakeoutOrderRepoImpl) WhereSearch(search string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if search != "" {
			return db.Where("short_order_number LIKE ? OR platform_order_id LIKE ?", "%"+search+"%", "%"+search+"%")
		}
		return db
	}
}

// Limit 限制查询数量
func (r *TakeoutOrderRepoImpl) Limit(limit int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if limit > 0 {
			return db.Limit(limit)
		}
		return db
	}
}

// Offset 偏移量
func (r *TakeoutOrderRepoImpl) Offset(offset int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if offset > 0 {
			return db.Offset(offset)
		}
		return db
	}
}
