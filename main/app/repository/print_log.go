package repository

import (
	"gorm.io/gorm"
	"time"
	"ttpos-server-go/app/model"
)

type IPrinterLogRepo interface {
	WithPrinter() DBOption            // 关联打印机
	WithPrinterPrinterType() DBOption // 关联打印机.打印机类型
	WithSaleBill() DBOption           // 关联销售账单
	WithSaleBillDesk() DBOption       // 关联销售账单.桌台

	WhereStatus(status uint8) DBOption     // 状态查询条件
	WhereType(typ uint8) DBOption          // 类型查询条件
	WhereUuid(uuid uint64) DBOption        // uuid 查询条件
	WhereCreatedBefore(days uint) DBOption // n天前的数据

	PaginateGet(page, pageSize int, opts ...DBOption) ([]model.PrinterLog, int64, error) // 分页获取

	GetPrinterLog(opts ...DBOption) model.PrinterLog
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
		return printerLog, 0, err
	}
	// 执行分页查询
	err = builder.Limit(pageSize).Offset((page - 1) * pageSize).Order("create_time desc").Find(&printerLog).Error
	if err != nil {
		return printerLog, 0, err
	}
	return printerLog, total, nil
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

func (r *printerLogRepo) WithPrinter() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Printer")
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

func (r *printerLogRepo) WithSaleBillDesk() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleBill.Desk")
	}
}

func (r *printerLogRepo) WhereStatus(status uint8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

func (r *printerLogRepo) WhereType(typ uint8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("type = ?", typ)
	}
}

func (r *printerLogRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
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
	return printerLog, err
}
