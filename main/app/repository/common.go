package repository

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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

// getCompanyUuid 从数据库名称中提取公司UUID
func GetCompanyUuid(db *gorm.DB) uint64 {
	var dbName string
	db.Raw("SELECT DATABASE()").Scan(&dbName)
	if len(dbName) > 4 && strings.HasPrefix(dbName, "shop") {
		uuidStr := strings.TrimPrefix(dbName, "shop")
		uuid, err := strconv.ParseUint(uuidStr, 10, 64)
		if err == nil {
			return uuid
		}
	}
	// NOTE sqlite - 离线版本要重新考虑
	return 0
}

// NotDeleted 筛选未被删除的
func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("delete_time = ?", constant.NotDeleted)
}

func Like(keyword string) string {
	specialChars := []string{"%", "_", "\\\\", "[", "]", "^", "|", "$", "(", ")"}

	keyword = strings.TrimSpace(keyword)
	for _, char := range specialChars {
		keyword = strings.ReplaceAll(keyword, char, "\\"+char)
	}

	return "%" + keyword + "%"
}

// ICommonRepo 公共仓库接口
type ICommonRepo interface {
	WhereByID(id uint) DBOption                                         // 根据ID查询
	WhereByUuid(uuid uint64) DBOption                                   // 根据UUID查询
	WhereByFromUnitUuid(fromUnitUuid uint64) DBOption                   // 根据来源单位UUID查询
	WhereInUuids(uuids []uint64) DBOption                               // 根据UUID列表查询
	WhereByMemberSaleOrderUuid(uuid uint64) DBOption                    // 根据会员端销售订单UUID查询
	WhereByDeskUuid(uuid uint64) DBOption                               // 根据桌台UUID查询
	WhereBySaleBillUuid(uuid uint64) DBOption                           // 根据销售单UUID查询
	WhereByAssociatedOrderUuid(uuid uint64) DBOption                    // 根据关联订单UUID查询
	WhereBySaleOrderUuid(uuid uint64) DBOption                          // 根据销售订单UUID查询
	WhereByH5OrderUuid(uuid uint64) DBOption                            // 根据h5订单UUID查询
	WhereByMemberUuid(uuid uint64) DBOption                             // 根据会员UUID查询
	WhereByNotRevoked() DBOption                                        // 未撤销的出库记录
	WhereByStatus(status uint) DBOption                                 // 根据状态查询
	WhereBySerialNumber(serialNo string) DBOption                       // 根据外送序号查询
	WhereByMultipleStatus(statusList []uint) DBOption                   // 根据多个状态查询
	WhereBySource(source uint) DBOption                                 // 根据来源查询
	WhereByRequirement(requirement string) DBOption                     // 根据requirement查询
	WhereByIsShowCashier(isShowCashier uint) DBOption                   // 根据是否显示收银机查询
	WhereByIsShowAssistant(isShow uint) DBOption                        // 根据是否显示点餐助手端查询
	WhereByIsShowTablet(isShow uint) DBOption                           // 根据是否显示平板端查询
	WhereByIsShowKitchen(isShow uint) DBOption                          // 根据是否显示厨显端查询
	WhereByIsShowH5(isShow uint) DBOption                               // 根据是否显示H5端查询
	WhereByIsShowMember(isShow uint) DBOption                           // 根据是否显示会员端查询
	WhereBySoftDelete() DBOption                                        // 根据软删除查询
	WhereByNotPackageSubProduct() DBOption                              // 根据不是套餐子商品查询
	WhereByNoSelectingTimeout() DBOption                                // 根据选购超时查询
	WhereByIsDefault(isDefault uint) DBOption                           // 根据是否默认查询
	WhereByCooking() DBOption                                           // 根据账单已经送厨房查询
	WhereByRelatedUuid(relatedUuid uint64) DBOption                     // 根据关联UUID查询
	WhereByRelatedType(relatedType uint) DBOption                       // 根据关联类型查询
	WhereByOrderNo(orderNo string) DBOption                             // 根据订单编号查询
	WhereByBillType(billType uint) DBOption                             // 根据账单类型查询
	WhereInBillType(billTypeList []uint) DBOption                       // 根据账单类型查询
	WhereByNotStatus(status uint) DBOption                              // 根据状态查询
	WhereByIsHide(isHide bool) DBOption                                 // 根据是否隐藏查询
	WhereByReduceStock(reduceStock uint) DBOption                       // 根据是否减库存查询
	WhereByAddStock(addStock uint) DBOption                             // 根据是否加库存查询
	WhereByDeviceUuid(deviceUuid uint64) DBOption                       // 根据设备uuid查询
	WhereByDeviceSn(deviceSn string) DBOption                           // 根据设备Sn查询
	WhereByNoDisable() DBOption                                         // 根据没禁用查询
	WhereByBuffetPackageUuid(buffetPackageUuid uint64) DBOption         // 根据自助餐套餐UUID查询
	WhereByCustomerTypeUuid(customerTypeUuid uint64) DBOption           // 根据顾客类型UUID查询
	WhereByProductPackageUuid(productPackUuid uint64) DBOption          // 根据产品套餐UUID查询
	WhereByPackageGroupUuid(packageGroupUuid uint64) DBOption           // 根据套餐分组UUID查询
	WhereByProductFlavorUuid(productFlavorUuid uint64) DBOption         // 根据产品口味UUID查询
	WhereBySign(sign string) DBOption                                   // 根据签名查询
	WhereByStaffUuid(staffUuid uint64) DBOption                         // 根据员工UUID查询
	WhereByShiftNo(shiftNo string) DBOption                             // 根据交班编号查询
	WhereByProcessedNot() DBOption                                      // 根据未处理查询
	WhereByRefundTime(refundTime int64) DBOption                        // 根据退款时间查询
	WhereByDutyNo(dutyNo string) DBOption                               // 根据班次编号查询
	WhereByShiftLogUuid(shiftLogUuid uint64) DBOption                   // 根据交班记录UUID查询
	WhereByAction(action string) DBOption                               // 根据操作查询
	WhereByOperatorUuid(operatorUuid uint64) DBOption                   // 根据操作员UUID查询
	WhereByIsVisitor(isVisitor uint) DBOption                           // 根据是否访客查询
	WhereLikeByName(name string) DBOption                               // 根据名称查询
	WhereBetweenByCreateTime(startTime int64, endTime int64) DBOption   // 根据创建时间查询
	WhereBetweenByPayTime(startTime int64, endTime int64) DBOption      // 根据支付时间查询
	WhereBetweenByCompleteTime(startTime int64, endTime int64) DBOption // 根据完成时间查询
	WhereGtUuid(uuid uint64) DBOption                                   // 根据UUID大于查询
	FilterSaleOrderProduct() DBOption                                   // 只查询常规的购物车商品
	FilterSaleOrderProductWithH5Order(h5OrderUuid uint64) DBOption      // 只查询常规的购物车商品、指定某个h5订单的商品
	FilterSaleOrderProductH5Unordered() DBOption                        // 只查询H5未下单的购物车商品
	FilterSaleOrderProductH5Ordered() DBOption                          // 只查询H5已下单的购物车商品
	FilterSaleOrderProductH5OrderedWithReject() DBOption                // 查询H5已下单的购物车商品.包括已送厨商品、已下单未接单的商品和被拒单的商品
	SortWithID(order string) DBOption                                   // 根据ID排序
	SortWithCreateTime(order string) DBOption                           // 根据创建时间排序
	SortWithSubmitPayTime(order string) DBOption                        // 根据提交支付时间排序
	SortWithPayTime(order string) DBOption                              // 根据支付时间排序
	SortWithHandleTime(order string) DBOption                           // 根据h5订单处理时间排序
	WhereCreateTimeGt(createTime int64) DBOption                        // 根据创建时间大于查询
	SortWithSort(order string) DBOption                                 // 根据Order By排序
	SortWithIsSpecial(order string) DBOption                            // 根据是否特殊排序
	Preload(preloads ...WithPreload) DBOption                           // 预加载
	IncrementNum(num uint) clause.Expr                                  // 增加商品数量
	DecrementNum(num uint) clause.Expr                                  // 减少商品数量
	WhereByProductIsAccept() DBOption                                   // 根据商品是否接单查询
	WhereByCountGtZero() DBOption                                       // 根据count大于零查询
	WhereByValidTime() DBOption                                         // 是否在有效期内
	WhereByStartTimeEndTime() DBOption                                  // 是否在有效期内
	WhereLike(field string, keyword string) DBOption                    // 根据字段模糊查询
	WhereByCategoryUuid(categoryUuid uint64) DBOption                   // 根据分类UUID查询
	DBOption(opt DBOption) func(*gorm.DB) *gorm.DB                      // 将DBOption转为func(*gorm.DB) *gorm.DB
	Transaction(db *gorm.DB, fn func(tx *gorm.DB) error) error          // 事务
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

// WhereByFromUnitUuid 根据来源单位UUID查询
func (r *commonRepo) WhereByFromUnitUuid(fromUnitUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("from_unit_uuid = ?", fromUnitUuid)
	}
}

// WhereByMemberSaleOrderUuid 根据会员销售订单UUID查询
func (r *commonRepo) WhereByMemberSaleOrderUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("member_sale_order_uuid = ?", uuid)
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

// WhereByAssociatedOrderUuid 根据关联订单UUID查询
func (r *commonRepo) WhereByAssociatedOrderUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("associated_order_uuid = ?", uuid)
	}
}

// WhereByStatus 根据状态查询
func (r *commonRepo) WhereByStatus(status uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WhereBySerialNumber 根据外送序号查询
func (r *commonRepo) WhereBySerialNumber(serialNo string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("serial_number like ?", "%"+serialNo+"%")
	}
}

// 根据多个状态查询
func (r *commonRepo) WhereByMultipleStatus(statusList []uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status IN (?)", statusList)
	}
}

// WhereBySource 根据来源查询
func (r *commonRepo) WhereBySource(source uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("source = ?", source)
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

// WhereByRequirement 根据requirement查询
func (r *commonRepo) WhereByRequirement(requirement string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("requirement = ?", requirement)
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

// WhereByIsShowH5 根据是否显示H5端查询
func (r *commonRepo) WhereByIsShowH5(isShow uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_show_h5 = ?", isShow)
	}
}

// WhereByIsShowMember 根据是否显示会员端查询
func (r *commonRepo) WhereByIsShowMember(isShow uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_show_delivery = ?", isShow)
	}
}

// WhereBySoftDelete 根据软删除查询
func (r *commonRepo) WhereBySoftDelete() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("delete_time = %d", constant.NotDeleted))
	}
}

// WhereByNotPackageSubProduct 根据不是套餐子商品查询
func (r *commonRepo) WhereByNotPackageSubProduct() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_type != ?", constant.ProductTypePackageSubProduct)
	}
}

// WhereByNoSelectingTimeout 根据选购超时查询
func (r *commonRepo) WhereByNoSelectingTimeout() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("cancel_scene != ?", constant.MemberSaleOrderSceneSelectingTimeout)
	}
}

// WhereByIsDefault 根据是否默认查询
func (r *commonRepo) WhereByIsDefault(isDefault uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_default = ?", isDefault)
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
		return db.Where("order_no like ?", "%"+orderNo+"%")
	}
}

// WhereByBillType 根据账单类型查询
func (r *commonRepo) WhereByBillType(billType uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("bill_type = ?", billType)
	}
}

// WhereInBillType 根据账单类型查询
func (r *commonRepo) WhereInBillType(billTypeList []uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("bill_type IN (?)", billTypeList)
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

// WhereByReduceStock 根据是否减库存查询
func (r *commonRepo) WhereByReduceStock(reduceStock uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("reduce_stock = ?", reduceStock)
	}
}

// WhereByAddStock 根据是否加库存查询
func (r *commonRepo) WhereByAddStock(addStock uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("add_stock = ?", addStock)
	}
}

func (r *commonRepo) WhereByDeviceUuid(deviceUuid uint64) DBOption {
	if deviceUuid == 0 {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
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

// WhereByPackageGroupUuid 根据套餐分组UUID查询
func (r *commonRepo) WhereByPackageGroupUuid(packageGroupUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("package_group_uuid = ?", packageGroupUuid)
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
func (r *commonRepo) WhereBetweenByCreateTime(startTime int64, endTime int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("create_time BETWEEN ? AND ?", startTime, endTime)
	}
}

// WhereBetweenByPayTime 根据支付时间查询
func (r *commonRepo) WhereBetweenByPayTime(startTime int64, endTime int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("pay_time BETWEEN ? AND ?", startTime, endTime)
	}
}

// WhereBetweenByCompleteTime 根据完成时间查询
func (r *commonRepo) WhereBetweenByCompleteTime(startTime int64, endTime int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("complete_time BETWEEN ? AND ?", startTime, endTime)
	}
}

// WhereGtUuid 根据UUID大于查询
func (r *commonRepo) WhereGtUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid > ?", uuid)
	}
}

// FilterSaleOrderProduct 只查询常规的购物车商品
func (r *commonRepo) FilterSaleOrderProduct() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("delete_time = ? AND is_accept_order = ?", constant.NotDeleted, constant.OrderProductIsAcceptOrderAccepted)
	}
}

// FilterSaleOrderProductWithH5Order 只查询常规的购物车商品、指定某个h5订单的商品
func (r *commonRepo) FilterSaleOrderProductWithH5Order(h5OrderUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`(
				( delete_time = ? AND is_accept_order = ? AND h5_order_uuid = ?) 
				OR 
				( delete_time = ? AND is_accept_order = ?)
			)`,
			constant.NotDeleted, constant.OrderProductIsAcceptOrderUnAccept, h5OrderUuid,
			constant.NotDeleted, constant.OrderProductIsAcceptOrderAccepted,
		)
	}
}

// FilterSaleOrderProductH5Unordered 只查询H5未下单的购物车商品
func (r *commonRepo) FilterSaleOrderProductH5Unordered() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("delete_time = ? AND is_accept_order = ? AND h5_order_uuid = ?", constant.NotDeleted, constant.OrderProductIsAcceptOrderUnAccept, constant.OptionalUuid)
	}
}

// FilterSaleOrderProductH5Ordered 只查询H5已下单的购物车商品.包括已送厨商品和已下单未接单的商品
func (r *commonRepo) FilterSaleOrderProductH5Ordered() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(`
				( delete_time = ? AND h5_order_uuid <> ? AND is_accept_order = ?) 
				OR 
				( delete_time = ? AND is_accept_order = ? AND status > ?)
			`,
			constant.NotDeleted, constant.OptionalUuid, constant.OrderProductIsAcceptOrderUnAccept,
			constant.NotDeleted, constant.OrderProductIsAcceptOrderAccepted, constant.SaleOrderProductStatusCooking,
		)
	}
}

// FilterSaleOrderProductH5OrderedWithReject 查询H5已下单的购物车商品.包括已送厨商品、已下单未接单的商品和被拒单的商品
func (r *commonRepo) FilterSaleOrderProductH5OrderedWithReject() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		sql := strings.ReplaceAll(`(
				(delete_time = ? AND h5_order_uuid <> ? AND is_accept_order = ?) OR 
				(delete_time <> ? AND h5_order_uuid <> ? AND is_accept_order = ?) OR 
				(delete_time = ? AND is_accept_order = ? AND status > ?)
			)`, "\t", "")
		return db.
			Where(sql, constant.NotDeleted, constant.OptionalUuid, constant.OrderProductIsAcceptOrderUnAccept,
				constant.NotDeleted, constant.OptionalUuid, constant.OrderProductIsAcceptOrderUnAccept,
				constant.NotDeleted, constant.OrderProductIsAcceptOrderAccepted, constant.SaleOrderProductStatusNormal,
			)
	}
}

func (r *commonRepo) WhereByProductIsAccept() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_accept_order = ?", constant.OrderProductIsAcceptOrderAccepted)
	}
}

func (r *commonRepo) WhereByCountGtZero() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("count > 0")
	}
}

func (r *commonRepo) WhereByValidTime() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		nowTimestamp := time.Now().Unix()
		return db.Where("valid_start_time <= ? AND valid_end_time >= ?", nowTimestamp, nowTimestamp)
	}
}

func (r *commonRepo) WhereByStartTimeEndTime() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		nowTimestamp := time.Now().Unix()
		return db.Where("start_time <= ? AND end_time >= ?", nowTimestamp, nowTimestamp)
	}
}

func (r *commonRepo) WhereLike(field string, keyword string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(field+" LIKE ?", Like(keyword))
	}
}

func (r *commonRepo) WhereByCategoryUuid(categoryUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("category_uuid = ?", categoryUuid)
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

// SortWithSubmitPayTime 根据提交支付时间排序
func (r *commonRepo) SortWithSubmitPayTime(order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("submit_pay_time " + order)
	}
}

// 按支付时间排序
func (r *commonRepo) SortWithPayTime(order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("pay_time " + order)
	}
}

// 按处理时间排序
func (r *commonRepo) SortWithHandleTime(order string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("handle_time " + order)
	}
}

// WhereCreateTimeGt 根据创建时间大于查询
func (r *commonRepo) WhereCreateTimeGt(createTime int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("create_time > ?", createTime)
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

// WhereByH5OrderUuid 根据h5订单UUID查询
func (r *commonRepo) WhereByH5OrderUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("h5_order_uuid = ?", uuid)
	}
}

// WhereByMemberUuid 根据会员UUID查询
func (r *commonRepo) WhereByMemberUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("member_uuid = ?", uuid)
	}
}

// WhereByNotRevoked 未撤销的出库记录
func (r *commonRepo) WhereByNotRevoked() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("revoke_time = 0")
	}
}

// WhereBySign 根据签名查询
func (r *commonRepo) WhereBySign(sign string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("sign = ?", sign)
	}
}

// WhereByStaffUuid 根据员工UUID查询
func (r *commonRepo) WhereByStaffUuid(staffUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("staff_uuid = ?", staffUuid)
	}
}

// WhereByShiftNo 根据交班编号查询
func (r *commonRepo) WhereByShiftNo(shiftNo string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("shift_no = ?", shiftNo)
	}
}

// WhereByProcessedNot 根据未处理查询
func (r *commonRepo) WhereByProcessedNot() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("processed = ?", constant.MemberPointLogOrBalanceProcessedNot)
	}
}

// WhereByRefundTime 根据退款时间查询
func (r *commonRepo) WhereByRefundTime(refundTime int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("refund_time = ?", refundTime)
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

// WhereByDutyNo 根据班次编号查询
func (r *commonRepo) WhereByDutyNo(dutyNo string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("duty_no = ?", dutyNo)
	}
}

// WhereByShiftLogUuid 根据交班记录UUID查询
func (r *commonRepo) WhereByShiftLogUuid(shiftLogUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("shift_log_uuid = ?", shiftLogUuid)
	}
}

// WhereByAction 根据操作查询
func (r *commonRepo) WhereByAction(action string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("action = ?", action)
	}
}

// WhereByOperatorUuid 根据操作员UUID查询
func (r *commonRepo) WhereByOperatorUuid(operatorUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("operator_uuid = ?", operatorUuid)
	}
}

// WhereByIsVisitor 根据是否访客查询
func (r *commonRepo) WhereByIsVisitor(isVisitor uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_visitor = ?", isVisitor)
	}
}
