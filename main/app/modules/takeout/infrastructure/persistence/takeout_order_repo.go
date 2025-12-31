package persistence

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/domain/value_object"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ITakeoutOrderRepo 外卖订单仓储接口
type ITakeoutOrderRepo interface {
	Create(order *model.TakeoutOrder) error
	Update(order *model.TakeoutOrder, options ...DBOption) error
	UpdateByMap(uuid uint64, data map[string]interface{}) error
	GetByUuid(uuid uint64, options ...DBOption) (*model.TakeoutOrder, error)
	GetByTakeoutOrderUuid(takeoutOrderUuid string, options ...DBOption) (*model.TakeoutOrder, error)
	GetByPlatformOrderId(platform, platformOrderId string, options ...DBOption) (*model.TakeoutOrder, error)
	GetList(options ...DBOption) ([]*model.TakeoutOrder, int64, error)
	GetCount(options ...DBOption) (int64, error) // 获取订单数量
	Delete(uuid uint64) error
	SetStaffShiftLogUuid(order *model.TakeoutOrder, staffUuid uint64) error     // 设置员工班次日志UUID
	BatchUpdateStaffShiftLogUuid(shiftLogUuid, staffUuid uint64) (int64, error) // 批量更新班次UUID
	GetErpOpenPosEntryNameByOrderUuid(orderUuid uint64) (string, error)         // 获取当前订单的ERP开账名称

	// 选项方法
	WithPreload(options ...DBOption) DBOption
	WithTakeoutOrderItems() DBOption
	WithTakeoutOrderItemModifiers() DBOption
	WithTakeoutOrderMaterials() DBOption
	WhereUuid(uuid uint64) DBOption
	WherePlatform(platform string) DBOption
	WhereOrderState(orderState int) DBOption
	WhereOrderStateIn(orderStates []int) DBOption // 根据多个订单状态筛选
	WhereTimeRange(startTime, endTime int64) DBOption
	WhereOrderTimeGt(orderTime int64) DBOption // 订单时间大于指定时间
	WherePendingOrAbnormal() DBOption          // 待接单或异常订单
	WherePendingOrAutoAccepted() DBOption      // 待接单或已自动接单
	WhereSearch(search string) DBOption
	WhereIsHistoryOrder(isHistory bool) DBOption
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

// GetByTakeoutOrderUuid 根据 TakeoutOrderUuid 字符串获取外卖订单
func (r *TakeoutOrderRepoImpl) GetByTakeoutOrderUuid(takeoutOrderUuid string, options ...DBOption) (*model.TakeoutOrder, error) {
	var order model.TakeoutOrder
	db := r.db.Model(&model.TakeoutOrder{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	err := db.Where("takeout_order_uuid = ?", takeoutOrderUuid).First(&order).Error

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

	// Preload 收货人信息
	err := db.Preload("TakeoutOrderItems").Order("order_time DESC").Find(&orders).Error

	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	return orders, total, nil
}

// GetCount 获取外卖订单数量
func (r *TakeoutOrderRepoImpl) GetCount(options ...DBOption) (int64, error) {
	var count int64
	db := r.db.Model(&model.TakeoutOrder{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	if err := db.Count(&count).Error; err != nil {
		return 0, errors.WithMessage(err)
	}

	return count, nil
}

// Delete 软删除外卖订单
func (r *TakeoutOrderRepoImpl) Delete(uuid uint64) error {
	return errors.WithMessage(
		r.db.Model(&model.TakeoutOrder{}).
			Where("uuid = ? AND delete_time = ?", uuid, constant.NotDeleted).
			Update("delete_time", time.Now().Unix()).Error,
	)
}

// SetStaffShiftLogUuid 设置员工班次日志UUID
func (r *TakeoutOrderRepoImpl) SetStaffShiftLogUuid(order *model.TakeoutOrder, staffUuid uint64) error {
	if order.StaffShiftLogUuid != 0 {
		return nil
	}
	if staffUuid == 0 {
		return nil
	}
	// 查询员工当前班次
	var shiftLog struct {
		Uuid uint64 `gorm:"column:uuid"`
	}
	err := r.db.Table("ttpos_staff_shift_log").
		Select("uuid").
		Where("staff_uuid = ? AND status = ?", staffUuid, constant.StaffNotHandedOver).
		Order("id DESC").
		First(&shiftLog).Error
	if err != nil {
		// 如果没有找到班次记录，不报错，返回 nil
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return errors.WithMessage(err)
	}
	// 设置订单的班次UUID
	order.StaffShiftLogUuid = shiftLog.Uuid
	return nil
}

// BatchUpdateStaffShiftLogUuid 批量更新班次UUID
// 将 staff_shift_log_uuid 为 0 的订单批量更新为当前班次
func (r *TakeoutOrderRepoImpl) BatchUpdateStaffShiftLogUuid(shiftLogUuid, staffUuid uint64) (int64, error) {
	result := r.db.Model(&model.TakeoutOrder{}).
		Where("staff_shift_log_uuid = ?", 0).
		Where("delete_time = ?", constant.NotDeleted).
		Updates(map[string]interface{}{
			"staff_shift_log_uuid": shiftLogUuid,
			"accepted_by":          staffUuid,
		})

	if result.Error != nil {
		return 0, errors.WithMessage(result.Error)
	}

	return result.RowsAffected, nil
}

// GetErpOpenPosEntryNameByOrderUuid 获取当前订单的ERP开账名称
func (r *TakeoutOrderRepoImpl) GetErpOpenPosEntryNameByOrderUuid(orderUuid uint64) (string, error) {
	var erpOpenPosEntryName string
	err := r.db.Table("ttpos_takeout_order as o").
		Select("s.erpnext_open_pos_entry_name").
		Joins("LEFT JOIN ttpos_staff_shift_log as s ON o.staff_shift_log_uuid = s.uuid").
		Where("o.uuid = ?", orderUuid).
		Where("o.delete_time = ?", constant.NotDeleted).
		Scan(&erpOpenPosEntryName).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", errors.WithMessage(err)
	}

	return erpOpenPosEntryName, nil
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
		if orderState >= 0 {
			return db.Where("order_state = ?", orderState)
		}
		return db
	}
}

// WhereOrderStateIn 根据多个订单状态筛选
func (r *TakeoutOrderRepoImpl) WhereOrderStateIn(orderStates []int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(orderStates) > 0 {
			return db.Where("order_state IN ?", orderStates)
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

// WhereSearch 根据关键词搜索（订单号、平台订单ID、收货人姓名、收货人电话）
func (r *TakeoutOrderRepoImpl) WhereSearch(search string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if search != "" {
			// 使用 LEFT JOIN 关联收货人表，搜索订单字段和收货人字段
			return db.Where(
				"short_order_number LIKE ? OR platform_order_id LIKE ? OR "+
					"EXISTS ("+
					"SELECT 1 FROM ttpos_takeout_order_receiver "+
					"WHERE ttpos_takeout_order_receiver.takeout_order_uuid = ttpos_takeout_order.uuid "+
					"AND ttpos_takeout_order_receiver.delete_time = 0 "+
					"AND (ttpos_takeout_order_receiver.receiver_name LIKE ? OR ttpos_takeout_order_receiver.receiver_phones LIKE ?)"+
					")",
				"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%",
			)
		}
		return db
	}
}

// WhereIsHistoryOrder 根据是否历史订单筛选
func (r *TakeoutOrderRepoImpl) WhereIsHistoryOrder(isHistory bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if !isHistory {
			return db.Where("order_state not in (4, 5)")
		}
		return db
	}
}

// WhereOrderTimeGt 订单时间大于指定时间
func (r *TakeoutOrderRepoImpl) WhereOrderTimeGt(orderTime int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if orderTime > 0 {
			return db.Where("order_time > ?", orderTime)
		}
		return db
	}
}

// WherePendingOrAbnormal 待接单或异常订单
func (r *TakeoutOrderRepoImpl) WherePendingOrAbnormal() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		// order_state = 0 (待接单) OR is_abnormal = 1 (异常订单)
		return db.Where("order_state = ? OR is_abnormal = ?", 0, 1)
	}
}

// WherePendingOrAutoAccepted 待接单或已自动接单
func (r *TakeoutOrderRepoImpl) WherePendingOrAutoAccepted() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		// order_state = 0 (待接单) OR (order_state = 1 AND order_accepted_type = 'AUTO')
		return db.Where("order_state = ? OR (order_state = ? AND order_accepted_type = ?)", 0, 1, value_object.TakeoutOrderAcceptedTypeAuto)
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

// WithPreload 预加载
func (r *TakeoutOrderRepoImpl) WithPreload(options ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		for _, option := range options {
			db = option(db)
		}
		return db
	}
}

// WithTakeoutOrderItems 预加载订单商品
func (r *TakeoutOrderRepoImpl) WithTakeoutOrderItems() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("TakeoutOrderItems")
	}
}

// WithTakeoutOrderItemModifiers 预加载订单商品修饰符
func (r *TakeoutOrderRepoImpl) WithTakeoutOrderItemModifiers() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("TakeoutOrderItems.TakeoutOrderItemModifiers")
	}
}

// WithTakeoutOrderMaterials 预加载订单原材料（含 Material 和 Unit 信息）
func (r *TakeoutOrderRepoImpl) WithTakeoutOrderMaterials() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("TakeoutOrderMaterials.Material.Unit")
	}
}
