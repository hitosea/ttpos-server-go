package req

// TakeoutMenuExportReq 外卖菜单导出请求
type TakeoutMenuExportReq struct {
	Platform    string `json:"platform" binding:"required"` // 平台名称：Grab, Lineman 等
	CompanyUuid uint64 `json:"companyUuid"`                 // 公司 UUID（可选，默认当前公司）
}

// TakeoutMenuImportReq 外卖菜单导入请求
type TakeoutMenuImportReq struct {
	MenuData interface{} `json:"menuData" binding:"required"` // 平台菜单 JSON 数据
}

// TakeoutMenuPreviewReq 外卖菜单预览请求
type TakeoutMenuPreviewReq struct {
	Platform    string `form:"platform" binding:"required"` // 平台名称
	CompanyUuid uint64 `form:"companyUuid"`                 // 公司 UUID（可选，默认当前公司）
}

// GrabCategoryBinding 分类绑定
type GrabCategoryBinding struct {
	GrabCategoryId string `json:"grabCategoryId"` // Grab 分类 ID
	CategoryUuid   uint64 `json:"categoryUuid"`   // 绑定的店内分类 UUID，空则创建
}

// GrabItemBinding 商品绑定
type GrabItemBinding struct {
	GrabItemId         string `json:"grabItemId"`         // Grab 商品 ID
	ProductPackageUuid uint64 `json:"productPackageUuid"` // 绑定的店内商品 UUID，空则创建
	CategoryUuid       uint64 `json:"categoryUuid"`       // 若需要强制绑定分类，可传；否则使用前端选择/自动创建
}

// GrabModifierBinding 修饰符/规格/加料/属性绑定
type GrabModifierBinding struct {
	GrabModifierId string `json:"grabModifierId"` // Grab 修饰符 ID
	FlavorUuid     uint64 `json:"flavorUuid"`     // 绑定到规格（可选）
	SauceUuid      uint64 `json:"sauceUuid"`      // 绑定到加料（可选）
	AttributeUuid  uint64 `json:"attributeUuid"`  // 绑定到属性值（可选）
}

// GrabMenuBinding 前端编排后的绑定结果
type GrabMenuBinding struct {
	Categories    []GrabCategoryBinding `json:"categories"`    // 分类绑定
	Items         []GrabItemBinding     `json:"items"`         // 商品绑定
	Modifiers     []GrabModifierBinding `json:"modifiers"`     // 规格/加料/属性绑定
	CreateMissing bool                  `json:"createMissing"` // 若未绑定则自动创建
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
