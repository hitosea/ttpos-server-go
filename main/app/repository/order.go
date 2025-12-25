package repository

import (
	goCtx "context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/domain/entity"
	"ttpos-server-go/app/modules/objectstorage/domain/repository"
	"ttpos-server-go/app/modules/objectstorage/domain/service"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/app/repository/ro"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// IOrderRepo 定义订单仓库接口
type IOrderRepo interface {
	IOrderQueryRepo
	CreateSaleBill(model model.SaleBill) (model.SaleBill, error)                                                               // 创建销售单
	CreateSaleBillSetting(model model.SaleBillSetting) (model.SaleBillSetting, error)                                          // 创建销售账单设置
	UpdateSaleBillSetting(obj model.SaleBillSetting) (model.SaleBillSetting, error)                                            // 更新销售账单设置
	CreateSaleOrder(model model.SaleOrder) (model.SaleOrder, error)                                                            // 创建订单
	CountSaleBill(opts ...DBOption) (int64, error)                                                                             // 统计销售单数据
	CreateSaleOrderBuffetCustomerType(model model.SaleOrderBuffetCustomerType) (model.SaleOrderBuffetCustomerType, error)      // 创建销售订单自助餐顾客类型
	DeleteSaleOrderBuffetCustomerType(saleOrderUuid uint64) error                                                              // 删除销售订单自助餐顾客类型
	CreateSaleOrderBuffetDelayProduct(model model.SaleOrderBuffetDelayProduct) (model.SaleOrderBuffetDelayProduct, error)      // 创建销售订单自助餐加钟
	UpdateSaleOrderBuffetDelayProductRecord(model model.SaleOrderBuffetDelayProduct) error                                     // 更新销售订单自助餐加钟
	CancelOrder(ctx context.Context, saleBillUuid uint64, deskUuid uint64, reason string) error                                // 取消订单
	CancelDeskOrder(ctx context.Context, deskUuid uint64, reason string) error                                                 // 取消桌台订单
	DeleteOrder(saleBillUuid uint64, saleOrderUuid uint64) error                                                               // 删除订单
	HideOrder(saleBillUuid uint64) error                                                                                       // 隐藏订单
	DeleteOrderProduct(saleBillUuid uint64, saleOrderUuid uint64, saleOrderProductUuid uint64, isPackageProduct bool) error    // 删除订单产品
	ChangePopulation(saleBillUuid uint64, population int) error                                                                // 修改订单人数
	ChangeProductRemark(saleBillUuid uint64, saleOrderUuid uint64, orderProductUuid uint64, remark string, sign string) error  // 修改订单商品备注
	SetLock(saleBillUuid uint64, isLock bool) error                                                                            // 设置订单锁定状态
	SaveOrUpdateInvoiceInfo(saleOrderUuid uint64, invoiceInfo model.SaleOrderInvoiceInfo) (*model.SaleOrderInvoiceInfo, error) // 设置订单发票信息
	UpdateErpDiscountAmount(saleOrderUuid uint64, erpDiscountAmount float64) error                                             // 更新订单应收优惠金额
}

// IOrderQueryRepo 订单查询
type IOrderQueryRepo interface {
	GetSaleBill(opts ...DBOption) (model.SaleBill, error)                                                                                      // 获取销售单
	IsOrderNoExists(orderNo string) (bool, error)                                                                                              // 查询orderNo是否存在
	GetInstantSaleBill(deviceUuid uint64) (*model.SaleBill, error)                                                                             // 获取待支付且未挂单的点餐订单
	GetOrderListWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.SaleBill, int64, error)                                    // 获取订单列表
	GetOrderNum(opts ...DBOption) (int64, error)                                                                                               // 获取订单数量
	GetCashierOrderListWithPagination(param GetCashierOrderListWithPaginationType, tz string) ([]model.SaleBill, int64, DBOption, error)       // 获取收银的订单列表
	GetCashierOrderExportListWithPagination(param GetCashierOrderListWithPaginationType, tz string) ([]model.SaleBill, int64, DBOption, error) // 获取收银的订单列表
	GetSaleBillInfo(saleBillUuid uint64, saleOrderUuid uint64) (model.SaleBill, error)                                                         // 获取销售账单详细信息
	GetSaleBillInfoByDesk(deskUuid, saleOrderUuid uint64) (model.SaleBill, error)                                                              // 获取桌台的销售账单详细信息
	GetOrderCartInfo(saleBillUuid uint64, opts ...OrderCartInfoOptionFunc) (*ro.ShopCartRepo, error)                                           // 获取点餐购物车信息
	GetOrderBuffetInfo(saleBillUuid, saleOrderUuid uint64) (model.SaleBill, error)                                                             // 获取订单自助餐信息
	GetSaleBillInfoAndProduct(saleBillUuid uint64, saleOrderUuid uint64, saleOrderProductUuid uint64) (model.SaleBill, error)                  // 获取销售账单详细信息-包含商品信息
	GetSaleBillInfoAndMember(saleBillUuid uint64) (model.SaleBill, error)                                                                      // 获取销售账单详细信息-包含会员信息
	GetSaleBillInfoAndPaymentOrders(saleBillUuid uint64, saleOrderUuid uint64, saleOrderPaymentUuid uint64) (model.SaleBill, error)            // 获取销售账单详细信息-包含商品信息
	GetSaleOrderProductListBySaleOrderProductUuids(saleOrderProductUuids []uint64) ([]model.SaleOrderProduct, error)                           // 根据销售订单商品uuid列表获取销售订单商品列表
	GetSaleBillDetails(saleBillUuid uint64, saleOrderUuid uint64) (model.SaleBill, error)                                                      // 获取销售账单详细信息-丰富的-几乎包含所有的关联
	IsPartiallyPaid(param any) bool                                                                                                            // 判断是否存在部分支付
	GetSaleOrderBomList(saleOrderUuid uint64) ([]model.SaleOrderProductBom, error)                                                             // 查询销售订单的所有bom
	GetSaleBillAllInfo(saleBillUuid uint64, opts ...GetSaleBillAllInfoOption) (*model.SaleBill, error)                                         // 获取销售账单所有信息
	GetSaleBillWithProducts(saleBillUuid uint64) (*model.SaleBill, error)                                                                      // 获取销售账单所有商品信息
	HasShowOrder(deviceUuid uint64) (uint64, error)                                                                                            // 判断该设备是否有未挂单的点餐订单
	GetSaleBillRecord(saleBillUuid uint64) (*model.SaleBill, error)                                                                            // 获取销售账单记录
	GetSaleBillSaleOrderRecord(saleOrderUuid uint64) (*model.SaleOrder, error)                                                                 // 获取销售账单记录
	GetInvoiceInfo(saleOrderUuid uint64) (*model.SaleOrderInvoiceInfo, error)                                                                  // 获取订单发票信息
	GetMonthlyOrderRanks(saleBillUuids []uint64) ([]MonthlyOrderRank, error)                                                                   // 获取订单的月排名信息（基于全表数据）
	GetSaleBillBatchCookingMode(saleBillUuid uint64) (string, error)                                                                           // 获取销售账单当前的分批送厨模式
	UpdateSaleBillOrderRemark(saleBillUuid uint64, orderRemark string) error                                                                   // 更新销售账单整单备注
	GetSaleOrderUuids(opts ...DBOption) []uint64
	GetSaleBillList(opts ...DBOption) []model.SaleBill // 获取销售账单列表
}

// orderRepo 订单仓库
type orderRepo struct {
	db *gorm.DB
}

// NewOrderRepo 创建新的订单仓库
func NewOrderRepo(db *gorm.DB) IOrderRepo {
	return NewOrderRepoImpl(db)
}

// NewOrderRepoImpl 创建新的订单仓库实现
func NewOrderRepoImpl(db *gorm.DB) IOrderRepo {
	return &orderRepo{db: db}
}

// CreateSaleBill 创建销售单
func (r *orderRepo) CreateSaleBill(model model.SaleBill) (model.SaleBill, error) {
	err := r.db.Create(&model).Error
	if err != nil {
		return model, fmt.Errorf("CreateSaleBill: %v", err)
	}

	return model, nil
}

// CreateSaleBillSetting 创建销售账单设置
func (r *orderRepo) CreateSaleBillSetting(model model.SaleBillSetting) (model.SaleBillSetting, error) {
	err := r.db.Create(&model).Error
	if err != nil {
		return model, fmt.Errorf("CreateSaleBillSetting: %v", err)
	}

	return model, nil
}

// UpdateSaleBillSetting 更新销售账单设置
func (r *orderRepo) UpdateSaleBillSetting(obj model.SaleBillSetting) (model.SaleBillSetting, error) {
	err := r.db.Model(&model.SaleBillSetting{}).Where("uuid = ?", obj.Uuid).Updates(map[string]interface{}{
		"service_fee_type":   obj.ServiceFeeType,
		"service_fee_value":  obj.ServiceFeeValue,
		"tax_fee_type":       obj.TaxFeeType,
		"service_apply":      obj.ServiceApply,
		"discount_type":      obj.DiscountType,
		"zero_rule":          obj.ZeroRule,
		"zero_checkout_rule": obj.ZeroCheckoutRule,
		"is_stat_gift":       obj.IsStatGift,
		"is_stat_free":       obj.IsStatFree,
	}).Error
	if err != nil {
		return obj, errors.WithMessage(err)
	}
	return obj, nil
}

// GetSaleBill 获取销售单
func (r *orderRepo) GetSaleBill(opts ...DBOption) (model.SaleBill, error) {
	var saleBill model.SaleBill
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&saleBill)
	if result.Error != nil {
		return saleBill, fmt.Errorf("GetSaleBill: %v", result.Error)
	}

	return saleBill, nil
}

// IsOrderNoExists 查询orderNo是否存在
func (r *orderRepo) IsOrderNoExists(orderNo string) (bool, error) {
	var countSaleBill int64
	var countSaleOrder int64

	if err := r.db.Model(&model.SaleBill{}).Where("order_no = ?", orderNo).Count(&countSaleBill).Error; err != nil {
		return false, errors.WithMessage(err)
	}
	if countSaleBill < 1 {
		return true, nil
	}
	if err := r.db.Model(&model.SaleOrder{}).Where("order_no = ?", orderNo).Count(&countSaleOrder).Error; err != nil {
		return false, errors.WithMessage(err)
	}
	if countSaleOrder < 1 {
		return true, nil
	}

	return false, errors.WithMessage(errors.New("订单编号生产异常"))
}

// GetInstantSaleBill 获取待支付且未挂单的点餐订单
func (r *orderRepo) GetInstantSaleBill(deviceUuid uint64) (*model.SaleBill, error) {
	saleBill, err := r.GetSaleBill(
		CommonRepo.WhereByBillType(constant.OrderSourceMapToBillType[constant.OrderSourceInstant]),
		CommonRepo.WhereByStatus(constant.SaleBillStatusPending),
		CommonRepo.WhereByIsHide(false),
		CommonRepo.WhereByDeviceUuid(deviceUuid),
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
		return nil, err
	}
	return &saleBill, nil
}

// CreateSaleOrder 创建销售订单
func (r *orderRepo) CreateSaleOrder(model model.SaleOrder) (model.SaleOrder, error) {
	err := r.db.Create(&model).Error
	if err != nil {
		return model, fmt.Errorf("CreateSaleOrder: %v", err)
	}
	return model, nil
}

// CountSaleBill 统计销售单数据
func (r *orderRepo) CountSaleBill(opts ...DBOption) (int64, error) {
	var count int64
	db := r.db.Model(&model.SaleBill{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&count).Error
	return count, err
}

// CreateSaleOrderBuffetCustomerType 创建销售订单自助餐顾客类型
func (r *orderRepo) CreateSaleOrderBuffetCustomerType(model model.SaleOrderBuffetCustomerType) (model.SaleOrderBuffetCustomerType, error) {
	err := r.db.Create(&model).Error
	if err != nil {
		return model, fmt.Errorf("CreateSaleOrderBuffetCustomerType: %v", err)
	}
	return model, nil
}

// DeleteSaleOrderBuffetCustomerType 删除销售订单自助餐顾客类型
func (r *orderRepo) DeleteSaleOrderBuffetCustomerType(saleOrderUuid uint64) error {
	err := r.db.Where("sale_order_uuid = ?", saleOrderUuid).Delete(&model.SaleOrderBuffetCustomerType{}).Error
	if err != nil {
		return fmt.Errorf("DeleteSaleOrderBuffetCustomerType: %v", err)
	}
	return nil
}

// CreateSaleOrderBuffetDelayProduct 创建销售订单自助餐加钟
func (r *orderRepo) CreateSaleOrderBuffetDelayProduct(obj model.SaleOrderBuffetDelayProduct) (model.SaleOrderBuffetDelayProduct, error) {
	obj.SetNil()
	err := r.db.Model(&model.SaleOrderBuffetDelayProduct{}).Create(&obj).Error
	if err != nil {
		return obj, errors.WithMessage(fmt.Errorf("CreateSaleOrderBuffetDelayProduct: %v", err))
	}
	return obj, nil
}

func (r *orderRepo) UpdateSaleOrderBuffetDelayProductRecord(obj model.SaleOrderBuffetDelayProduct) error {
	obj.SetNil()
	if obj.NoPrimaryKey() {
		return errors.WithMessage(errors.New("UUID或ID不能为0"))
	}
	return r.db.Model(&model.SaleOrderBuffetDelayProduct{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(obj).Error
}

// GetOrderListWithPagination 获取订单列表
func (r *orderRepo) GetOrderListWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.SaleBill, int64, error) {
	var orders []model.SaleBill
	var total int64

	db := r.db.Model(&model.SaleBill{}).Session(&gorm.Session{})

	for _, opt := range opts {
		db = opt(db)
	}

	// 获取列表
	err := db.Count(&total).Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&orders).Error

	return orders, total, errors.WithMessage(err)
}

// GetOrderNum 获取订单的数量
func (r *orderRepo) GetOrderNum(opts ...DBOption) (int64, error) {
	var count int64
	db := r.db.Model(&model.SaleBill{}).Session(&gorm.Session{})

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Count(&count)
	err := result.Error
	if err != nil {
		return 0, fmt.Errorf("GetOrderNum: %v", err)
	}
	return count, nil
}

// GetCashierOrderListWithPaginationType 获取收银台订单列表-参数
type GetCashierOrderListWithPaginationType struct {
	PageNo              int    // 页码
	PageSize            int    // 页面大小
	OrderNo             string // 订单编号
	DateType            int    // 时间类型,-1=全部、0=今天、1=昨天、2=本周
	EnableCreateTime    bool   // 是否启用创建时间
	EnablePayTime       bool   // 是否启用支付时间
	QueryStartTime      uint   // 查询开始时间
	QueryEndTime        uint   // 查询结束时间
	Status              int    // 订单状态,-1=全部、0=待支付、1=已支付、2=已取消、3=已完成
	BillType            int    // 订单类型,-1=全部、0=餐单、1=外卖
	DiningMethod        int    // 用餐方式,-1=全都、 0-堂食 1-打包
	SaleBillUuids       string // 销售账单UUID列表，多个UUID用逗号分隔
	IsOnlyDataManage    int    // 是否只包含数据管理, 0-不包含、1-包含
	IsContainDataManage int    // 是否包含数据管理, 0-不包含、1-包含
}

// MonthlyOrderRank 每月订单排名信息
type MonthlyOrderRank struct {
	MonthYear          string `json:"month_year"`           // 年-月格式，如 "2024-09"
	FirstOrderUuid     uint64 `json:"first_order_uuid"`     // 该月的第一条订单ID（基于全表数据）
	OrderUuid          uint64 `json:"order_uuid"`           // 订单ID
	OrderNo            string `json:"order_no"`             // 订单编号
	MonthlyOrderNumber int    `json:"monthly_order_number"` // 该订单在当月是第几条数据（基于全表数据）
}

func (r *orderRepo) getOrderListDBOption(param GetCashierOrderListWithPaginationType, tz string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		// 订单编号
		if param.OrderNo != "" {
			db = db.Where("order_no like ?", "%"+param.OrderNo+"%")
		}
		// 账单类型
		if param.BillType != -1 {
			db = db.Where("bill_type = ?", param.BillType)
		}
		if param.DiningMethod != -1 {
			db = db.Where("dining_method = ?", param.DiningMethod)
		}
		//  日期类型 -1-全都 1-今天 2-昨天 3-本周 4-本月 5-本年 6-近7天 7-上个月
		if param.DateType >= 0 && param.DateType <= 3 {
			var startTime, endTime int64
			switch param.DateType {
			case constant.OrderDateTypeToday: // 今天
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeToday)
			case constant.OrderDateTypeYesterday: // 昨天
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeYesterday)
			case constant.OrderDateTypeWeek: // 本周
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeThisWeek)
			case constant.OrderDateTypeMonth: // 本月
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeThisMonth)
			case constant.OrderDateTypeYear: // 本年
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeThisYear)
			case constant.OrderDateTypeLastWeek: // 近7天
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeLastWeek)
			case constant.OrderDateTypeLastMonth: // 上个月
				startTime, endTime, _ = utils.SetTimezone(tz).GetTimeRange(utils.DayTypeLastMonth)
			}
			// db = db.Where("create_time BETWEEN ? AND ?", startTime, endTime)
			param.QueryStartTime = uint(startTime)
			param.QueryEndTime = uint(endTime)
		}
		// 日期范围
		if param.QueryStartTime != 0 || param.QueryEndTime != 0 {
			timeFields := []string{}
			if param.EnableCreateTime || !param.EnablePayTime {
				timeFields = append(timeFields, "create_time")
			}
			if param.EnablePayTime {
				timeFields = append(timeFields, "finish_time")
			}
			// 开始时间
			endTime := uint(0)
			if param.QueryEndTime != 0 {
				endTime = param.QueryEndTime
			}
			//
			query := ""
			args := []interface{}{}
			for i, field := range timeFields {
				if i > 0 {
					query += " OR "
				}
				if param.QueryStartTime > 0 && endTime > 0 {
					query += fmt.Sprintf("(%s BETWEEN ? AND ?)", field)
					args = append(args, param.QueryStartTime, endTime)
				} else if param.QueryStartTime > 0 {
					query += fmt.Sprintf("(%s > ?)", field)
					args = append(args, param.QueryStartTime)
				} else if endTime > 0 {
					query += fmt.Sprintf("(%s < ? AND %s > 0)", field, field)
					args = append(args, endTime)
				}
			}
			if query != "" {
				db = db.Where(query, args...)
			}
		}
		//
		return db
	}
}

// GetCashierOrderListWithPagination 获取收银台订单列表
func (r *orderRepo) GetCashierOrderListWithPagination(param GetCashierOrderListWithPaginationType, tz string) (lists []model.SaleBill, total int64, dbOption DBOption, err error) {
	// 额外条件
	dbOption = r.getOrderListDBOption(param, tz)

	opts := []DBOption{
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders.PaymentMethod",
			},
			WithPreload{
				Query: "SaleOrders.ReturnOrders",
			},
			WithPreload{
				Query: "SaleOrders.Member",
			},
			WithPreload{
				Query: "DataManage",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereByType(model.DataManageTypeOrder)),
				},
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByCooking(),
		CommonRepo.WhereInBillType([]uint{constant.SaleBillTypeDesk, constant.SaleBillTypeInstant}),
		CommonRepo.SortWithID("DESC"),
		dbOption,
		//
		func() DBOption {
			return func(db *gorm.DB) *gorm.DB {
				//  账单状态
				if param.Status != -1 {
					db = db.Where("status = ?", uint(param.Status))
				}
				//
				return db
			}
		}(),
	}
	if param.IsOnlyDataManage == 1 {
		uuidList := strings.Split(param.SaleBillUuids, ",")
		uuids := []uint64{}
		for _, uuid := range uuidList {
			uuid, _ := strconv.ParseUint(uuid, 10, 64)
			uuids = append(uuids, uint64(uuid))
		}
		opts = append(opts, CommonRepo.WhereInUuids(uuids))
	}
	if param.IsOnlyDataManage == 0 && param.IsContainDataManage == 0 {
		opts = append(opts,
			func() DBOption {
				return func(db *gorm.DB) *gorm.DB {
					return db.Where("uuid NOT IN (SELECT data_uuid FROM ttpos_data_manage WHERE type = ?)", model.DataManageTypeOrder)
				}
			}(),
		)
	}

	//
	lists, total, err = r.GetOrderListWithPagination(
		param.PageNo,
		param.PageSize,
		opts...,
	)
	if err != nil {
		return nil, 0, dbOption, fmt.Errorf("GetCashierOrderListWithPagination: %v", err)
	}
	return lists, total, dbOption, nil
}

// GetCashierOrderExportListWithPagination 获取收银台导出订单列表
func (r *orderRepo) GetCashierOrderExportListWithPagination(param GetCashierOrderListWithPaginationType, tz string) (lists []model.SaleBill, total int64, dbOption DBOption, err error) {
	// 额外条件
	dbOption = r.getOrderListDBOption(param, tz)
	//
	lists, total, err = r.GetOrderListWithPagination(
		param.PageNo,
		param.PageSize,
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders.PaymentMethod",
			},
			WithPreload{
				Query: "SaleOrders.ReturnOrders",
			},
			WithPreload{
				Query: "SaleOrders.Member",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.CancelReasons.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ReturnOrderProducts",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetDelayProducts.ReturnOrderProducts",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.BuffetPackage.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.ReturnOrderProducts",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.BuffetCustomerTypePrice.BuffetCustomerType",
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByCooking(),
		CommonRepo.SortWithID("DESC"),
		dbOption,
		//
		func() DBOption {
			return func(db *gorm.DB) *gorm.DB {
				//  账单状态
				if param.Status != -1 {
					db = db.Where("status = ?", uint(param.Status))
				}
				return db
			}
		}(),
	)
	if err != nil {
		return nil, 0, dbOption, fmt.Errorf("GetCashierOrderListWithPagination: %v", err)
	}
	return lists, total, dbOption, nil
}

// GetSaleBillInfo 获取销售账单详细信息
func (r *orderRepo) GetSaleBillInfo(saleBillUuid uint64, saleOrderUuid uint64) (model.SaleBill, error) {
	info, err := r.GetSaleBill(
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleBillSetting",
			},
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						db = db.Where("delete_time = ?", constant.NotDeleted)
						if saleOrderUuid > constant.OptionalUuid {
							db = db.Where("uuid = ?", saleOrderUuid)
						}
						return db
					},
				},
			},
			WithPreload{
				Query: "SaleOrders.ReturnOrders",
			},
			WithPreload{
				Query: "SaleOrders.FreeReasons.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.Member",
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ImageFile",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.CancelReasons.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ReturnOrderProducts",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetDelayProducts.ReturnOrderProducts",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.BuffetPackage.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.ReturnOrderProducts",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.BuffetCustomerTypePrice.BuffetCustomerType",
			},
			WithPreload{
				Query: "Desk",
			},
			WithPreload{
				Query: "OrderSource.MultiLanguageName",
			},
			WithPreload{
				Query: "Nationality.MultiLanguageName",
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByUuid(saleBillUuid),
	)
	if err != nil {
		return model.SaleBill{}, fmt.Errorf("GetSaleBillInfo: %v", err)
	}
	return info, nil
}

// GetSaleBillInfoByDesk 获取桌台的账单信息
func (r *orderRepo) GetSaleBillInfoByDesk(deskUuid uint64, saleOrderUuid uint64) (model.SaleBill, error) {
	info, err := r.GetSaleBill(
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						db = db.Where("delete_time = ?", constant.NotDeleted)
						if saleOrderUuid > constant.OptionalUuid {
							db = db.Where("uuid = ?", saleOrderUuid)
						}
						return db
					},
				},
			},
			WithPreload{
				Query: "SaleBillSetting",
			},
			WithPreload{
				Query: "OrderSource.MultiLanguageName",
				Args: []any{
					func(db *gorm.DB) *gorm.DB {
						return db.Where("delete_time = ?", constant.NotDeleted)
					},
				},
			},
			WithPreload{
				Query: "Nationality.MultiLanguageName",
				Args: []any{
					func(db *gorm.DB) *gorm.DB {
						return db.Where("delete_time = ?", constant.NotDeleted)
					},
				},
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByDeskUuid(deskUuid),
	)
	if err != nil {
		return model.SaleBill{}, fmt.Errorf("GetSaleBillInfoByDesk: %v", err)
	}
	return info, nil
}

type OrderCartInfoOption struct {
	UnorderedH5Product   int    // 1-查询H5未下单的商品 2-查询H5已下单的商品 3-查询H5已下单的商品和被拒单的商品
	H5OrderUuid          uint64 // 指定某个h5订单
	FilterEndStatus      bool   // 指定传入的salebill状态
	NotDeleted           bool   // 查询未被删除的商品
	NoQueryMustPlan      bool   // 不查询必点信息
	H5AutoAdd            bool   // 是否是H5自动添加的商品
	NoAutoAdd            bool   // 是否不自动添加的商品。如果为true，则不自动加购的商品，平板上操作时不自动加购
	CanCloseMustPlanView bool   // 是否可以关闭必点弹窗
	CompanyUuid          uint64 // 公司UUID，用于控制是否使用对象存储层
}

const (
	UnorderedH5Product         = 1
	OrderedH5Product           = 2
	OrderedH5ProductWithReject = 3
)

type OrderCartInfoOptionFunc func(option *OrderCartInfoOption)

// WithUnorderedH5Product 查询H5未下单的商品
func WithUnorderedH5Product() OrderCartInfoOptionFunc {
	return func(option *OrderCartInfoOption) {
		option.UnorderedH5Product = UnorderedH5Product
	}
}

// WithOrderedH5Product 查询H5已下单的商品
func WithOrderedH5Product() OrderCartInfoOptionFunc {
	return func(option *OrderCartInfoOption) {
		option.UnorderedH5Product = OrderedH5Product
	}
}

// WithOrderedH5ProductWithReject 查询H5已下单的商品和被拒单的商品
func WithOrderedH5ProductWithReject() OrderCartInfoOptionFunc {
	return func(option *OrderCartInfoOption) {
		option.UnorderedH5Product = OrderedH5ProductWithReject
	}
}

// WithH5OrderUuid 指定某个h5订单
func WithH5OrderUuid(h5OrderUuid uint64) OrderCartInfoOptionFunc {
	return func(option *OrderCartInfoOption) {
		option.H5OrderUuid = h5OrderUuid
	}
}

// WithNotDeleted 查询未被删除的商品
func WithNotDeleted() OrderCartInfoOptionFunc {
	return func(option *OrderCartInfoOption) {
		option.NotDeleted = true
	}
}

// FilterEndStatus 过滤结束状态
func FilterEndStatus() OrderCartInfoOptionFunc {
	return func(option *OrderCartInfoOption) {
		option.FilterEndStatus = true
	}
}

// WithNoQueryMustPlan 不查询必点信息
func WithNoQueryMustPlan() OrderCartInfoOptionFunc {
	return func(option *OrderCartInfoOption) {
		option.NoQueryMustPlan = true
	}
}

// WithH5AutoAdd 是否是H5自动添加的商品.如果需要自动加购时，加购的商品应该为未下单商品
func WithH5AutoAdd() OrderCartInfoOptionFunc {
	return func(option *OrderCartInfoOption) {
		option.H5AutoAdd = true
	}
}

// WithNoAutoAdd 是否不自动添加的商品。如果为true，则不自动加购的商品，平板上操作时不自动加购
func WithNoAutoAdd() OrderCartInfoOptionFunc {
	return func(option *OrderCartInfoOption) {
		option.NoAutoAdd = true
	}
}

// WithCanCloseMustPlanView 是否可以关闭必点弹窗
func WithCanCloseMustPlanView() OrderCartInfoOptionFunc {
	return func(option *OrderCartInfoOption) {
		option.CanCloseMustPlanView = true
	}
}

// WithCompanyUuid 设置公司UUID
func WithCompanyUuid(companyUuid uint64) OrderCartInfoOptionFunc {
	return func(option *OrderCartInfoOption) {
		option.CompanyUuid = companyUuid
	}
}

// GetOrderCartInfo 获取购物车信息
func (r *orderRepo) GetOrderCartInfo(saleBillUuid uint64, opts ...OrderCartInfoOptionFunc) (*ro.ShopCartRepo, error) {
	option := &OrderCartInfoOption{}
	for _, opt := range opts {
		opt(option)
	}

	// 只查询常规的购物车商品
	filterProduct := CommonRepo.DBOption(CommonRepo.FilterSaleOrderProduct())
	if option.UnorderedH5Product == UnorderedH5Product {
		// 只查询H5未下单的商品
		filterProduct = CommonRepo.DBOption(CommonRepo.FilterSaleOrderProductH5Unordered())
	} else if option.UnorderedH5Product == OrderedH5Product {
		// 只查询H5已下单的商品
		filterProduct = CommonRepo.DBOption(CommonRepo.FilterSaleOrderProductH5Ordered())
	} else if option.UnorderedH5Product == OrderedH5ProductWithReject {
		// 只查询H5已下单的商品和被拒单的商品
		filterProduct = CommonRepo.DBOption(CommonRepo.FilterSaleOrderProductH5OrderedWithReject())
	} else if option.H5OrderUuid > constant.OptionalUuid {
		// 查询常规的购物车商品、某个h5订单的商品
		filterProduct = CommonRepo.DBOption(CommonRepo.FilterSaleOrderProductWithH5Order(option.H5OrderUuid))
	} else if option.NotDeleted {
		// 查询未被删除的商品
		filterProduct = CommonRepo.DBOption(CommonRepo.WhereBySoftDelete())
	}

	repo := NewSaleBillRepo(r.db)
	saleBill, err := repo.GetSaleBill(
		CommonRepo.WhereByUuid(saleBillUuid),
		// CommonRepo.WhereBySoftDelete(), // 软删除的账单也查询出来，告诉前端该订单已经删除
		CommonRepo.WhereByIsHide(false),
	)
	if err != nil {
		return nil, errors.WithMessage(fmt.Errorf("GetOrderCartInfo: %v, saleBillUuid: %d", err, saleBillUuid))
	}

	if saleBill.IsDeskSaleBill() {
		if saleBill.IsBuffetSaleBill() {
			// 当销售账单是自助餐账单时，额外查询自助餐信息
			{
				saleBill, errDesk := repo.GetSaleBill(
					CommonRepo.WhereByUuid(saleBillUuid),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleBillSetting",
						},
						WithPreload{
							Query: "Desk",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "BuffetPackage1.MultiLanguageName",
						},
						WithPreload{
							Query: "BuffetPackage1.BuffetProducts",
						},
						WithPreload{
							Query: "BuffetPackage2.BuffetProducts",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "BuffetPackage2.MultiLanguageName",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders",
							Args: []interface{}{
								func(db *gorm.DB) *gorm.DB {
									return db.Where("delete_time = ?", constant.NotDeleted)
								},
							},
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts",
							Args:  []any{filterProduct},
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.H5Order",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductPackage",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductPackage.ProductCategory",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductPackage.ProductBoms",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductPackage.ProductPackageAttributeGroups",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.OrderItemRemarks",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductionOrderProduct",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductMustPlan",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes",
							Args: []any{
								CommonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
									return db.Order("id asc")
								}),
								CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
							},
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms",
							Args: []any{
								CommonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
									return db.Order("product_bom_uuid asc")
								}),
								CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
							},
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.BatchTag.MultiLanguageName",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderBuffetCustomerTypes.BuffetPackage.MultiLanguageName",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderBuffetCustomerTypes.BuffetCustomerTypePrice.BuffetCustomerType",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderBuffetDelayProducts",
						},
					),
				)
				if errDesk != nil {
					return nil, errors.WithMessage(errDesk)
				}

				bill := &saleBill
				// 计算一次金额，避免错误
				bill.CalcAll()
				return &ro.ShopCartRepo{SaleBill: bill}, nil
			}
		} else {
			// 当销售账单是桌台订单时，额外查询桌台信息
			if option.CompanyUuid == 7709131161600000 {
				// 先获取 SaleBill（不使用 Preload）
				saleBill, errDesk := repo.GetSaleBill(
					CommonRepo.WhereByUuid(saleBillUuid),
					// 只 Preload SaleOrders 和 SaleOrderProducts（这些是必须的，因为需要过滤条件）
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders",
							Args: []interface{}{
								func(db *gorm.DB) *gorm.DB {
									return db.Where("delete_time = ?", constant.NotDeleted)
								},
							},
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts",
							Args:  []any{filterProduct},
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.H5Order",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductMustPlan",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.OrderItemRemarks",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes",
							Args: []any{
								CommonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
									return db.Order("id asc")
								}),
								CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
							},
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms",
							Args: []any{
								CommonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
									return db.Order("product_bom_uuid asc")
								}),
								CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
							},
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductionOrderProduct",
						},
					),
				)
				if errDesk != nil {
					return nil, errors.WithMessage(errDesk)
				}

				// 获取 company uuid（优先使用 option 中的，否则从 db 获取）
				companyUuid := option.CompanyUuid
				if companyUuid == 0 {
					companyUuid = GetCompanyUuid(r.db)
				}

				// 只有 companyUuid 为 7709131161600000 时才使用对象存储层逻辑
				const targetCompanyUuid = 7709131161600000
				if companyUuid == targetCompanyUuid {
					// 使用对象存储层自动注入关联对象
					ctx := context.NewContext(context.WithCompanyUuid(companyUuid), context.WithContext(goCtx.Background()))

					// 创建缓存组配置
					groupConfig := cache.GroupConfig{
						Name:             "object-storage",
						EnableLocalCache: true,             // 开启 L1 本地缓存
						EnableRedisCache: true,             // 开启 L2 Redis 缓存
						NegativeTTL:      30 * time.Second, // 负缓存 30 秒
					}

					// 获取关联配置（传入 groupConfig 和 underlyingCache，为每种对象类型创建具体的缓存适配器）
					associations := getSaleBillAssociationsForOrderCart(ctx, r.db, groupConfig, cache.Global)

					// 创建对象存储实例（使用 any 类型用于 PreloadWithConfig，但内部 QueryFunc 使用具体类型）
					// 注意：PreloadWithConfig 只需要 associations，不需要 config
					objectStorage := persistence.NewObjectStorage[any](&service.Config[any]{
						TTL:          5 * time.Minute,
						DisableCache: false,
						CacheLayer:   adapter.NewCacheGroupAdapter[any](groupConfig, cache.Global, 5*time.Minute),
					})

					// 注入关联对象
					if err := objectStorage.PreloadWithConfig(ctx, &saleBill, associations); err != nil {
						// 如果对象存储层注入失败，记录错误但不影响主流程
						return nil, errors.WithMessage(fmt.Errorf("对象存储层注入失败: %v", err))
					}
				}

				bill := &saleBill
				// 计算一次金额，避免错误
				bill.CalcAll()
				return &ro.ShopCartRepo{SaleBill: bill}, nil
			} else // 当销售账单是桌台订单时，额外查询桌台信息
			{
				saleBill, errDesk := repo.GetSaleBill(
					CommonRepo.WhereByUuid(saleBillUuid),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleBillSetting",
						},
						WithPreload{
							Query: "Desk",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders",
							Args: []interface{}{
								func(db *gorm.DB) *gorm.DB {
									return db.Where("delete_time = ?", constant.NotDeleted)
								},
							},
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts",
							Args:  []any{filterProduct},
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.H5Order",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductMustPlan",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.OrderItemRemarks",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductPackage",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductPackage.ProductCategory",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductPackage.ProductBoms",
						},
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductPackage.ProductPackageAttributeGroups",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.ProductionOrderProduct",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes",
							Args: []any{
								CommonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
									return db.Order("id asc")
								}),
								CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
							},
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms",
							Args: []any{
								CommonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
									return db.Order("product_bom_uuid asc")
								}),
								CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
							},
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName",
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.BatchTag.MultiLanguageName",
						},
					),
				)
				if errDesk != nil {
					return nil, errors.WithMessage(errDesk)
				}
				bill := &saleBill
				// 计算一次金额，避免错误
				bill.CalcAll()
				return &ro.ShopCartRepo{SaleBill: bill}, nil
			}
		}
	} else {
		// 当销售账单是点餐订单时，只查询账单信息
		{ // 通过销售订单ID得到订单商品列表、订单金额信息、账单的销售订单列表
			saleBill, errSaleBill := r.GetSaleBill(
				CommonRepo.WhereByUuid(saleBillUuid),
				CommonRepo.WhereBySoftDelete(),
				CommonRepo.Preload(
					WithPreload{
						Query: "SaleBillSetting",
					},
					WithPreload{
						Query: "Desk",
						Args: []interface{}{
							func(db *gorm.DB) *gorm.DB {
								return db.Where("delete_time = ?", constant.NotDeleted)
							},
						},
					},
				),
				CommonRepo.Preload(
					WithPreload{
						Query: "SaleOrders",
						Args: []interface{}{
							func(db *gorm.DB) *gorm.DB {
								return db.Where("delete_time = ?", constant.NotDeleted)
							},
						},
					},
				),
				CommonRepo.Preload(
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts",
						Args: []interface{}{
							func(db *gorm.DB) *gorm.DB {
								return db.Where("delete_time = ? AND is_accept_order = ?", constant.NotDeleted, constant.OrderProductIsAcceptOrderAccepted)
							},
						},
					},
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.H5Order",
					},
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.ProductMustPlan",
					},
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.OrderItemRemarks",
					},
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.ProductPackage",
					},
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.ProductPackage.ProductCategory",
					},
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.ProductPackage.ProductBoms",
					},
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.ProductPackage.ProductPackageAttributeGroups",
					},
				),
				CommonRepo.Preload(
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.ProductionOrderProduct",
					},
				),
				CommonRepo.Preload(
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
					},
				),
				CommonRepo.Preload(
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes",
						Args: []any{
							CommonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
								return db.Order("id asc")
							}),
							CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
						},
					},
				),
				CommonRepo.Preload(
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName",
					},
				),
				CommonRepo.Preload(
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms",
						Args: []any{
							CommonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
								return db.Order("product_bom_uuid asc")
							}),
							CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
						},
					},
				),
				CommonRepo.Preload(
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom",
					},
				),
				CommonRepo.Preload(
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName",
					},
				),
				CommonRepo.Preload(
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName",
					},
				),
				CommonRepo.Preload(
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.BatchTag.MultiLanguageName",
					},
				),
			)
			if errSaleBill != nil {
				return nil, errors.WithMessage(fmt.Errorf("GetOrderCartInfo errSaleBill: %v", errSaleBill))
			}
			bill := &saleBill
			// 计算一次金额，避免错误
			bill.CalcAll()
			return &ro.ShopCartRepo{SaleBill: bill}, nil
		}
	}
}

// GetOrderBuffetInfo 获取订单自助餐信息
func (r *orderRepo) GetOrderBuffetInfo(saleBillUuid, saleOrderUuid uint64) (model.SaleBill, error) {
	return NewSaleBillRepo(r.db).GetSaleBill(
		CommonRepo.WhereByUuid(saleBillUuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						return db.Where("uuid = ? AND delete_time = ?", saleOrderUuid, constant.NotDeleted)
					},
				},
			},
		),
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.BuffetPackage.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.BuffetCustomerTypePrice.BuffetCustomerType",
			},
		),
	)
}

// GetSaleBillInfoAndProduct 获取销售账单详细信息-包含商品信息
func (r *orderRepo) GetSaleBillInfoAndProduct(saleBillUuid uint64, saleOrderUuid uint64, saleOrderProductUuid uint64) (model.SaleBill, error) {
	info, err := r.GetSaleBill(
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						db = db.Where("delete_time = ?", 0)
						if saleOrderUuid > 0 {
							db = db.Where("uuid = ?", saleOrderUuid)
						}
						return db
					},
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						db = db.Where("delete_time = ?", 0)
						if saleOrderProductUuid > 0 {
							db = db.Where("uuid = ?", saleOrderProductUuid)
						}
						return db
					},
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetDelayProducts",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						db = db.Where("delete_time = ?", 0)
						return db
					},
				},
			},
			WithPreload{
				Query: "SaleBillSetting",
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByUuid(saleBillUuid),
	)
	if err != nil {
		return model.SaleBill{}, fmt.Errorf("GetSaleBillInfoAndProduct: %v", err)
	}
	return info, nil
}

// GetSaleBillInfoAndMember 获取销售账单详细信息-包含会员信息
func (r *orderRepo) GetSaleBillInfoAndMember(saleBillUuid uint64) (model.SaleBill, error) {
	info, err := r.GetSaleBill(
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						db = db.Where("delete_time = ?", 0)
						return db
					},
				},
			},
			WithPreload{
				Query: "SaleOrders.Member",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						db = db.Where("delete_time = ?", 0)
						return db
					},
				},
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByUuid(saleBillUuid),
	)
	if err != nil {
		return model.SaleBill{}, fmt.Errorf("GetSaleBillInfoAndMember: %v", err)
	}
	return info, nil
}

// GetSaleBillInfoAndPaymentOrders 获取销售账单详细信息-包含支付信息
func (r *orderRepo) GetSaleBillInfoAndPaymentOrders(saleBillUuid uint64, saleOrderUuid uint64, saleOrderPaymentUuid uint64) (model.SaleBill, error) {
	info, err := r.GetSaleBill(
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						db = db.Where("delete_time = ?", 0)
						if saleOrderUuid > 0 {
							db = db.Where("uuid = ?", saleOrderUuid)
						}
						return db
					},
				},
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						db = db.Where("delete_time = ?", 0)
						if saleOrderPaymentUuid > 0 {
							db = db.Where("uuid = ?", saleOrderPaymentUuid)
						}
						return db
					},
				},
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByUuid(saleBillUuid),
	)
	if err != nil {
		return model.SaleBill{}, fmt.Errorf("GetSaleBillInfoAndProduct: %v", err)
	}
	return info, nil
}

// GetSaleOrderProductListBySaleOrderProductUuids 根据销售订单商品uuid列表获取销售订单商品列表
func (r *orderRepo) GetSaleOrderProductListBySaleOrderProductUuids(saleOrderProductUuids []uint64) ([]model.SaleOrderProduct, error) {
	products := make([]model.SaleOrderProduct, 0)
	err := r.db.Model(&model.SaleOrderProduct{}).Where("uuid in ?", saleOrderProductUuids).Find(&products).Error
	if err != nil {
		return nil, fmt.Errorf("GetSaleOrderProductListBySaleOrderProductUuids: %v", err)
	}
	return products, nil
}

// GetSaleBillDetails 获取销售账单详细信息 - 几乎包含所有的关联
func (r *orderRepo) GetSaleBillDetails(saleBillUuid uint64, saleOrderUuid uint64) (model.SaleBill, error) {
	info, err := r.GetSaleBill(
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						db = db.Where("delete_time = ?", constant.NotDeleted)
						if saleOrderUuid > 0 {
							db = db.Where("uuid = ?", saleOrderUuid)
						}
						return db
					},
				},
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders",
			},
			WithPreload{
				Query: "SaleOrders.Member",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.ReturnOrders",
			},
			WithPreload{
				Query: "Cashier",
			},
			WithPreload{
				Query: "OrderSource.MultiLanguageName",
				// 移除 delete_time 过滤，保证历史订单可显示已删除的配置名称
			},
			WithPreload{
				Query: "Nationality.MultiLanguageName",
				// 移除 delete_time 过滤，保证历史订单可显示已删除的配置名称
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByUuid(saleBillUuid),
	)
	if err != nil {
		return model.SaleBill{}, fmt.Errorf("GetSaleBillDetails: %v", err)
	}
	return info, nil
}

// CancelOrder 取消订单
func (r *orderRepo) CancelOrder(ctx context.Context, saleBillUuid uint64, deskUuid uint64, reason string) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.SaleOrder{}).Where("sale_bill_uuid = ?", saleBillUuid).Where("status = ?", constant.SaleOrderStatusPending).Update("status", constant.SaleOrderStatusCanceled).Error
		if err != nil {
			return errors.WithMessage(err)
		}
		// 获取当前员工信息
		staff := ctx.GetStaff()
		//
		return tx.Model(&model.SaleBill{}).
			Where("uuid = ?", saleBillUuid).
			Where("status = ?", constant.SaleBillStatusPending).
			Updates(map[string]interface{}{
				"desk_uuid":    deskUuid,
				"uuid":         saleBillUuid,
				"status":       constant.SaleBillStatusCanceled,
				"reason":       reason,
				"duty_no":      staff.DutyNo,
				"cashier_uuid": staff.Uuid,
				"cashier_name": staff.GetUserName(),
			}).Error
	})
	if err != nil {
		return fmt.Errorf("CancelOrder: %v", err)
	}
	return nil
}

// CancelDeskOrder 关闭桌台订单
func (r *orderRepo) CancelDeskOrder(ctx context.Context, deskUuid uint64, reason string) error {
	var saleBill model.SaleBill
	if err := r.db.Model(&model.SaleBill{}).
		Where("status = ?", constant.SaleBillStatusPending).
		Where("desk_uuid = ?", deskUuid).
		First(&saleBill).Error; err != nil {
		return fmt.Errorf("CancelDeskOrder: %v", err)
	}
	return r.CancelOrder(ctx, saleBill.Uuid, deskUuid, reason)
}

// DeleteOrder 软删除订单
func (r *orderRepo) DeleteOrder(saleBillUuid uint64, saleOrderUuid uint64) error {
	tx := r.db.Begin()
	now := uint(time.Now().Unix())
	// 如果是删除全部订单或只剩最后一个订单,则同时删除主订单
	if saleOrderUuid == 0 {
		// 删除销售订单
		if err := tx.Model(&model.SaleOrder{}).Where("sale_bill_uuid = ?", saleBillUuid).Update("delete_time", now).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("DeleteOrder: %v", err)
		}
		// 删除销售账单
		if err := tx.Model(&model.SaleBill{}).Where("uuid = ? AND status = ?", saleBillUuid, constant.SaleBillStatusCanceled).Update("delete_time", now).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("DeleteOrder: %v", err)
		}
	} else {
		// 删除销售订单
		if err := tx.Model(&model.SaleOrder{}).
			Where("sale_bill_uuid = ? AND status = ?", saleBillUuid, constant.SaleOrderStatusCanceled).
			Where("uuid = ?", saleOrderUuid).
			Update("delete_time", now).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("DeleteOrder: %v", err)
		}
	}

	err := tx.Commit().Error
	if err != nil {
		return fmt.Errorf("DeleteOrder: %v", err)
	}
	return nil
}

// HideOrder 隐藏订单
func (r *orderRepo) HideOrder(saleBillUuid uint64) error {
	now := uint(time.Now().Unix())
	tx := r.db.Begin()
	// 删除销售订单
	if err := tx.Model(&model.SaleBill{}).
		Where("uuid = ? AND status = ?", saleBillUuid, constant.SaleBillStatusPending).
		Update("hide_bill_time", now).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("HideOrder: %v", err)
	}
	err := tx.Commit().Error
	if err != nil {
		return fmt.Errorf("HideOrder: %v", err)
	}
	return nil
}

// GetSaleBillRecord 获取销售账单记录
func (r *orderRepo) GetSaleBillRecord(saleBillUuid uint64) (*model.SaleBill, error) {
	var saleBill model.SaleBill
	if err := r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBillUuid).First(&saleBill).Error; err != nil {
		return nil, fmt.Errorf("GetSaleBillRecord: %v", err)
	}
	return &saleBill, nil
}

// GetSaleBillSaleOrderRecord 获取销售账单记录
func (r *orderRepo) GetSaleBillSaleOrderRecord(saleOrderUuid uint64) (*model.SaleOrder, error) {
	var saleOrder model.SaleOrder
	if err := r.db.Model(&model.SaleOrder{}).Where("uuid = ?", saleOrderUuid).First(&saleOrder).Error; err != nil {
		return nil, fmt.Errorf("GetSaleBillSaleOrderRecord: %v", err)
	}
	return &saleOrder, nil
}

// GetSaleBillAllInfo 获取销售账单所有信息
func (r *orderRepo) GetCacheSaleBillAllInfo(saleBillUuid uint64) (*model.SaleBill, error) {
	cacheDataRepo := NewCacheDataRepo(r.db)
	multiLanguageNameUuids := []uint64{}

	// 查询销售账单
	saleBill, err := r.GetSaleBill(
		CommonRepo.Preload(
			// ==================== 销售账单的收银员信息 ====================
			WithPreload{
				Query: "Cashier",
			},
			// ==================== 销售账单的账单设置 ====================
			WithPreload{
				Query: "SaleBillSetting",
			},
			WithPreload{
				Query: "SaleOrders",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			// ==================== 销售账单的优惠券信息 ====================
			WithPreload{
				Query: "SaleOrders.Coupons",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.Coupons.MarketingCoupon",
			},
			WithPreload{
				Query: "SaleOrders.Coupons.MemberCoupon",
			},
			// ==================== 销售账单的支付信息 ====================
			WithPreload{
				Query: "SaleOrders.PaymentOrders",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
					CommonRepo.DBOption(CommonRepo.WhereByStatus(constant.PaymentOrderStatusPaid)),
				},
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders.PaymentMethod",
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders.ReturnOrderAmounts",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders.ReturnOrderAmounts.PaymentMethod",
			},
			// ==================== 销售账单的会员信息 ====================
			WithPreload{
				Query: "SaleOrders.Member.MemberLevel",
			},
			WithPreload{
				Query: "SaleOrders.Member.MemberCard.MemberCardType",
			},
			WithPreload{
				Query: "SaleOrders.Member.MemberBalanceLog",
			},
			// ==================== 销售账单的商品信息 ====================
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ReturnOrderProducts",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.CancelReasons",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes",
				Args: []any{
					CommonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
						return db.Order("id asc")
					}),
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				// 用于检查商品包是否下架
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms",
				Args: []any{
					CommonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
						return db.Order("product_bom_uuid asc")
					}),
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			// ==================== 销售账单的退款信息 ====================
			WithPreload{
				Query: "SaleOrders.ReturnOrders",
			},
			// ==================== 销售账单的会员积分信息 ====================
			WithPreload{
				Query: "SaleOrders.MemberPointLogs",
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByUuid(saleBillUuid),
	)
	if err != nil {
		return &saleBill, fmt.Errorf("GetSaleBill: %v", err)
	}

	// ==================== 查询桌台信息 ====================
	// 查询桌台
	if saleBill.DeskUuid != 0 {
		if desk, err := cacheDataRepo.GetCacheDesk(saleBill.DeskUuid); err == nil {
			saleBill.Desk = &desk
		} else {
			return &saleBill, fmt.Errorf("GetSaleBill: %v", err)
		}
	}
	// 查询自助餐套餐1
	if saleBill.BuffetPackage1Uuid != 0 {
		if buffetPackage1, err := cacheDataRepo.GetCacheBuffetPackage(saleBill.BuffetPackage1Uuid); err == nil {
			multiLanguageNameUuids = append(multiLanguageNameUuids, buffetPackage1.MultiLanguageNameUuid)
			// 加载自助餐套餐1顾客类型价格。用于判断顾客价格是否改变
			buffetCustomerTypePrices, err := cacheDataRepo.GetCacheBuffetCustomerTypePrices(buffetPackage1.Uuid)
			if err != nil {
				return &saleBill, fmt.Errorf("GetSaleBill: %v", err)
			}
			buffetPackage1.BuffetCustomerTypePrices = buffetCustomerTypePrices

			// 加载自助餐套餐1商品。用于判断加购的商品是否属于自助餐套餐1
			buffetProducts, err := cacheDataRepo.GetCacheBuffetProducts(buffetPackage1.Uuid)
			if err != nil {
				return &saleBill, fmt.Errorf("GetSaleBill: %v", err)
			}
			buffetPackage1.BuffetProducts = buffetProducts

			// 赋值自助餐套餐1
			saleBill.BuffetPackage1 = &buffetPackage1
		} else {
			return &saleBill, fmt.Errorf("GetSaleBill: %v", err)
		}
	}
	// 查询自助餐套餐2
	if saleBill.BuffetPackage2Uuid != 0 {
		if buffetPackage2, err := cacheDataRepo.GetCacheBuffetPackage(saleBill.BuffetPackage2Uuid); err == nil {
			multiLanguageNameUuids = append(multiLanguageNameUuids, buffetPackage2.MultiLanguageNameUuid)
			// 加载自助餐套餐2顾客类型价格。用于判断顾客价格是否改变
			buffetCustomerTypePrices, err := cacheDataRepo.GetCacheBuffetCustomerTypePrices(buffetPackage2.Uuid)
			if err != nil {
				return &saleBill, fmt.Errorf("GetSaleBill: %v", err)
			}
			buffetPackage2.BuffetCustomerTypePrices = buffetCustomerTypePrices

			// 加载自助餐套餐2商品。用于判断加购的商品是否属于自助餐套餐2
			buffetProducts, err := cacheDataRepo.GetCacheBuffetProducts(buffetPackage2.Uuid)
			if err != nil {
				return &saleBill, fmt.Errorf("GetSaleBill: %v", err)
			}
			buffetPackage2.BuffetProducts = buffetProducts

			// 赋值自助餐套餐2
			saleBill.BuffetPackage2 = &buffetPackage2
		} else {
			return &saleBill, fmt.Errorf("GetSaleBill: %v", err)
		}
	}
	// 查询销售订单商品
	if len(saleBill.SaleOrders) > 0 {
		productPackageUuids := []uint64{}
		productBomUuids := []uint64{}
		productAttributeUuids := []uint64{}
		for _, saleOrder := range saleBill.SaleOrders {

			// 查询自助餐套餐1、2
			if saleBill.IsBuffet == constant.SaleBillIsBuffetYes {
				// 查询自助餐套餐1、2顾客类型
				saleOrderBuffetCustomerTypes, err := NewSaleOrderBuffetCustomerTypeRepo(r.db).GetSaleOrderBuffetCustomerTypes(
					CommonRepo.Preload(
						WithPreload{
							Query: "BuffetPackage.MultiLanguageName",
						},
						WithPreload{
							Query: "BuffetCustomerTypePrice.BuffetCustomerType",
						},
						WithPreload{
							Query: "ReturnOrderProducts",
							Args: []any{
								CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
							},
						},
					),
					CommonRepo.WhereBySaleOrderUuid(saleOrder.Uuid),
					CommonRepo.WhereBySoftDelete(),
				)
				if err != nil {
					return &saleBill, fmt.Errorf("GetSaleBill: %v", err)
				}
				saleOrder.SaleOrderBuffetCustomerTypes = saleOrderBuffetCustomerTypes

				// 查询自助餐套餐加钟
				saleOrderBuffetDelayProducts, err := NewSaleOrderBuffetDelayProductRepo(r.db).GetSaleOrderBuffetDelayProducts(
					CommonRepo.Preload(
						WithPreload{
							Query: "ReturnOrderProducts",
							Args: []any{
								CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
							},
						},
					),
					CommonRepo.WhereBySaleOrderUuid(saleOrder.Uuid),
					CommonRepo.WhereBySoftDelete(),
				)
				if err != nil {
					return &saleBill, fmt.Errorf("GetSaleBill: %v", err)
				}
				saleOrder.SaleOrderBuffetDelayProducts = saleOrderBuffetDelayProducts
			}

			// 查询商品包、商品BOM、商品属性
			for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
				if saleOrderProduct.ProductPackageUuid != 0 {
					multiLanguageNameUuids = append(multiLanguageNameUuids, saleOrderProduct.MultiLanguageNameUuid)
					productPackageUuids = append(productPackageUuids, saleOrderProduct.ProductPackageUuid)
					if len(saleOrderProduct.SaleOrderProductBoms) > 0 {
						for _, saleOrderProductBom := range saleOrderProduct.SaleOrderProductBoms {
							if saleOrderProductBom.IsDelete() {
								continue
							}
							productBomUuids = append(productBomUuids, saleOrderProductBom.ProductBomUuid)
						}
					}
					if len(saleOrderProduct.SaleOrderProductAttributes) > 0 {
						for _, saleOrderProductAttribute := range saleOrderProduct.SaleOrderProductAttributes {
							if saleOrderProductAttribute.IsDelete() {
								continue
							}
							productAttributeUuids = append(productAttributeUuids, saleOrderProductAttribute.ProductAttributeUuid)
						}
					}
				}
			}
			// 商品包
			if len(productPackageUuids) > 0 {
				productPackages, err := cacheDataRepo.GetCacheAllProductPackageByUuids(productPackageUuids)
				if err != nil {
					return &saleBill, fmt.Errorf("GetSaleBill: %v", err)
				}
				for _, productPackage := range productPackages {
					for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
						if saleOrderProduct.ProductPackageUuid == productPackage.Uuid {
							saleOrderProduct.ProductPackage = productPackage
						}
					}
				}
			}
			// 商品BOM
			if len(productBomUuids) > 0 {
				productBoms, err := cacheDataRepo.GetCacheAllProductBomByUuids(productBomUuids)
				if err != nil {
					return &saleBill, fmt.Errorf("GetSaleBill: %v", err)
				}
				for _, productBom := range productBoms {
					for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
						if len(saleOrderProduct.SaleOrderProductBoms) > 0 {
							for _, saleOrderProductBom := range saleOrderProduct.SaleOrderProductBoms {
								if saleOrderProductBom.IsDelete() {
									continue
								}
								if saleOrderProductBom.ProductBomUuid == productBom.Uuid {
									saleOrderProductBom.ProductBom = *productBom
								}
							}
						}
					}
				}
			}
			// 商品属性
			if len(productAttributeUuids) > 0 {
				productAttributes, err := cacheDataRepo.GetCacheAllProductAttributeByUuids(productAttributeUuids)
				if err != nil {
					return &saleBill, fmt.Errorf("GetSaleBill: %v", err)
				}
				for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
					if len(saleOrderProduct.SaleOrderProductAttributes) > 0 {
						for _, saleOrderProductAttribute := range saleOrderProduct.SaleOrderProductAttributes {
							if saleOrderProductAttribute.IsDelete() {
								continue
							}
							for _, productAttribute := range productAttributes {
								if saleOrderProductAttribute.ProductAttributeUuid == productAttribute.Uuid {
									saleOrderProductAttribute.ProductAttribute = *productAttribute
								}
							}
						}
					}
				}
			}
		}
	}

	// 查询语言
	if len(multiLanguageNameUuids) > 0 {
		var multiLanguageNames []model.MultiLanguageName
		db := r.db.Model(&model.MultiLanguageName{}).Where("uuid in ?", multiLanguageNameUuids)
		if result := db.Find(&multiLanguageNames); result.Error != nil {
			return &saleBill, fmt.Errorf("GetSaleBill: %v", result.Error)
		}
		for _, multiLanguageName := range multiLanguageNames {
			if saleBill.BuffetPackage1 != nil && saleBill.BuffetPackage1.MultiLanguageNameUuid == multiLanguageName.Uuid {
				saleBill.BuffetPackage1.MultiLanguageName = multiLanguageName
			}
			if saleBill.BuffetPackage2 != nil && saleBill.BuffetPackage2.MultiLanguageNameUuid == multiLanguageName.Uuid {
				saleBill.BuffetPackage2.MultiLanguageName = multiLanguageName
			}
			// 商品多语言名称
			for _, saleOrder := range saleBill.SaleOrders {
				for _, saleOrderProduct := range saleOrder.SaleOrderProducts {
					if saleOrderProduct.MultiLanguageNameUuid == multiLanguageName.Uuid {
						saleOrderProduct.MultiLanguageName = &multiLanguageName
					}
				}
			}
		}
	}
	return &saleBill, nil
}

type GetSaleBillAllInfoOption func(option *GetSaleBillAllInfoOptions)

type GetSaleBillAllInfoOptions struct {
	MemberSaleOrderUuid uint64 `json:"member_sale_order_uuid"` // 会员端销售订单UUID. 不为0时，根据会员端销售订单UUID查询
}

func WithMemberSaleOrderUuid(uuid uint64) func(option *GetSaleBillAllInfoOptions) {
	return func(option *GetSaleBillAllInfoOptions) {
		option.MemberSaleOrderUuid = uuid
	}
}

// GetSaleBillAllInfo 获取销售账单所有信息
func (r *orderRepo) GetSaleBillAllInfo(saleBillUuid uint64, opts ...GetSaleBillAllInfoOption) (*model.SaleBill, error) {
	option := &GetSaleBillAllInfoOptions{}
	for _, opt := range opts {
		opt(option)
	}
	uuidFilter := CommonRepo.WhereByUuid(saleBillUuid) // 默认根据销售账单UUID查询
	if option.MemberSaleOrderUuid != 0 {
		uuidFilter = CommonRepo.WhereByMemberSaleOrderUuid(option.MemberSaleOrderUuid) // 根据会员端销售订单UUID查询
	}
	info, err := r.GetSaleBill(
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			// ==================== 销售账单的桌台信息 ====================
			// 加载桌台信息
			WithPreload{
				// 使用场景：获取反结账信息，判断桌台是否空闲
				Query: "Desk",
			},
			// ==================== 销售账单的自助餐套餐1、2信息 ====================
			// 加载自助餐套餐1多语言名称
			WithPreload{
				Query: "BuffetPackage1.MultiLanguageName",
			},
			// 加载自助餐套餐1顾客类型价格。用于判断顾客价格是否改变
			WithPreload{
				Query: "BuffetPackage1.BuffetCustomerTypePrices",
			},
			// 加载自助餐套餐1商品bom。用于判断加购的商品是否属于自助餐套餐1
			WithPreload{
				Query: "BuffetPackage1.BuffetProducts.ProductPackage",
			},
			// 加载自助餐套餐2多语言名称
			WithPreload{
				Query: "BuffetPackage2.MultiLanguageName",
			},
			// 加载自助餐套餐2顾客类型价格。用于判断顾客价格是否改变
			WithPreload{
				Query: "BuffetPackage2.BuffetCustomerTypePrices",
			},
			// 加载自助餐套餐2商品bom。用于判断加购的商品是否属于自助餐套餐2
			WithPreload{
				Query: "BuffetPackage2.BuffetProducts.ProductPackage",
			},
			// ==================== 销售账单的账单设置 ====================
			WithPreload{
				Query: "SaleBillSetting",
			},
			// ==================== 销售账单的优惠券信息 ====================
			WithPreload{
				Query: "SaleOrders.Coupons",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.Coupons.MarketingCoupon",
			},
			WithPreload{
				Query: "SaleOrders.Coupons.MemberCoupon",
			},
			// ==================== 销售账单的订单信息 ====================
			WithPreload{
				Query: "SaleOrders.PaymentOrders",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
					CommonRepo.DBOption(CommonRepo.WhereByStatus(constant.PaymentOrderStatusPaid)),
				},
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders.PaymentMethod",
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders.ReturnOrderAmounts",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders.ReturnOrderAmounts.PaymentMethod",
			},
			WithPreload{
				Query: "SaleOrders.Member.MemberLevel",
			},
			WithPreload{
				Query: "SaleOrders.Member.MemberCard.MemberCardType",
			},
			WithPreload{
				Query: "SaleOrders.Member.MemberBalanceLog",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ImageFile",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ReturnOrderProducts",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.ReturnOrderProducts",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetDelayProducts.ReturnOrderProducts",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.CancelReasons",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.OrderItemRemarks",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ProductPackage.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ProductPackage.DineTax",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ProductPackage.TakeoutTax",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ProductPackage.ProductUnit",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ProductPackage.ProductUnit.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ProductPackage.ProductCategory",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes",
				Args: []any{
					CommonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
						return db.Order("id asc")
					}),
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName",
			},
			WithPreload{
				// 用于检查商品包是否下架
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms",
				Args: []any{
					CommonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
						return db.Order("product_bom_uuid asc")
					}),
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				// 用于检查商品包是否下架
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductPackage",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.FlavorMaterials",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.FlavorMaterials.Material.WarehouseItems",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.SauceMaterials.Material.WarehouseItems",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.ProductBomCard.RelatedMaterials.Material.WarehouseItems",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductBomCard.RelatedMaterials.Material.WarehouseItems",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.BatchTag",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.BuffetPackage.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.BuffetCustomerTypePrice.BuffetCustomerType",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetDelayProducts",
			},
			WithPreload{
				Query: "SaleOrders.ReturnOrders",
			},
			WithPreload{
				Query: "SaleOrders.MemberPointLogs",
			},
			// ==================== 销售账单的收银员信息 ====================
			WithPreload{
				Query: "Cashier",
			},
			// ==================== 销售账单的订单来源和国籍信息 ====================
			WithPreload{
				Query: "OrderSource.MultiLanguageName",
				// 移除 delete_time 过滤，保证历史订单可显示已删除的配置名称
			},
			WithPreload{
				Query: "Nationality.MultiLanguageName",
				// 移除 delete_time 过滤，保证历史订单可显示已删除的配置名称
			},
		),
		CommonRepo.WhereBySoftDelete(),
		uuidFilter,
	)
	if err != nil {
		return nil, fmt.Errorf("GetSaleBillAllInfo: %v", err)
	}
	return &info, nil
}

// GetSaleBillWithProducts 获取销售账单所有商品信息
func (r *orderRepo) GetSaleBillWithProducts(saleBillUuid uint64) (*model.SaleBill, error) {
	info, err := r.GetSaleBill(
		CommonRepo.Preload(
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
					CommonRepo.DBOption(CommonRepo.WhereBySaleBillUuid(saleBillUuid)),
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ProductionOrderProduct",
			},
			WithPreload{
				Query: "OrderSource.MultiLanguageName",
				// 移除 delete_time 过滤，保证历史订单可显示已删除的配置名称
			},
			WithPreload{
				Query: "Nationality.MultiLanguageName",
				// 移除 delete_time 过滤，保证历史订单可显示已删除的配置名称
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByUuid(saleBillUuid),
	)
	if err != nil {
		return nil, fmt.Errorf("GetSaleBillWithProducts: %v", err)
	}
	return &info, nil
}

// IsPartiallyPaid 是否已经被部分支付
func (r *orderRepo) IsPartiallyPaid(param any) bool {
	var info model.SaleBill
	var ok bool
	switch v := param.(type) {
	case model.SaleBill:
		info = v
		ok = true
	case uint64:
		info, _ = r.GetSaleBillInfo(v, constant.OptionalUuid)
		ok = true
	default:
		return false
	}
	if !ok || len(info.SaleOrders) == 0 {
		return false
	}
	//
	uuids := make([]uint64, len(info.SaleOrders))
	for i, v := range info.SaleOrders {
		uuids[i] = v.Uuid
	}
	repo := NewPaymentOrderRepo(r.db)
	paymentOrders, err := repo.GetPaymentOrderList(repo.WhereRelatedUuids(uuids))
	if err != nil {
		return false
	}
	for _, v := range paymentOrders {
		if v.Status != 2 {
			return true
		}
	}
	return false
}

// GetSaleOrderBomList 查询销售订单的所有bom
func (r *orderRepo) GetSaleOrderBomList(saleOrderUuid uint64) ([]model.SaleOrderProductBom, error) {
	var saleOrderProductBoms []model.SaleOrderProductBom
	if err := r.db.Preload("ProductBom.ProductPackage").Preload("SaleOrderProduct").Model(&model.SaleOrderProductBom{}).Where("sale_order_uuid = ? AND delete_time = ?", saleOrderUuid, constant.NotDeleted).Find(&saleOrderProductBoms).Error; err != nil {
		return nil, fmt.Errorf("GetSaleOrderBomList: %v", err)
	}
	return saleOrderProductBoms, nil
}

// DeleteOrderProduct 删除订单产品
func (r *orderRepo) DeleteOrderProduct(saleBillUuid uint64, saleOrderUuid uint64, saleOrderProductUuid uint64, isPackageProduct bool) error {
	timeNow := uint(time.Now().Unix())
	// 删除关联关系
	err := r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.SaleOrderProductBom{}).Where("sale_order_product_uuid = ?", saleOrderProductUuid).Update("delete_time", timeNow).Error
		if err != nil {
			return errors.WithMessage(err)
		}
		err = tx.Model(&model.SaleOrderProductAttribute{}).Where("sale_order_product_uuid = ?", saleOrderProductUuid).Update("delete_time", timeNow).Error
		if err != nil {
			return errors.WithMessage(err)
		}
		err = tx.Model(&model.ProductionOrderProduct{}).Where("sale_order_product_uuid = ?", saleOrderProductUuid).Update("delete_time", timeNow).Error
		if err != nil {
			return errors.WithMessage(err)
		}
		if isPackageProduct {
			// 套餐子商品送厨单商品标记删除 - 分批处理避免锁风暴
			var subProductUuids []uint64
			err = tx.Model(&model.SaleOrderProduct{}).
				Where("package_uuid = ?", saleOrderProductUuid).
				Pluck("uuid", &subProductUuids).Error
			if err != nil {
				return errors.WithMessage(err)
			}
			// 如果存在套餐子商品，分批更新生产订单商品
			if len(subProductUuids) > 0 {
				const batchSize = 200 // 每批处理200条，避免一次性锁定大量数据（UPDATE操作会产生行锁和gap lock）
				for i := 0; i < len(subProductUuids); i += batchSize {
					end := i + batchSize
					if end > len(subProductUuids) {
						end = len(subProductUuids)
					}
					batch := subProductUuids[i:end]
					err = tx.Model(&model.ProductionOrderProduct{}).
						Where("sale_order_product_uuid IN (?)", batch).
						Update("delete_time", timeNow).Error
					if err != nil {
						return errors.WithMessage(err, fmt.Sprintf("分批更新生产订单商品失败，批次: %d-%d", i, end-1))
					}
				}
			}
		}
		return tx.Model(&model.SaleOrderProduct{}).
			Where("(status != ? or cancel_time != 0)", constant.OrderProductStatusSentKitchen).
			Where("delete_time = ?", 0).
			Where("sale_bill_uuid = ? AND sale_order_uuid = ? AND uuid = ?", saleBillUuid, saleOrderUuid, saleOrderProductUuid).
			Update("delete_time", uint(time.Now().Unix())).
			Error
	})
	if err != nil {
		return fmt.Errorf("DeleteOrderProduct: %v", err)
	}
	return nil
}

// ChangePopulation 修改订单人数
func (r *orderRepo) ChangePopulation(saleBillUuid uint64, population int) error {
	err := r.db.Model(&model.SaleBill{}).
		Where("delete_time = ?", 0).
		Where("uuid = ?", saleBillUuid).
		Updates(map[string]interface{}{
			"uuid":     saleBillUuid, // NOTE 不要删除，事件钩子需要
			"meal_num": population,
		}).Error
	if err != nil {
		return fmt.Errorf("ChangePopulation: %v", err)
	}
	return nil
}

// ChangeProductRemark 修改订单商品备注
func (r *orderRepo) ChangeProductRemark(saleBillUuid uint64, saleOrderUuid uint64, orderProductUuid uint64, remark string, sign string) error {
	err := r.db.Model(&model.SaleOrderProduct{}).
		Where("delete_time = ?", constant.NotDeleted).
		Where("sale_bill_uuid = ? AND sale_order_uuid = ? AND uuid = ?", saleBillUuid, saleOrderUuid, orderProductUuid).
		Updates(map[string]interface{}{
			"remark": remark,
			"sign":   sign,
		}).Error
	if err != nil {
		return fmt.Errorf("ChangeProductRemark: %v", err)
	}
	return nil
}

// HasShowOrder 判断该设备是否有未挂单的点餐订单
func (r *orderRepo) HasShowOrder(deviceUuid uint64) (uint64, error) {
	var saleBill model.SaleBill
	if err := r.db.Where("device_uuid = ? AND status = ? AND hide_bill_time = ? AND delete_time = ?", deviceUuid, constant.SaleBillStatusPending, 0, constant.NotDeleted).First(&saleBill).Error; err != nil {
		if utils.IsNotFoundRecord(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("HasShowOrder: %v", err)
	}
	return saleBill.Uuid, nil
}

// SetLock 设置订单锁定状态
func (r *orderRepo) SetLock(saleBillUuid uint64, isLock bool) error {
	isLockInt := 0
	lockTime := 0
	if isLock {
		isLockInt = 1
		lockTime = int(time.Now().Unix())
	}
	err := r.db.Model(&model.SaleBill{}).
		Where("delete_time = ? AND uuid = ?", constant.NotDeleted, saleBillUuid).
		Updates(map[string]interface{}{
			"is_lock":   isLockInt,
			"lock_time": lockTime,
		}).Error
	if err != nil {
		return fmt.Errorf("SetLock: %v", err)
	}
	return nil
}

// SaveOrUpdateInvoiceInfo 保存或更新订单发票信息
func (r *orderRepo) SaveOrUpdateInvoiceInfo(saleOrderUuid uint64, invoiceInfo model.SaleOrderInvoiceInfo) (*model.SaleOrderInvoiceInfo, error) {
	// 直接尝试查询现有发票信息
	var dbInvoiceInfo model.SaleOrderInvoiceInfo
	result := r.db.Where("sale_order_uuid = ?", saleOrderUuid).First(&dbInvoiceInfo)

	// 如果记录不存在，创建新记录
	if result.Error != nil {
		// 如果是记录不存在的错误
		if result.Error == gorm.ErrRecordNotFound {
			// 创建新记录
			invoiceInfo.SaleOrderUuid = saleOrderUuid
			invoiceInfo.PrintNum = 1
			err := r.db.Create(&invoiceInfo).Error
			if err != nil {
				return nil, fmt.Errorf("创建发票信息失败: %v", err)
			}
			return &invoiceInfo, nil
		} else {
			// 其他查询错误
			return nil, fmt.Errorf("查询发票信息失败: %v", result.Error)
		}
	}

	// 记录存在，更新现有记录
	// 打印次数加一
	invoiceInfo.PrintNum = dbInvoiceInfo.PrintNum + 1

	// 更新现有记录
	err := r.db.Model(&model.SaleOrderInvoiceInfo{}).
		Where("sale_order_uuid = ?", saleOrderUuid).
		Updates(invoiceInfo).Error
	if err != nil {
		return nil, fmt.Errorf("更新发票信息失败: %v", err)
	}

	return &invoiceInfo, nil
}

// GetInvoiceInfo 获取订单发票信息
func (r *orderRepo) GetInvoiceInfo(saleOrderUuid uint64) (*model.SaleOrderInvoiceInfo, error) {
	var invoiceInfo model.SaleOrderInvoiceInfo
	result := r.db.Where("sale_order_uuid = ?", saleOrderUuid).First(&invoiceInfo)
	return &invoiceInfo, result.Error
}

// UpdateErpDiscountAmount 更新订单应收优惠金额
func (r *orderRepo) UpdateErpDiscountAmount(saleOrderUuid uint64, erpDiscountAmount float64) error {
	return r.db.Model(&model.SaleOrder{}).Where("uuid = ?", saleOrderUuid).Updates(map[string]interface{}{
		"erp_discount_amount": erpDiscountAmount,
	}).Error
}

// GetMonthlyOrderRanks 获取订单的月排名信息（基于全表数据）
// 结合SQL查询：按月分组，取出每月第一条订单ID（基于全表），和当前订单是这个月的第几条数据（基于全表）
func (r *orderRepo) GetMonthlyOrderRanks(saleBillUuids []uint64) ([]MonthlyOrderRank, error) {
	var results []MonthlyOrderRank

	// 如果没有传入saleBillUuids，返回空结果
	if len(saleBillUuids) == 0 {
		return results, nil
	}

	// 构建IN条件
	saleBillUuidsStr := make([]string, len(saleBillUuids))
	for i, uuid := range saleBillUuids {
		saleBillUuidsStr[i] = fmt.Sprintf("%d", uuid)
	}
	inClause := strings.Join(saleBillUuidsStr, ",")

	// 构建SQL查询
	// 基于全表数据计算每个订单在当月的排名和每月第一条订单ID
	query := fmt.Sprintf(`
		SELECT 
			t.month_year,
			t.first_order_uuid,
			so.uuid as order_uuid,
			so.order_no,
			t.monthly_order_number
		FROM ttpos_sale_order so
		INNER JOIN (
			SELECT 
				DATE_FORMAT(FROM_UNIXTIME(ttpos_sale_order.create_time), '%%Y%%m') as month_year,
				FIRST_VALUE(ttpos_sale_order.uuid) OVER (PARTITION BY DATE_FORMAT(FROM_UNIXTIME(ttpos_sale_order.create_time), '%%Y%%m') ORDER BY ttpos_sale_order.create_time) as first_order_uuid,
				ttpos_sale_order.uuid,
				ROW_NUMBER() OVER (PARTITION BY DATE_FORMAT(FROM_UNIXTIME(ttpos_sale_order.create_time), '%%Y%%m') ORDER BY ttpos_sale_order.create_time) as monthly_order_number
			FROM ttpos_sale_order LEFT JOIN ttpos_sale_bill sb on ttpos_sale_order.sale_bill_uuid = sb.uuid 
			WHERE ttpos_sale_order.delete_time = 0 AND sb.production_time > 0
		) t ON so.uuid = t.uuid
		WHERE so.delete_time = 0 
			AND so.sale_bill_uuid IN (%s)
		ORDER BY t.month_year, t.monthly_order_number
	`, inClause)

	// 执行原生SQL查询
	err := r.db.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("GetMonthlyOrderRanks: %v", err)
	}

	return results, nil
}

// UpdateSaleBillOrderRemark 更新销售账单整单备注
func (r *orderRepo) UpdateSaleBillOrderRemark(saleBillUuid uint64, orderRemark string) error {
	return r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBillUuid).Updates(model.SaleBill{
		OrderRemark: orderRemark,
	}).Error
}

// 获取订单当前的分批送厨模式
func (r *orderRepo) GetSaleBillBatchCookingMode(saleBillUuid uint64) (string, error) {
	var result struct {
		BatchCookingMode string `json:"batch_cooking_mode"`
	}
	if err := r.db.Model(&model.SaleBillSetting{}).Where("sale_bill_uuid = ?", saleBillUuid).Select("batch_cooking_mode").Scan(&result).Error; err != nil {
		return "", fmt.Errorf("GetSaleBillBatchCookingMode: %v", err)
	}
	return result.BatchCookingMode, nil
}

// GetSaleOrderUuids 获取销售订单UUID列表
func (r *orderRepo) GetSaleOrderUuids(opts ...DBOption) []uint64 {
	var saleOrderUuids []uint64
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	db.Model(&model.SaleOrder{}).Select("uuid").Pluck("uuid", &saleOrderUuids)
	return saleOrderUuids
}

// GetSaleBillList 获取销售账单列表
func (r *orderRepo) GetSaleBillList(opts ...DBOption) []model.SaleBill {
	var saleBills []model.SaleBill
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	db.Find(&saleBills)
	return saleBills
}

// extractUUIDFromKey 从缓存 key 中提取 UUID
// Key 格式：{system_prefix}:{company_uuid}:{object_type}:{object_uuid}
func extractUUIDFromKey(key string) (uint64, error) {
	parts := strings.Split(key, ":")
	if len(parts) >= 4 {
		return strconv.ParseUint(parts[3], 10, 64) // UUID 在索引 3 的位置
	}
	return 0, fmt.Errorf("invalid key format: %s", key)
}

// extractUUIDsFromKeys 从缓存 keys 中提取 UUIDs
func extractUUIDsFromKeys(keys []string) []uint64 {
	uuids := make([]uint64, 0, len(keys))
	for _, key := range keys {
		if uuid, err := extractUUIDFromKey(key); err == nil {
			uuids = append(uuids, uuid)
		}
	}
	return uuids
}

// convertBatchResultToUUIDMap 将批量查询结果从 map[string]any 转换为 map[uint64]interface{}
// convertBatchResultToUUIDMap 将 map[string]T 转换为 map[uint64]interface{}
func convertBatchResultToUUIDMap[T any](batchResult map[string]T) map[uint64]interface{} {
	result := make(map[uint64]interface{})
	for key, val := range batchResult {
		if uuid, err := extractUUIDFromKey(key); err == nil {
			result[uuid] = val
		}
	}
	return result
}

// cacheGroupSingleton 缓存组单例存储
// 使用 sync.Map 存储每种类型的 cacheGroup 创建函数，确保 L1 缓存可以跨请求共享
// key: 类型名称字符串, value: func() cache.ICacheGroup[T]
var (
	cacheGroupSingletons sync.Map // map[string]func() any
	cacheGroupMutex      sync.Mutex
)

// getCacheLayerForObjectType 根据对象类型获取对应的缓存层
// 使用单例模式确保每种对象类型的 cacheGroup 实例是唯一的，从而让 L1 缓存可以跨请求共享
func getCacheLayerForObjectType[T any](groupConfig cache.GroupConfig, underlyingCache cache.Cache, defaultTTL time.Duration) repository.CacheLayer[T] {
	// 使用类型名称作为 key，确保每种类型只有一个 cacheGroup 实例
	typeName := reflect.TypeOf((*T)(nil)).Elem().String()

	// 尝试从单例池中获取 cacheGroup 创建函数
	var group cache.ICacheGroup[T]
	if cached, ok := cacheGroupSingletons.Load(typeName); ok {
		// 使用类型断言获取创建函数
		if createFunc, ok := cached.(func() cache.ICacheGroup[T]); ok {
			group = createFunc()
		}
	}

	// 如果不存在，创建新的 cacheGroup 实例和创建函数
	if group == nil {
		cacheGroupMutex.Lock()
		// Double check
		if cached, ok := cacheGroupSingletons.Load(typeName); ok {
			if createFunc, ok := cached.(func() cache.ICacheGroup[T]); ok {
				group = createFunc()
			}
		}
		if group == nil {
			// 创建 cacheGroup 实例
			group = cache.NewCacheGroup[T](groupConfig)
			// 存储创建函数（返回同一个实例）
			createFunc := func() cache.ICacheGroup[T] {
				return group
			}
			cacheGroupSingletons.Store(typeName, createFunc)
		}
		cacheGroupMutex.Unlock()
	}

	// 使用已有的 cacheGroup 实例创建 adapter
	return adapter.NewCacheGroupAdapterWithGroup[T](group, underlyingCache, defaultTTL)
}

// getSaleBillAssociationsForOrderCart 获取 SaleBill 的关联配置（用于订单购物车场景）
// 这个函数在 repository 层定义，避免循环依赖
func getSaleBillAssociationsForOrderCart(ctx goCtx.Context, db *gorm.DB, groupConfig cache.GroupConfig, underlyingCache cache.Cache) []entity.Association {
	deskRepo := NewDeskRepo(db)
	productPackageRepo := NewProductPackageRepo(db)
	multiLanguageNameRepo := NewMultiLanguageNameRepo(db)
	productCategoryRepo := NewProductCategoryRepo(db)
	productBomRepo := NewProductBomRepo(db)
	productAttributeRepo := NewProductAttributeRepo(db)
	productFlavorRepo := NewProductFlavorRepo(db)
	productSauceRepo := NewProductSauceRepo(db)
	batchTagRepo := NewBatchTagRepo(db)

	return []entity.Association{
		// 一对一关系：SaleBillSetting
		{
			Path:       "SaleBillSetting",
			ObjectType: "sale_bill_setting",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(model.SaleBill).Uuid
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存
				cacheLayer := getCacheLayerForObjectType[*model.SaleBillSetting](groupConfig, underlyingCache, 10*time.Minute)
				key := persistence.BuildKey(ctx, "sale_bill_setting", uuid)
				result, err := cacheLayer.GET(key, func() (*model.SaleBillSetting, error) {
					var setting model.SaleBillSetting
					err := db.Where("sale_bill_uuid = ?", uuid).First(&setting).Error
					if err != nil {
						return nil, err
					}
					return &setting, nil
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
		},
		// 一对一关系：Desk
		{
			Path:       "Desk",
			ObjectType: "desk",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(model.SaleBill).DeskUuid
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存（使用具体的类型 *model.Desk）
				cacheLayer := getCacheLayerForObjectType[*model.Desk](groupConfig, underlyingCache, 10*time.Minute)
				key := persistence.BuildKey(ctx, "desk", uuid)
				result, err := cacheLayer.GET(key, func() (*model.Desk, error) {
					desk, err := deskRepo.GetDesk(
						CommonRepo.WhereByUuid(uuid),
						CommonRepo.WhereBySoftDelete(),
					)
					if err != nil {
						return nil, err
					}
					return &desk, nil
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.ProductPackage（支持批量优化）
		{
			Path:       "SaleOrders.SaleOrderProducts.ProductPackage",
			ObjectType: "product_package",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(*model.SaleOrderProduct).ProductPackageUuid
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存
				cacheLayer := getCacheLayerForObjectType[*model.ProductPackage](groupConfig, underlyingCache, 5*time.Minute)
				key := persistence.BuildKey(ctx, "product_package", uuid)
				result, err := cacheLayer.GET(key, func() (*model.ProductPackage, error) {
					pkg, err := productPackageRepo.GetProductPackage(
						CommonRepo.WhereByUuid(uuid),
						CommonRepo.WhereBySoftDelete(),
					)
					if err != nil {
						return nil, err
					}
					return pkg, nil
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
			BatchQueryFunc: func(ctx goCtx.Context, uuids []uint64) (map[uint64]interface{}, error) {
				// 构建批量查询的 keys
				keys := make([]string, 0, len(uuids))
				for _, uuid := range uuids {
					keys = append(keys, persistence.BuildKey(ctx, "product_package", uuid))
				}
				cacheLayer := getCacheLayerForObjectType[*model.ProductPackage](groupConfig, underlyingCache, 5*time.Minute)
				// 使用对象存储层的批量缓存
				batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.ProductPackage, error) {
					packages, err := productPackageRepo.GetProductPackageListByUuids(uuids)
					if err != nil {
						return nil, err
					}
					result := make(map[string]*model.ProductPackage)
					for _, pkg := range packages {
						if pkg != nil {
							key := persistence.BuildKey(ctx, "product_package", pkg.Uuid)
							result[key] = pkg
						}
					}
					return result, nil
				})
				if err != nil {
					return nil, err
				}
				return convertBatchResultToUUIDMap(batchResult), nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.MultiLanguageName（支持批量优化）
		{
			Path:       "SaleOrders.SaleOrderProducts.MultiLanguageName",
			ObjectType: "multi_language_name",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(*model.SaleOrderProduct).MultiLanguageNameUuid
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存
				cacheLayer := getCacheLayerForObjectType[*model.MultiLanguageName](groupConfig, underlyingCache, 5*time.Minute)
				key := persistence.BuildKey(ctx, "multi_language_name", uuid)
				result, err := cacheLayer.GET(key, func() (*model.MultiLanguageName, error) {
					return multiLanguageNameRepo.GetMultiLanguageName(
						CommonRepo.WhereByUuid(uuid),
					)
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
			BatchQueryFunc: func(ctx goCtx.Context, uuids []uint64) (map[uint64]interface{}, error) {
				// 构建批量查询的 keys
				keys := make([]string, 0, len(uuids))
				for _, uuid := range uuids {
					keys = append(keys, persistence.BuildKey(ctx, "multi_language_name", uuid))
				}
				cacheLayer := getCacheLayerForObjectType[*model.MultiLanguageName](groupConfig, underlyingCache, 5*time.Minute)
				// 使用对象存储层的批量缓存
				batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.MultiLanguageName, error) {
					result := make(map[string]*model.MultiLanguageName)
					for _, uuid := range uuids {
						name, err := multiLanguageNameRepo.GetMultiLanguageName(
							CommonRepo.WhereByUuid(uuid),
						)
						if err == nil {
							key := persistence.BuildKey(ctx, "multi_language_name", uuid)
							result[key] = name
						}
					}
					return result, nil
				})
				if err != nil {
					return nil, err
				}
				return convertBatchResultToUUIDMap(batchResult), nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.ProductPackage.ProductCategory
		{
			Path:       "SaleOrders.SaleOrderProducts.ProductPackage.ProductCategory",
			ObjectType: "product_category",
			GetUUID: func(obj interface{}) uint64 {
				// obj 是 *ProductPackage 类型
				if pkg, ok := obj.(*model.ProductPackage); ok && pkg != nil {
					return pkg.CategoryUuid
				}
				return 0
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存
				cacheLayer := getCacheLayerForObjectType[*model.ProductCategory](groupConfig, underlyingCache, 5*time.Minute)
				key := persistence.BuildKey(ctx, "product_category", uuid)
				result, err := cacheLayer.GET(key, func() (*model.ProductCategory, error) {
					return productCategoryRepo.GetProductCategory(
						CommonRepo.WhereByUuid(uuid),
					)
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.ProductPackage.ProductBoms
		{
			Path:       "SaleOrders.SaleOrderProducts.ProductPackage.ProductBoms",
			ObjectType: "product_bom",
			GetUUID: func(obj interface{}) uint64 {
				// obj 是 *ProductPackage 类型（因为 SaleOrderProduct.ProductPackage 是指针类型）
				if pkg, ok := obj.(*model.ProductPackage); ok && pkg != nil {
					return pkg.Uuid
				}
				return 0
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存（注意：这里缓存的是 ProductPackage，然后提取 ProductBoms）
				cacheLayer := getCacheLayerForObjectType[*model.ProductPackage](groupConfig, underlyingCache, 5*time.Minute)
				key := persistence.BuildKey(ctx, "product_package", uuid)
				result, err := cacheLayer.GET(key, func() (*model.ProductPackage, error) {
					productPackage, err := productPackageRepo.GetProductPackage(
						CommonRepo.WhereByUuid(uuid),
						CommonRepo.Preload(WithPreload{Query: "ProductBoms"}),
					)
					if err != nil {
						return nil, err
					}
					return productPackage, nil
				})
				if err != nil {
					return nil, err
				}
				return result.ProductBoms, nil
			},
			BatchQueryFunc: func(ctx goCtx.Context, uuids []uint64) (map[uint64]interface{}, error) {
				// 构建批量查询的 keys
				keys := make([]string, 0, len(uuids))
				for _, uuid := range uuids {
					keys = append(keys, persistence.BuildKey(ctx, "product_package", uuid))
				}
				cacheLayer := getCacheLayerForObjectType[*model.ProductPackage](groupConfig, underlyingCache, 5*time.Minute)
				// 使用对象存储层的批量缓存
				batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.ProductPackage, error) {
					packages, err := productPackageRepo.GetProductPackageList(
						CommonRepo.WhereInUuids(uuids),
						CommonRepo.Preload(WithPreload{Query: "ProductBoms"}),
					)
					if err != nil {
						return nil, err
					}
					result := make(map[string]*model.ProductPackage)
					for _, pkg := range packages {
						key := persistence.BuildKey(ctx, "product_package", pkg.Uuid)
						result[key] = pkg
					}
					return result, nil
				})
				if err != nil {
					return nil, err
				}
				// 转换为 map[uint64]interface{}，提取 ProductBoms
				result := make(map[uint64]interface{})
				for key, val := range batchResult {
					if uuid, err := extractUUIDFromKey(key); err == nil {
						result[uuid] = val.ProductBoms
					}
				}
				return result, nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.ProductPackage.ProductPackageAttributeGroups
		{
			Path:       "SaleOrders.SaleOrderProducts.ProductPackage.ProductPackageAttributeGroups",
			ObjectType: "product_package_attribute_group",
			GetUUID: func(obj interface{}) uint64 {
				// obj 是 *ProductPackage 类型
				if pkg, ok := obj.(*model.ProductPackage); ok && pkg != nil {
					return pkg.Uuid
				}
				return 0
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存（注意：这里缓存的是 ProductPackage，然后提取 ProductPackageAttributeGroups）
				cacheLayer := getCacheLayerForObjectType[*model.ProductPackage](groupConfig, underlyingCache, 5*time.Minute)
				key := persistence.BuildKey(ctx, "product_package", uuid)
				result, err := cacheLayer.GET(key, func() (*model.ProductPackage, error) {
					productPackage, err := productPackageRepo.GetProductPackage(
						CommonRepo.WhereByUuid(uuid),
						CommonRepo.Preload(WithPreload{Query: "ProductPackageAttributeGroups"}),
					)
					if err != nil {
						return nil, err
					}
					return productPackage, nil
				})
				if err != nil {
					return nil, err
				}
				return result.ProductPackageAttributeGroups, nil
			},
			BatchQueryFunc: func(ctx goCtx.Context, uuids []uint64) (map[uint64]interface{}, error) {
				// 构建批量查询的 keys
				keys := make([]string, 0, len(uuids))
				for _, uuid := range uuids {
					keys = append(keys, persistence.BuildKey(ctx, "product_package", uuid))
				}
				cacheLayer := getCacheLayerForObjectType[*model.ProductPackage](groupConfig, underlyingCache, 5*time.Minute)
				// 使用对象存储层的批量缓存
				batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.ProductPackage, error) {
					packages, err := productPackageRepo.GetProductPackageList(
						CommonRepo.WhereInUuids(uuids),
						CommonRepo.Preload(WithPreload{Query: "ProductPackageAttributeGroups"}),
					)
					if err != nil {
						return nil, err
					}
					result := make(map[string]*model.ProductPackage)
					for _, pkg := range packages {
						key := persistence.BuildKey(ctx, "product_package", pkg.Uuid)
						result[key] = pkg
					}
					return result, nil
				})
				if err != nil {
					return nil, err
				}
				// 转换为 map[uint64]interface{}，提取 ProductPackageAttributeGroups
				result := make(map[uint64]interface{})
				for key, val := range batchResult {
					if uuid, err := extractUUIDFromKey(key); err == nil {
						result[uuid] = val.ProductPackageAttributeGroups
					}
				}
				return result, nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute
		{
			Path:       "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute",
			ObjectType: "product_attribute",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(*model.SaleOrderProductAttribute).ProductAttributeUuid
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存
				key := persistence.BuildKey(ctx, "product_attribute", uuid)
				cacheLayer := getCacheLayerForObjectType[*model.ProductAttribute](groupConfig, underlyingCache, 5*time.Minute)
				result, err := cacheLayer.GET(key, func() (*model.ProductAttribute, error) {
					attr, err := productAttributeRepo.GetProductAttribute(
						CommonRepo.WhereByUuid(uuid),
						CommonRepo.WhereBySoftDelete(),
					)
					if err != nil {
						return nil, err
					}
					return attr, nil
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
			BatchQueryFunc: func(ctx goCtx.Context, uuids []uint64) (map[uint64]interface{}, error) {
				// 构建批量查询的 keys
				keys := make([]string, 0, len(uuids))
				for _, uuid := range uuids {
					keys = append(keys, persistence.BuildKey(ctx, "product_attribute", uuid))
				}
				cacheLayer := getCacheLayerForObjectType[*model.ProductAttribute](groupConfig, underlyingCache, 5*time.Minute)
				// 使用对象存储层的批量缓存
				batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.ProductAttribute, error) {
					missedUuids := extractUUIDsFromKeys(keys)
					result := make(map[string]*model.ProductAttribute)
					for _, uuid := range missedUuids {
						attr, err := productAttributeRepo.GetProductAttribute(
							CommonRepo.WhereByUuid(uuid),
							CommonRepo.WhereBySoftDelete(),
						)
						if err == nil {
							key := persistence.BuildKey(ctx, "product_attribute", uuid)
							result[key] = attr
						}
					}
					return result, nil
				})
				if err != nil {
					return nil, err
				}
				result := make(map[uint64]interface{})
				for key, val := range batchResult {
					if uuid, err := extractUUIDFromKey(key); err == nil {
						result[uuid] = val
					}
				}
				return result, nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName
		{
			Path:       "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName",
			ObjectType: "multi_language_name",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(model.ProductAttribute).MultiLanguageNameUuid
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存
				cacheLayer := getCacheLayerForObjectType[*model.MultiLanguageName](groupConfig, underlyingCache, 5*time.Minute)
				key := persistence.BuildKey(ctx, "multi_language_name", uuid)
				result, err := cacheLayer.GET(key, func() (*model.MultiLanguageName, error) {
					return multiLanguageNameRepo.GetMultiLanguageName(
						CommonRepo.WhereByUuid(uuid),
					)
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
			BatchQueryFunc: func(ctx goCtx.Context, uuids []uint64) (map[uint64]interface{}, error) {
				// 构建批量查询的 keys
				keys := make([]string, 0, len(uuids))
				for _, uuid := range uuids {
					keys = append(keys, persistence.BuildKey(ctx, "multi_language_name", uuid))
				}
				cacheLayer := getCacheLayerForObjectType[*model.MultiLanguageName](groupConfig, underlyingCache, 5*time.Minute)
				// 使用对象存储层的批量缓存
				batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.MultiLanguageName, error) {
					result := make(map[string]*model.MultiLanguageName)
					for _, uuid := range uuids {
						name, err := multiLanguageNameRepo.GetMultiLanguageName(
							CommonRepo.WhereByUuid(uuid),
						)
						if err == nil {
							key := persistence.BuildKey(ctx, "multi_language_name", uuid)
							result[key] = name
						}
					}
					return result, nil
				})
				if err != nil {
					return nil, err
				}
				return convertBatchResultToUUIDMap(batchResult), nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom
		{
			Path:       "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom",
			ObjectType: "product_bom",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(*model.SaleOrderProductBom).ProductBomUuid
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存
				cacheLayer := getCacheLayerForObjectType[*model.ProductBom](groupConfig, underlyingCache, 5*time.Minute)
				key := persistence.BuildKey(ctx, "product_bom", uuid)
				result, err := cacheLayer.GET(key, func() (*model.ProductBom, error) {
					bom, err := productBomRepo.GetProductBom(
						CommonRepo.WhereByUuid(uuid),
						CommonRepo.WhereBySoftDelete(),
					)
					if err != nil {
						return nil, err
					}
					return bom, nil
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
			BatchQueryFunc: func(ctx goCtx.Context, uuids []uint64) (map[uint64]interface{}, error) {
				// 构建批量查询的 keys
				keys := make([]string, 0, len(uuids))
				for _, uuid := range uuids {
					keys = append(keys, persistence.BuildKey(ctx, "product_bom", uuid))
				}
				cacheLayer := getCacheLayerForObjectType[*model.ProductBom](groupConfig, underlyingCache, 5*time.Minute)
				// 使用对象存储层的批量缓存
				batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.ProductBom, error) {
					result := make(map[string]*model.ProductBom)
					for _, uuid := range uuids {
						bom, err := productBomRepo.GetProductBom(
							CommonRepo.WhereByUuid(uuid),
							CommonRepo.WhereBySoftDelete(),
						)
						if err == nil {
							key := persistence.BuildKey(ctx, "product_bom", uuid)
							result[key] = bom
						}
					}
					return result, nil
				})
				if err != nil {
					return nil, err
				}
				return convertBatchResultToUUIDMap(batchResult), nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductFlavor
		{
			Path:       "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductFlavor",
			ObjectType: "product_flavor",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(model.ProductBom).ProductFlavorUuid
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存
				cacheLayer := getCacheLayerForObjectType[*model.ProductFlavor](groupConfig, underlyingCache, 5*time.Minute)
				key := persistence.BuildKey(ctx, "product_flavor", uuid)
				result, err := cacheLayer.GET(key, func() (*model.ProductFlavor, error) {
					flavor, err := productFlavorRepo.GetProductFlavor(
						CommonRepo.WhereByUuid(uuid),
						CommonRepo.WhereBySoftDelete(),
					)
					if err != nil {
						return nil, err
					}
					return &flavor, nil
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
			BatchQueryFunc: func(ctx goCtx.Context, uuids []uint64) (map[uint64]interface{}, error) {
				// 构建批量查询的 keys
				keys := make([]string, 0, len(uuids))
				for _, uuid := range uuids {
					keys = append(keys, persistence.BuildKey(ctx, "product_flavor", uuid))
				}
				cacheLayer := getCacheLayerForObjectType[*model.ProductFlavor](groupConfig, underlyingCache, 5*time.Minute)
				// 使用对象存储层的批量缓存
				batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.ProductFlavor, error) {
					result := make(map[string]*model.ProductFlavor)
					for _, uuid := range uuids {
						flavor, err := productFlavorRepo.GetProductFlavor(
							CommonRepo.WhereByUuid(uuid),
							CommonRepo.WhereBySoftDelete(),
						)
						if err == nil {
							key := persistence.BuildKey(ctx, "product_flavor", uuid)
							result[key] = &flavor
						}
					}
					return result, nil
				})
				if err != nil {
					return nil, err
				}
				return convertBatchResultToUUIDMap(batchResult), nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName
		{
			Path:       "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductFlavor.MultiLanguageName",
			ObjectType: "multi_language_name",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(model.ProductFlavor).MultiLanguageNameUuid
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存
				cacheLayer := getCacheLayerForObjectType[*model.MultiLanguageName](groupConfig, underlyingCache, 5*time.Minute)
				key := persistence.BuildKey(ctx, "multi_language_name", uuid)
				result, err := cacheLayer.GET(key, func() (*model.MultiLanguageName, error) {
					return multiLanguageNameRepo.GetMultiLanguageName(
						CommonRepo.WhereByUuid(uuid),
					)
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
			BatchQueryFunc: func(ctx goCtx.Context, uuids []uint64) (map[uint64]interface{}, error) {
				// 构建批量查询的 keys
				keys := make([]string, 0, len(uuids))
				for _, uuid := range uuids {
					keys = append(keys, persistence.BuildKey(ctx, "multi_language_name", uuid))
				}
				cacheLayer := getCacheLayerForObjectType[*model.MultiLanguageName](groupConfig, underlyingCache, 5*time.Minute)
				// 使用对象存储层的批量缓存
				batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.MultiLanguageName, error) {
					result := make(map[string]*model.MultiLanguageName)
					for _, uuid := range uuids {
						name, err := multiLanguageNameRepo.GetMultiLanguageName(
							CommonRepo.WhereByUuid(uuid),
						)
						if err == nil {
							key := persistence.BuildKey(ctx, "multi_language_name", uuid)
							result[key] = name
						}
					}
					return result, nil
				})
				if err != nil {
					return nil, err
				}
				return convertBatchResultToUUIDMap(batchResult), nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce
		{
			Path:       "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce",
			ObjectType: "product_sauce",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(model.ProductBom).ProductSauceUuid
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存
				cacheLayer := getCacheLayerForObjectType[*model.ProductSauce](groupConfig, underlyingCache, 5*time.Minute)
				key := persistence.BuildKey(ctx, "product_sauce", uuid)
				result, err := cacheLayer.GET(key, func() (*model.ProductSauce, error) {
					sauce, err := productSauceRepo.GetProductSauce(
						CommonRepo.WhereByUuid(uuid),
						CommonRepo.WhereBySoftDelete(),
					)
					if err != nil {
						return nil, err
					}
					return sauce, nil
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
			BatchQueryFunc: func(ctx goCtx.Context, uuids []uint64) (map[uint64]interface{}, error) {
				// 构建批量查询的 keys
				keys := make([]string, 0, len(uuids))
				for _, uuid := range uuids {
					keys = append(keys, persistence.BuildKey(ctx, "product_sauce", uuid))
				}
				cacheLayer := getCacheLayerForObjectType[*model.ProductSauce](groupConfig, underlyingCache, 5*time.Minute)
				// 使用对象存储层的批量缓存
				batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.ProductSauce, error) {
					result := make(map[string]*model.ProductSauce)
					for _, uuid := range uuids {
						sauce, err := productSauceRepo.GetProductSauce(
							CommonRepo.WhereByUuid(uuid),
							CommonRepo.WhereBySoftDelete(),
						)
						if err == nil {
							key := persistence.BuildKey(ctx, "product_sauce", uuid)
							result[key] = sauce
						}
					}
					return result, nil
				})
				if err != nil {
					return nil, err
				}
				return convertBatchResultToUUIDMap(batchResult), nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName
		{
			Path:       "SaleOrders.SaleOrderProducts.SaleOrderProductBoms.ProductBom.ProductSauce.MultiLanguageName",
			ObjectType: "multi_language_name",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(model.ProductSauce).MultiLanguageNameUuid
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存
				cacheLayer := getCacheLayerForObjectType[*model.MultiLanguageName](groupConfig, underlyingCache, 5*time.Minute)
				key := persistence.BuildKey(ctx, "multi_language_name", uuid)
				result, err := cacheLayer.GET(key, func() (*model.MultiLanguageName, error) {
					return multiLanguageNameRepo.GetMultiLanguageName(
						CommonRepo.WhereByUuid(uuid),
					)
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
			BatchQueryFunc: func(ctx goCtx.Context, uuids []uint64) (map[uint64]interface{}, error) {
				// 构建批量查询的 keys
				keys := make([]string, 0, len(uuids))
				for _, uuid := range uuids {
					keys = append(keys, persistence.BuildKey(ctx, "multi_language_name", uuid))
				}
				cacheLayer := getCacheLayerForObjectType[*model.MultiLanguageName](groupConfig, underlyingCache, 5*time.Minute)
				// 使用对象存储层的批量缓存
				batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.MultiLanguageName, error) {
					result := make(map[string]*model.MultiLanguageName)
					for _, uuid := range uuids {
						name, err := multiLanguageNameRepo.GetMultiLanguageName(
							CommonRepo.WhereByUuid(uuid),
						)
						if err == nil {
							key := persistence.BuildKey(ctx, "multi_language_name", uuid)
							result[key] = name
						}
					}
					return result, nil
				})
				if err != nil {
					return nil, err
				}
				return convertBatchResultToUUIDMap(batchResult), nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.BatchTag
		{
			Path:       "SaleOrders.SaleOrderProducts.BatchTag",
			ObjectType: "batch_tag",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(*model.SaleOrderProduct).BatchTagUuid
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存
				cacheLayer := getCacheLayerForObjectType[*model.BatchTag](groupConfig, underlyingCache, 5*time.Minute)
				key := persistence.BuildKey(ctx, "batch_tag", uuid)
				result, err := cacheLayer.GET(key, func() (*model.BatchTag, error) {
					tag, err := batchTagRepo.GetBatchTag(
						CommonRepo.WhereByUuid(uuid),
						CommonRepo.WhereBySoftDelete(),
					)
					if err != nil {
						return nil, err
					}
					return tag, nil
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
			BatchQueryFunc: func(ctx goCtx.Context, uuids []uint64) (map[uint64]interface{}, error) {
				// 构建批量查询的 keys
				keys := make([]string, 0, len(uuids))
				for _, uuid := range uuids {
					keys = append(keys, persistence.BuildKey(ctx, "batch_tag", uuid))
				}
				cacheLayer := getCacheLayerForObjectType[*model.BatchTag](groupConfig, underlyingCache, 5*time.Minute)
				// 使用对象存储层的批量缓存
				batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.BatchTag, error) {
					result := make(map[string]*model.BatchTag)
					for _, uuid := range uuids {
						tag, err := batchTagRepo.GetBatchTag(
							CommonRepo.WhereByUuid(uuid),
							CommonRepo.WhereBySoftDelete(),
						)
						if err == nil {
							key := persistence.BuildKey(ctx, "batch_tag", uuid)
							result[key] = tag
						}
					}
					return result, nil
				})
				if err != nil {
					return nil, err
				}
				return convertBatchResultToUUIDMap(batchResult), nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.BatchTag.MultiLanguageName
		{
			Path:       "SaleOrders.SaleOrderProducts.BatchTag.MultiLanguageName",
			ObjectType: "multi_language_name",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(model.BatchTag).MultiLanguageNameUuid
			},
			QueryFunc: func(ctx goCtx.Context, uuid uint64) (interface{}, error) {
				// 使用对象存储层的缓存
				cacheLayer := getCacheLayerForObjectType[*model.MultiLanguageName](groupConfig, underlyingCache, 5*time.Minute)
				key := persistence.BuildKey(ctx, "multi_language_name", uuid)
				result, err := cacheLayer.GET(key, func() (*model.MultiLanguageName, error) {
					return multiLanguageNameRepo.GetMultiLanguageName(
						CommonRepo.WhereByUuid(uuid),
					)
				})
				if err != nil {
					return nil, err
				}
				return result, nil
			},
			BatchQueryFunc: func(ctx goCtx.Context, uuids []uint64) (map[uint64]interface{}, error) {
				// 构建批量查询的 keys
				keys := make([]string, 0, len(uuids))
				for _, uuid := range uuids {
					keys = append(keys, persistence.BuildKey(ctx, "multi_language_name", uuid))
				}
				cacheLayer := getCacheLayerForObjectType[*model.MultiLanguageName](groupConfig, underlyingCache, 5*time.Minute)
				// 使用对象存储层的批量缓存
				batchResult, err := cacheLayer.BATCH_GET(keys, func([]string) (map[string]*model.MultiLanguageName, error) {
					result := make(map[string]*model.MultiLanguageName)
					for _, uuid := range uuids {
						name, err := multiLanguageNameRepo.GetMultiLanguageName(
							CommonRepo.WhereByUuid(uuid),
						)
						if err == nil {
							key := persistence.BuildKey(ctx, "multi_language_name", uuid)
							result[key] = name
						}
					}
					return result, nil
				})
				if err != nil {
					return nil, err
				}
				return convertBatchResultToUUIDMap(batchResult), nil
			},
		},
	}
}
