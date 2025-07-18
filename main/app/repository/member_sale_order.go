package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMemberSaleOrderRepo interface {
	IQueryMemberSaleOrderRepo
	CreateMemberSaleOrder(memberSaleOrder model.MemberSaleOrder) error                // 创建会员端销售订单
	UpdateMemberSaleOrderVerifiedPhoneStatus(memberSaleOrderUuid uint64) error        // 更新会员端销售订单的手机号验证状态为已验证
	UpdateMemberSaleOrderPendingPayment(memberSaleOrder *model.MemberSaleOrder) error // 更新会员端销售订单为待支付状态
	UpdateMemberSaleOrderPendingMerchantAccept(memberSaleOrderUuid uint64) error      // 更新会员端销售订单为“待商家接单”状态，表示订单支付成功，等待商家接单
}

type IQueryMemberSaleOrderRepo interface {
	GetMemberSaleOrder(opts ...DBOption) (*model.MemberSaleOrder, error)                                                                                             // 获取会员端销售订单
	GetMemberSaleOrderRecord(uuid uint64, opts ...DBOption) (*model.MemberSaleOrder, error)                                                                          // 获取会员端销售订单记录
	PaginateGetMemberSaleOrder(pageNo, pageSize int, opts ...DBOption) ([]model.MemberSaleOrder, int64, error)                                                       // 分页获取会员端销售订单
	GetCashierMemberSaleOrderList(pageNo, pageSize int, statusList []uint) ([]model.MemberSaleOrder, int64, error)                                                   // 获取收银台"外送"订单列表
	GetCashierMemberSaleOrderManageList(pageNo, pageSize int, statusList []uint, req GetCashierMemberSaleOrderManageListReq) ([]model.MemberSaleOrder, int64, error) // 获取收银台"外送"订单管理列表
	UpdateMemberSaleOrderAccept(memberSaleOrder model.MemberSaleOrder) error                                                                                         // 更新会员端销售订单-接单
	UpdateMemberSaleOrderReject(memberSaleOrder model.MemberSaleOrder) error                                                                                         // 更新会员端销售订单-拒单
	UpdateMemberSaleOrderCookFinish(memberSaleOrder model.MemberSaleOrder) error                                                                                     // 更新会员端销售订单-备餐完成
	UpdateDeliveryDistance(memberSaleOrderUuid uint64, distance float64) error                                                                                       // 更新会员端销售订单的配送距离
	UpdateMemberSaleOrder(memberSaleOrder model.MemberSaleOrder) error                                                                                               // 更新会员端销售订单

	PaginateGet(pageNo, pageSize int, opts ...DBOption) ([]model.MemberSaleOrder, int64, error)
	WhereStatusIn(status []uint) DBOption
	WhereNotStatusIn(status []uint) DBOption
	WhereUpdateTimeGt(ts int64) DBOption
	GetOrderCount(opts ...DBOption) (int64, error)

	WithSaleBillSaleOrderProduct(preloads ...WithPreload) DBOption
}

func NewMemberSaleOrderRepo(db *gorm.DB) IMemberSaleOrderRepo {
	return NewMemberSaleOrderRepoImpl(db)
}

type MemberSaleOrderRepo struct {
	db *gorm.DB
}

func NewMemberSaleOrderRepoImpl(db *gorm.DB) *MemberSaleOrderRepo {
	return &MemberSaleOrderRepo{db: db}
}

func (r *MemberSaleOrderRepo) GetMemberSaleOrder(opts ...DBOption) (*model.MemberSaleOrder, error) {
	var memberSaleOrder model.MemberSaleOrder
	db := r.db.Model(&model.MemberSaleOrder{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&memberSaleOrder).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &memberSaleOrder, nil
}

func (r *MemberSaleOrderRepo) GetMemberSaleOrderRecord(uuid uint64, opts ...DBOption) (*model.MemberSaleOrder, error) {
	memberSaleOrder, err := r.GetMemberSaleOrder(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "Address",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "PaymentMethod",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "Address.MemberAddress",
			},
			WithPreload{
				Query: "Address.Member",
			},
			WithPreload{
				Query: "SaleBill.SaleOrders",
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return memberSaleOrder, nil
}

func (r *MemberSaleOrderRepo) CreateMemberSaleOrder(memberSaleOrder model.MemberSaleOrder) error {
	memberSaleOrder.SetNil()
	if err := r.db.Create(&memberSaleOrder).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderVerifiedPhoneStatus 更新会员端销售订单的手机号验证状态为已验证
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderVerifiedPhoneStatus(memberSaleOrderUuid uint64) error {
	if err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrderUuid).Update("is_verified_phone", 1).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderPendingPayment 更新会员端销售订单为待支付状态
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderPendingPayment(memberSaleOrder *model.MemberSaleOrder) error {
	if err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrder.Uuid).Updates(model.MemberSaleOrder{
		PaymentMethodUuid: memberSaleOrder.PaymentMethodUuid,            // 更新支付方式UUID
		Status:            constant.MemberSaleOrderStatusPendingPayment, // 更新订单状态为待支付
		Remark:            memberSaleOrder.Remark,                       // 更新订单备注
		ProductNum:        memberSaleOrder.ProductNum,                   // 更新商品数量
		ProductAmount:     memberSaleOrder.ProductAmount,                // 更新商品金额
		MemberDiscountFee: memberSaleOrder.MemberDiscountFee,            // 更新会员折扣
		Amount:            memberSaleOrder.Amount,                       // 更新订单总金额
	}).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderPendingMerchantAccept 更新会员端销售订单为“待商家接单”状态，表示订单支付成功，等待商家接单
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderPendingMerchantAccept(memberSaleOrderUuid uint64) error {
	if err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrderUuid).Updates(model.MemberSaleOrder{
		Status:  constant.MemberSaleOrderStatusPendingMerchantAccept, // 更新订单状态为待支付
		PayTime: time.Now().Unix(),
	}).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *MemberSaleOrderRepo) PaginateGetMemberSaleOrder(pageNo, pageSize int, opts ...DBOption) ([]model.MemberSaleOrder, int64, error) {
	var memberSaleOrders []model.MemberSaleOrder
	var total int64

	db := r.db.Model(&model.MemberSaleOrder{}).Session(&gorm.Session{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&memberSaleOrders).Error
	return memberSaleOrders, total, errors.WithMessage(err)
}

// GetCashierMemberSaleOrderList 获取收银端"外送"订单列表
func (r *MemberSaleOrderRepo) GetCashierMemberSaleOrderList(pageNo, pageSize int, statusList []uint) ([]model.MemberSaleOrder, int64, error) {
	opts := []DBOption{
		CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
		CommonRepo.DBOption(CommonRepo.SortWithPayTime("desc")),
	}

	// 根据状态列表筛选
	// 默认状态列表为空时，查询待接单状态
	statusFilter := CommonRepo.DBOption(CommonRepo.WhereByMultipleStatus([]uint{
		constant.MemberSaleOrderStatusPendingMerchantAccept,
	}))
	if len(statusList) == 1 {
		statusFilter = CommonRepo.DBOption(CommonRepo.WhereByStatus(statusList[0]))
	} else if len(statusList) > 1 {
		statusFilter = CommonRepo.DBOption(CommonRepo.WhereByMultipleStatus(statusList))
	}
	opts = append(opts, statusFilter)
	return r.PaginateGetMemberSaleOrder(pageNo, pageSize, opts...)
}

type GetCashierMemberSaleOrderManageListReq struct {
	OrderNo    *string
	SerialNo   *string
	TimeFilter *req.TimeFilterParams
}

// GetCashierMemberSaleOrderManageList 获取收银端"外送"订单管理列表
func (r *MemberSaleOrderRepo) GetCashierMemberSaleOrderManageList(pageNo, pageSize int, statusList []uint, req GetCashierMemberSaleOrderManageListReq) ([]model.MemberSaleOrder, int64, error) {

	opts := []DBOption{
		CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
		CommonRepo.DBOption(CommonRepo.SortWithPayTime("desc")),
		CommonRepo.Preload(
			WithPreload{
				Query: "PaymentMethod",
			},
			WithPreload{
				Query: "Address",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "Address.Member",
			},
		),
	}

	// 根据状态列表筛选
	if len(statusList) == 1 {
		statusFilter := CommonRepo.DBOption(CommonRepo.WhereByStatus(statusList[0]))
		opts = append(opts, statusFilter)
	} else if len(statusList) > 1 {
		statusFilter := CommonRepo.DBOption(CommonRepo.WhereByMultipleStatus(statusList))
		opts = append(opts, statusFilter)

	}
	// 根据时间筛选
	if req.TimeFilter != nil {
		if req.TimeFilter.TimeType == 1 {
			// 下单时间
			timeFilter := CommonRepo.DBOption(CommonRepo.WhereBetweenByCreateTime(req.TimeFilter.QueryStartTime, req.TimeFilter.QueryEndTime))
			opts = append(opts, timeFilter)
		} else if req.TimeFilter.TimeType == 2 {
			// 支付时间
			timeFilter := CommonRepo.DBOption(CommonRepo.WhereBetweenByPayTime(req.TimeFilter.QueryStartTime, req.TimeFilter.QueryEndTime))
			opts = append(opts, timeFilter)
		}
	}

	return r.PaginateGetMemberSaleOrder(pageNo, pageSize, opts...)
}

func (r *MemberSaleOrderRepo) PaginateGet(pageNo, pageSize int, opts ...DBOption) ([]model.MemberSaleOrder, int64, error) {
	var memberSaleOrders []model.MemberSaleOrder
	var total int64
	db := r.db.Model(&model.MemberSaleOrder{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err)
	}
	// 获取分页数据
	err := db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Order("update_time desc").Find(&memberSaleOrders).Error
	return memberSaleOrders, total, errors.WithMessage(err)
}

func (r *MemberSaleOrderRepo) WhereStatusIn(status []uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if len(status) == 0 {
			return db
		}
		return db.Where("status IN (?)", status)
	}
}

func (r *MemberSaleOrderRepo) WhereNotStatusIn(status []uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status NOT IN (?)", status)
	}
}

func (r *MemberSaleOrderRepo) WhereUpdateTimeGt(ts int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("create_time > ?", ts)
	}
}

// GetOrderCount 获取外送订单数量
func (r *MemberSaleOrderRepo) GetOrderCount(opts ...DBOption) (int64, error) {
	var total int64
	db := r.db.Model(&model.MemberSaleOrder{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	if err := db.Count(&total).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	return total, nil
}

// UpdateMemberSaleOrderReject 更新会员端销售订单-拒单
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderReject(memberSaleOrder model.MemberSaleOrder) error {
	err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrder.Uuid).Updates(model.MemberSaleOrder{
		Status:       memberSaleOrder.Status,
		CancelScene:  memberSaleOrder.CancelScene,
		CancelTime:   memberSaleOrder.CancelTime,
		CancelReason: memberSaleOrder.CancelReason,
	}).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderAccept 更新会员端销售订单-接单
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderAccept(memberSaleOrder model.MemberSaleOrder) error {
	err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrder.Uuid).Updates(model.MemberSaleOrder{
		Status:     memberSaleOrder.Status,
		AcceptTime: memberSaleOrder.AcceptTime,
	}).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderCookFinish 更新会员端销售订单-备餐完成
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderCookFinish(memberSaleOrder model.MemberSaleOrder) error {
	err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrder.Uuid).Updates(model.MemberSaleOrder{
		Status:   memberSaleOrder.Status,
		CookTime: memberSaleOrder.CookTime,
	}).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateDeliveryDistance 更新会员端销售订单的配送距离
func (r *MemberSaleOrderRepo) UpdateDeliveryDistance(memberSaleOrderUuid uint64, distance float64) error {
	err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrderUuid).Updates(model.MemberSaleOrder{
		DeliveryDistance: distance,
	}).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrder 更新会员端销售订单
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrder(memberSaleOrder model.MemberSaleOrder) error {
	memberSaleOrder.SetNil()
	err := r.db.Model(&model.MemberSaleOrder{}).Select("*").Where("uuid = ?", memberSaleOrder.Uuid).Updates(memberSaleOrder).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *MemberSaleOrderRepo) WithSaleBillSaleOrderProduct(preloads ...WithPreload) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleBill.SaleOrders.SaleOrderProducts.ImageFile").
			Preload("SaleBill.SaleOrders.SaleOrderProducts.MultiLanguageName")
	}
}
