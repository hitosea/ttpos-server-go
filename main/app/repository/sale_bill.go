package repository

import (
	"context"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type ISaleBillRepo interface {
	ISaleBillQueryRepo
	UpdateSaleBill(saleBill *model.SaleBill) error
	UpdateSaleBillRecord(saleBill model.SaleBill) error
	UpdateOrCreateSaleBillRecord(saleBill model.SaleBill) error
	UpdateSaleBillShowMustPlan(saleBillUuid uint64) error                      // 确认必点
	UpdateSaleBillAutoAddMustProduct(saleBillUuid uint64) error                // 完成自动加购
	DeleteSaleBill(saleBillUuid uint64) error                                  // 软删除销售账单
	UpdateDutyNo(saleBillUuid uint64, dutyNo string) error                     // 更新销售账单的当班编号
	UpdateSaleBillSerialNo(saleBillUuid uint64, serialNo string) error         // 更新销售账单的流水号
	UpdateSaleBillBatchTagUuid(saleBillUuid uint64, batchTagUuid uint64) error // 更新销售账单的分批类型UUID
	UpdateOrderSource(saleBillUuid uint64, orderSourceUuid uint64) error
	UpdateNationality(saleBillUuid uint64, nationalityUuid uint64) error
}

// ISaleBillQueryRepo 销售账单的查询接口。
type ISaleBillQueryRepo interface {
	GetSaleBill(opts ...DBOption) (model.SaleBill, error)
	GetSaleBillList(opts ...DBOption) ([]*model.SaleBill, error)
	GetSaleBillListPage(pageNo, pageSize int, opts ...DBOption) ([]*model.SaleBill, int64, error)
	GetSaleBillByUuid(uuid uint64) (*model.SaleBill, error)
	GetSaleBillByDeviceUuid(deviceSn uint64) (*model.SaleBill, error)
	GetSaleOrderIndexByUuid(saleBillUuid, saleOrderUuid uint64) (int, error)                       // 获取销售订单的拆单序号。用于操作日志展示
	GetHideSaleBillList(pageNo, pageSize int, deviceUuid uint64) ([]*model.SaleBill, int64, error) // 获取挂单销售账单列表
	GetInstantSaleBillLatest() (*model.SaleBill, error)                                            // 获取最新的一条点餐销售账单
	GetMemberSaleBillLatest() (*model.SaleBill, error)                                             // 获取最新的一条会员端销售账单
	GetSaleBillBuffetProductList(saleBillUuid uint64) (*model.SaleBill, error)                     // 获取销售账单的自助餐商品列表
	GetSaleBillRecord(uuid uint64) (*model.SaleBill, error)
	GetDeskSaleBillUnPay() ([]*model.SaleBill, error) // 获取所有未付款的桌台账单
	GetCompleteTotal() (int64, error)                 // 获取总数量
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
		return nil, 0, errors.WithMessage(err)
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

// GetSaleBillByDeviceUuid 通过deviceSn查询点餐页面未挂单、未结账的账单
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

func (r *saleBillRepo) GetSaleOrderIndexByUuid(saleBillUuid, saleOrderUuid uint64) (int, error) {
	saleBill, err := r.GetSaleBill(
		CommonRepo.WhereByUuid(saleBillUuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
		),
	)
	if err != nil {
		// 如果销售账单不存在，返回-1 . 表示销售账单不存在或已经删除，则不展示这条订单操作记录
		if strings.Contains(err.Error(), gorm.ErrRecordNotFound.Error()) {
			return -1, nil
		}
		return 0, errors.WithMessage(err)
	}
	saleOrder := saleBill.GetSaleOrder(saleOrderUuid)
	if saleOrder == nil {
		// 如果销售订单不存在，返回-1 . 表示销售订单不存在或已经删除，则不展示这条订单操作记录
		return -1, nil
	}
	if !saleBill.IsSplit() {
		// 如果销售单没有拆单，返回0。操作记录不显示拆单序号前缀
		return 0, nil
	}

	// 获取销售单的拆单序号
	return saleOrder.GetIndex(), nil
}

func (r *saleBillRepo) UpdateSaleBill(saleBill *model.SaleBill) error {
	return r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBill.Uuid).Updates(saleBill).Error
}

// UpdateSaleBillRecord 仅更新sale_bill表
func (r *saleBillRepo) UpdateSaleBillRecord(saleBill model.SaleBill) error {
	saleBill.SetNil() // 将关联对象置空，为了不更新这些关联的对象
	if saleBill.NoPrimaryKey() {
		return errors.New("SaleBill不能没有ID或UUID")
	}
	// 2. 创建上下文
	ctx := context.WithValue(context.Background(), constant.OrderOperateSource, saleBill.GetOperateSource())
	// 3. 更新
	return r.db.WithContext(ctx).Model(&model.SaleBill{}).Select("*").Where("uuid = ?", saleBill.Uuid).Updates(&saleBill).Error
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

// UpdateSaleBillAutoAddMustProduct 设置账单已经完成自动加购
func (r *saleBillRepo) UpdateSaleBillAutoAddMustProduct(saleBillUuid uint64) error {
	return r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBillUuid).Update("auto_add_must_product", 0).Error
}

func (r *saleBillRepo) GetHideSaleBillList(pageNo, pageSize int, deviceUuid uint64) ([]*model.SaleBill, int64, error) {
	var saleBills []*model.SaleBill
	saleBills, total, err := r.GetSaleBillListPage(pageNo, pageSize,
		CommonRepo.WhereByIsHide(true),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByStatus(constant.SaleBillStatusPending),
		CommonRepo.WhereByDeviceUuid(deviceUuid),
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
		return nil, 0, errors.WithMessage(err)
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

func (r *saleBillRepo) GetMemberSaleBillLatest() (*model.SaleBill, error) {
	saleBill, err := r.GetSaleBill(
		CommonRepo.WhereByBillType(constant.OrderSourceMapToBillType[constant.OrderSourceMember]),
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

func (r *saleBillRepo) GetSaleBillBuffetProductList(saleBillUuid uint64) (*model.SaleBill, error) {
	saleBill, err := r.GetSaleBill(
		CommonRepo.WhereByUuid(saleBillUuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "BuffetPackage1.BuffetProducts.ProductPackage",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "BuffetPackage1.BuffetProducts",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "BuffetPackage1.BuffetProducts.ProductPackage",
			},
			WithPreload{
				Query: "BuffetPackage2",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "BuffetPackage2.BuffetProducts",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "BuffetPackage2.BuffetProducts.ProductPackage",
			},
		),
	)
	if err != nil {
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

func (r *saleBillRepo) DeleteSaleBill(saleBillUuid uint64) error {
	return r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBillUuid).Update("delete_time", time.Now().Unix()).Error
}

// GetDeskSaleBillUnPay 获取所有待付款的桌台账单
func (r *saleBillRepo) GetDeskSaleBillUnPay() ([]*model.SaleBill, error) {
	saleBills, err := r.GetSaleBillList(
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByStatus(constant.SaleBillStatusPending),
		CommonRepo.WhereByBillType(constant.OrderSourceMapToBillType[constant.OrderSourceDesk]),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return saleBills, nil
}

// GetCompleteTotal 获取总数量
func (r *saleBillRepo) GetCompleteTotal() (int64, error) {
	var total int64
	err := r.db.Model(&model.SaleBill{}).Where("delete_time = 0").Where("status = ?", constant.SaleBillStatusComplete).Count(&total).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return total, nil
}

// UpdateDutyNo 更新销售账单的当班编号
func (r *saleBillRepo) UpdateDutyNo(saleBillUuid uint64, dutyNo string) error {
	if err := r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBillUuid).Updates(model.SaleBill{
		DutyNo: dutyNo,
	}).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateSaleBillSerialNo 更新销售账单的流水号
func (r *saleBillRepo) UpdateSaleBillSerialNo(saleBillUuid uint64, serialNo string) error {
	return r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBillUuid).Updates(model.SaleBill{
		SerialNo: serialNo,
	}).Error
}

// UpdateSaleBillBatchTagUuid 更新销售账单的分批类型UUID
func (r *saleBillRepo) UpdateSaleBillBatchTagUuid(saleBillUuid uint64, batchTagUuid uint64) error {
	return r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBillUuid).Updates(model.SaleBill{
		BatchTagUuid: batchTagUuid,
	}).Error
}

// UpdateOrderSource 更新销售账单的订单来源
func (r *saleBillRepo) UpdateOrderSource(saleBillUuid uint64, orderSourceUuid uint64) error {
	return r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBillUuid).Updates(model.SaleBill{
		OrderSourceUuid: orderSourceUuid,
	}).Error
}

// UpdateNationality 更新销售账单的国籍
func (r *saleBillRepo) UpdateNationality(saleBillUuid uint64, nationalityUuid uint64) error {
	return r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBillUuid).Updates(model.SaleBill{
		NationalityUuid: nationalityUuid,
	}).Error
}
