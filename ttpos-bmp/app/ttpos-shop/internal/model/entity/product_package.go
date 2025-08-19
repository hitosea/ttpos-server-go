// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProductPackage is the golang structure for table product_package.
type ProductPackage struct {
	Id                    uint    `json:"id"                    orm:"id"                       description:"自增ID"`                    // 自增ID
	Uuid                  uint64  `json:"uuid"                  orm:"uuid"                     description:"商品包ID"`                   // 商品包ID
	Name                  string  `json:"name"                  orm:"name"                     description:"商品包名称"`                   // 商品包名称
	MultiLanguageNameUuid uint64  `json:"multiLanguageNameUuid" orm:"multi_language_name_uuid" description:"多语言名称ID"`                 // 多语言名称ID
	ImageName             string  `json:"imageName"             orm:"image_name"               description:"图片名称"`                    // 图片名称
	ImageFileUuid         uint64  `json:"imageFileUuid"         orm:"image_file_uuid"          description:"图片ID"`                    // 图片ID
	DeductStockType       int     `json:"deductStockType"       orm:"deduct_stock_type"        description:"库存计算方法, 0-付款减库存 1-下单减库存"` // 库存计算方法, 0-付款减库存 1-下单减库存
	UnitUuid              uint64  `json:"unitUuid"              orm:"unit_uuid"                description:"单位UUID"`                  // 单位UUID
	DineTaxUuid           uint64  `json:"dineTaxUuid"           orm:"dine_tax_uuid"            description:"堂食税UUID"`                 // 堂食税UUID
	CategoryUuid          uint64  `json:"categoryUuid"          orm:"category_uuid"            description:"类别UUID"`                  // 类别UUID
	SpecialCategoryUuid   uint64  `json:"specialCategoryUuid"   orm:"special_category_uuid"    description:"特殊类别UUID"`                // 特殊类别UUID
	TakeoutTaxUuid        uint64  `json:"takeoutTaxUuid"        orm:"takeout_tax_uuid"         description:"外卖税UUID"`                 // 外卖税UUID
	PrinterTagUuid        uint64  `json:"printerTagUuid"        orm:"printer_tag_uuid"         description:"打印机标签UUID"`               // 打印机标签UUID
	SupplierUuid          uint64  `json:"supplierUuid"          orm:"supplier_uuid"            description:"供应商UUID"`                 // 供应商UUID
	Status                int     `json:"status"                orm:"status"                   description:"状态,0-下架 1-上架"`            // 状态,0-下架 1-上架
	IsShowCashier         int     `json:"isShowCashier"         orm:"is_show_cashier"          description:"是否在收银设备显示, 0-否 1-是"`      // 是否在收银设备显示, 0-否 1-是
	IsShowTablet          int     `json:"isShowTablet"          orm:"is_show_tablet"           description:"是否在平板设备显示, 0-否 1-是"`      // 是否在平板设备显示, 0-否 1-是
	IsShowKitchen         int     `json:"isShowKitchen"         orm:"is_show_kitchen"          description:"是否在厨房设备显示, 0-否 1-是"`      // 是否在厨房设备显示, 0-否 1-是
	IsShowAssistant       int     `json:"isShowAssistant"       orm:"is_show_assistant"        description:"是否在助手设备显示, 0-否 1-是"`      // 是否在助手设备显示, 0-否 1-是
	IsShowH5              int     `json:"isShowH5"              orm:"is_show_h5"               description:"是否在H5设备显示, 0-否 1-是"`      // 是否在H5设备显示, 0-否 1-是
	Sort                  int     `json:"sort"                  orm:"sort"                     description:"排序"`                      // 排序
	LimitNum              int     `json:"limitNum"              orm:"limit_num"                description:"限购数量"`                    // 限购数量
	SauceRequired         int     `json:"sauceRequired"         orm:"sauce_required"           description:"是否必选小料, 0-否 1-是"`         // 是否必选小料, 0-否 1-是
	SauceMaxSelection     int     `json:"sauceMaxSelection"     orm:"sauce_max_selection"      description:"小料最大选择数量"`                // 小料最大选择数量
	Describe              string  `json:"describe"              orm:"describe"                 description:"卖点描述"`                    // 卖点描述
	OpenDiscount          int     `json:"openDiscount"          orm:"open_discount"            description:"是否开启会员折扣, 0-否 1-是"`       // 是否开启会员折扣, 0-否 1-是
	ActualSaleNum         float64 `json:"actualSaleNum"         orm:"actual_sale_num"          description:"实际销量。每次卖出时,实际销量增加"`       // 实际销量。每次卖出时,实际销量增加
	CreateTime            uint    `json:"createTime"            orm:"create_time"              description:"创建时间(时间戳)"`               // 创建时间(时间戳)
	UpdateTime            uint    `json:"updateTime"            orm:"update_time"              description:"更新时间(时间戳)"`               // 更新时间(时间戳)
	DeleteTime            uint    `json:"deleteTime"            orm:"delete_time"              description:"删除时间(时间戳)"`               // 删除时间(时间戳)
}
