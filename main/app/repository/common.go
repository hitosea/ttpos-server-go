package repository

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"ttpos-server-go/app/constant"
)

type Where func(*gorm.DB) *gorm.DB

func handleWiths(db *gorm.DB, withs []With) *gorm.DB {
	if len(withs) == 0 {
		return db
	}
	for _, with := range withs {
		db = with(db)
	}
	return db
}

type With func(*gorm.DB) *gorm.DB
type DBOption func(*gorm.DB) *gorm.DB

// WithPreload 预加载
type WithPreload struct {
	Query string
	Args  []interface{}
}

// NotDeleted 筛选未被删除的
func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where(fmt.Sprintf("delete_time = %d", constant.NotDeleted))
}

func Like(keyword string) string {
	return "%" + keyword + "%"
}

// ICommonRepo 公共仓库接口
type ICommonRepo interface {
	WhereByID(id uint) DBOption                                     // 根据ID查询
	WhereByUuid(uuid uint64) DBOption                               // 根据UUID查询
	WhereByDeskUuid(uuid uint64) DBOption                           // 根据桌台UUID查询
	WhereBySaleBillUuid(uuid uint64) DBOption                       // 根据销售单UUID查询
	WhereBySaleOrderUuid(uuid uint64) DBOption                      // 根据销售订单UUID查询
	WhereByStatus(status uint) DBOption                             // 根据状态查询
	WhereByIsShowCashier(isShowCashier uint) DBOption               // 根据是否显示收银机查询
	WhereBySoftDelete() DBOption                                    // 根据软删除查询
	WhereByOrderNo(orderNo string) DBOption                         // 根据订单编号查询
	WhereByBillType(billType uint) DBOption                         // 根据账单类型查询
	WhereByIsHide(isHide bool) DBOption                             // 根据是否隐藏查询
	WhereByBuffetPackageUuid(buffetPackageUuid uint64) DBOption     // 根据自助餐套餐UUID查询
	WhereByCustomerTypeUuid(customerTypeUuid uint64) DBOption       // 根据顾客类型UUID查询
	WhereByProductPackageUuid(productPackUuid uint64) DBOption      // 根据产品套餐UUID查询
	WhereByProductFlavorUuid(productFlavorUuid uint64) DBOption     // 根据产品口味UUID查询
	WhereBySign(sign string) DBOption                               // 根据签名查询
	WhereLikeByName(name string) DBOption                           // 根据名称查询
	WhereBetweenByCreateTime(startTime uint, endTime uint) DBOption // 根据创建时间查询
	SortWithID(order string) DBOption                               // 根据ID排序
	SortWithSort(order string) DBOption                             // 根据Order By排序
	SortWithIsSpecial(order string) DBOption                        // 根据是否特殊排序
	Preload(preloads ...WithPreload) DBOption                       // 预加载
	IncrementNum(num uint) clause.Expr                              // 增加商品数量
	DecrementNum(num uint) clause.Expr                              // 减少商品数量
}

// commonRepo 公共仓库实现
type commonRepo struct{}

var CommonRepo ICommonRepo

func init() {
	CommonRepo = NewCommonRepo()
}

// NewCommonRepo 创建新的公共仓库
func NewCommonRepo() ICommonRepo {
	return NewCommonRepoImpl()
}

// NewCommonRepoImpl 创建新的公共仓库实现
func NewCommonRepoImpl() ICommonRepo {
	return &commonRepo{}
}

// WhereByID 根据ID查询
func (r *commonRepo) WhereByID(id uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	}
}

// WhereByUuid 根据UUID查询
func (r *commonRepo) WhereByUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

func (r *commonRepo) WhereByDeskUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("desk_uuid = ?", uuid)
	}
}

// WhereBySaleBillUuid 根据销售单UUID查询
func (r *commonRepo) WhereBySaleBillUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereByStatus 根据状态查询
func (r *commonRepo) WhereByStatus(status uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WhereByIsShowCashier 根据是否显示收银机查询
func (r *commonRepo) WhereByIsShowCashier(isShowCashier uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_show_cashier = ?", isShowCashier)
	}
}

// WhereBySoftDelete 根据软删除查询
func (r *commonRepo) WhereBySoftDelete() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("delete_time = %d", constant.NotDeleted))
	}
}

// WhereByOrderNo 根据订单编号查询
func (r *commonRepo) WhereByOrderNo(orderNo string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("order_no = ?", orderNo)
	}
}

// WhereByBillType 根据账单类型查询
func (r *commonRepo) WhereByBillType(billType uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("bill_type = ?", billType)
	}
}

// WhereByIsHide 根据是否隐藏查询
func (r *commonRepo) WhereByIsHide(isHide bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if isHide {
			return db.Where("hide_bill_time > 0")
		}
		return db.Where("hide_bill_time = 0")
	}
}

// WhereByBuffetPackageUuid 根据自助餐套餐UUID查询
func (r *commonRepo) WhereByBuffetPackageUuid(buffetPackageUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("buffet_package_uuid = ?", buffetPackageUuid)
	}
}

// WhereByCustomerTypeUuid 根据顾客类型UUID查询
func (r *commonRepo) WhereByCustomerTypeUuid(customerTypeUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("customer_type_uuid = ?", customerTypeUuid)
	}
}

// WhereByProductPackageUuid 根据产品套餐UUID查询
func (r *commonRepo) WhereByProductPackageUuid(productPackUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_package_uuid = ?", productPackUuid)
	}
}

// WhereByProductFlavorUuid 根据产品口味UUID查询
func (r *commonRepo) WhereByProductFlavorUuid(productFlavorUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_flavor_uuid = ?", productFlavorUuid)
	}
}

// WhereLikeByName 根据名称查询
func (r *commonRepo) WhereLikeByName(name string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name LIKE ?", "%"+name+"%")
	}
}

// WhereBetweenByCreateTime 根据创建时间查询
func (r *commonRepo) WhereBetweenByCreateTime(startTime uint, endTime uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("create_time BETWEEN ? AND ?", startTime, endTime)
	}
}

// SortWithID 根据ID排序
func (r *commonRepo) SortWithID(order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("id " + order)
	}
}

// SortWithSort 根据Order By排序
func (r *commonRepo) SortWithSort(order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("sort " + order)
	}
}

// SortWithIsSpecial 根据是否特殊排序
func (r *commonRepo) SortWithIsSpecial(order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("is_special " + order)
	}
}

// Preload 预加载
func (r *commonRepo) Preload(preloads ...WithPreload) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		for _, preload := range preloads {
			db = db.Preload(preload.Query, preload.Args...)
		}
		return db
	}
}

// WhereBySaleOrderUuid 根据销售订单UUID查询
func (r *commonRepo) WhereBySaleOrderUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sale_order_uuid = ?", uuid)
	}
}

// WhereBySign 根据签名查询
func (r *commonRepo) WhereBySign(sign string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sign = ?", sign)
	}
}

// IncrementNum 增加商品数量
func (r *commonRepo) IncrementNum(num uint) clause.Expr {
	return gorm.Expr("num + ?", num)
}

// DecrementNum 减少商品数量
func (r *commonRepo) DecrementNum(num uint) clause.Expr {
	return gorm.Expr("num - ?", num)
}
