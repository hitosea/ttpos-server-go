package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IDeskRepo 桌台
type IDeskRepo interface {
	GetDeskList(pageNo, pageSize int) ([]model.Desk, int64, error)
	GetDeskAndSaleBillByDeskUuid(deskUuid uint64) (model.Desk, error) // 通过桌台ID获取桌台信息和销售账单信息
	GetClientDeskList(pageNo, pageSize int) ([]model.Desk, int64, error)
	GetDesk(opts ...DBOption) (model.Desk, error) // 获取桌台
	GetDesks(opts ...DBOption) ([]model.Desk, error)
	GetDeskInfo(deskUuid uint64, opts ...DBOption) (model.Desk, error)
	UpdateDesk(deskUuid uint64, desk model.Desk) error
	UpdateDeskByMap(deskUuid uint64, vars map[string]any) error // 更新桌台
	UnbindDesk(deskUuid, deviceUuid uint64) error               // 平板端解绑桌台
	CreateDesk(desk model.Desk) (uint64, error)
	DeleteDesk(deskUuid uint64) error
	CloseDesk(deskUuid uint64, reason string) error
	WhereUuid(uuid uint64) DBOption
	WhereDeviceUuid(uuid uint64) DBOption
	WhereIsBind() DBOption
	WhereIsNotDisable() DBOption
}

func NewDeskRepo(db *gorm.DB) IDeskRepo {
	return NewDeskRepoImpl(db)
}

// NewDeskRepoImpl 创建新的桌台仓库实现
func NewDeskRepoImpl(db *gorm.DB) IDeskRepo {
	return &deskRepo{db: db}
}

type deskRepo struct {
	db *gorm.DB
}

// GetDeskList 获取桌台列表，排除逻辑删除的桌台
func (r *deskRepo) GetDeskList(pageNo, pageSize int) ([]model.Desk, int64, error) {
	var desks []model.Desk
	var total int64

	query := r.db.Model(&model.Desk{}).Where("delete_time = ?", 0)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&desks).Error
	return desks, total, err
}

// GetClientDeskList 获取客户端桌台列表，排除逻辑删除的桌台，排除被禁用的桌台
func (r *deskRepo) GetClientDeskList(pageNo, pageSize int) ([]model.Desk, int64, error) {
	var desks []model.Desk
	var total int64

	query := r.db.Model(&model.Desk{}).Preload("SaleBill").Where("delete_time = ?", 0)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Order("sort asc").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&desks).Error

	return desks, total, err
}

func (r *deskRepo) GetDesk(opts ...DBOption) (model.Desk, error) {
	var desk model.Desk
	db := r.db.Model(&model.Desk{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&desk).Error
	return desk, err
}

func (r *deskRepo) GetDesks(opts ...DBOption) ([]model.Desk, error) {
	var desks []model.Desk
	db := r.db.Model(&model.Desk{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("sort desc").Find(&desks).Error
	return desks, err
}

// GetDeskInfo 获取桌台信息
func (r *deskRepo) GetDeskInfo(deskUuid uint64, opts ...DBOption) (model.Desk, error) {
	var desk model.Desk

	db := r.db.Model(&model.Desk{}).Where("uuid = ?", deskUuid)

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Preload("SaleBill", "status = ?", constant.SaleBillStatusPending).First(&desk)
	if result.Error != nil {
		return desk, result.Error
	}

	return desk, nil
}

// UpdateDesk 更新桌台
func (r *deskRepo) UpdateDesk(deskUuid uint64, desk model.Desk) error {
	desk.SetNil()
	if err := r.db.Model(&model.Desk{}).Where("uuid = ?", deskUuid).Updates(desk).Error; err != nil {
		return err
	}
	return nil
}

// UpdateDeskByMap 更新桌台
func (r *deskRepo) UpdateDeskByMap(deskUuid uint64, vars map[string]any) error {
	if err := r.db.Model(&model.Desk{}).Where("uuid = ?", deskUuid).Updates(vars).Error; err != nil {
		return err
	}
	return nil
}

// UnbindDesk 解绑桌台
func (r *deskRepo) UnbindDesk(deskUuid, deviceUuid uint64) error {
	if err := r.db.Model(&model.Desk{}).Where("uuid = <> ? AND device_uuid = ?", deskUuid, deviceUuid).
		Updates(map[string]any{"device_uuid": 0}).Error; err != nil {
		return err
	}
	return nil
}

// CreateDesk 创建桌台
func (r *deskRepo) CreateDesk(desk model.Desk) (uint64, error) {
	// 创建桌台
	if err := r.db.Create(&desk).Error; err != nil {
		return 0, err
	}
	return desk.Uuid, nil
}

// DeleteDesk 软删除桌台
func (r *deskRepo) DeleteDesk(deskUuid uint64) error {
	return r.db.Model(&model.Desk{}).Where("uuid = ?", deskUuid).Update("delete_time", uint(time.Now().Unix())).Error
}

// CloseDesk 关闭桌台
func (r *deskRepo) CloseDesk(deskUuid uint64, reason string) error {
	err := NewOrderRepo(r.db).CancelDeskOrder(deskUuid, reason)
	if err != nil {
		return err
	}
	return r.db.Model(&model.Desk{}).Where("uuid = ?", deskUuid).Updates(map[string]any{
		"status":         constant.DeskStatusClose,
		"sale_bill_uuid": 0,
	}).Error
}

// WhereUuid 桌台uuid条件
func (r *deskRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereDeviceUuid 桌台绑定的设备uuid条件
func (r *deskRepo) WhereDeviceUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("device_uuid = ?", uuid)
	}
}

// WhereIsBind 桌台绑定条件
func (r *deskRepo) WhereIsBind() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("device_uuid = 0")
	}
}

// WhereIsNotDisable 开关开启，桌台未被禁用
func (r *deskRepo) WhereIsNotDisable() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_disable = 0")
	}
}

func (r *deskRepo) GetDeskAndSaleBillByDeskUuid(deskUuid uint64) (model.Desk, error) {
	var desk model.Desk
	return desk, r.db.Model(&model.Desk{}).Where("uuid = ? AND delete_time = ?", deskUuid, constant.NotDeleted).Preload("SaleBill", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", constant.SaleBillStatusPending)
	}).First(&desk).Error
}
