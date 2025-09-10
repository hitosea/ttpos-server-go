package repository

import (
	"encoding/json"
	"fmt"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IPrinterLogRepo interface {
	WithPrinter() DBOption             // 关联打印机
	WithPrinterLogData() DBOption      // 关联打印日志数据
	WithPrinterPrinterType() DBOption  // 关联打印机.打印机类型
	WithSaleBill() DBOption            // 关联销售单
	WithSaleOrder() DBOption           // 关联销售账单
	WithProductPrinter() DBOption      // 关联销售账单.商品打印
	WithMemberRechargeOrder() DBOption // 关联充值订单

	WhereLimit(limit int) DBOption                   // 限制查询数量
	WhereStatus(status uint8) DBOption               // 状态查询条件
	WhereCreateTimeGt(ts int64) DBOption             // 创建时间大于
	WhereType(typ uint8) DBOption                    // 类型查询条件
	WhereDataType(dataType uint8) DBOption           // 数据类型查询条件
	WhereFirstExecution(firstExecution int) DBOption // 是否首次执行
	WhereUuid(uuid uint64) DBOption                  // uuid 查询条件
	WhereCreatedBefore(days uint) DBOption           // n天前的数据
	WherePrinterTime() DBOption                      // 打印时间查询条件
	WherePrintMethod(printMethod int) DBOption       // 打印方式查询条件
	WhereTimeRange(startTime, endTime uint) DBOption // 时间范围查询条件

	PaginateGet(page, pageSize int, opts ...DBOption) ([]model.PrinterLog, int64, error) // 分页获取
	GetPrintLogCount(opts ...DBOption) (int64, error)                                    // 根据条件或者打印日志数量

	GetPrinterLog(opts ...DBOption) model.PrinterLog
	GetPrinterLogList(opts ...DBOption) ([]model.PrinterLog, error)
	GetPrinterData(deviceSn string, opts ...DBOption) ([]model.PrinterLog, error)
	GetByUuids(uuids []uint64) ([]model.PrinterLog, error)
	GetShiftPrinterData(deviceSn string, opts ...DBOption) *resp.PrinterData

	GetPrinter(opts ...DBOption) *model.Printer

	Update(uuid uint64, vars map[string]any) error
	UpdateByWhere(vars map[string]any, opts ...DBOption) error
	BatchUpdate(logs []model.PrinterLog) error

	Create(printerLog model.PrinterLog) (model.PrinterLog, error)
	Delete7DaysAgo() error
}

func NewPrinterLogRepo(db *gorm.DB) IPrinterLogRepo {
	return NewPrinterLogRepoImpl(db)
}

type printerLogRepo struct {
	db *gorm.DB
}

func NewPrinterLogRepoImpl(db *gorm.DB) IPrinterLogRepo {
	return &printerLogRepo{db: db}
}

func (r *printerLogRepo) PaginateGet(page, pageSize int, opts ...DBOption) ([]model.PrinterLog, int64, error) {
	var printerLog []model.PrinterLog
	var total int64

	// 构建基础查询
	builder := r.db.Model(&model.PrinterLog{}).Scopes(NotDeleted)
	for _, opt := range opts {
		builder = opt(builder)
	}
	// 获取总记录数
	err := builder.Count(&total).Error
	if err != nil {
		return printerLog, 0, errors.WithMessage(err)
	}
	// 执行分页查询
	err = builder.Limit(pageSize).Offset((page - 1) * pageSize).Order("create_time desc").Find(&printerLog).Error
	if err != nil {
		return printerLog, 0, errors.WithMessage(err)
	}
	return printerLog, total, nil
}

// GetPrintLogCount 根据条件或者打印日志数量
func (r *printerLogRepo) GetPrintLogCount(opts ...DBOption) (int64, error) {
	var total int64
	// 构建基础查询
	builder := r.db.Model(&model.PrinterLog{}).Scopes(NotDeleted)
	for _, opt := range opts {
		builder = opt(builder)
	}
	// 获取总记录数
	err := builder.Count(&total).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return total, nil
}

func (r *printerLogRepo) GetPrinterLog(opts ...DBOption) model.PrinterLog {
	var printerLog model.PrinterLog
	db := r.db.Model(&model.PrinterLog{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.First(&printerLog)
	return printerLog
}

func (r *printerLogRepo) GetPrinterLogList(opts ...DBOption) ([]model.PrinterLog, error) {
	var printerLog []model.PrinterLog
	db := r.db.Model(&model.PrinterLog{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Find(&printerLog)
	return printerLog, nil
}

// GetPrinterData 获取打印数据
func (r *printerLogRepo) GetPrinterData(deviceSn string, opts ...DBOption) ([]model.PrinterLog, error) {
	printerLogList, err := r.GetPrinterLogList(
		r.WithPrinter(),
		r.WithPrinterLogData(),
		r.WithPrinterPrinterType(),
		r.WhereType(1),
		r.WhereStatus(1),
		r.WhereLimit(50),
		r.WhereFirstExecution(0),
		func(db *gorm.DB) *gorm.DB {
			// 相同设备的
			db.Where("(cashier_device_id = ? OR cashier_device_id = '')", deviceSn)
			// 0次或1次
			db.Where("(num in (0, 1))")
			// 1天内
			db.Where("(create_time > UNIX_TIMESTAMP() - 86400)")
			// 未读
			db.Where("read_device_id = ''")
			//
			return db
		},
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 标记已读
	if len(printerLogList) > 0 {
		// 提取打印日志ID列表
		var logIds []uint
		for _, log := range printerLogList {
			logIds = append(logIds, log.ID)
		}
		// 更新打印日志为已读
		err = r.UpdateByWhere(
			map[string]any{
				"read_device_id": deviceSn,
			},
			func(db *gorm.DB) *gorm.DB {
				return db.Where("id in (?)", logIds)
			},
		)
		if err != nil {
			logger.Logger.Error("更新打印日志失败", zap.Error(err))
		}
	}

	return printerLogList, nil
}

// GetShiftPrinterData 获取交班打印数据
func (r *printerLogRepo) GetShiftPrinterData(deviceSn string, opts ...DBOption) *resp.PrinterData {
	printerLog := r.GetPrinterLog(
		r.WithPrinter(),
		r.WithPrinterPrinterType(),
		r.WhereType(1),
		r.WhereDataType(constant.PrinterTemplateHandoverSheet),
		r.WhereStatus(1),
		r.WhereFirstExecution(0),
		func(db *gorm.DB) *gorm.DB {
			// 相同设备的
			db.Where("(cashier_device_id = ? OR cashier_device_id = '')", deviceSn)
			// 0次或1次
			db.Where("(num in (0, 1))")
			// 1天内
			db.Where("(create_time > UNIX_TIMESTAMP() - 86400)")
			// 未读
			db.Where("read_device_id = ''")
			//
			return db
		},
	)

	// 标记已读
	if printerLog.ID > 0 {
		// 更新打印日志为已读
		err := r.UpdateByWhere(
			map[string]any{
				"read_device_id": deviceSn,
				"status":         2,
			},
			func(db *gorm.DB) *gorm.DB {
				return db.Where("id = ?", printerLog.ID)
			},
		)
		if err != nil {
			logger.Logger.Error("更新打印日志失败", zap.Error(err))
		}
		// 返回打印数据
		return &resp.PrinterData{
			Uuid:             printerLog.Uuid,
			Data:             printerLog.Data,
			PrintMethod:      printerLog.PrintMethod,
			PrinterType:      printerLog.PrinterType,
			IsCashierPrinter: printerLog.IsCashierPrinter(),
			IsUsbPrinter:     printerLog.IsUsbPrinter(),
			Copies:           printerLog.GetCopies(),
			PrinterConfig: func() string {
				if printerLog.Printer == nil {
					return ""
				}
				configJson, err := json.Marshal(printerLog.Printer.GetConfigJson())
				if err != nil {
					return ""
				}
				return string(configJson)
			}(),
			PrintingTime: printerLog.PrintingTime,
			EnableStatusCheck: func() int {
				if printerLog.Printer == nil {
					return 0
				}
				return printerLog.Printer.EnableStatusCheck
			}(),
		}
	}

	return nil
}

// GetPrinter 获取打印机
func (r *printerLogRepo) GetPrinter(opts ...DBOption) *model.Printer {
	var printer model.Printer
	db := r.db.Model(&model.Printer{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.First(&printer)
	return &printer
}

// GetByUuids 根据UUID列表获取打印日志
func (r *printerLogRepo) GetByUuids(uuids []uint64) ([]model.PrinterLog, error) {
	var printerLogs []model.PrinterLog
	if len(uuids) == 0 {
		return printerLogs, nil
	}

	err := r.db.Model(&model.PrinterLog{}).Scopes(NotDeleted).Select("id", "uuid", "num", "status", "reason").Where("uuid IN (?)", uuids).Find(&printerLogs).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return printerLogs, nil
}

func (r *printerLogRepo) WithPrinter() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Printer")
	}
}
func (r *printerLogRepo) WithMemberRechargeOrder() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberRechargeOrder")
	}
}
func (r *printerLogRepo) WithPrinterLogData() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PrinterLogData")
	}
}
func (r *printerLogRepo) WithPrinterPrinterType() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Printer.PrinterType")
	}
}
func (r *printerLogRepo) WithSaleBill() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleBill")
	}
}

// WithSaleOrder 关联销售账单
func (r *printerLogRepo) WithSaleOrder() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleOrder.SaleBill")
	}
}

func (r *printerLogRepo) WithProductPrinter() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPrinter")
	}
}

func (r *printerLogRepo) WhereStatus(status uint8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

func (r *printerLogRepo) WhereCreateTimeGt(ts int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("create_time > ?", ts)
	}
}

func (r *printerLogRepo) WhereLimit(limit int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

func (r *printerLogRepo) WhereType(typ uint8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ?", typ)
	}
}

func (r *printerLogRepo) WhereDataType(dataType uint8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("data_type = ?", dataType)
	}
}

func (r *printerLogRepo) WherePrintMethod(printMethod int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("print_method = ?", printMethod)
	}
}

func (r *printerLogRepo) WherePrinterTime() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("((printer_time + 10) < UNIX_TIMESTAMP() OR num = 0)")
	}
}

func (r *printerLogRepo) WhereFirstExecution(firstExecution int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("first_execution = ?", firstExecution)
	}
}

func (r *printerLogRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereTimeRange 创建时间范围查询条件
func (r *printerLogRepo) WhereTimeRange(startTime, endTime uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if startTime > 0 {
			db = db.Where("create_time >= ?", startTime)
		}
		if endTime > 0 {
			db = db.Where("create_time <= ?", endTime)
		}
		return db
	}
}

func (r *printerLogRepo) WhereCreatedBefore(days uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("create_time < ?", time.Now().Add(time.Duration(-days*24)*time.Hour).Unix())
	}
}

func (r *printerLogRepo) Update(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.PrinterLog{}).Where("uuid = ?", uuid).Updates(vars).Error
}

func (r *printerLogRepo) UpdateByWhere(vars map[string]any, opts ...DBOption) error {
	db := r.db.Model(&model.PrinterLog{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(vars).Error
}

func (r *printerLogRepo) BatchUpdate(logs []model.PrinterLog) error {
	if len(logs) == 0 {
		return nil
	}

	// 每批处理的数据量
	batchSize := 50

	// 分批处理
	for i := 0; i < len(logs); i += batchSize {
		end := i + batchSize
		if end > len(logs) {
			end = len(logs)
		}
		batch := logs[i:end]

		// 构建当前批次的 CASE WHEN 语句
		var uuids []uint64
		statusCases := "CASE uuid"
		reasonCases := "CASE uuid"
		numCases := "CASE uuid"

		for _, log := range batch {
			uuids = append(uuids, log.Uuid)
			statusCases += fmt.Sprintf(" WHEN %d THEN %d", log.Uuid, log.Status)
			reasonCases += fmt.Sprintf(" WHEN %d THEN '%s'", log.Uuid, log.Reason)
			numCases += fmt.Sprintf(" WHEN %d THEN %d", log.Uuid, log.Num)
		}

		statusCases += " ELSE status END"
		reasonCases += " ELSE reason END"
		numCases += " ELSE num END"

		// 执行当前批次的批量更新
		err := r.db.Model(&model.PrinterLog{}).
			Where("uuid IN (?)", uuids).
			Updates(map[string]interface{}{
				"status": gorm.Expr(statusCases),
				"reason": gorm.Expr(reasonCases),
				"num":    gorm.Expr(numCases),
			}).Error

		if err != nil {
			return errors.WithMessage(err)
		}
	}

	return nil
}

func (r *printerLogRepo) Create(printerLog model.PrinterLog) (model.PrinterLog, error) {
	err := r.db.Model(&model.PrinterLog{}).Create(&printerLog).Error
	return printerLog, errors.WithMessage(err)
}

// 删除7天前的数据 - 从数据库上删除
func (r *printerLogRepo) Delete7DaysAgo() error {
	return r.db.Model(&model.PrinterLog{}).Where("create_time < ?", time.Now().Add(time.Duration(-7*24)*time.Hour).Unix()).Delete(&model.PrinterLog{}).Error
}
