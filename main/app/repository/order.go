package repository

import (
	"fmt"
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IOrderRepo 定义订单仓库接口
type IOrderRepo interface {
	CreateSaleBill(model model.SaleBill) (model.SaleBill, error)                                                    // 创建销售单
	GetSaleBill(opts ...DBOption) (model.SaleBill, error)                                                           // 获取销售单
	CreateSaleOrder(model model.SaleOrder) (model.SaleOrder, error)                                                 // 创建订单
	GetOrderListWithPagination(pageNo int, pageSize int, opts ...DBOption) ([]model.SaleBill, int64, error)         // 获取订单列表
	GetOrderNum(opts ...DBOption) (int64, error)                                                                    // 获取订单数量
	GetCashierOrderListWithPagination(param GetCashierOrderListWithPaginationType) ([]model.SaleBill, int64, error) // 获取收银的订单列表
	GetSaleBillDetail(saleBillUuid uint64, saleOrderUuid uint64) (model.SaleBill, error)                            // 获取销售账单详细信息
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

// GetSaleBill 获取销售单
func (r *orderRepo) GetSaleBill(opts ...DBOption) (model.SaleBill, error) {
	var model model.SaleBill
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&model)
	if result.Error != nil {
		return model, result.Error
	}

	return model, nil
}

// CreateSaleOrder 创建销售订单
func (r *orderRepo) CreateSaleOrder(model model.SaleOrder) (model.SaleOrder, error) {
	err := r.db.Create(&model).Error
	if err != nil {
		return model, err
	}
	return model, nil
}

// GetOrderList 获取订单列表
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

// 获取订单的数量
func (r *orderRepo) GetOrderNum(opts ...DBOption) (int64, error) {
	var count int64
	db := r.db.Model(&model.SaleBill{}).Session(&gorm.Session{})

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Count(&count)
	return count, result.Error
}

// GetCashierOrderListWithPagination 获取收银台订单列表-参数
type GetCashierOrderListWithPaginationType struct {
	PageNo           int
	PageSize         int
	OrderNo          string
	DateType         int
	EnableCreateTime bool
	EnablePayTime    bool
	QueryStartTime   uint
	QueryEndTime     uint
	Status           int
	BillType         int
}

// GetCashierOrderListWithPagination 获取收银台订单列表
func (r *orderRepo) GetCashierOrderListWithPagination(param GetCashierOrderListWithPaginationType) (lists []model.SaleBill, total int64, err error) {
	commonRepo := NewCommonRepo()
	lists, total, err = r.GetOrderListWithPagination(
		param.PageNo,
		param.PageSize,
		commonRepo.Preload(
			WithPreload{
				Query: "SaleOrders",
				Args: []interface{}{
					func(db *gorm.DB) *gorm.DB {
						return db.Where("delete_time = ?", 0)
					},
				},
			},
			WithPreload{
				Query: "SaleOrders.PaymentOrders.PaymentMethod",
			},
		),
		commonRepo.WhereBySoftDelete(),
		commonRepo.SortWithID("DESC"),
		// 额外条件
		func() DBOption {
			return func(db *gorm.DB) *gorm.DB {
				// 订单编号
				if param.OrderNo != "" {
					db = db.Where("order_no = ?", param.OrderNo)
				}
				// 账单类型
				if param.BillType != -1 {
					db = db.Where("bill_type = ?", param.BillType)
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

// GetSaleBillDetail 获取销售账单详细信息
func (r *orderRepo) GetSaleBillDetail(saleBillUuid uint64, saleOrderUuid uint64) (model.SaleBill, error) {
	commonRepo := NewCommonRepo()
	info, err := r.GetSaleBill(
		commonRepo.Preload(
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
				Query: "SaleOrders.SaleOrderProducts.SaleOrderProductBoms",
			},
		),
		commonRepo.WhereByUuid(saleBillUuid),
	)
	if err != nil {
		return model.SaleBill{}, err
	}
	return info, nil
}
