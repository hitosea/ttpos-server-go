package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IDeviceRepo interface {
	WhereSn(deviceId string) DBOption   // sn/device_id 条件
	WhereUuid(uuid uint64) DBOption     // uuid 条件
	WhereSource(source string) DBOption // 来源（cashier、tablet、kitchen、assistant）条件
	WhereMain() DBOption                // 主设备条件

	GetDevice(opts ...DBOption) (model.Device, error)       // 获取设备 - 不包含软删除
	GetDeviceAll(opts ...DBOption) (model.Device, error)    // 获取设备 - 所有 - 包含软删除
	GetDeviceBySn(sn string) (*model.Device, error)         // 根据sn获取设备
	GetDeviceByUuid(uuid uint64) (*model.Device, error)     // 根据uuid获取设备
	GetDeviceBrand(opts ...DBOption) string                 // 获取设备品牌
	GetDeviceSn(opts ...DBOption) string                    // 获取设备sn
	GetBindCountBySource(source string) uint                // 根据来源获取设备绑定数量
	IsExistCashierMain(source string) bool                  // 是否存在主设备
	CreateDevice(device model.Device) (model.Device, error) // 创建设备
	UpdateDevice(uuid uint64, vars map[string]any) error    // 更新设备
	GetDeviceList(opts ...DBOption) ([]model.Device, error) // 获取设备列表

}

func NewDeviceRepo(db *gorm.DB) IDeviceRepo {
	return NewDeviceRepoImpl(db)
}

type deviceRepo struct {
	db *gorm.DB
}

func NewDeviceRepoImpl(db *gorm.DB) IDeviceRepo {
	return &deviceRepo{db: db}
}

func (r *deviceRepo) GetDeviceAll(opts ...DBOption) (model.Device, error) {
	var device model.Device
	db := r.db.Model(&model.Device{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&device).Error
	return device, errors.WithMessage(err, "db.First failed")
}

func (r *deviceRepo) GetDevice(opts ...DBOption) (model.Device, error) {
	var device model.Device
	db := r.db.Model(&model.Device{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&device).Error
	return device, errors.WithMessage(err, "db.First failed")
}

// GetDeviceBySn 根据sn获取设备
func (r *deviceRepo) GetDeviceBySn(sn string) (*model.Device, error) {
	device, err := r.GetDevice(
		CommonRepo.WhereByDeviceSn(sn),
	)
	return &device, errors.WithMessage(err, "r.GetDevice failed")
}

func (r *deviceRepo) GetDeviceByUuid(uuid uint64) (*model.Device, error) {
	device, err := r.GetDevice(
		CommonRepo.WhereByDeviceUuid(uuid),
	)
	return &device, errors.WithMessage(err, "r.GetDevice failed")
}

func (r *deviceRepo) GetDeviceBrand(opts ...DBOption) string {
	var brand string
	db := r.db.Model(&model.Device{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Select("brand").Scan(&brand)
	return brand
}

func (r *deviceRepo) GetDeviceSn(opts ...DBOption) string {
	var sn string
	db := r.db.Model(&model.Device{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Select("device_id").Scan(&sn)
	return sn
}

func (r *deviceRepo) WhereSn(deviceId string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("device_id = ?", deviceId)
	}
}

func (r *deviceRepo) WhereSource(source string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("source = ?", source)
	}
}

func (r *deviceRepo) WhereMain() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_main = 1")
	}
}

func (r *deviceRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

func (r *deviceRepo) GetBindCountBySource(source string) uint {
	var count int64
	r.db.Model(&model.Device{}).Scopes(NotDeleted).Where("source = ? AND finally_login_uuid > 0", source).Count(&count)
	return uint(count)
}

func (r *deviceRepo) IsExistCashierMain(source string) bool {
	var count int64
	r.db.Model(&model.Device{}).Scopes(NotDeleted).Where("source = ? AND is_main = 1", source).Count(&count)
	if count > 0 {
		return true
	}
	return false
}

func (r *deviceRepo) UpdateDevice(uuid uint64, vars map[string]interface{}) error {
	return r.db.Model(&model.Device{}).Where("uuid = ?", uuid).Updates(vars).Error
}

func (r *deviceRepo) CreateDevice(device model.Device) (model.Device, error) {
	err := r.db.Model(&model.Device{}).Create(&device).Error
	return device, errors.WithMessage(err)
}

func (r *deviceRepo) GetDeviceList(opts ...DBOption) ([]model.Device, error) {
	var devices []model.Device
	db := r.db.Model(&model.Device{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&devices).Error
	return devices, errors.WithMessage(err, "db.Find failed")
}
