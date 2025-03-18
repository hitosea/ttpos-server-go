package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IPrinterLogRepo interface {
	WithPrinter() DBOption             // 关联打印机
	WithPrinterPrinterType() DBOption  // 关联打印机.打印机类型
	WithSaleBill() DBOption            // 关联销售单
	WithSaleOrder() DBOption           // 关联销售账单
	WithProductPrinter() DBOption      // 关联销售账单.商品打印
	WithMemberRechargeOrder() DBOption // 关联充值订单

	WhereLimit(limit int) DBOption                   // 限制查询数量
	WhereStatus(status uint8) DBOption               // 状态查询条件
	WhereType(typ uint8) DBOption                    // 类型查询条件
	WhereFirstExecution(firstExecution int) DBOption // 是否首次执行
	WhereUuid(uuid uint64) DBOption                  // uuid 查询条件
	WhereCreatedBefore(days uint) DBOption           // n天前的数据
	WherePrinterTime() DBOption                      // 打印时间查询条件
	WherePrintMethod(printMethod int) DBOption       // 打印方式查询条件
	WhereTimeRange(startTime, endTime uint) DBOption // 时间范围查询条件

	PaginateGet(page, pageSize int, opts ...DBOption) ([]model.PrinterLog, int64, error) // 分页获取
	GetPrintLogCount(opts ...DBOption) (int64, error)

	GetPrinterLog(opts ...DBOption) model.PrinterLog
	GetPrinterLogList(opts ...DBOption) ([]model.PrinterLog, error)
	GetPrinterData(deviceSn string, opts ...DBOption) ([]model.PrinterLog, error)

	//
	Update(uuid uint64, vars map[string]any) error
	UpdateByWhere(vars map[string]any, opts ...DBOption) error

	Create(printerLog model.PrinterLog) (model.PrinterLog, error)
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
		r.WithPrinterPrinterType(),
		r.WhereType(1),
		r.WhereStatus(0),
		r.WhereLimit(5),
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
			map[string]any{"read_device_id": deviceSn},
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
		return db.Where("create_time < ?", time.Now().Add(time.Duration(-days*24)*time.Hour))
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

func (r *printerLogRepo) Create(printerLog model.PrinterLog) (model.PrinterLog, error) {
	err := r.db.Model(&model.PrinterLog{}).Create(&printerLog).Error
	return printerLog, errors.WithMessage(err)
}
