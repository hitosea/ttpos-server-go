package repository

import (
	"fmt"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/context"

	"gorm.io/gorm"
)

// IDeskRepo 桌台
type IDeskRepo interface {
	IDeskQueryRepo
	UpdateDesk(deskUuid uint64, desk model.Desk) error
	UpdateDeskRecord(desk model.Desk) error
	UpdateDeskByMap(deskUuid uint64, vars map[string]any) error // 更新桌台
	UnbindDesk(deskUuid, deviceUuid uint64) error               // 平板端解绑桌台
	CreateDesk(desk model.Desk) (uint64, error)
	DeleteDesk(deskUuid uint64) error
	CloseDesk(ctx context.Context, deskUuid, saleBillUuid uint64, reason string) error
	WhereUuid(uuid uint64) DBOption
	WhereDeviceUuid(uuid uint64) DBOption
	WhereIsDisable(isDisable int) DBOption
}

// IDeskQueryRepo 桌台查询
type IDeskQueryRepo interface {
	GetDeskList(pageNo, pageSize int) ([]model.Desk, int64, error)
	GetDeskAndSaleBillByDeskUuid(deskUuid uint64) (model.Desk, error) // 通过桌台ID获取桌台信息和销售账单信息
	GetSaleBillUuidByDeskUuid(deskUuid uint64) (uint64, error)        // 通过桌台ID获取销售账单UUID（仅用于锁机制）
	GetClientDeskList(source string, status, isBuffet, pageNo, pageSize int) ([]model.Desk, int64, error)
	GetDesk(opts ...DBOption) (model.Desk, error) // 获取桌台
	GetDesks(opts ...DBOption) ([]*model.Desk, error)
	GetSaleBillUuidAndSaleOrderUuid(deskUuid uint64) (uint64, uint64, error) // 获取桌台的账单uuid和第一子单的uuid
	GetAvailableDeskList() ([]*model.Desk, error)                            // 获取所有空闲的桌台
	GetDeskInfo(deskUuid uint64, opts ...DBOption) (model.Desk, error)       //
	GetDeskRecord(deskUuid uint64) (*model.Desk, error)                      // 通过uuid获取桌台的记录信息
	GetDeskCountsByRegion() (map[uint64]int64, error)                        // 获取按区域分组的桌台数量
	GetDesksByRegionUuid(regionUuid uint64) ([]model.Desk, error)            // 按区域查询桌台列表
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
		return nil, 0, errors.WithMessage(err)
	}

	// 获取分页数据
	err := query.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&desks).Error
	return desks, total, errors.WithMessage(err)
}

// GetClientDeskList 获取客户端桌台列表，排除逻辑删除的桌台，排除被禁用的桌台
func (r *deskRepo) GetClientDeskList(source string, status, isBuffet, pageNo, pageSize int) ([]model.Desk, int64, error) {
	var desks []model.Desk
	var total int64

	tablePrefix := config.Database.TablePrefix

	query := r.db.Model(&model.Desk{}).
		Joins(fmt.Sprintf("LEFT JOIN %ssale_bill ON %sdesk.sale_bill_uuid = %ssale_bill.uuid", tablePrefix, tablePrefix, tablePrefix)).
		Preload("SaleBill").
		Preload("SaleBill.BatchTag").
		Preload("SaleBill.SaleOrders", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted).Order("create_time asc")
		}).
		Where(fmt.Sprintf("%sdesk.delete_time = ?", tablePrefix), 0).
		Where(fmt.Sprintf("%sdesk.is_disable = ?", tablePrefix), constant.DeskEnable)

	if status != -1 {
		if status == 2 {
			query = query.Where(fmt.Sprintf("%sdesk.status = ?", tablePrefix), constant.DeskStatusOpen)
			query = query.Where(fmt.Sprintf("%sdesk.sale_bill_uuid <> ?", tablePrefix), 0)
		} else {
			query = query.Where(fmt.Sprintf("%sdesk.status = ?", tablePrefix), status)
		}
	}

	if isBuffet != -1 {
		if isBuffet == 1 {
			query = query.Where(fmt.Sprintf("%ssale_bill.is_buffet = ?", tablePrefix), 1)
		} else {
			query = query.Where(fmt.Sprintf("%ssale_bill.is_buffet = ?", tablePrefix), 0)
		}
	}

	// 平板端，筛选未绑定桌台的
	if source == constant.SourceTablet {
		query = query.Where(fmt.Sprintf("%sdesk.device_uuid = 0", tablePrefix))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	err := query.Order(fmt.Sprintf("%sdesk.sort asc", tablePrefix)).
		Select(fmt.Sprintf("%sdesk.*", tablePrefix)).
		Offset((pageNo - 1) * pageSize).
		Limit(pageSize).
		Find(&desks).Error

	return desks, total, errors.WithMessage(err)
}

func (r *deskRepo) GetDesk(opts ...DBOption) (model.Desk, error) {
	var desk model.Desk
	db := r.db.Model(&model.Desk{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&desk).Error
	return desk, errors.WithMessage(err)
}

func (r *deskRepo) GetDesks(opts ...DBOption) ([]*model.Desk, error) {
	var desks []*model.Desk
	db := r.db.Model(&model.Desk{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("sort asc").Find(&desks).Error
	return desks, errors.WithMessage(err)
}

// GetSaleBillUuidAndSaleOrderUuid 获取桌台的账单uuid和第一子单的uuid
func (r *deskRepo) GetSaleBillUuidAndSaleOrderUuid(deskUuid uint64) (uint64, uint64, error) {
	var desk model.Desk
	desk, err := r.GetDesk(
		CommonRepo.WhereByUuid(deskUuid),
		CommonRepo.WhereByStatus(constant.DeskStatusOpen), // 只查询开台的桌台.只有开台的桌台才有账单和子单
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleBill.SaleOrders",
			},
		),
	)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return 0, 0, errors.WithMessage(errors.New("桌台已关闭"))
		}
		return 0, 0, errors.WithMessage(err)
	}
	if desk.SaleBill == nil {
		return 0, 0, errors.WithMessage(errors.New("桌台没有账单"))
	}
	if len(desk.SaleBill.SaleOrders) == 0 {
		return 0, 0, errors.WithMessage(errors.New("桌台没有子单"))
	}
	return desk.SaleBill.Uuid, desk.SaleBill.SaleOrders[0].Uuid, nil
}

// GetAvailableDeskList 获取所有空闲的桌台
func (r *deskRepo) GetAvailableDeskList() ([]*model.Desk, error) {
	desks, err := r.GetDesks(CommonRepo.WhereByStatus(constant.DeskStatusClose))
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return desks, nil
}

// GetDeskInfo 获取桌台信息
func (r *deskRepo) GetDeskInfo(deskUuid uint64, opts ...DBOption) (model.Desk, error) {
	var desk model.Desk

	db := r.db.Model(&model.Desk{}).Where("uuid = ?", deskUuid)

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Preload("SaleBill.SaleBillSetting").Preload("SaleBill.BatchTag").Preload("SaleBill.SaleOrders", func(db *gorm.DB) *gorm.DB {
		return db.Scopes(NotDeleted).Order("create_time asc")
	}).First(&desk)
	if result.Error != nil {
		return desk, result.Error
	}

	return desk, nil
}

func (r *deskRepo) GetDeskRecord(deskUuid uint64) (*model.Desk, error) {
	desk, err := r.GetDesk(
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByUuid(deskUuid),
		//CommonRepo.WhereByStatus(constant.DeskStatusClose),
		CommonRepo.WhereByNoDisable(),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &desk, nil
}

// UpdateDesk 更新桌台
func (r *deskRepo) UpdateDesk(deskUuid uint64, desk model.Desk) error {
	desk.SetNil()
	if err := r.db.Model(&model.Desk{}).Where("uuid = ?", deskUuid).Updates(desk).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateDeskRecord 更新桌台记录
func (r *deskRepo) UpdateDeskRecord(desk model.Desk) error {
	desk.SetNil() // 将关联对象置空，为了不更新这些关联的对象
	if desk.NoPrimaryKey() {
		return errors.New("Desk不能没有ID或UUID")
	}
	return r.db.Model(&model.Desk{}).Select("*").Where("uuid = ?", desk.Uuid).Updates(&desk).Error
}

// UpdateDeskByMap 更新桌台
func (r *deskRepo) UpdateDeskByMap(deskUuid uint64, vars map[string]any) error {
	if err := r.db.Model(&model.Desk{}).Where("uuid = ?", deskUuid).Updates(vars).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UnbindDesk 解绑桌台
func (r *deskRepo) UnbindDesk(deskUuid, deviceUuid uint64) error {
	if err := r.db.Model(&model.Desk{}).Where("uuid <> ? AND device_uuid = ?", deskUuid, deviceUuid).
		Updates(map[string]any{"device_uuid": 0}).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// CreateDesk 创建桌台
func (r *deskRepo) CreateDesk(desk model.Desk) (uint64, error) {
	// 创建桌台
	if err := r.db.Create(&desk).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	return desk.Uuid, nil
}

// DeleteDesk 软删除桌台
func (r *deskRepo) DeleteDesk(deskUuid uint64) error {
	return r.db.Model(&model.Desk{}).Where("uuid = ?", deskUuid).Update("delete_time", uint(time.Now().Unix())).Error
}

// CloseDesk 关闭桌台
func (r *deskRepo) CloseDesk(ctx context.Context, deskUuid, saleBillUuid uint64, reason string) error {
	err := NewOrderRepo(r.db).CancelDeskOrder(ctx, deskUuid, reason)
	if err != nil {
		return errors.WithMessage(err)
	}
	// 删除销售账单
	if saleBillUuid > 0 {
		if err := r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBillUuid).Update("delete_time", uint(time.Now().Unix())).Error; err != nil {
			return errors.WithMessage(err)
		}
	}
	// 更新桌台
	return r.db.Model(&model.Desk{}).Where("uuid = ?", deskUuid).Updates(map[string]any{
		"uuid":           deskUuid,
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

// WhereIsDisable 开关开启，桌台未被禁用
func (r *deskRepo) WhereIsDisable(isDisable int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_disable = ?", isDisable)
	}
}

func (r *deskRepo) GetDeskAndSaleBillByDeskUuid(deskUuid uint64) (model.Desk, error) {
	var desk model.Desk
	return desk, r.db.Model(&model.Desk{}).Where("uuid = ? AND delete_time = ?", deskUuid, constant.NotDeleted).Preload("SaleBill", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", constant.SaleBillStatusPending)
	}).First(&desk).Error
}

// GetSaleBillUuidByDeskUuid 通过桌台ID获取销售账单UUID（仅用于锁机制）
// 这个方法只查询 sale_bill_uuid 字段，避免加载完整对象，提高性能
func (r *deskRepo) GetSaleBillUuidByDeskUuid(deskUuid uint64) (uint64, error) {
	var saleBillUuid uint64
	err := r.db.Model(&model.Desk{}).
		Select("sale_bill_uuid").
		Where("uuid = ? AND delete_time = ?", deskUuid, constant.NotDeleted).
		Scan(&saleBillUuid).Error
	if err != nil {
		return 0, errors.WithMessage(err, "获取桌台的销售账单UUID失败")
	}
	return saleBillUuid, nil
}

// GetDeskCountsByRegion 获取按区域分组的桌台数量
func (r *deskRepo) GetDeskCountsByRegion() (map[uint64]int64, error) {
	var results []struct {
		RegionUuid uint64
		Count      int64
	}

	err := r.db.Model(&model.Desk{}).
		Select("region_uuid, COUNT(*) as count").
		Where("delete_time = ?", 0).
		Group("region_uuid").
		Find(&results).Error

	if err != nil {
		return nil, errors.WithMessage(err, "查询桌台数量失败")
	}

	// 转换为 map
	countMap := make(map[uint64]int64)
	for _, result := range results {
		countMap[result.RegionUuid] = result.Count
	}

	return countMap, nil
}

// GetDesksByRegionUuid 按区域查询桌台列表
func (r *deskRepo) GetDesksByRegionUuid(regionUuid uint64) ([]model.Desk, error) {
	var desks []model.Desk

	err := r.db.Model(&model.Desk{}).
		Where("region_uuid = ? AND delete_time = ?", regionUuid, 0).
		Find(&desks).Error

	if err != nil {
		return nil, errors.WithMessage(err, "查询桌台列表失败")
	}

	return desks, nil
}
