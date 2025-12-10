package req

// TakeoutMenuExportReq 外卖菜单导出请求
type TakeoutMenuExportReq struct {
	Platform       string   `json:"platform" binding:"required"` // 平台名称：grab, lineman 等
	CompanyUuid    uint64   `json:"companyUuid"`                 // 公司 UUID（可选，默认当前公司）
	CategoryIDs    []uint64 `json:"categoryIds"`                 // 分类 ID 列表（可选，为空则导出所有）
	SellingTimeIDs []uint64 `json:"sellingTimeIds"`              // 售卖时段 ID 列表（可选）
}

// TakeoutMenuImportReq 外卖菜单导入请求
type TakeoutMenuImportReq struct {
	Platform          string      `json:"platform" binding:"required"` // 平台名称
	CompanyUuid       uint64      `json:"companyUuid"`                 // 公司 UUID（可选，默认当前公司）
	MenuData          interface{} `json:"menuData" binding:"required"` // 平台菜单 JSON 数据
	SyncMode          string      `json:"syncMode"`                    // 同步模式：full / incremental
	OverwriteExisting bool        `json:"overwriteExisting"`           // 是否覆盖已存在的数据
}

// GrabMenuImportItem 逐项导入的占位（可扩展）
type GrabMenuImportItem struct {
	// TODO: 如需分项导入，可在此定义
}

// TakeoutMenuPreviewReq 外卖菜单预览请求
type TakeoutMenuPreviewReq struct {
	Platform    string `form:"platform" binding:"required"` // 平台名称
	CompanyUuid uint64 `form:"companyUuid"`                 // 公司 UUID（可选，默认当前公司）
}

// GrabProductImportItem 单个 Grab 商品映射项
type GrabProductImportItem struct {
	GrabProductId       string  `json:"grabProductId" binding:"required"` // Grab 商品唯一ID
	ProductPackageUuid  uint64  `json:"productPackageUuid"`               // 选择的店内商品UUID，未选则自动创建
	CategoryUuid        uint64  `json:"categoryUuid"`                     // 选择的店内分类，未选则自动创建
	SkuName             string  `json:"skuName"`                          // 创建店内商品时使用
	SkuPrice            float64 `json:"skuPrice"`                         // 价格（单位：主货币）
	SkuUnitUuid         uint64  `json:"skuUnitUuid"`                      // 单位UUID，未选则自动创建
	SkuUnitName         string  `json:"skuUnitName"`                      // 单位名称，用于自动创建
	AttributeGroupName  string  `json:"attributeGroupName"`               // 属性组名称（可选，用于自动创建）
	AttributeValueName  string  `json:"attributeValueName"`               // 属性值名称（可选，用于自动创建）
	SellingTimeUuid     uint64  `json:"sellingTimeUuid"`                  // 可选，售卖时段
	NeedTranslateFields bool    `json:"needTranslateFields"`              // 是否需要翻译（默认 true）
}

// GrabProductImportReq Grab 商品导入请求
type GrabProductImportReq struct {
	CompanyUuid uint64                  `json:"companyUuid"`              // 公司 UUID（可选，默认当前公司）
	Items       []GrabProductImportItem `json:"items" binding:"required"` // 映射项列表
	Overwrite   bool                    `json:"overwrite"`                // 是否覆盖已有映射
}

// GrabBindingLinkReq 获取 Grab 绑定链接请求
type GrabBindingLinkReq struct {
	CompanyUuid uint64 `json:"company_uuid"` // 公司 UUID（可选，默认当前公司）
}

// GrabBindingStatusReq 检查 Grab 绑定状态请求
type GrabBindingStatusReq struct {
	CompanyUuid uint64 `json:"company_uuid"` // 公司 UUID（可选，默认当前公司）
}

// GrabMenuReq 获取 Grab 菜单请求
type GrabMenuReq struct {
	CompanyUuid uint64 `json:"company_uuid"` // 公司 UUID（可选，默认当前公司）
}
