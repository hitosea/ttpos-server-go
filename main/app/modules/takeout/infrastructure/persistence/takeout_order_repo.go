package persistence

import (
	"time"
	"ttpos-server-go/app/errors"
	appModel "ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/takeout/domain/model"
	"ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
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
	SetStaffShiftLogUuid(order *model.TakeoutOrder) error                                   // 设置员工班次日志UUID
	BatchUpdateStaffShiftLogUuid(shiftLogUuid, staffUuid uint64) (int64, error)             // 批量更新班次UUID
	GetPendingShiftLogOrders() ([]*model.TakeoutOrder, error)                               // 获取待分配班次的订单（staff_shift_log_uuid = 0，排除已取消和已拒单）
	BatchAssignShiftLog(shiftLogUuid, staffUuid uint64, orderUuids []uint64) (int64, error) // 批量分配班次给订单
	GetErpOpenPosEntryNameByOrderUuid(orderUuid uint64) (string, error)                     // 获取当前订单的ERP开账名称
	GetShiftLogByOrderUuid(orderUuid uint64) (*appModel.StaffShiftLog, error)               // 根据订单UUID获取班次信息
	DeleteRelatedData(orderUuid uint64) error                                               // 删除订单关联数据（商品、修饰符、收货人、活动、促销、物料）

	// 选项方法
	WithPreload(options ...DBOption) DBOption
	WithTakeoutOrderItems() DBOption
	WithTakeoutOrderItemModifiers() DBOption
	WithTakeoutOrderUpdateLogs(opts ...DBOption) DBOption
	WithTakeoutOrderMaterials() DBOption
	WithTakeoutOrderReceiver() DBOption
	WithTakeoutOrderCampaigns() DBOption
	WithTakeoutOrderPromos() DBOption
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
	Select(fields ...string) DBOption
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
	db := r.db.Model(&model.TakeoutOrder{}).Where("delete_time = ?", 0)

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
	db := r.db.Model(&model.TakeoutOrder{}).Where("delete_time = ?", 0)

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
	db := r.db.Model(&model.TakeoutOrder{}).Where("delete_time = ?", 0)

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

	db := r.db.Model(&model.TakeoutOrder{}).Where("delete_time = ?", 0)

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
	db := r.db.Model(&model.TakeoutOrder{}).Where("delete_time = ?", 0)

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
			Where("uuid = ? AND delete_time = ?", uuid, 0).
			Update("delete_time", time.Now().Unix()).Error,
	)
}

// SetStaffShiftLogUuid 设置员工班次日志UUID
func (r *TakeoutOrderRepoImpl) SetStaffShiftLogUuid(order *model.TakeoutOrder) error {
	shiftLogUuid, staffUuid, err := r.GetAvailableShiftLogUuid()
	if err != nil {
		logger.Logger.Warn("获取可用班次失败", zap.Error(err), zap.Uint64("orderUuid", order.Uuid))
	} else if shiftLogUuid > 0 {
		order.StaffShiftLogUuid = shiftLogUuid
		order.AcceptedBy = staffUuid
	}
	return nil
}

// BatchUpdateStaffShiftLogUuid 批量更新班次UUID
// 将 staff_shift_log_uuid 为 0 的订单批量更新为当前班次
func (r *TakeoutOrderRepoImpl) BatchUpdateStaffShiftLogUuid(shiftLogUuid, staffUuid uint64) (int64, error) {
	result := r.db.Model(&model.TakeoutOrder{}).
		Where("staff_shift_log_uuid = ?", 0).
		Where("delete_time = ?", 0).
		Updates(map[string]interface{}{
			"staff_shift_log_uuid": shiftLogUuid,
			"accepted_by":          staffUuid,
		})

	if result.Error != nil {
		return 0, errors.WithMessage(result.Error)
	}

	return result.RowsAffected, nil
}

// GetPendingShiftLogOrders 获取待分配班次的订单（staff_shift_log_uuid = 0，排除已取消和已拒单）
func (r *TakeoutOrderRepoImpl) GetPendingShiftLogOrders() ([]*model.TakeoutOrder, error) {
	var orders []*model.TakeoutOrder
	err := r.db.Model(&model.TakeoutOrder{}).
		Where("staff_shift_log_uuid = ?", 0).
		Where("delete_time = ?", 0).
		Find(&orders).Error

	if err != nil {
		return nil, errors.WithMessage(err, "查询待分配班次的订单失败")
	}

	return orders, nil
}

// BatchAssignShiftLog 批量分配班次给订单
func (r *TakeoutOrderRepoImpl) BatchAssignShiftLog(shiftLogUuid, staffUuid uint64, orderUuids []uint64) (int64, error) {
	if len(orderUuids) == 0 {
		return 0, nil
	}

	result := r.db.Model(&model.TakeoutOrder{}).
		Where("uuid IN ?", orderUuids).
		Where("staff_shift_log_uuid = ?", 0).
		Where("delete_time = ?", 0).
		Updates(map[string]interface{}{
			"staff_shift_log_uuid": shiftLogUuid,
			"accepted_by":          staffUuid,
		})

	if result.Error != nil {
		return 0, errors.WithMessage(result.Error, "批量分配班次失败")
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
		Where("o.delete_time = ?", 0).
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
			return db.Where("order_state not in (?, ?, ?)",
				value_object.TakeoutOrderStateCompleted,
				value_object.TakeoutOrderStateRejected,
				value_object.TakeoutOrderStateCanceled,
			)
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
		return db.Where("order_state = ? OR is_abnormal = ?", value_object.TakeoutOrderStatePending, 1)
	}
}

// WherePendingOrAutoAccepted 待接单或已自动接单
func (r *TakeoutOrderRepoImpl) WherePendingOrAutoAccepted() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		// order_state = 0 (待接单) OR (order_state	= 10 AND order_accepted_type = 'AUTO')
		return db.Where("order_state = ? OR (order_state = ? AND order_accepted_type = ?)", value_object.TakeoutOrderStatePending, value_object.TakeoutOrderStateAccepted, value_object.TakeoutOrderAcceptedTypeAuto)
	}
}

// Select 选择查询字段
func (r *TakeoutOrderRepoImpl) Select(fields ...string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Select(fields)
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

// WithTakeoutOrderUpdateLogs 预加载订单更新日志
func (r *TakeoutOrderRepoImpl) WithTakeoutOrderUpdateLogs(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Preload("TakeoutOrderUpdateLogs", func(db *gorm.DB) *gorm.DB {
			for _, opt := range opts {
				db = opt(db)
			}
			return db.Order("create_time DESC")
		})
		return db
	}
}

// WithTakeoutOrderMaterials 预加载订单原材料（含 Material 和 Unit 信息）
func (r *TakeoutOrderRepoImpl) WithTakeoutOrderMaterials() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("TakeoutOrderMaterials.Material.Unit")
	}
}

// WithTakeoutOrderReceiver 预加载订单收货人信息
func (r *TakeoutOrderRepoImpl) WithTakeoutOrderReceiver() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("TakeoutOrderReceiver")
	}
}

// WithTakeoutOrderCampaigns 预加载订单活动信息
func (r *TakeoutOrderRepoImpl) WithTakeoutOrderCampaigns() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("TakeoutOrderCampaigns")
	}
}

// WithTakeoutOrderPromos 预加载订单促销信息
func (r *TakeoutOrderRepoImpl) WithTakeoutOrderPromos() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("TakeoutOrderPromos")
	}
}

// DeleteRelatedData 删除订单关联数据（商品、修饰符、收货人、活动、促销、物料）
// 使用物理删除，按依赖顺序：先删子表再删主表
func (r *TakeoutOrderRepoImpl) DeleteRelatedData(orderUuid uint64) error {
	// 1. 删除商品修饰符（依赖商品）
	if err := r.db.Unscoped().Where("takeout_order_item_uuid IN (SELECT uuid FROM ttpos_takeout_order_item WHERE takeout_order_uuid = ?)", orderUuid).Delete(&model.TakeoutOrderItemModifier{}).Error; err != nil {
		return errors.WithMessage(err, "删除订单商品修饰符失败")
	}

	// 2. 删除商品
	if err := r.db.Unscoped().Where("takeout_order_uuid = ?", orderUuid).Delete(&model.TakeoutOrderItem{}).Error; err != nil {
		return errors.WithMessage(err, "删除订单商品失败")
	}

	// 3. 删除收货人
	if err := r.db.Unscoped().Where("takeout_order_uuid = ?", orderUuid).Delete(&model.TakeoutOrderReceiver{}).Error; err != nil {
		return errors.WithMessage(err, "删除订单收货人失败")
	}

	// 4. 删除活动
	if err := r.db.Unscoped().Where("takeout_order_uuid = ?", orderUuid).Delete(&model.TakeoutOrderCampaign{}).Error; err != nil {
		return errors.WithMessage(err, "删除订单活动失败")
	}

	// 5. 删除促销
	if err := r.db.Unscoped().Where("takeout_order_uuid = ?", orderUuid).Delete(&model.TakeoutOrderPromo{}).Error; err != nil {
		return errors.WithMessage(err, "删除订单促销失败")
	}

	// 6. 删除物料
	if err := r.db.Unscoped().Where("takeout_order_uuid = ?", orderUuid).Delete(&model.TakeoutOrderMaterial{}).Error; err != nil {
		return errors.WithMessage(err, "删除订单物料失败")
	}

	return nil
}

// GetShiftLogByOrderUuid 根据订单UUID获取班次信息
// @version v2.12.0
// @spec story-erp-grab-invoice-sync
func (r *TakeoutOrderRepoImpl) GetShiftLogByOrderUuid(orderUuid uint64) (*appModel.StaffShiftLog, error) {
	// 1. 查询订单获取 staff_shift_log_uuid
	var order model.TakeoutOrder
	err := r.db.Model(&model.TakeoutOrder{}).
		Where("uuid = ?", orderUuid).
		Where("delete_time = ?", 0).
		Select("staff_shift_log_uuid").
		First(&order).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.WithMessage(errors.New("订单不存在"), "查询订单失败")
		}
		return nil, errors.WithMessage(err, "查询订单失败")
	}

	// 2. 检查订单是否有班次信息
	if order.StaffShiftLogUuid == 0 {
		return nil, errors.WithMessage(errors.New("订单没有班次信息"), "订单未关联班次")
	}

	// 3. 查询班次信息
	var shiftLog appModel.StaffShiftLog
	err = r.db.Model(&appModel.StaffShiftLog{}).
		Where("uuid = ?", order.StaffShiftLogUuid).
		First(&shiftLog).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.WithMessage(errors.New("班次不存在"), "查询班次失败")
		}
		return nil, errors.WithMessage(err, "查询班次失败")
	}

	return &shiftLog, nil
}

// getAvailableShiftLogUuid 获取可用的班次UUID
// 优先级规则（只管班次，不管登录状态）：
// 1. 主收银机的班次（优先级最高）
// 2. 非主收银机中，最先登录的班次（按 finally_login_time ASC 排序）
// 3. 若都没有未交班的班次，返回0（等待下次登录时处理）
func (r *TakeoutOrderRepoImpl) GetAvailableShiftLogUuid() (uint64, uint64, error) {
	db := r.db

	// 定义查询结果结构体
	type ShiftLogWithDevice struct {
		appModel.StaffShiftLog
		IsMain           uint8 `gorm:"column:is_main"`
		FinallyLoginTime int64 `gorm:"column:finally_login_time"`
	}

	// 一次性查询所有未交班的班次（关联设备信息）
	var shiftLogs []ShiftLogWithDevice
	err := db.Model(&appModel.StaffShiftLog{}).
		Select("ttpos_staff_shift_log.*, ttpos_device.is_main, ttpos_device.finally_login_time").
		Joins("INNER JOIN ttpos_staff ON ttpos_staff_shift_log.staff_uuid = ttpos_staff.uuid").
		Joins("INNER JOIN ttpos_device ON ttpos_staff.bind_key = ttpos_device.device_id").
		Where("ttpos_staff_shift_log.status = ?", 0).
		Where("ttpos_device.source = ?", "cashier").
		Where("ttpos_device.delete_time = ?", 0).
		Where("ttpos_staff.delete_time = ?", 0).
		Order("ttpos_device.is_main DESC").           // 主收银机优先（1 在前，0 在后）
		Order("ttpos_device.finally_login_time ASC"). // 同为非主收银机时，按登录时间升序
		Order("ttpos_staff_shift_log.uuid DESC").     // 同一设备多个班次时，取最新的
		Find(&shiftLogs).Error

	if err != nil {
		logger.Logger.Error("查询未交班班次失败", zap.Error(err))
		return 0, 0, err
	}

	if len(shiftLogs) == 0 {
		logger.Logger.Info("未找到可用班次，等待下次登录时处理")
		return 0, 0, nil
	}

	// 返回优先级最高的班次（已按规则排序，直接取第一个）
	shiftLog := shiftLogs[0]

	return shiftLog.Uuid, shiftLog.StaffUuid, nil
}
