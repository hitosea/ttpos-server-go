package model

// 商品必选商品计划表 `ttpos_product_must_plan`
type ProductMustPlan struct {
	BaseModel
	Name         string `gorm:"column:name;type:varchar(255);not null;default:'';comment:'方案名称'"`
	UseChannel   string `gorm:"column:use_channel;type:varchar(255);not null;default:'';comment:'使用渠道 10-点餐方式 20-桌台方式'"`
	MustType     uint   `gorm:"column:must_type;default:1;comment:'必点类型 1-每人必点1份 2-每笔订单必点1份'"`
	MustRule     uint   `gorm:"column:must_rule;default:1;comment:'必点规则 1-固定商品 2-可选商品'"`
	Status       uint   `gorm:"column:status;default:1;comment:'状态,1-开启 0-关闭'"`
	AutoCart     uint   `gorm:"column:auto_cart;default:1;comment:'自动加入购物车 1-是 0-否'"`
	AutoChange   uint   `gorm:"column:auto_change;default:1;comment:'顾客可修改必点数量 1-是 0-否'"`
	AutoCheck    uint   `gorm:"column:auto_check;default:1;comment:'下单时检查必点商品 1-是 0-否'"`
	AutoCheckout uint   `gorm:"column:auto_checkout;default:1;comment:'结账时检查必点商品 1-是 0-否'"`

	ProductMustPlanItem []ProductMustPlanItem `gorm:"foreignKey:ProductMustPlanUuid;references:Uuid"`
}

// 商品必选商品计划区域表 `ttpos_product_must_plan_region`
type ProductMustPlanRegion struct {
	BaseModel
	ProductMustPlanUuid uint64 `gorm:"column:product_must_plan_uuid;type:bigint unsigned;not null;default:0;comment:'商品必选商品计划ID'"`
	DeskRegionUuid      uint64 `gorm:"column:desk_region_uuid;type:bigint unsigned;not null;default:0;comment:'桌台区域ID'"`

	ProductMustPlan ProductMustPlan `gorm:"foreignKey:ProductMustPlanUuid;references:Uuid"`
}

// 商品必点计划商品表 `ttpos_product_must_plan_item`
type ProductMustPlanItem struct {
	BaseModel
	ProductMustPlanUuid uint64 `gorm:"column:product_must_plan_uuid;type:bigint unsigned;not null;default:0;comment:'商品必选商品计划ID'"`
	ProductPackageUuid  uint64 `gorm:"column:product_package_uuid;type:bigint unsigned;not null;default:0;comment:'商品包ID'"`

	ProductMustPlan ProductMustPlan `gorm:"foreignKey:ProductMustPlanUuid;references:Uuid"`
}
