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
	GetDesk(opts ...DBOption) (*model.Desk, error) // 获取桌台
	GetDeskInfo(deskUuid uint64, opts ...DBOption) (model.Desk, error)
	UpdateDesk(deskUuid uint64, desk model.Desk) error
	CreateDesk(desk model.Desk) (uint64, error)
	DeleteDesk(deskUuid uint64) error
	CloseDesk(deskUuid uint64, reason string) error
}

func NewDeskRepo(db *gorm.DB) IDeskRepo {
	return NewDeskRepoImpl(db)
}

// NewDeskRepoImpl 创建新的桌台仓库实现
func NewDeskRepoImpl(db *gorm.DB) *DeskRepoImpl {
	return &DeskRepoImpl{db: db}
}

type DeskRepoImpl struct {
	db *gorm.DB
}

// GetDeskList 获取桌台列表，排除逻辑删除的桌台
func (r *DeskRepoImpl) GetDeskList(pageNo, pageSize int) ([]model.Desk, int64, error) {
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
func (r *DeskRepoImpl) GetClientDeskList(pageNo, pageSize int) ([]model.Desk, int64, error) {
	var desks []model.Desk
	var total int64

	query := r.db.Model(&model.Desk{}).Preload("SaleBill").Where("delete_time = ?", 0).Where("is_disable = ?", 1)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := query.Order("sort asc").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&desks).Error

	return desks, total, err
}

func (r *DeskRepoImpl) GetDesk(opts ...DBOption) (*model.Desk, error) {
	var desk model.Desk
	db := r.db.Model(&model.Desk{})

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&desk)
	if result.Error != nil {
		return nil, result.Error
	}

	return &desk, nil
}

// GetDeskInfo 获取桌台信息
func (r *DeskRepoImpl) GetDeskInfo(deskUuid uint64, opts ...DBOption) (model.Desk, error) {
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
func (r *DeskRepoImpl) UpdateDesk(deskUuid uint64, desk model.Desk) error {
	if err := r.db.Model(&model.Desk{}).Where("uuid = ?", deskUuid).Updates(desk).Error; err != nil {
		return err
	}
	return nil
}

// CreateDesk 创建桌台
func (r *DeskRepoImpl) CreateDesk(desk model.Desk) (uint64, error) {
	// 创建桌台
	if err := r.db.Create(&desk).Error; err != nil {
		return 0, err
	}
	return desk.Uuid, nil
}

// DeleteDesk 软删除桌台
func (r *DeskRepoImpl) DeleteDesk(deskUuid uint64) error {
	return r.db.Model(&model.Desk{}).Where("uuid = ?", deskUuid).Update("delete_time", uint(time.Now().Unix())).Error
}

// CloseDesk 关闭桌台
func (r *DeskRepoImpl) CloseDesk(deskUuid uint64, reason string) error {
	err := NewOrderRepo(r.db).CancelDeskOrder(deskUuid, reason)
	if err != nil {
		return err
	}
	return r.db.Model(&model.Desk{}).Where("uuid = ?", deskUuid).Update("status", 1).Error
}

func (r *DeskRepoImpl) GetDeskAndSaleBillByDeskUuid(deskUuid uint64) (model.Desk, error) {
	var desk model.Desk
	return desk, r.db.Model(&model.Desk{}).Where("uuid = ? AND delete_time = ?", deskUuid, constant.NotDeleted).Preload("SaleBill", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", constant.SaleBillStatusPending)
	}).First(&desk).Error
}
