package repository

import (
	"fmt"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/config"

	"gorm.io/gorm"
)

type IMemberSaleOrderRepo interface {
	IQueryMemberSaleOrderRepo
	CreateMemberSaleOrder(memberSaleOrder model.MemberSaleOrder) error                        // 创建会员端销售订单
	UpdateMemberSaleOrderVerifiedPhoneStatus(memberSaleOrderUuid uint64) error                // 更新会员端销售订单的手机号验证状态为已验证
	UpdateMemberSaleOrderPendingPayment(memberSaleOrder *model.MemberSaleOrder) error         // 更新会员端销售订单为待支付状态
	UpdateMemberSaleOrderPendingMerchantAccept(memberSaleOrderUuid uint64) error              // 更新会员端销售订单为"待商家接单"状态，表示订单支付成功，等待商家接单
	UpdateMemberSaleOrderSubmitPayTime(memberSaleOrderUuid uint64, submitPayTime int64) error // 更新会员端销售订单的提交支付时间戳
}

type IQueryMemberSaleOrderRepo interface {
	GetMemberSaleOrder(opts ...DBOption) (*model.MemberSaleOrder, error)                                                                                             // 获取会员端销售订单
	GetMemberSaleOrderRecord(uuid uint64, opts ...DBOption) (*model.MemberSaleOrder, error)                                                                          // 获取会员端销售订单记录
	GetMemberSaleOrderRecordOnly(uuid uint64) (*model.MemberSaleOrder, error)                                                                                        // 获取会员端销售订单记录，不包含关联数据
	GetMemberSaleOrderRecordOnlyBySaleBillUuid(saleBillUuid uint64) (*model.MemberSaleOrder, error)                                                                  // 获取会员端销售订单记录，不包含关联数据，并包含配送费
	PaginateGetMemberSaleOrder(pageNo, pageSize int, opts ...DBOption) ([]model.MemberSaleOrder, int64, error)                                                       // 分页获取会员端销售订单
	GetCashierMemberSaleOrderList(pageNo, pageSize int, statusList []uint) ([]model.MemberSaleOrder, int64, error)                                                   // 获取收银台"外送"订单列表
	GetCashierMemberSaleOrderManageList(pageNo, pageSize int, statusList []uint, req GetCashierMemberSaleOrderManageListReq) ([]model.MemberSaleOrder, int64, error) // 获取收银台"外送"订单管理列表
	GetCashierMemberSaleOrderNum(statusList []uint, timeFilter *req.TimeFilterParams, opts ...DBOption) (int64, error)                                               // 获取收银台"外送"订单数量
	UpdateMemberSaleOrderAccept(memberSaleOrder model.MemberSaleOrder) error                                                                                         // 更新会员端销售订单-接单
	UpdateMemberSaleOrderReject(memberSaleOrder model.MemberSaleOrder) error                                                                                         // 更新会员端销售订单-拒单
	UpdateMemberSaleOrderCookFinish(memberSaleOrder model.MemberSaleOrder) error                                                                                     // 更新会员端销售订单-备餐完成
	UpdateDeliveryDistance(memberSaleOrderUuid uint64, distance float64) error                                                                                       // 更新会员端销售订单的配送距离
	UpdateMemberSaleOrder(memberSaleOrder model.MemberSaleOrder) error                                                                                               // 更新会员端销售订单
	UpdateMemberSaleOrderRiderAccept(memberSaleOrder model.MemberSaleOrder) error                                                                                    // 更新会员端销售订单-骑手接单
	UpdateMemberSaleOrderRiderDelivery(memberSaleOrder model.MemberSaleOrder) error                                                                                  // 更新会员端销售订单-骑手配送中
	UpdateMemberSaleOrderRiderCompleted(memberSaleOrder model.MemberSaleOrder) error                                                                                 // 更新会员端销售订单-骑手配送完成
	UpdateMemberSaleOrderProviderInfo(memberSaleOrder model.MemberSaleOrder) error                                                                                   // 更新会员端销售订单-配送 provider 信息
	UpdateMemberSaleOrderAddress(memberSaleOrder model.MemberSaleOrder) error                                                                                        // 更新会员端销售订单-配送地址
	GetMemberSaleOrderByContactNameAndContactPhoneSuffix(contactName string, contactPhoneSuffix string) ([]*model.MemberSaleOrder, error)                            // 根据联系人姓名和手机号后缀查询会员端销售订单
	GetOrderNum(status []uint) (int64, error)                                                                                                                        // 获取订单数量
	GetMemberSaleOrderLatest() (*model.MemberSaleOrder, error)                                                                                                       // 获取最新的一条会员端销售订单(已提交支付的)
	UpdateMemberSaleOrderRefundAmount(memberSaleOrderUuid uint64, refundAmount float64) error                                                                        // 更新会员端销售订单的退款金额
	UpdateMemberSaleOrderSort(memberSaleOrderUuid uint64, sort int) error                                                                                            // 更新会员端销售订单的sort排序

	GetForCall(opts ...DBOption) ([]model.MemberSaleOrder, error)
	WhereStatusIn(status []uint) DBOption
	WhereNotStatusIn(status []uint) DBOption
	WhereUpdateTimeGt(ts int64) DBOption
	WhereKeyword(keyword string, language string) DBOption
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
				Query: "Address",
			},
			WithPreload{
				Query: "Member",
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

// GetMemberSaleOrderRecordOnly 获取会员端销售订单记录，不包含关联数据
func (r *MemberSaleOrderRepo) GetMemberSaleOrderRecordOnly(uuid uint64) (*model.MemberSaleOrder, error) {
	var memberSaleOrder model.MemberSaleOrder
	db := r.db.Model(&model.MemberSaleOrder{})

	err := db.Where("uuid = ?", uuid).First(&memberSaleOrder).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &memberSaleOrder, nil
}

// GetMemberSaleOrderRecordOnlyBySaleBillUuid 获取会员端销售订单记录，不包含关联数据，并包含配送费
func (r *MemberSaleOrderRepo) GetMemberSaleOrderRecordOnlyBySaleBillUuid(saleBillUuid uint64) (*model.MemberSaleOrder, error) {
	var memberSaleOrder model.MemberSaleOrder
	db := r.db.Model(&model.MemberSaleOrder{})

	err := db.Where("sale_bill_uuid = ?", saleBillUuid).First(&memberSaleOrder).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &memberSaleOrder, nil
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
		SerialNumber:        memberSaleOrder.SerialNumber,                 // 更新订单流水号
		PaymentMethodUuid:   memberSaleOrder.PaymentMethodUuid,            // 更新支付方式UUID
		Status:              constant.MemberSaleOrderStatusPendingPayment, // 更新订单状态为待支付
		Remark:              memberSaleOrder.Remark,                       // 更新订单备注
		ProductNum:          memberSaleOrder.ProductNum,                   // 更新商品数量
		ProductAmount:       memberSaleOrder.ProductAmount,                // 更新商品金额
		OriginProductAmount: memberSaleOrder.OriginProductAmount,          // 更新商品原价
		MemberDiscountFee:   memberSaleOrder.MemberDiscountFee,            // 更新会员折扣
		Amount:              memberSaleOrder.Amount,                       // 更新订单总金额
		SubmitPayTime:       memberSaleOrder.SubmitPayTime,                // 更新提交支付时间戳
	}).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderPendingMerchantAccept 更新会员端销售订单为"待商家接单"状态，表示订单支付成功，等待商家接单
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderPendingMerchantAccept(memberSaleOrderUuid uint64) error {
	if err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrderUuid).Updates(model.MemberSaleOrder{
		Status:  constant.MemberSaleOrderStatusPendingMerchantAccept, // 更新订单状态为待商家接单
		PayTime: time.Now().Unix(),
	}).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderSubmitPayTime 更新会员端销售订单的提交支付时间戳
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderSubmitPayTime(memberSaleOrderUuid uint64, submitPayTime int64) error {
	if err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrderUuid).Update("submit_pay_time", submitPayTime).Error; err != nil {
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

	err = db.Offset((pageNo - 1) * pageSize).Order("id desc").Limit(pageSize).Find(&memberSaleOrders).Error
	return memberSaleOrders, total, errors.WithMessage(err)
}

// GetCashierMemberSaleOrderList 获取收银端"外送"订单列表
func (r *MemberSaleOrderRepo) GetCashierMemberSaleOrderList(pageNo, pageSize int, statusList []uint) ([]model.MemberSaleOrder, int64, error) {
	opts := []DBOption{
		CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
		CommonRepo.DBOption(CommonRepo.SortWithPayTime("desc")),
		CommonRepo.DBOption(CommonRepo.WhereByNoSelectingTimeout()),
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
		CommonRepo.DBOption(CommonRepo.WhereByNoSelectingTimeout()),
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
		// status = 6(骑手配送中) 排在前面，其他按照create_time 倒序
		func(db *gorm.DB) *gorm.DB {
			return db.Order(fmt.Sprintf("status = %d desc, create_time desc", constant.MemberSaleOrderStatusDelivering))
		},
	}

	// 根据外送序号搜索
	if req.SerialNo != nil {
		serialNoFilter := CommonRepo.DBOption(CommonRepo.WhereBySerialNumber(*req.SerialNo))
		opts = append(opts, serialNoFilter)
	}

	// 根据订单号搜索
	if req.OrderNo != nil {
		orderNoFilter := CommonRepo.DBOption(CommonRepo.WhereByOrderNo(*req.OrderNo))
		opts = append(opts, orderNoFilter)
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

// GetCashierMemberSaleOrderNum 获取收银端"外送"订单数量
func (r *MemberSaleOrderRepo) GetCashierMemberSaleOrderNum(statusList []uint, timeFilter *req.TimeFilterParams, opts ...DBOption) (int64, error) {

	// 根据状态列表筛选
	if len(statusList) == 1 {
		statusFilter := CommonRepo.DBOption(CommonRepo.WhereByStatus(statusList[0]))
		opts = append(opts, statusFilter)
	} else if len(statusList) > 1 {
		statusFilter := CommonRepo.DBOption(CommonRepo.WhereByMultipleStatus(statusList))
		opts = append(opts, statusFilter)
	}
	// 根据时间筛选
	if timeFilter != nil {
		if timeFilter.TimeType == 1 {
			// 下单时间
			timeFilter := CommonRepo.DBOption(CommonRepo.WhereBetweenByCreateTime(timeFilter.QueryStartTime, timeFilter.QueryEndTime))
			opts = append(opts, timeFilter)
		} else if timeFilter.TimeType == 2 {
			// 支付时间
			timeFilter := CommonRepo.DBOption(CommonRepo.WhereBetweenByPayTime(timeFilter.QueryStartTime, timeFilter.QueryEndTime))
			opts = append(opts, timeFilter)
		}
	}

	var total int64
	db := r.db.Model(&model.MemberSaleOrder{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return total, nil
}

func (r *MemberSaleOrderRepo) GetForCall(opts ...DBOption) ([]model.MemberSaleOrder, error) {
	var memberSaleOrders []model.MemberSaleOrder
	db := r.db.Model(&model.MemberSaleOrder{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	// 1、会员付款，非自动接单
	// 2、已自动接单
	// 3、会员取消
	db = db.Where("( status = ? AND is_auto_accept = 0 ) OR ( status = ? AND is_auto_accept = 1 ) OR ( status = ? AND cancel_scene = ? )",
		constant.MemberSaleOrderStatusPendingMerchantAccept, constant.MemberSaleOrderStatusCooking, constant.MemberSaleOrderStatusCancelled, constant.MemberSaleOrderSceneMemberCancel)
	// 获取10条数据
	err := db.Limit(10).Order("update_time desc").Find(&memberSaleOrders).Error
	return memberSaleOrders, errors.WithMessage(err)
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

// languageColumnMap 语言代码到 multi_language_name 表列名的安全映射
var languageColumnMap = map[string]string{
	"zh":   "zh_name",
	"en":   "en_name",
	"th":   "th_name",
	"zhtw": "zh_tw_name",
	"ja":   "ja_name",
	"ko":   "ko_name",
	"my":   "my_name",
	"tr":   "tr_name",
	"sv":   "sv_name",
	"de":   "en_name", // 德语暂无独立列，回退到英语
}

// WhereKeyword 根据订单号或商品名称搜索
func (r *MemberSaleOrderRepo) WhereKeyword(keyword string, language string) DBOption {
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

		// 子查询：查找包含指定商品名称的销售订单UUID
		subQuery := r.db.Table(prefix+"sale_order_product sop").
			Select("DISTINCT sop.sale_order_uuid").
			Joins("LEFT JOIN "+prefix+"multi_language_name so ON sop.multi_language_name_uuid = so.uuid").
			Where("sop.delete_time = ?", 0).
			Where("so."+columnName+" LIKE ?", "%"+keyword+"%")

		return db.Where("(sale_order_uuid IN (?) OR order_no LIKE ?)", subQuery, "%"+keyword+"%")
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
		Status:             memberSaleOrder.Status,
		AcceptTime:         memberSaleOrder.AcceptTime,
		RelatedOrderType:   memberSaleOrder.RelatedOrderType,
		RelatedOrderNo:     memberSaleOrder.RelatedOrderNo,
		ExpectedFinishTime: memberSaleOrder.ExpectedFinishTime,
		IsAutoAccept:       memberSaleOrder.IsAutoAccept,
		PayTime:            memberSaleOrder.PayTime,
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
		DeliveryDistance:     distance,
		IsDistanceCalculated: constant.DistanceCalculated,
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

// UpdateMemberSaleOrderRiderAccept 更新会员端销售订单-骑手接单
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderRiderAccept(memberSaleOrder model.MemberSaleOrder) error {
	err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrder.Uuid).Updates(model.MemberSaleOrder{
		Status:          memberSaleOrder.Status,
		RiderAcceptTime: memberSaleOrder.RiderAcceptTime,
		RiderName:       memberSaleOrder.RiderName,
		RiderPhone:      memberSaleOrder.RiderPhone,
		Location:        memberSaleOrder.Location,
	}).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderRiderDelivery 更新会员端销售订单-骑手配送中
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderRiderDelivery(memberSaleOrder model.MemberSaleOrder) error {
	err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrder.Uuid).Updates(model.MemberSaleOrder{
		RiderName:      memberSaleOrder.RiderName,
		RiderPhone:     memberSaleOrder.RiderPhone,
		Location:       memberSaleOrder.Location,
		Status:         memberSaleOrder.Status,
		RiderStartTime: memberSaleOrder.RiderStartTime,
	}).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderRiderCompleted 更新会员端销售订单-骑手配送完成
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderRiderCompleted(memberSaleOrder model.MemberSaleOrder) error {
	data := model.MemberSaleOrder{
		Status:     memberSaleOrder.Status,
		FinishTime: memberSaleOrder.FinishTime,
		Sort:       memberSaleOrder.Sort,
	}
	if memberSaleOrder.RiderName != "" {
		data.RiderName = memberSaleOrder.RiderName
	}
	if memberSaleOrder.RiderPhone != "" {
		data.RiderPhone = memberSaleOrder.RiderPhone
	}
	if memberSaleOrder.Location != "" {
		data.Location = memberSaleOrder.Location
	}
	err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrder.Uuid).Updates(data).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderProviderInfo 更新会员端销售订单-配送 provider 信息
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderProviderInfo(memberSaleOrder model.MemberSaleOrder) error {
	err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrder.Uuid).Updates(model.MemberSaleOrder{
		RelatedOrderType: memberSaleOrder.RelatedOrderType,
		RelatedOrderNo:   memberSaleOrder.RelatedOrderNo,
	}).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderAddress 更新会员端销售订单-配送地址
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderAddress(memberSaleOrder model.MemberSaleOrder) error {
	err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrder.Uuid).Updates(model.MemberSaleOrder{
		MemberAddressUuid:    memberSaleOrder.MemberAddressUuid,
		ContactLocation:      memberSaleOrder.ContactLocation,
		ContactAddress:       memberSaleOrder.ContactAddress,
		ContactAddressDetail: memberSaleOrder.ContactAddressDetail,
		ContactName:          memberSaleOrder.ContactName,
		ContactPhone:         memberSaleOrder.ContactPhone,
		ContactPhonePrefix:   memberSaleOrder.ContactPhonePrefix,
		ContactGender:        memberSaleOrder.ContactGender,
	}).Error
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

// 根据联系人姓名和联系人手机号后四位查询外送订单
func (r *MemberSaleOrderRepo) GetMemberSaleOrderByContactNameAndContactPhoneSuffix(contactName string, contactPhoneSuffix string) ([]*model.MemberSaleOrder, error) {
	var memberSaleOrders []*model.MemberSaleOrder
	err := r.db.Model(&model.MemberSaleOrder{}).
		Where("delete_time = ?", 0).
		Where("status in ?", []int{ // 只查询以下状态的订单
			constant.MemberSaleOrderStatusPendingMerchantAccept, // 待商家接单
			constant.MemberSaleOrderStatusCooking,               // 商家备餐中
			constant.MemberSaleOrderStatusPendingRiderPickup,    // 待骑手接单
			constant.MemberSaleOrderStatusPendingRiderDelivery,  // 骑手正在赶往商家
			constant.MemberSaleOrderStatusDelivering,            // 骑手配送中
			constant.MemberSaleOrderStatusCompleted,             // 已完成
			constant.MemberSaleOrderStatusCancelled,             // 已取消
		}).
		Where("contact_name LIKE ? OR contact_phone LIKE ?", "%"+contactName+"%", "%"+contactPhoneSuffix).Find(&memberSaleOrders).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return memberSaleOrders, nil
}

// GetOrderNum 获取某些状态的订单数量
func (r *MemberSaleOrderRepo) GetOrderNum(status []uint) (int64, error) {
	var num int64
	err := r.db.Model(&model.MemberSaleOrder{}).
		Where("delete_time = ?", 0).
		Where("cancel_scene != ?", constant.MemberSaleOrderSceneSelectingTimeout).
		Where("status in ?", status).
		Count(&num).Error
	if err != nil {
		return 0, errors.WithMessage(err)
	}
	return num, nil
}

func (r *MemberSaleOrderRepo) GetMemberSaleOrderLatest() (*model.MemberSaleOrder, error) {
	memberSaleOrder, err := r.GetMemberSaleOrder(
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.SortWithSubmitPayTime("desc"),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return memberSaleOrder, nil
}

// 更新会员端销售订单的退款金额
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderRefundAmount(memberSaleOrderUuid uint64, refundAmount float64) error {
	err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrderUuid).Update("refund_amount", gorm.Expr("refund_amount + ?", refundAmount)).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderSort 更新会员端销售订单的sort排序
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderSort(memberSaleOrderUuid uint64, sort int) error {
	err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrderUuid).Update("sort", sort).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}
