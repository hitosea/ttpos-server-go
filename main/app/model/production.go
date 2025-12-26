package model

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/modules/takeout/domain/model"
)

// ProductionOrder 生产单 `ttpos_production_order`
type ProductionOrder struct {
	BaseModel
	DeskUuid         uint64 `gorm:"column:desk_uuid;type:bigint(20) unsigned;default:0;comment:桌台ID;NOT NULL" json:"desk_uuid"`
	SaleOrderUuid    uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;default:0;comment:销售订单ID;NOT NULL" json:"sale_order_uuid"`
	SaleBillUuid     uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单ID;NOT NULL" json:"sale_bill_uuid"`
	TakeoutOrderUuid uint64 `gorm:"column:takeout_order_uuid;type:bigint(20) unsigned;default:0;comment:外卖订单UUID（关联 ttpos_takeout_order.uuid）;NOT NULL" json:"takeout_order_uuid"`
	Source           string `gorm:"column:source;type:varchar(255);comment:操作来源 shop-商家、cashier-收银机、tablet-平板端、kitchen-厨显端、assistant-点餐助手、h5-H5、grab、lineman" json:"source"`

	ProductionOrderProducts []*ProductionOrderProduct `gorm:"foreignKey:ProductionOrderUuid;references:Uuid" json:"production_order_products"`
}

// 判断是否为外卖订单
func (p *ProductionOrder) IsTakeoutOrder() bool {
	return p.TakeoutOrderUuid > 0
}

// 获取来源平台
func (p *ProductionOrder) GetPlatform() string {
	if p.TakeoutOrderUuid > 0 {
		return p.Source // grab, lineman
	}
	return ""
}

// ProductionOrderProduct 生产单商品 `ttpos_production_order_product`
type ProductionOrderProduct struct {
	BaseModel
	Name                  string  `gorm:"column:name;type:varchar(255);comment:名称;NOT NULL" json:"name"`
	Num                   float64 `gorm:"column:num;type:decimal(12,2);default:0;comment:商品数量;NOT NULL" json:"num"`
	InitNum               float64 `gorm:"column:init_num;type:decimal(12,2);default:0;comment:送厨时商品数量;NOT NULL" json:"init_num"`
	FlavorName            string  `gorm:"column:flavor_name;type:text;comment:规格名称,不随后台改变;" json:"flavor_name"`
	ProductBomUuid        uint64  `gorm:"column:product_bom_uuid;type:bigint(20) unsigned;default:0;comment:商品BOM ID;NOT NULL" json:"product_bom_uuid"`
	ProductAttributeNames string  `gorm:"column:product_attribute_names;type:varchar(255);comment:商品属性名称,多个属性名用逗号分隔,不随后台改变;NOT NULL" json:"product_attribute_names"`
	ProductSaucesNames    string  `gorm:"column:product_sauces_names;type:varchar(255);comment:商品加料名称,多个加料名用逗号分隔,不随后台改变;NOT NULL" json:"product_sauces_names"`
	Status                int     `gorm:"column:status;type:tinyint(1);default:0;comment:状态, 0-待制作 1-制作中 2-已完成 3-已退菜;NOT NULL" json:"status"`
	Remark                string  `gorm:"column:remark;type:varchar(255);comment:商品备注;NOT NULL" json:"remark"`
	HasMaterial           int     `gorm:"column:has_material;type:tinyint(1);default:0;comment:是否无原料, 0-无原料,商品没有关联原料 1-有原料;NOT NULL" json:"has_material"`
	SaleBillUuid          uint64  `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单ID;NOT NULL" json:"sale_bill_uuid"`
	ProductPackageUuid    uint64  `gorm:"column:product_package_uuid;type:bigint(20) unsigned;default:0;comment:商品包ID;NOT NULL" json:"product_package_uuid"`
	SaleOrderProductUuid  uint64  `gorm:"column:sale_order_product_uuid;type:bigint(20) unsigned;default:0;comment:销售订单商品ID;NOT NULL" json:"sale_order_product_uuid"`
	ProductionOrderUuid   uint64  `gorm:"column:production_order_uuid;type:bigint(20) unsigned;default:0;comment:生产订单ID;NOT NULL" json:"production_order_uuid"`
	TakeoutOrderUuid      uint64  `gorm:"column:takeout_order_uuid;type:bigint(20) unsigned;default:0;comment:外卖订单UUID（关联 ttpos_takeout_order.uuid）;NOT NULL" json:"takeout_order_uuid"`
	TakeoutOrderItemUuid  uint64  `gorm:"column:takeout_order_item_uuid;type:bigint(20) unsigned;default:0;comment:外卖订单商品UUID（关联 ttpos_takeout_order_item.uuid）;NOT NULL" json:"takeout_order_item_uuid"`
	FirstCategoryUuid     uint64  `gorm:"column:first_category_uuid;type:bigint(20) unsigned;default:0;comment:一级分类ID;NOT NULL" json:"first_category_uuid"`
	FinishedTime          uint    `gorm:"column:finished_time;type:int(10) unsigned;default:0;comment:完成时间(时间戳);NOT NULL" json:"finished_time"`
	MakeStatus            int     `gorm:"column:make_status;type:tinyint(1);default:0;comment:制作状态 0-默认，未制作完成，1-已制作完成，2-已恢复到制作中;NOT NULL" json:"make_status"`
	MadeTime              int64   `gorm:"column:made_time;type:int(10) unsigned;default:0;comment:制作完成时间(时间戳);NOT NULL" json:"made_time"`

	// 分批相关
	BatchTagUuid uint64 `gorm:"column:batch_tag_uuid;type:bigint(20) unsigned;default:0;comment:分批类型UUID;NOT NULL" json:"batch_tag_uuid"`
	BatchTime    int64  `gorm:"column:batch_time;type:int(10) unsigned;default:0;comment:分批时间(时间戳)，表示该商品实际送厨到厨房的时间;NOT NULL" json:"batch_time"`
	IsBatch      uint8  `gorm:"column:is_batch;type:tinyint(1);default:0;comment:是否是分批商品, 0-否 1-是;NOT NULL" json:"is_batch"`
	// 效率分析相关
	MakeDuration    int64   `gorm:"column:make_duration;type:int(10) unsigned;default:0;comment:制作时长(秒);NOT NULL" json:"make_duration"`
	SendDuration    int64   `gorm:"column:send_duration;type:int(10) unsigned;default:0;comment:传菜时长(秒);NOT NULL" json:"send_duration"`
	AllDuration     int64   `gorm:"column:all_duration;type:int(10) unsigned;default:0;comment:总时长(秒);NOT NULL" json:"all_duration"`
	AvgMakeDuration float64 `gorm:"column:avg_make_duration;type:decimal(22,4);default:0.00;comment:制作时长平均值(秒);NOT NULL" json:"avg_make_duration"`
	AvgSendDuration float64 `gorm:"column:avg_send_duration;type:decimal(22,4);default:0.00;comment:传菜时长平均值(秒);NOT NULL" json:"avg_send_duration"`
	AvgAllDuration  float64 `gorm:"column:avg_all_duration;type:decimal(22,4);default:0.00;comment:总时长平均值(秒);NOT NULL" json:"avg_all_duration"`

	ProductionOrderMaterials []*ProductionOrderMaterial `gorm:"foreignKey:ProductionOrderProductUuid;references:Uuid" json:"production_order_materials"`
	ProductionOrder          ProductionOrder            `gorm:"foreignKey:ProductionOrderUuid;references:Uuid" json:"production_order"`
	SaleOrderProduct         SaleOrderProduct           `gorm:"foreignKey:SaleOrderProductUuid;references:Uuid" json:"sale_order_product"`
	SaleBill                 SaleBill                   `gorm:"foreignKey:SaleBillUuid;references:Uuid" json:"sale_bill"`
	ProductCategory          ProductCategory            `gorm:"foreignKey:FirstCategoryUuid;references:Uuid" json:"product_category"`
	BatchTag                 *BatchTag                  `gorm:"foreignKey:BatchTagUuid;references:Uuid" json:"batch_tag"`
	TakeoutOrder             *model.TakeoutOrder        `gorm:"foreignKey:Uuid;references:TakeoutOrderUuid" json:"takeout_order"`
	TakeoutOrderItem         *model.TakeoutOrderItem    `gorm:"foreignKey:TakeoutOrderItemUuid;references:Uuid" json:"takeout_order_item"`
}

// 获取商品的送厨时间
func (p *ProductionOrderProduct) GetCreateTime() int64 {
	if p.IsBatchBool() {
		return p.BatchTime // 如果商品是分批商品,则送厨时间是分批送厨的时间
	}
	return p.CreateTime
}

// 只能厨显模式下,获取商品的制作完成时间
func (p *ProductionOrderProduct) GetMadeTime() int64 {
	if p.MadeTime == 0 { // 如果商品没有制作完成时间,则使用送厨时间
		return p.GetCreateTime()
	}
	return p.MadeTime
}

// 是否是分批商品
func (p *ProductionOrderProduct) IsBatchBool() bool {
	return p.IsBatch == 1
}

// 是否处于预送厨阶段
func (p *ProductionOrderProduct) IsPreCooking() bool {
	return p.BatchTagUuid == 0
}

// 获取生产单商品的打包状态：0-堂食、1-打包
func (p *ProductionOrderProduct) GetWrapStatus() uint {
	// 如果商品是点餐订单的商品，则根据sale_bill是否打包来判断商品是否打包
	if p.SaleBill.IsInstantBill() {
		return p.SaleBill.DiningMethod
	}
	// 如果商品是桌台订单的商品，则根据订单商品的打包状态来判断商品是否打包
	if p.SaleBill.IsDeskBill() && p.SaleOrderProduct.IsWrapProduct() {
		return constant.SaleBillDiningMethodTakeout // 打包
	}
	return constant.SaleBillDiningMethodDineIn // 堂食
}

// IsTakeoutOrder 是否是外卖订单
func (p *ProductionOrderProduct) IsTakeoutOrder() bool {
	return p.TakeoutOrderUuid > 0
}

// ProductionOrderMaterial 生产单原料 `ttpos_production_order_material`
type ProductionOrderMaterial struct {
	BaseModel
	Name                       string `gorm:"column:name;type:varchar(255);comment:原料名称,不随后台改变;NOT NULL" json:"name"`
	MaterialUuid               uint64 `gorm:"column:material_uuid;type:bigint(20) unsigned;default:0;comment:原料ID;NOT NULL" json:"material_uuid"`
	Num                        int    `gorm:"column:num;type:int(11);default:0;comment:原料数量;NOT NULL" json:"num"`
	IsProductBom               int    `gorm:"column:is_product_bom;type:tinyint(1);default:0;comment:是否为商品BOM, 0-否 1-是, 没有原料的规格商品为1;NOT NULL" json:"is_product_bom"`
	Unit                       string `gorm:"column:unit;type:varchar(255);comment:单位,不随后台改变;NOT NULL" json:"unit"`
	ProductionOrderProductUuid uint64 `gorm:"column:production_order_product_uuid;type:bigint(20) unsigned;default:0;comment:生产订单商品ID;NOT NULL" json:"production_order_product_uuid"`
	SaleOrderProductUuid       uint64 `gorm:"column:sale_order_product_uuid;type:bigint(20) unsigned;default:0;comment:销售订单商品ID;NOT NULL" json:"sale_order_product_uuid"`
	TakeoutOrderItemUuid       uint64 `gorm:"column:takeout_order_item_uuid;type:bigint(20) unsigned;default:0;comment:外卖订单商品UUID（关联 ttpos_takeout_order_item.uuid）;NOT NULL" json:"takeout_order_item_uuid"`
}
