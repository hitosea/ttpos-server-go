package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"

	"gorm.io/gorm"
)

// ISettingRepository 设置存储接口
type ISettingRepository interface {
	// 基础CRUD操作
	FindByKey(ctx context.Context, key string) (*model.Setting, error)
	FindAll(ctx context.Context) ([]*model.Setting, error)
	Save(ctx context.Context, setting *model.Setting) error
	Update(ctx context.Context, setting *model.Setting) error
	Delete(ctx context.Context, key string) error

	// 批量操作
	SaveBatch(ctx context.Context, settings []*model.Setting) error
	UpdateBatch(ctx context.Context, settings []*model.Setting) error

	// 条件查询
	FindByKeys(ctx context.Context, keys []string) ([]*model.Setting, error)
	Exists(ctx context.Context, key string) (bool, error)

	// 公司设置相关
	GetCompanySetting(ctx context.Context) (*model.CompanySetting, error)

	// 支付方式相关
	GetPaymentMethodList(ctx context.Context) ([]model.PaymentMethod, error)

	// 缓存相关
	GetFromCache(ctx context.Context, key string) (string, bool)
	SetToCache(ctx context.Context, key, value string, ttl int) error
	DeleteFromCache(ctx context.Context, key string) error
	ClearCache(ctx context.Context) error

	// 数据库连接相关
	GetRepository(db *gorm.DB) *gorm.DB
}

// ICloudSettingRepository 云端设置存储接口
type ICloudSettingRepository interface {
	// 云端设置获取
	GetCloudBasicInfo(ctx context.Context) (string, error)
}

// IPrinterRepository 打印机存储接口
type IPrinterRepository interface {
	FindByUUID(ctx context.Context, uuid uint64) (*model.Printer, error)
	FindBySN(ctx context.Context, sn string) (*model.Printer, error)
}

// ICompanySettingRepository 公司设置存储接口
type ICompanySettingRepository interface {
	GetCompanySetting(ctx context.Context) (*model.CompanySetting, error)
}

// IDeviceRepository 设备存储接口
type IDeviceRepository interface {
	FindBySN(ctx context.Context, sn string) (*model.Device, error)
}

// 使用现有的model实体
type Printer = model.Printer
type PrinterType = model.PrinterType
type CompanySetting = model.CompanySetting
type Device = model.Device
