package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type ISaleBillRepo interface {
	GetSaleBill(opts ...DBOption) (model.SaleBill, error)
	GetSaleBillByUuid(uuid uint64) (*model.SaleBill, error)
	GetSaleBillByDeviceUuid(deviceSn uint64) (*model.SaleBill, error)
	UpdateSaleBill(saleBill *model.SaleBill) error
	UpdateSaleBillRecord(saleBill model.SaleBill) error
	UpdateOrCreateSaleBillRecord(saleBill model.SaleBill) error
	UpdateSaleBillShowMustPlan(saleBillUuid uint64) error
	GetHideSaleBillList(pageNo, pageSize int) ([]*model.SaleBill, int64, error) // 获取挂单销售账单列表
	GetInstantSaleBillLatest() (*model.SaleBill, error)                         // 获取最新的一条点餐销售账单
	GetSaleBillRecord(uuid uint64) (*model.SaleBill, error)
}

type saleBillRepo struct {
	db *gorm.DB
}

func NewSaleBillRepo(db *gorm.DB) ISaleBillRepo {
	return &saleBillRepo{db: db}
}

func (r *saleBillRepo) GetSaleBill(opts ...DBOption) (model.SaleBill, error) {
	var saleBill model.SaleBill
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&saleBill)
	if result.Error != nil {
		return saleBill, result.Error
	}

	return saleBill, nil
}

func (r *saleBillRepo) GetSaleBillList(opts ...DBOption) ([]*model.SaleBill, error) {
	var saleBills []*model.SaleBill
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&saleBills)
	if result.Error != nil {
		return saleBills, result.Error
	}

	return saleBills, nil
}

func (r *saleBillRepo) GetSaleBillListPage(pageNo, pageSize int, opts ...DBOption) ([]*model.SaleBill, int64, error) {
	var saleBills []*model.SaleBill
	var total int64
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	if err := db.Model(&model.SaleBill{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	result := db.Order("create_time asc").Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&saleBills)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return saleBills, total, nil
}

func (r *saleBillRepo) GetSaleBillByUuid(uuid uint64) (*model.SaleBill, error) {
	saleBill, err := r.GetSaleBill(CommonRepo.WhereByUuid(uuid))
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &saleBill, nil
}

// 通过deviceSn查询点餐页面未挂单、未结账的账单
func (r *saleBillRepo) GetSaleBillByDeviceUuid(deviceUuid uint64) (*model.SaleBill, error) {
	saleBill, err := r.GetSaleBill(
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByIsHide(false),
		CommonRepo.WhereByDeviceUuid(deviceUuid),
		CommonRepo.WhereByStatus(constant.SaleBillStatusPending))
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &saleBill, nil
}

func (r *saleBillRepo) UpdateSaleBill(saleBill *model.SaleBill) error {
	return r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBill.Uuid).Updates(saleBill).Error
}

// 仅更新sale_bill表
func (r *saleBillRepo) UpdateSaleBillRecord(saleBill model.SaleBill) error {
	saleBill.SetNil() // 将关联对象置空，为了不更新这些关联的对象
	if saleBill.NoPrimaryKey() {
		return errors.New("SaleBill不能没有ID或UUID")
	}
	return r.db.Model(&model.SaleBill{}).Select("*").Where("uuid = ?", saleBill.Uuid).Updates(&saleBill).Error
}

func (r *saleBillRepo) UpdateOrCreateSaleBillRecord(saleBill model.SaleBill) error {
	saleBill.SetNil() // 将关联对象置空，为了不更新这些关联的对象
	if saleBill.ID == 0 {
		return r.db.Model(&model.SaleBill{}).Create(&saleBill).Error
	}
	return r.db.Model(&model.SaleBill{}).Select("*").Where("uuid = ?", saleBill.Uuid).Updates(&saleBill).Error
}

func (r *saleBillRepo) UpdateSaleBillShowMustPlan(saleBillUuid uint64) error {
	return r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBillUuid).Update("show_must_plan", constant.SaleBillShowMustPlanNo).Error
}

func (r *saleBillRepo) GetHideSaleBillList(pageNo, pageSize int) ([]*model.SaleBill, int64, error) {
	var saleBills []*model.SaleBill
	saleBills, total, err := r.GetSaleBillListPage(pageNo, pageSize,
		CommonRepo.WhereByIsHide(true),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
					CommonRepo.DBOption(CommonRepo.WhereByProductIsAccept()),
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
			},
		),
	)
	if err != nil {
		return nil, 0, err
	}
	return saleBills, total, nil
}

func (r *saleBillRepo) GetInstantSaleBillLatest() (*model.SaleBill, error) {
	saleBill, err := r.GetSaleBill(
		CommonRepo.WhereByBillType(constant.OrderSourceMapToBillType[constant.OrderSourceInstant]),
		CommonRepo.SortWithCreateTime("desc"),
	)
	if err != nil {
		if utils.IsNotFoundRecord(err) {
			return nil, nil
		}
		return nil, errors.WithMessage(err)
	}
	return &saleBill, nil
}

func (r *saleBillRepo) GetSaleBillRecord(uuid uint64) (*model.SaleBill, error) {
	saleBill, err := r.GetSaleBill(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &saleBill, nil
}
