package req

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// MaterialCategoryAddReq 创建物品类别请求
type MaterialCategoryAddReq struct {
	LocaleName dto.LocaleResponse `json:"locale_name"` // 物品类别名称
}

func (r *MaterialCategoryAddReq) Validate() error {
	if r.LocaleName.IsNull() {
		return errors.WithMessage(errors.New("名称不能为空"))
	}
	return nil
}

// MaterialCategoryListReq 获取物品类别列表请求
type MaterialCategoryListReq struct {
}

// MaterialListReq 物品列表查询
type MaterialListReq struct {
	dto.PageReq            // 分页参数
	Keyword       string   `form:"keyword" json:"keyword"`               // 关键字
	Status        int      `form:"status" json:"status"`                 // 状态，0-全部 1-启用 2-停用
	CategoryUuids []uint64 `form:"category_uuids" json:"category_uuids"` // 分类UUID列表,多选时
}

func (r *MaterialListReq) GetCategoryUuids() []uint64 {
	list := make([]uint64, 0)
	for i := range r.CategoryUuids {
		if r.CategoryUuids[i] != 0 {
			list = append(list, r.CategoryUuids[i])
		}
	}
	return list
}

// MaterialDetailReq 物品详情查询
type MaterialDetailReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 物品UUID
}

// MaterialAddReq 添加物品请求
type MaterialAddReq struct {
	LocaleName       dto.LocaleResponse `json:"locale_name"`        // 物品名称
	CategoryUuid     uint64             `json:"category_uuid"`      // 分类UUID
	Status           int                `json:"status"`             // 状态，1-启用 0-停用
	Valuation        float64            `json:"valuation"`          // 估值率
	InitStock        float64            `json:"init_stock"`         // 期初库存
	BarcodeValue     string             `json:"barcode_value"`      // 条形码值
	UnitUuid         uint64             `json:"unit_uuid"`          // 基准单位UUID
	UnitList         []MaterialUnitReq  `json:"unit_list"`          // 单位列表
	PurchaseUnitUuid uint64             `json:"purchase_unit_uuid"` // 采购单位UUID
	CostUnitUuid     uint64             `json:"cost_unit_uuid"`     // 成本单位UUID
}

func (r *MaterialAddReq) Validate() error {
	if r.LocaleName.IsNull() {
		return errors.WithMessage(errors.New("名称不能为空"))
	}
	if r.CategoryUuid == 0 {
		return errors.WithMessage(errors.New("分类不能为空"))
	}
	if r.UnitUuid == 0 {
		return errors.WithMessage(errors.New("基准单位不能为空"))
	}
	if r.PurchaseUnitUuid == 0 {
		return errors.WithMessage(errors.New("采购单位不能为空"))
	}
	if r.CostUnitUuid == 0 {
		return errors.WithMessage(errors.New("成本单位不能为空"))
	}
	if r.BarcodeValue != "" {
		// 检查条形码长度（条形码长度为1-13位）
		if len(r.BarcodeValue) < 1 || len(r.BarcodeValue) > 13 {
			return errors.WithMessage(errors.New("条形码长度应为1-13位"))
		}
		// 检查是否只包含数字
		for _, char := range r.BarcodeValue {
			if char < '0' || char > '9' {
				return errors.WithMessage(errors.New("条形码只能包含数字"))
			}
		}
	}
	// 创建的时候期初库存跟估值率要大于0
	if r.Valuation <= 0 {
		return errors.WithMessage(errors.New("估值率需大于零"))
	}
	if r.InitStock <= 0 {
		return errors.WithMessage(errors.New("期初库存需大于零"))
	}
	return nil
}

type MaterialAddErpReq struct {
	ItemCode      string           `json:"item_code" binding:"required"`      // 物品编码, 如果为空，则为新增；如果非空，则为编辑
	ItemName      string           `json:"item_name" binding:"required"`      // 物品名称, 英文
	StockUom      string           `json:"stock_uom" binding:"required"`      // 基准库存单位, 英文
	Disabled      bool             `json:"disabled" binding:"required"`       // 是否禁用
	BarcodeValue  string           `json:"barcode_value" binding:"required"`  // 条形码值
	ValuationRate float64          `json:"valuation_rate" binding:"required"` // 估值率
	OpeningStock  float64          `json:"opening_stock" binding:"required"`  // 期初库存
	Uoms          []MaterialUomReq `json:"uoms" binding:"required,dive"`      // 单位列表
}

type ProductAddErpReq struct {
	ItemName          string `json:"item_name" binding:"required"`          // 商品名称, 英文
	StockUom          string `json:"stock_uom" binding:"required"`          // 商品单位, 英文
	ItemCode          string `json:"item_code" binding:"required"`          // 商品编码，如果为空，则为新增；如果非空，则为编辑
	TemplateItemCode  string `json:"template_item_code" binding:"required"` // 模版商品编码
	ItemSpecification string `json:"item_specification" binding:"required"` // 商品规格
	BarcodeValue      string `json:"barcode_value" binding:"required"`      // 条形码值
}

type PackageAddErpReq struct {
	ItemName string `json:"item_name" binding:"required"` // 套餐名称, 英文
	StockUom string `json:"stock_uom" binding:"required"` // 套餐单位, 英文
	ItemCode string `json:"item_code" binding:"required"` // 套餐编码，如果为空，则为新增；如果非空，则为编辑
}

type MaterialUomReq struct {
	Uom            string  `json:"uom" binding:"required"`                   // 单位, 英文
	ConversionRate float64 `json:"conversion_rate" binding:"required,min=0"` // 转换率
}

// MaterialUnitReq 物品单位请求
type MaterialUnitReq struct {
	Uuid           uint64  `json:"uuid" binding:"required"`                  // 单位UUID
	ConversionRate float64 `json:"conversion_rate" binding:"required,min=0"` // 转换率
}

// MaterialEditReq 编辑物品请求
type MaterialEditReq struct {
	Uuid             uint64             `json:"uuid"`               // 物品UUID
	LocaleName       dto.LocaleResponse `json:"locale_name"`        // 物品名称
	CategoryUuid     uint64             `json:"category_uuid"`      // 分类UUID
	Status           int                `json:"status"`             // 状态，1-启用 0-停用
	Valuation        float64            `json:"valuation"`          // 估值率
	BarcodeValue     string             `json:"barcode_value"`      // 条形码值
	UnitList         []MaterialUnitReq  `json:"unit_list"`          // 单位列表,新增的非基准单位
	PurchaseUnitUuid uint64             `json:"purchase_unit_uuid"` // 采购单位UUID
	CostUnitUuid     uint64             `json:"cost_unit_uuid"`     // 成本单位UUID
}

func (r *MaterialEditReq) Validate() error {
	if r.Uuid == 0 {
		return errors.WithMessage(errors.New("物品ID不能为空"))
	}
	if r.LocaleName.IsNull() {
		return errors.WithMessage(errors.New("名称不能为空"))
	}
	if r.CategoryUuid == 0 {
		return errors.WithMessage(errors.New("分类不能为空"))
	}
	if r.PurchaseUnitUuid == 0 {
		return errors.WithMessage(errors.New("采购单位不能为空"))
	}
	if r.CostUnitUuid == 0 {
		return errors.WithMessage(errors.New("成本单位不能为空"))
	}
	if r.BarcodeValue != "" {
		// 检查条形码长度（条形码长度为1-13位）
		if len(r.BarcodeValue) < 1 || len(r.BarcodeValue) > 13 {
			return errors.WithMessage(errors.New("条形码长度应为1-13位"))
		}
		// 检查是否只包含数字
		for _, char := range r.BarcodeValue {
			if char < '0' || char > '9' {
				return errors.WithMessage(errors.New("条形码只能包含数字"))
			}
		}
	}
	return nil
}

// MaterialDeleteReq 删除物品请求
type MaterialDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 物品UUID
}

// MaterialStatusReq 修改物品状态请求
type MaterialStatusReq struct {
	Uuids  []uint64 `json:"uuids" binding:"required"` // 物品UUID
	Status int      `json:"status"`                   // 状态，1-启用 0-停用
}

// MaterialUnitListReq 获取物品单位列表请求
type MaterialUnitListReq struct {
	Uuid uint64 `form:"uuid" json:"uuid" binding:"required"` // 物品UUID
}

// ProductBomCardAddReq 添加成本卡请求
type ProductBomCardAddReq struct {
	RelatedUuid uint64                        `json:"related_uuid" binding:"required"`   // 关联UUID,给规格商品或加料绑定成本卡。规格商品时，关联UUID为规格商品UUID；加料时，关联UUID为加料UUID
	RelatedType uint8                         `json:"related_type" binding:"required"`   // 关联类型,1-规格商品 2-加料
	Num         int                           `json:"num" binding:"required"`            // 加工份数
	Materials   ProductBomCardMaterialListReq `json:"materials" binding:"required,dive"` // 材料列表
}

type ProductBomCardMaterialListReq struct {
	List []ProductBomCardMaterialReq `json:"list" binding:"required,dive"` // 材料列表
}

type ProductBomCardMaterialReq struct {
	MaterialUuid uint64  `json:"material_uuid" binding:"required"` // 材料UUID
	Num          float64 `json:"num" binding:"required"`           // 数量
	UnitUuid     uint64  `json:"unit_uuid" binding:"required"`     // 成本单位UUID
}

// ProductBomCardDetailReq 规格商品成本卡详情请求
type ProductBomCardDetailReq struct {
	Uuid uint64 `form:"uuid" binding:"required"` // 成本卡UUID product_bom_card_uuid
}

// ProductBomCardUnlinkReq 解除成本卡关联请求
type ProductBomCardUnlinkReq struct {
	RelatedUuid uint64 `json:"related_uuid" binding:"required"` // 关联UUID,给规格商品或加料绑定成本卡。规格商品时，关联UUID为规格商品UUID；加料时，关联UUID为加料UUID
	RelatedType uint8  `json:"related_type" binding:"required"` // 关联类型,1-规格商品 2-加料
}

// ProductBomCardCopyReq 复制成本卡请求
type ProductBomCardCopyReq struct {
	RelatedUuid        uint64 `json:"related_uuid" binding:"required"`          // 关联UUID,给规格商品或加料绑定成本卡。规格商品时，关联UUID为规格商品UUID；加料时，关联UUID为加料UUID
	RelatedType        uint8  `json:"related_type" binding:"required"`          // 关联类型,1-规格商品 2-加料
	ProductBomCardUuid uint64 `json:"product_bom_card_uuid" binding:"required"` // 成本卡UUID
}

// ProductBomCardImportReq 从菜品导入成本卡请求
type ProductBomCardImportReq struct {
	RelatedUuid uint64 `json:"related_uuid" binding:"required"` // 关联UUID,给规格商品绑定成本卡。规格商品时，关联UUID为规格商品UUID；
	// 创建物品
	MaterialAddReq
	// 创建成本卡
	Num float64 `json:"num" binding:"required"` // 净耗量
}

// MaterialImportListItemReq 导入物品项请求
type MaterialImportListItemReq struct {
	LocaleName   dto.LocaleResponse `json:"locale_name"`   // 名称
	CategoryName string             `json:"category_name"` // 分类名称
	BarcodeValue string             `json:"barcode_value"` // 条形码值
	Status       int                `json:"status"`        // 状态，1-启用 0-停用
	UnitName     string             `json:"unit_name"`     // 基准单位名称
	Valuation    float64            `json:"valuation"`     // 估值率
	InitStock    float64            `json:"init_stock"`    // 期初库存
	Row          int                `json:"row"`           // 行号
}

// MaterialImportListReq 导入物品列表请求
type MaterialImportListReq struct {
	List []MaterialImportListItemReq `json:"list" binding:"required,dive"` // 物品列表
}

// ProductImportItemReq 导入商品项请求
type MaterialImportItemReq struct {
	LocaleName   dto.LocaleResponse `json:"locale_name"`   // 物品名称
	CategoryUuid uint64             `json:"category_uuid"` // 分类UUID
	UnitUuid     uint64             `json:"unit_uuid"`     // 单位UUID
	Valuation    float64            `json:"valuation"`     // 估值率
	InitStock    float64            `json:"init_stock"`    // 期初库存
	BarcodeValue string             `json:"barcode_value"` // 条形码值
	Status       int                `json:"status"`        // 状态，1-启用 0-停用
	Row          int                `json:"row"`           // excel表的行编号
}

// ProductImportReq 导入商品请求
type MaterialImportReq struct {
	List []MaterialImportItemReq `json:"list" binding:"required,dive"` // 商品列表
}

// DeleteProductErpReq 删除商品请求
type DeleteProductErpReq struct {
	Items []DeleteProductErpItemReq `json:"items"` // 商品列表
}

type DeleteProductErpItemReq struct {
	ItemCode string `json:"item_code" binding:"required"` // 商品编码
	ItemName string `json:"item_name" binding:"required"` // 商品名称, 英文.没办法，接口要传
	StockUom string `json:"stock_uom" binding:"required"` // 商品单位, 英文.没办法，接口要传
}
