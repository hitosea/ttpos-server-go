package repository

import (
	"fmt"
	"ttpos-server-go/app/constant"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DBOption func(*gorm.DB) *gorm.DB

// WithPreload 预加载
type WithPreload struct {
	Query string
	Args  []interface{}
}

// NotDeleted 筛选未被删除的
func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("delete_time = ?", constant.NotDeleted)
}

func Like(keyword string) string {
	return "%" + keyword + "%"
}

// ICommonRepo 公共仓库接口
type ICommonRepo interface {
	WhereByID(id uint) DBOption                                     // 根据ID查询
	WhereByUuid(uuid uint64) DBOption                               // 根据UUID查询
	WhereInUuids(uuids []uint64) DBOption                           // 根据UUID列表查询
	WhereByDeskUuid(uuid uint64) DBOption                           // 根据桌台UUID查询
	WhereBySaleBillUuid(uuid uint64) DBOption                       // 根据销售单UUID查询
	WhereBySaleOrderUuid(uuid uint64) DBOption                      // 根据销售订单UUID查询
	WhereByStatus(status uint) DBOption                             // 根据状态查询
	WhereByIsShowCashier(isShowCashier uint) DBOption               // 根据是否显示收银机查询
	WhereByIsShowAssistant(isShow uint) DBOption                    // 根据是否显示点餐助手端查询
	WhereByIsShowTablet(isShow uint) DBOption                       // 根据是否显示平板端查询
	WhereByIsShowKitchen(isShow uint) DBOption                      // 根据是否显示厨显端查询
	WhereBySoftDelete() DBOption                                    // 根据软删除查询
	WhereByCooking() DBOption                                       // 根据账单已经送厨房查询
	WhereByRelatedUuid(relatedUuid uint64) DBOption                 // 根据关联UUID查询
	WhereByRelatedType(relatedType uint) DBOption                   // 根据关联类型查询
	WhereByOrderNo(orderNo string) DBOption                         // 根据订单编号查询
	WhereByBillType(billType uint) DBOption                         // 根据账单类型查询
	WhereByNotStatus(status uint) DBOption                          // 根据状态查询
	WhereByIsHide(isHide bool) DBOption                             // 根据是否隐藏查询
	WhereByDeviceUuid(deviceUuid uint64) DBOption                   // 根据设备uuid查询
	WhereByDeviceSn(deviceSn string) DBOption                       // 根据设备Sn查询
	WhereByNoDisable() DBOption                                     // 根据没禁用查询
	WhereByBuffetPackageUuid(buffetPackageUuid uint64) DBOption     // 根据自助餐套餐UUID查询
	WhereByCustomerTypeUuid(customerTypeUuid uint64) DBOption       // 根据顾客类型UUID查询
	WhereByProductPackageUuid(productPackUuid uint64) DBOption      // 根据产品套餐UUID查询
	WhereByProductFlavorUuid(productFlavorUuid uint64) DBOption     // 根据产品口味UUID查询
	WhereBySign(sign string) DBOption                               // 根据签名查询
	WhereLikeByName(name string) DBOption                           // 根据名称查询
	WhereBetweenByCreateTime(startTime uint, endTime uint) DBOption // 根据创建时间查询
	SortWithID(order string) DBOption                               // 根据ID排序
	SortWithCreateTime(order string) DBOption                       // 根据创建时间排序
	SortWithSort(order string) DBOption                             // 根据Order By排序
	SortWithIsSpecial(order string) DBOption                        // 根据是否特殊排序
	Preload(preloads ...WithPreload) DBOption                       // 预加载
	IncrementNum(num uint) clause.Expr                              // 增加商品数量
	DecrementNum(num uint) clause.Expr                              // 减少商品数量
	WhereByProductIsAccept() DBOption                               // 根据商品是否接单查询
	DBOption(opt DBOption) func(*gorm.DB) *gorm.DB                  // 将DBOption转为func(*gorm.DB) *gorm.DB
	Transaction(db *gorm.DB, fn func(tx *gorm.DB) error) error      // 事务
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

func (r *commonRepo) DBOption(opt DBOption) func(*gorm.DB) *gorm.DB {
	f := (func(*gorm.DB) *gorm.DB)(opt)
	return f
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

// WhereInUuids 根据UUID列表查询
func (r *commonRepo) WhereInUuids(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid IN (?)", uuids)
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
		return db.Where("sale_bill_uuid = ?", uuid)
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

// WhereByIsShowAssistant 根据是否显示点餐助手端查询
func (r *commonRepo) WhereByIsShowAssistant(isShow uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_show_assistant = ?", isShow)
	}
}

// WhereByIsShowTablet 根据是否显示平板端查询
func (r *commonRepo) WhereByIsShowTablet(isShow uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_show_tablet = ?", isShow)
	}
}

// WhereByIsShowKitchen 根据是否显示厨显端查询
func (r *commonRepo) WhereByIsShowKitchen(isShow uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_show_kitchen = ?", isShow)
	}
}

// WhereBySoftDelete 根据软删除查询
func (r *commonRepo) WhereBySoftDelete() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("delete_time = %d", constant.NotDeleted))
	}
}

func (r *commonRepo) WhereByCooking() DBOption {
	//首次送厨时间大于0，表示该销售账单已经送厨
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("production_time > 0")
	}
}

// WhereByRelatedUuid 根据关联UUID查询
func (r *commonRepo) WhereByRelatedUuid(relatedUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("related_uuid = ?", relatedUuid)
	}
}

// WhereByRelatedType 根据关联类型查询
func (r *commonRepo) WhereByRelatedType(relatedType uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("related_type = ?", relatedType)
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

func (r *commonRepo) WhereByNotStatus(status uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status <> ?", status)
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

func (r *commonRepo) WhereByDeviceUuid(deviceUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("device_uuid = ?", deviceUuid)
	}
}

func (r *commonRepo) WhereByDeviceSn(deviceSn string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("device_id = ?", deviceSn)
	}
}

func (r *commonRepo) WhereByNoDisable() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("is_disable = %d", constant.DeskEnable))
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

func (r *commonRepo) WhereByProductIsAccept() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_accept_order = ?", constant.OrderProductIsAcceptOrderAccepted)
	}
}

// SortWithID 根据ID排序
func (r *commonRepo) SortWithID(order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("id " + order)
	}
}

func (r *commonRepo) SortWithCreateTime(order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("create_time " + order)
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

// Transaction 事务
func (r *commonRepo) Transaction(db *gorm.DB, fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}
