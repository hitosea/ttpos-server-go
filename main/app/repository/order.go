package repository

import (
	"errors"
	"fmt"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/ro"

	"gorm.io/gorm"
)

// IOrderRepo 定义订单仓库接口
type IOrderRepo interface {
	CreateSaleBill(model model.SaleBill) (model.SaleBill, error)                                                              // 创建销售单
	CreateSaleBillSetting(model model.SaleBillSetting) (model.SaleBillSetting, error)                                         // 创建销售账单设置
	GetSaleBill(opts ...DBOption) (model.SaleBill, error)                                                                     // 获取销售单
	CreateSaleOrder(model model.SaleOrder) (model.SaleOrder, error)                                                           // 创建订单
	GetOrderListWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.SaleBill, int64, error)                   // 获取订单列表
	GetOrderNum(opts ...DBOption) (int64, error)                                                                              // 获取订单数量
	GetCashierOrderListWithPagination(param GetCashierOrderListWithPaginationType) ([]model.SaleBill, int64, error)           // 获取收银的订单列表
	GetSaleBillInfo(saleBillUuid uint64, saleOrderUuid uint64) (model.SaleBill, error)                                        // 获取销售账单详细信息
	GetSaleBillInfoByDesk(deskUuid, saleOrderUuid uint64) (model.SaleBill, error)                                             // 获取桌台的销售账单详细信息
	GetSaleBillProductInfoByDesk(deskUuid uint64) (model.SaleBill, error)                                                     // 获取桌台的销售账单详细信息
	GetOrderCartInfo(saleBillUuid uint64) (*ro.ShopCartRepo, error)                                                           // 获取点餐购物车信息
	GetSaleBillInfoAndProduct(saleBillUuid uint64, saleOrderUuid uint64, saleOrderProductUuid uint64) (model.SaleBill, error) // 获取销售账单详细信息-包含商品信息
	GetSaleOrderProductListBySaleOrderProductUuids(saleOrderProductUuids []uint64) ([]model.SaleOrderProduct, error)          // 根据销售订单商品uuid列表获取销售订单商品列表
	GetSaleBillDetails(saleBillUuid uint64, saleOrderUuid uint64) (model.SaleBill, error)                                     // 获取销售账单详细信息-丰富的-几乎包含所有的关联
	CreateSaleOrderBuffetCustomerType(model model.SaleOrderBuffetCustomerType) (model.SaleOrderBuffetCustomerType, error)     // 创建销售订单自助餐顾客类型
	DeleteSaleOrderBuffetCustomerType(saleOrderUuid uint64) error                                                             // 删除销售订单自助餐顾客类型
	CancelOrder(saleBillUuid uint64, reason string) error                                                                     // 取消订单
	CancelDeskOrder(deskUuid uint64, reason string) error                                                                     // 取消桌台订单
	DeleteOrder(saleBillUuid uint64, saleOrderUuid uint64) error                                                              // 删除订单
	IsPartiallyPaid(param any) bool                                                                                           // 判断是否存在部分支付
	HideOrder(saleBillUuid uint64) error                                                                                      // 隐藏订单
	DeleteOrderProduct(saleBillUuid uint64, saleOrderUuid uint64, saleOrderProductUuid uint64) error                          // 删除订单产品
	GetSaleOrderBomList(saleOrderUuid uint64) ([]model.SaleOrderProductBom, error)                                            // 查询销售订单的所有bom
	ChangeProductPrice(saleBillUuid uint64, saleOrderUuid uint64, saleOrderProductUuid uint64, price float64) error           // 修改订单商品价格
	ChangePopulation(saleBillUuid uint64, population int) error                                                               // 修改订单人数
	ChangeBuffetCustomerType(saleBillUuid uint64, population int) error                                                       // 调整自助餐
	ChangeProductRemark(saleBillUuid uint64, saleOrderUuid uint64, orderProductUuid uint64, remark string) error              // 修改订单商品备注
	GetSaleBillAllInfo(saleBillUuid uint64) (*model.SaleBill, error)                                                          // 获取销售账单所有信息
	HasShowOrder(deviceUuid uint64) (bool, error)                                                                             // 判断该设备是否有未挂单的点餐订单
	GetSaleBillRecord(saleBillUuid uint64) (*model.SaleBill, error)                                                           // 获取销售账单记录,不进行关联查询
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
		return model, err
	}

	return model, nil
}

// CreateSaleBillSetting 创建销售账单设置
func (r *orderRepo) CreateSaleBillSetting(model model.SaleBillSetting) (model.SaleBillSetting, error) {
	err := r.db.Create(&model).Error
	if err != nil {
		return model, err
	}

	return model, nil
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
		return saleBill, result.Error
	}

	return saleBill, nil
}

// CreateSaleOrder 创建销售订单
func (r *orderRepo) CreateSaleOrder(model model.SaleOrder) (model.SaleOrder, error) {
	err := r.db.Create(&model).Error
	if err != nil {
		return model, err
	}
	return model, nil
}

// CreateSaleOrderBuffetCustomerType 创建销售订单自助餐顾客类型
func (r *orderRepo) CreateSaleOrderBuffetCustomerType(model model.SaleOrderBuffetCustomerType) (model.SaleOrderBuffetCustomerType, error) {
	err := r.db.Create(&model).Error
	if err != nil {
		return model, err
	}
	return model, nil
}

// DeleteSaleOrderBuffetCustomerType 删除销售订单自助餐顾客类型
func (r *orderRepo) DeleteSaleOrderBuffetCustomerType(saleOrderUuid uint64) error {
	err := r.db.Where("sale_order_uuid = ?", saleOrderUuid).Delete(&model.SaleOrderBuffetCustomerType{}).Error
	if err != nil {
		return err
	}
	return nil
}

// GetOrderListWithPagination 获取订单列表
func (r *orderRepo) GetOrderListWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.SaleBill, int64, error) {
	var orders []model.SaleBill
	var total int64

	db := r.db.Model(&model.SaleBill{}).Session(&gorm.Session{})

	for _, opt := range opts {
		db = opt(db)
	}

	// 获取总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&orders).Error

	return orders, total, err
}

// GetOrderNum 获取订单的数量
func (r *orderRepo) GetOrderNum(opts ...DBOption) (int64, error) {
	var count int64
	db := r.db.Model(&model.SaleBill{}).Session(&gorm.Session{})

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Count(&count)
	return count, result.Error
}

// GetCashierOrderListWithPaginationType 获取收银台订单列表-参数
type GetCashierOrderListWithPaginationType struct {
	PageNo           int    // 页码
	PageSize         int    // 页面大小
	OrderNo          string // 订单编号
	DateType         int    // 时间类型,-1=全部、0=今天、1=本周、2=本月、3=本年
	EnableCreateTime bool   // 是否启用创建时间
	EnablePayTime    bool   // 是否启用支付时间
	QueryStartTime   uint   // 查询开始时间
	QueryEndTime     uint   // 查询结束时间
	Status           int    // 订单状态,-1=全部、0=待支付、1=已支付、2=已取消、3=已完成
	BillType         int    // 订单类型,-1=全部、0=餐单、1=外卖
	DiningMethod     int    // 用餐方式,-1=全都、 0-堂食 1-打包
}

// GetCashierOrderListWithPagination 获取收银台订单列表
func (r *orderRepo) GetCashierOrderListWithPagination(param GetCashierOrderListWithPaginationType) (lists []model.SaleBill, total int64, err error) {
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
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByNotStatus(constant.SaleBillStatusNoCooking),
		CommonRepo.SortWithID("DESC"),
		// 额外条件
		func() DBOption {
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
				//  账单状态
				if param.Status != -1 {
					db = db.Where("status = ?", uint(param.Status))
				}
				//  日期类型 -1-全都 1-今天 2-昨天 3-本周
				if param.DateType >= 0 && param.DateType <= 3 {
					now := time.Now()
					var startTime, endTime time.Time
					switch param.DateType {
					case 1: // 今天
						startTime = now.Truncate(24 * time.Hour)
						endTime = startTime.Add(24*time.Hour - time.Second)
					case 2: // 昨天
						startTime = now.AddDate(0, 0, -1).Truncate(24 * time.Hour)
						endTime = startTime.Add(24*time.Hour - time.Second)
					case 3: // 本周
						weekday := int(now.Weekday())
						if weekday == 0 {
							weekday = 7
						}
						startTime = now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
						endTime = startTime.AddDate(0, 0, 7).Add(-time.Second)
					}
					db = db.Where("create_time BETWEEN ? AND ?", startTime.Unix(), endTime.Unix())
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
						endTime = param.QueryEndTime + 86399
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
		}(),
	)
	//
	return lists, total, err
}

// GetSaleBillInfo 获取销售账单详细信息
func (r *orderRepo) GetSaleBillInfo(saleBillUuid uint64, saleOrderUuid uint64) (model.SaleBill, error) {
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
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByUuid(saleBillUuid),
	)
	if err != nil {
		return model.SaleBill{}, err
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
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByDeskUuid(deskUuid),
	)
	if err != nil {
		return model.SaleBill{}, err
	}
	return info, nil
}

// GetOrderCartInfo 获取购物车信息
func (r *orderRepo) GetOrderCartInfo(saleBillUuid uint64) (*ro.ShopCartRepo, error) {

	repo := NewSaleBillRepo(r.db)
	saleBill, err := repo.GetSaleBill(
		CommonRepo.WhereByUuid(saleBillUuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByIsHide(false),
	)
	if err != nil {
		return nil, err
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
							Args: []interface{}{
								func(db *gorm.DB) *gorm.DB {
									return db.Where("delete_time = ? AND is_accept_order = ?", constant.NotDeleted, constant.OrderProductIsAcceptOrderAccepted)
								},
							},
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
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
					return nil, errDesk
				}

				bill := &saleBill
				// 计算一次金额，避免错误
				bill.CalcAll()
				return &ro.ShopCartRepo{SaleBill: bill}, nil
			}
		} else {
			// 当销售账单是桌台订单时，额外查询桌台信息
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
							Args: []interface{}{
								func(db *gorm.DB) *gorm.DB {
									return db.Where("delete_time = ? AND is_accept_order = ?", constant.NotDeleted, constant.OrderProductIsAcceptOrderAccepted)
								},
							},
						},
					),
					CommonRepo.Preload(
						WithPreload{
							Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
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
				)
				if errDesk != nil {
					return nil, errDesk
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
				),
				CommonRepo.Preload(
					WithPreload{
						Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
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
			)
			if errSaleBill != nil {
				return nil, errSaleBill
			}
			bill := &saleBill
			// 计算一次金额，避免错误
			bill.CalcAll()
			return &ro.ShopCartRepo{SaleBill: bill}, nil
		}
	}
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
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByUuid(saleBillUuid),
	)
	if err != nil {
		return model.SaleBill{}, err
	}
	return info, nil
}

// GetSaleOrderProductListBySaleOrderProductUuids 根据销售订单商品uuid列表获取销售订单商品列表
func (r *orderRepo) GetSaleOrderProductListBySaleOrderProductUuids(saleOrderProductUuids []uint64) ([]model.SaleOrderProduct, error) {
	products := make([]model.SaleOrderProduct, 0)
	err := r.db.Model(&model.SaleOrderProduct{}).Where("uuid in ?", saleOrderProductUuids).Find(&products).Error
	if err != nil {
		return nil, err
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
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms",
			},
			WithPreload{
				Query: "SaleOrders.ReturnOrders",
			},
			WithPreload{
				Query: "Cashier",
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByUuid(saleBillUuid),
	)
	if err != nil {
		return model.SaleBill{}, err
	}
	return info, nil
}

// CancelOrder 取消订单
func (r *orderRepo) CancelOrder(saleBillUuid uint64, reason string) error {
	timeNow := uint(time.Now().Unix())
	saleOrder := &model.SaleOrder{}
	where := "sale_order_uuid in (select uuid from " + saleOrder.TableName() + " where sale_bill_uuid = ?)"
	//
	return r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.SaleOrder{}).Where("sale_bill_uuid = ?", saleBillUuid).Where("status = ?", constant.SaleBillStatusPending).Update("status", constant.SaleBillStatusCanceled).Error
		if err != nil {
			return err
		}
		err = tx.Model(&model.SaleOrderProduct{}).Where(where, saleBillUuid).Update("delete_time", timeNow).Error
		if err != nil {
			return err
		}
		err = tx.Model(&model.SaleOrderProductBom{}).Where(where, saleBillUuid).Update("delete_time", timeNow).Error
		if err != nil {
			return err
		}
		err = tx.Model(&model.SaleOrderBuffetCustomerType{}).Where(where, saleBillUuid).Update("delete_time", timeNow).Error
		if err != nil {
			return err
		}
		err = tx.Model(&model.SaleOrderDiscountStrategy{}).Where(where, saleBillUuid).Update("delete_time", timeNow).Error
		if err != nil {
			return err
		}
		err = tx.Model(&model.SaleOrderProductAttribute{}).Where(where, saleBillUuid).Update("delete_time", timeNow).Error
		if err != nil {
			return err
		}
		err = tx.Model(&model.ProductionOrder{}).Where(where, saleBillUuid).Update("delete_time", timeNow).Error // TODO: 有可能会有多个生产单).Error
		if err != nil {
			return err
		}
		//
		return tx.Model(&model.SaleBill{}).
			Where("uuid = ?", saleBillUuid).
			Where("status = ?", constant.SaleBillStatusPending).
			Updates(map[string]interface{}{
				"status":      constant.SaleBillStatusCanceled,
				"delete_time": timeNow,
				"reason":      reason,
			}).Error
	})
}

// CancelDeskOrder 关闭桌台订单
func (r *orderRepo) CancelDeskOrder(deskUuid uint64, reason string) error {
	var saleBill model.SaleBill
	if err := r.db.Model(&model.SaleBill{}).
		Where("status = ?", constant.SaleBillStatusPending).
		Where("desk_uuid = ?", deskUuid).
		First(&saleBill).Error; err != nil {
		return err
	}
	return r.CancelOrder(saleBill.Uuid, reason)
}

// DeleteOrder 软删除订单
func (r *orderRepo) DeleteOrder(saleBillUuid uint64, saleOrderUuid uint64) error {
	now := uint(time.Now().Unix())
	tx := r.db.Begin()

	// 删除销售订单
	if err := tx.Model(&model.SaleOrder{}).
		Where("sale_bill_uuid = ? AND status = ?", saleBillUuid, constant.SaleBillStatusPending).
		Where(map[string]interface{}{"uuid": saleOrderUuid}).
		Update("delete_time", now).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 如果是删除全部订单或只剩最后一个订单,则同时删除主订单
	var count int64
	if saleOrderUuid == 0 || tx.Model(&model.SaleOrder{}).Where("sale_bill_uuid = ? AND delete_time = 0", saleBillUuid).Count(&count).Error == nil && count <= 1 {
		if err := tx.Model(&model.SaleBill{}).
			Where("uuid = ? AND status = ?", saleBillUuid, constant.SaleBillStatusCanceled).
			Update("delete_time", now).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
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
		return err
	}
	return tx.Commit().Error
}

// GetSaleBillRecord 获取销售账单记录
func (r *orderRepo) GetSaleBillRecord(saleBillUuid uint64) (*model.SaleBill, error) {
	var saleBill model.SaleBill
	if err := r.db.Model(&model.SaleBill{}).Where("uuid = ?", saleBillUuid).First(&saleBill).Error; err != nil {
		return nil, err
	}
	return &saleBill, nil
}

// GetSaleBillAllInfo 获取销售账单所有信息
func (r *orderRepo) GetSaleBillAllInfo(saleBillUuid uint64) (*model.SaleBill, error) {
	info, err := r.GetSaleBill(
		CommonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						db = db.Where("delete_time = ?", constant.NotDeleted)
						return db
					},
				},
			},
			// ==================== 销售账单的桌台信息 ====================
			// 加载桌台信息
			WithPreload{
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
			// ==================== 销售账单的订单信息 ====================
			WithPreload{
				Query: "SaleOrders.PaymentOrders",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
					CommonRepo.DBOption(CommonRepo.WhereByStatus(constant.PaymentOrderStatusPaid)),
				},
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders.PaymentMethod",
			},
			WithPreload{
				Query: "SaleOrders.Member.MemberLevel",
			},
			WithPreload{
				Query: "SaleOrders.Member.MemberCard.MemberCardType",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts",
				Args: []interface{}{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
					CommonRepo.DBOption(CommonRepo.WhereBySaleBillUuid(saleBillUuid)),
				},
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.ProductPackage",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductAttributes.ProductAttribute.MultiLanguageName",
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
				Query: "SaleOrders.SaleOrderBuffetCustomerTypes.BuffetPackage.MultiLanguageName",
			},
			WithPreload{
				Query: "SaleOrders.SaleOrderBuffetDelayProducts",
			},
			WithPreload{
				Query: "SaleOrders.ReturnOrders",
			},
			// ==================== 销售账单的收银员信息 ====================
			WithPreload{
				Query: "Cashier",
			},
		),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.WhereByUuid(saleBillUuid),
	)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return saleOrderProductBoms, nil
}

// DeleteOrderProduct 删除订单产品
func (r *orderRepo) DeleteOrderProduct(saleBillUuid uint64, saleOrderUuid uint64, saleOrderProductUuid uint64) error {
	timeNow := uint(time.Now().Unix())
	// 删除关联关系
	return r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.SaleOrderProductBom{}).Where("sale_order_product_uuid = ?", saleOrderProductUuid).Update("delete_time", timeNow).Error
		if err != nil {
			return err
		}
		err = tx.Model(&model.SaleOrderProductAttribute{}).Where("sale_order_product_uuid = ?", saleOrderProductUuid).Update("delete_time", timeNow).Error
		if err != nil {
			return err
		}
		err = tx.Model(&model.ProductionOrderProduct{}).Where("sale_order_product_uuid = ?", saleOrderProductUuid).Update("delete_time", timeNow).Error
		if err != nil {
			return err
		}
		return tx.Model(&model.SaleOrderProduct{}).
			Where("(status != ? or cancel_time != 0)", constant.OrderProductStatusSentKitchen).
			Where("delete_time = ?", 0).
			Where("sale_bill_uuid = ? AND sale_order_uuid = ? AND uuid = ?", saleBillUuid, saleOrderUuid, saleOrderProductUuid).
			Update("delete_time", uint(time.Now().Unix())).
			Error
	})
}

// ChangeProductPrice 修改订单商品价格
func (r *orderRepo) ChangeProductPrice(saleBillUuid uint64, saleOrderUuid uint64, saleOrderProductUuid uint64, price float64) error {
	return r.db.Model(&model.SaleOrderProduct{}).
		Where("delete_time = ?", 0).
		Where("sale_bill_uuid = ? AND sale_order_uuid = ? AND uuid = ?", saleBillUuid, saleOrderUuid, saleOrderProductUuid).
		Updates(map[string]interface{}{
			"is_custom_price": 1,
			"product_price":   price,
		}).Error
}

// ChangePopulation 修改订单人数
func (r *orderRepo) ChangePopulation(saleBillUuid uint64, population int) error {
	return r.db.Model(&model.SaleBill{}).
		Where("delete_time = ?", 0).
		Where("uuid = ?", saleBillUuid).
		Updates(map[string]interface{}{
			"meal_num": population,
		}).Error
}

// ChangeBuffetCustomerType 修改订单人数
func (r *orderRepo) ChangeBuffetCustomerType(saleBillUuid uint64, population int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// saleOrder := &model.SaleOrder{}
		// where := "sale_order_uuid in (select uuid from " + saleOrder.TableName() + " where sale_bill_uuid = ?)"
		// //
		// err := tx.Model(&model.SaleOrderBuffetCustomerType{}).
		// 	Where("delete_time = ?", 0).
		// 	Where(where, saleBillUuid).
		// 	Updates(map[string]interface{}{
		// 		"num":    population,
		// 		"remark": remark,
		// 	}).Error
		// if err != nil {
		// 	return err
		// }
		// 	//
		return tx.Model(&model.SaleBill{}).
			Where("delete_time = ?", 0).
			Where("uuid = ?", saleBillUuid).
			Updates(map[string]interface{}{
				"meal_num": population,
			}).Error
	})
}

// ChangeProductRemark 修改订单商品备注
func (r *orderRepo) ChangeProductRemark(saleBillUuid uint64, saleOrderUuid uint64, orderProductUuid uint64, remark string) error {
	return r.db.Model(&model.SaleOrderProduct{}).
		Where("delete_time = ?", constant.NotDeleted).
		Where("sale_bill_uuid = ? AND sale_order_uuid = ? AND uuid = ?", saleBillUuid, saleOrderUuid, orderProductUuid).
		Updates(map[string]interface{}{
			"remark": remark,
		}).Error
}

// GetSaleBillProductInfoByDesk 获取桌台的账单商品信息
func (r *orderRepo) GetSaleBillProductInfoByDesk(deskUuid uint64) (model.SaleBill, error) {
	var saleBill model.SaleBill
	if err := r.db.Preload("SaleOrders", func(db *gorm.DB) *gorm.DB {
		return db.Where("delete_time = ?", constant.NotDeleted)
	}).Preload("SaleOrders.SaleOrderProducts", func(db *gorm.DB) *gorm.DB {
		return db.Where("delete_time = ? AND is_accept_order = ?", constant.NotDeleted, constant.OrderProductIsAcceptOrderAccepted)
	}).Preload("SaleOrders.SaleOrderProducts.MultiLanguageName").Preload("SaleOrders.SaleOrderProducts.SaleOrderProductAttributes").Model(&model.SaleBill{}).Where("desk_uuid = ?", deskUuid).Find(&saleBill).Error; err != nil {
		return model.SaleBill{}, err
	}
	return saleBill, nil
}

// HasShowOrder 判断该设备是否有未挂单的点餐订单
func (r *orderRepo) HasShowOrder(deviceUuid uint64) (bool, error) {
	var saleBill model.SaleBill
	if err := r.db.Where("device_uuid = ? AND status = ? AND hide_bill_time = ? AND delete_time = ?", deviceUuid, constant.SaleBillStatusPending, 0, constant.NotDeleted).First(&saleBill).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return saleBill.Uuid != 0, nil
}
