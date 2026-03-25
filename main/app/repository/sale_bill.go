package repository

import (
	"context"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type ISaleBillRepo interface {
	ISaleBillQueryRepo
	UpdateSaleBill(saleBill *model.SaleBill) error
	UpdateSaleBillRecord(saleBill model.SaleBill) error
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
	WhereKeyword(keyword string, language string) DBOption
	GetSaleBillByUuid(uuid uint64) (*model.SaleBill, error)
	GetSaleBillByDeviceUuid(deviceSn uint64) (*model.SaleBill, error)
	GetSaleOrderIndexByUuid(saleBillUuid, saleOrderUuid uint64) (int, error)                                         // 获取销售订单的拆单序号。用于操作日志展示
	GetHideSaleBillList(pageNo, pageSize int, deviceUuid uint64, opts ...DBOption) ([]*model.SaleBill, int64, error) // 获取挂单销售账单列表
	WhereBySerialNo(keyword string) DBOption                                                                         // 根据流水号模糊搜索
	GetInstantSaleBillLatest() (*model.SaleBill, error)                                                              // 获取最新的一条点餐销售账单
	GetMemberSaleBillLatest() (*model.SaleBill, error)                                                               // 获取最新的一条会员端销售账单
	GetSaleBillBuffetProductList(saleBillUuid uint64) (*model.SaleBill, error)                                       // 获取销售账单的自助餐商品列表
	GetSaleBillRecord(uuid uint64) (*model.SaleBill, error)
	GetDeskSaleBillUnPay() ([]*model.SaleBill, error)                               // 获取所有未付款的桌台账单
	GetCompleteTotal() (int64, error)                                               // 获取总数量
	PluckOrderNos(orderNoPrefix string, startTime, endTime int64) ([]string, error) // 查询订单编号
	PluckPendingUuidsWithBatchTag() ([]uint64, error)                               // 获取有分批类型的进行中账单UUID列表
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

func (r *saleBillRepo) GetHideSaleBillList(pageNo, pageSize int, deviceUuid uint64, opts ...DBOption) ([]*model.SaleBill, int64, error) {
	var saleBills []*model.SaleBill
	queryOpts := []DBOption{
		CommonRepo.WhereByIsHide(true),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByStatus(constant.SaleBillStatusPending),
		// CommonRepo.WhereByDeviceUuid(deviceUuid), 产品定义会员端堂食订单先下单后付款时需要在挂单列表中显示
	}
	queryOpts = append(queryOpts, opts...)
	queryOpts = append(queryOpts, CommonRepo.Preload(
		WithPreload{
			Query: "SaleOrders",
			Args: []any{
				CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
			},
		},
		WithPreload{
			Query: "SaleOrders.SaleOrderProducts",
			Args: []any{
				CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				CommonRepo.DBOption(CommonRepo.WhereByProductIsAccept()),
			},
		},
		WithPreload{
			Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
		},
	))
	saleBills, total, err := r.GetSaleBillListPage(pageNo, pageSize, queryOpts...)
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

// PluckOrderNos 查询指定前缀和时间范围内的订单编号
func (r *saleBillRepo) PluckOrderNos(orderNoPrefix string, startTime, endTime int64) ([]string, error) {
	var orderNos []string
	err := r.db.Model(&model.SaleBill{}).
		Where("order_no LIKE ?", orderNoPrefix+"%").
		Where("create_time >= ? AND create_time <= ?", startTime, endTime).
		Pluck("order_no", &orderNos).Error
	return orderNos, err
}

// PluckPendingUuidsWithBatchTag 获取有分批类型的进行中账单UUID列表
func (r *saleBillRepo) PluckPendingUuidsWithBatchTag() ([]uint64, error) {
	var uuids []uint64
	err := r.db.Model(&model.SaleBill{}).
		Where("status = ?", constant.SaleBillStatusPending).
		Where("delete_time = ?", 0).
		Where("batch_tag_uuid <> ?", 0).
		Select("uuid").Scan(&uuids).Error
	return uuids, err
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
// JSON 方案：同时保存 order_source_uuid 和 order_source_name 快照（包含所有语言）
// Requirement: story-main-order-source-snapshot-fix
func (r *saleBillRepo) UpdateOrderSource(saleBillUuid uint64, orderSourceUuid uint64) error {
	saleBill := model.SaleBill{OrderSourceUuid: orderSourceUuid}
	// 表示取消设置外卖来源. 方法可以给外卖来源UUID为0，表示取消设置外卖来源. 非0表示设置外卖来源.
	if orderSourceUuid == 0 {
		saleBill.OrderSourceName = ""
	} else {
		// 1. 查询外卖来源信息（包括多语言数据）
		orderSourceRepo := NewOrderSourceRepo(r.db)
		orderSource, err := orderSourceRepo.FindByUuid(orderSourceUuid)
		if err != nil {
			return errors.WithMessage(err, "查询外卖来源信息失败")
		}
		if orderSource == nil {
			return errors.WithMessage(errors.New("外卖来源不存在"))
		}

		// 2. 序列化为 JSON 快照
		if err := saleBill.SetOrderSourceNameSnapshot(orderSource.MultiLanguageName); err != nil {
			return errors.WithMessage(err, "序列化外卖来源快照失败")
		}
	}

	// 3. 同时更新 order_source_uuid 和 order_source_name
	return r.db.Model(&model.SaleBill{}).
		Select("order_source_uuid", "order_source_name").
		Where("uuid = ?", saleBillUuid).
		Updates(saleBill).Error
}

// UpdateNationality 更新销售账单的国籍
// JSON 方案：同时保存 nationality_uuid 和 nationality_name 快照（包含所有语言）
// Requirement: story-main-nationality-snapshot-fix
func (r *saleBillRepo) UpdateNationality(saleBillUuid uint64, nationalityUuid uint64) error {
	saleBill := model.SaleBill{NationalityUuid: nationalityUuid}
	// 表示取消设置国籍. 方法可以给国籍UUID为0，表示取消设置国籍. 非0表示设置国籍.
	if nationalityUuid == 0 {
		saleBill.NationalityName = ""
	} else {
		// 1. 查询国籍信息（包括多语言数据）
		nationalityRepo := NewNationalityRepo(r.db)
		nationality, err := nationalityRepo.FindByUuid(nationalityUuid)
		if err != nil {
			return errors.WithMessage(err, "查询国籍信息失败")
		}
		if nationality == nil {
			return errors.WithMessage(errors.New("国籍不存在"))
		}

		// 2. 序列化为 JSON 快照
		if err := saleBill.SetNationalityNameSnapshot(nationality.MultiLanguageName); err != nil {
			return errors.WithMessage(err, "序列化国籍快照失败")
		}
	}

	// 3. 同时更新 nationality_uuid 和 nationality_name
	return r.db.Model(&model.SaleBill{}).
		Select("nationality_uuid", "nationality_name").
		Where("uuid = ?", saleBillUuid).
		Updates(saleBill).Error
}

// WhereKeyword 根据订单号或商品名称搜索销售账单
// 支持搜索：1) 正常订单的商品名称 2) 拒单订单的商品快照名称
func (r *saleBillRepo) WhereKeyword(keyword string, language string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" {
			return db
		}

		// 安全列名映射，防止 SQL 注入
		columnName, ok := languageColumnMap[language]
		if !ok {
			columnName = "en_name"
		}

		// 获取表前缀
		prefix := config.Database.TablePrefix

		// 子查询1：通过商品名称查找对应的 sale_bill uuid（正常订单）
		// sale_order_product -> sale_order.sale_bill_uuid
		subQuery1 := r.db.Table(prefix+"sale_order so").
			Select("DISTINCT so.sale_bill_uuid").
			Joins("INNER JOIN "+prefix+"sale_order_product sop ON sop.sale_order_uuid = so.uuid AND sop.delete_time = 0").
			Joins("LEFT JOIN "+prefix+"multi_language_name mln ON sop.multi_language_name_uuid = mln.uuid").
			Where("so.delete_time = ?", 0).
			Where("mln."+columnName+" LIKE ?", "%"+keyword+"%")

		// 子查询2：通过 H5 订单商品快照名称查找（拒单订单）
		// h5_order_product.name 存储的是拒单时的商品名称快照
		subQuery2 := r.db.Table(prefix+"h5_order_product hop").
			Select("DISTINCT hop.sale_bill_uuid").
			Where("hop.delete_time = ?", 0).
			Where("hop.name LIKE ?", "%"+keyword+"%")

		return db.Where("(uuid IN (?) OR uuid IN (?) OR order_no LIKE ?)", subQuery1, subQuery2, "%"+keyword+"%")
	}
}

// WhereBySerialNo 根据流水号模糊搜索
func (r *saleBillRepo) WhereBySerialNo(keyword string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" {
			return db
		}
		return db.Where("serial_no LIKE ?", Like(keyword))
	}
}
