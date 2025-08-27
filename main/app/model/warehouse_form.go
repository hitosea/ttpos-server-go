package model

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/pkg/utils"
)

// WarehouseForm 入库单表 `ttpos_warehouse_form`
type WarehouseForm struct {
	BaseModel
	FormNo string `gorm:"column:form_no;type:varchar(255);default:'';comment:编号"`
	Scene  int    `gorm:"column:scene;type:tinyint(2);default:0;comment:交易类型,0-purchase采购入库 1-add添加入库 2-adjust调整入库 3-退菜入库"`
	Num    int    `gorm:"column:num;type:int(11);default:0;comment:数量"`
	Remark string `gorm:"column:remark;type:varchar(255);default:'';comment:备注"`
	Status int    `gorm:"column:status;type:tinyint(1);default:0;comment:状态,0-success已入库 1-canceled已撤销"`
	// 关联uuid
	ProductBomUuid    uint64 `gorm:"column:product_bom_uuid;type:bigint(20) unsigned;default:0;comment:商品BOM表uuid"`
	MaterialUuid      uint64 `gorm:"column:material_uuid;type:bigint(20) unsigned;default:0;comment:材料uuid"`
	PurchaseOrderUuid uint64 `gorm:"column:purchase_order_uuid;type:bigint(20) unsigned;default:0;comment:采购订单uuid"`
	OperatorUuid      uint64 `gorm:"column:operator_uuid;type:bigint(20) unsigned;default:0;comment:操作员uuid"`
	// 时间相关
	RevokeTime int `gorm:"column:revoke_time;type:int(10);default:0;comment:撤销时间(时间戳)"`

	// 关联模型
	WarehouseFormItems []*WarehouseFormItem `gorm:"foreignKey:WarehouseFormUuid;references:Uuid"`
}

func (model *WarehouseForm) SetNil() {
	model.WarehouseFormItems = nil
}

// WarehouseFormItem 入库单明细表 `ttpos_warehouse_form_item`
type WarehouseFormItem struct {
	BaseModel
	Num                  float64 `gorm:"column:num;type:decimal(12,4);not null;default:0;comment:入库数量"`
	Scene                int     `gorm:"column:scene;type:tinyint(2);not null;default:0;comment:场景,0-采购 1-添加入库 2-调整入库 3-退菜入库,这个场景不显示在入库记录页面"`
	AddStock             int     `gorm:"column:add_stock;type:tinyint(1);not null;default:0;comment:是否已经加库存,0-未加库存 1-已加库存。用于判断该入库记录是否已经将对应的货物加库存，若没加库存将在下次检查时加该货物的库存"`
	MaterialUuid         uint64  `gorm:"column:material_uuid;type:bigint unsigned;not null;default:0;comment:材料uuid"`
	ProductBomUuid       uint64  `gorm:"column:product_bom_uuid;type:bigint unsigned;not null;default:0;comment:商品BOM表uuid"`
	WarehouseFormUuid    uint64  `gorm:"column:warehouse_form_uuid;type:bigint unsigned;not null;default:0;comment:入库单uuid"`
	SaleOrderProductUuid uint64  `gorm:"column:sale_order_product_uuid;type:bigint unsigned;not null;default:0;comment:销售订单商品uuid,用于退菜入库"`
	SaleBillUuid         uint64  `gorm:"column:sale_bill_uuid;type:bigint unsigned;not null;default:0;comment:销售账单uuid,用于退菜入库"`

	// 关联模型
	ProductBom *ProductBom `gorm:"foreignKey:ProductBomUuid;references:Uuid"`
	Material   *Material   `gorm:"foreignKey:MaterialUuid;references:Uuid"`
}

func (model *WarehouseFormItem) SetNil() {
	model.ProductBom = nil
	model.Material = nil
}

// IsMaterial 是否是原材料
func (model *WarehouseFormItem) IsMaterial() bool {
	return model.MaterialUuid != 0
}

// IsProductBom 是否是规格商品或小料
func (model *WarehouseFormItem) IsProductBom() bool {
	return model.ProductBomUuid != 0
}

// WarehouseOutForm 出库单表 `ttpos_warehouse_out_form`
type WarehouseOutForm struct {
	BaseModel
	FormNo     string `gorm:"column:form_no;type:varchar(255);default:'';comment:编号"`
	Scene      int    `gorm:"column:scene;type:tinyint(2);default:0;comment:出库类型,0-sales销售出库 1-adjust调整出库 2-loss损耗出库 3-lost丢失出库 4-delete删除出库"`
	Remark     string `gorm:"column:remark;type:varchar(255);default:'';comment:备注"`
	Status     int    `gorm:"column:status;type:tinyint(1);default:0;comment:状态,0-success已出库 1-canceled已撤销"`
	RevokeTime int64  `gorm:"column:revoke_time;type:int(10);default:0;comment:撤销时间(时间戳)"`

	// 关联uuid
	OperatorUuid        uint64 `gorm:"column:operator_uuid;type:bigint(20) unsigned;default:0;comment:操作员uuid"`
	AssociatedOrderUuid uint64 `gorm:"column:associated_order_uuid;type:bigint(20) unsigned;default:0;comment:关联订单uuid"` // sale_bill_uuid

	// 关联模型
	WarehouseOutFormItems []*WarehouseOutFormItem `gorm:"foreignKey:WarehouseOutFormUuid;references:Uuid"`
}

func (model *WarehouseOutForm) SetNil() {
	model.WarehouseOutFormItems = nil
}

// 撤销出库。使用场景：反结账时，撤销出库记录，将库存退还
func (model *WarehouseOutForm) RevokeForm() {
	model.Status = constant.WarehouseOutFormStatusCanceled
	model.RevokeTime = time.Now().Unix()
	for _, item := range model.WarehouseOutFormItems {
		item.RevokeTime = time.Now().Unix()
	}
}

// WarehouseOutFormItem 出库单明细表 `ttpos_warehouse_out_form_item`
type WarehouseOutFormItem struct {
	BaseModel
	Num         float64 `gorm:"column:num;type:decimal(12,4);default:0;comment:数量"`
	Scene       int     `gorm:"column:scene;type:tinyint(2);default:0;comment:场景,0-销售出库 1-adjust调整 2-loss损耗 3-lost丢失 4-delete删除"`
	Status      int     `gorm:"column:status;type:tinyint(1);default:0;comment:状态,0-预出库 1-已出库。预出库时，表示库存扣减但未在出库记录页面显示.已出库时才在出库记录页面显示."`
	ReduceStock int     `gorm:"column:reduce_stock;type:tinyint(1);default:0;comment:是否已经减库存,0-未减库存 1-已减库存。用于判断该出库记录是否已经将对应的货物减库存，若没减库存将在下次检查时减该货物的库存"`
	RevokeTime  int64   `gorm:"column:revoke_time;type:int(10);default:0;comment:撤销时间(时间戳)"`
	// 关联uuid
	WarehouseOutFormUuid uint64 `gorm:"column:warehouse_out_form_uuid;type:bigint(20) unsigned;default:0;comment:出库单uuid"`
	ProductBomUuid       uint64 `gorm:"column:product_bom_uuid;type:bigint(20) unsigned;default:0;comment:商品BOM表uuid, 规格商品或小料"`
	MaterialUuid         uint64 `gorm:"column:material_uuid;type:bigint(20) unsigned;default:0;comment:材料uuid，原材料"`
	SaleOrderProductUuid uint64 `gorm:"column:sale_order_product_uuid;type:bigint(20) unsigned;default:0;comment:销售订单商品uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录"`
	SaleOrderUuid        uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;default:0;comment:销售订单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录"`
	SaleBillUuid         uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单uuid,用于结账完成时判断订单的每个商品是否都已有对应的出库记录"`

	// 关联模型
	ProductBom *ProductBom `gorm:"foreignKey:ProductBomUuid;references:Uuid"` // 出库的规格商品或小料
	Material   *Material   `gorm:"foreignKey:MaterialUuid;references:Uuid"`   // 出库的原材料
}

func (model *WarehouseOutFormItem) SetNil() {

}

// IsMaterial 是否是原材料
func (model *WarehouseOutFormItem) IsMaterial() bool {
	return model.MaterialUuid != 0
}

// IsProductBom 是否是规格商品或小料
func (model *WarehouseOutFormItem) IsProductBom() bool {
	return model.ProductBomUuid != 0
}

type Product struct {
	SaleOrderProductUuid uint64                 `json:"sale_order_product_uuid"` // 销售订单商品uuid
	ProductBomUuid       uint64                 `json:"product_bom_uuid"`        // 规格商品或小料的uuid
	SaleOrderUuid        uint64                 `json:"sale_order_uuid"`         // 销售订单uuid
	Num                  float64                `json:"num"`                     // 数量
	ProductBomMaterials  []*ProductBomMaterials `json:"product_bom_materials"`   // 规格商品或小料的材料
}

type ProductBomMaterials struct {
	MaterialUuid  uint64  `json:"material_uuid"`
	Num           float64 `json:"num"`
	SaleOrderUuid uint64  `json:"sale_order_uuid"` // 销售订单uuid
}

type ErpProductBomMaterials struct {
	ErpCode string  `json:"erp_code"` // 原材料erp编码
	Num     float64 `json:"num"`      // 原材料数量
	Uom     string  `json:"uom"`      // 单位
}

type ProductList []*Product

func (p ProductList) GetProductBomMaterials() []*ProductBomMaterials {
	materials := make([]*ProductBomMaterials, 0)
	for _, product := range p {
		materials = append(materials, product.ProductBomMaterials...)
	}
	return materials
}

// NewWarehouseOutForm 创建出库单.
// 使用场景：
// 1. 送厨时，下单减库存，创建出库单
// 2. 结账时，判断订单的每个商品是否都已有对应的出库记录，如果没有，则创建出库单
func NewWarehouseOutForm(list ProductList, isCheckout bool, saleBillUuid uint64, staffUuid uint64) []*WarehouseOutForm {
	newForm := func() *WarehouseOutForm {
		uuid, _ := utils.GetID()
		form := &WarehouseOutForm{BaseModel: BaseModel{Uuid: uuid}}
		form.FormNo = "CK" + time.Now().Format("20060102150405")
		form.Scene = constant.WarehouseOutFormSceneSales // 销售出库
		form.AssociatedOrderUuid = saleBillUuid
		form.OperatorUuid = staffUuid
		return form
	}

	status := constant.WarehouseOutFormItemStatusPre
	if isCheckout {
		status = constant.WarehouseOutFormItemStatusSuccess
	}

	forms := make([]*WarehouseOutForm, 0)
	// 规格商品或小料出库记录
	for _, item := range list {
		form := newForm()

		items := make([]*WarehouseOutFormItem, 0)
		items = append(items, &WarehouseOutFormItem{
			WarehouseOutFormUuid: form.Uuid,
			ProductBomUuid:       item.ProductBomUuid,
			Num:                  item.Num,
			Scene:                constant.WarehouseOutFormSceneSales, // 销售出库
			Status:               status,
			SaleOrderUuid:        item.SaleOrderUuid,
			SaleOrderProductUuid: item.SaleOrderProductUuid,
			SaleBillUuid:         saleBillUuid,
		})
		form.WarehouseOutFormItems = items
		forms = append(forms, form)
	}

	// 原材料出库记录
	materials := list.GetProductBomMaterials()
	for _, material := range materials {
		form := newForm()

		items := make([]*WarehouseOutFormItem, 0)
		items = append(items, &WarehouseOutFormItem{
			WarehouseOutFormUuid: form.Uuid,
			MaterialUuid:         material.MaterialUuid,
			Num:                  material.Num,
			Scene:                constant.WarehouseOutFormSceneSales, // 销售出库
			Status:               status,
			SaleOrderUuid:        material.SaleOrderUuid,
			SaleBillUuid:         saleBillUuid,
		})
		form.WarehouseOutFormItems = items
		forms = append(forms, form)
	}

	return forms
}

// NewWarehouseForm 创建入库单.
// 使用场景：
// 1. 退菜时，创建入库单
func NewWarehouseForm(list ProductList, saleBillUuid uint64) *WarehouseForm {
	uuid, _ := utils.GetID()
	form := &WarehouseForm{BaseModel: BaseModel{Uuid: uuid}}
	form.FormNo = "RK" + time.Now().Format("20060102150405")
	form.Scene = constant.WarehouseFormSceneReturn // 退菜入库

	items := make([]*WarehouseFormItem, 0)
	// 规格商品或小料出库记录
	for _, item := range list {
		items = append(items, &WarehouseFormItem{
			Num:                  float64(item.Num),
			Scene:                constant.WarehouseFormSceneReturn, // 退菜入库
			AddStock:             constant.WarehouseOutFormItemReduceStockNotProcessed,
			ProductBomUuid:       item.ProductBomUuid,
			WarehouseFormUuid:    form.Uuid,
			SaleOrderProductUuid: item.SaleOrderProductUuid,
			SaleBillUuid:         saleBillUuid,
		})
	}

	// 原材料出库记录
	materials := list.GetProductBomMaterials()
	for _, material := range materials {
		items = append(items, &WarehouseFormItem{
			Num:               material.Num,
			Scene:             constant.WarehouseFormSceneReturn, // 退菜入库
			AddStock:          constant.WarehouseOutFormItemReduceStockNotProcessed,
			MaterialUuid:      material.MaterialUuid,
			WarehouseFormUuid: form.Uuid,
			SaleBillUuid:      saleBillUuid,
		})
	}
	form.WarehouseFormItems = items
	return form
}
