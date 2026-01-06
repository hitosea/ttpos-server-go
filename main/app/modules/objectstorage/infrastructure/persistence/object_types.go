package persistence

// ObjectType 对象类型常量定义
// 用于统一管理缓存 key 中的对象类型字符串，避免硬编码
const (
	// ProductList 商品列表
	ObjectTypeProductList = "product_list"

	// ProductBom 普通商品BOM
	ObjectTypeProductBom = "product_bom"

	// ProductFlavor 商品口味
	ObjectTypeProductFlavor = "product_flavor"

	// ProductBomFlavor 规格商品BOM（包含预加载的规格信息）
	ObjectTypeProductBomFlavor = "product_bom_flavor"

	// ProductBomSauce 小料商品BOM（包含预加载的小料信息）
	ObjectTypeProductBomSauce = "product_bom_sauce"

	// ProductBomBaseInfo 商品BOM基础信息（包含预加载的完整关联数据）
	ObjectTypeProductBomBaseInfo = "product_bom_base_info"

	// ProductPackage 商品包
	ObjectTypeProductPackage = "product_package"

	// ProductAttribute 商品属性
	ObjectTypeProductAttribute = "product_attribute"

	// SaleBillSetting 销售单设置
	ObjectTypeSaleBillSetting = "sale_bill_setting"

	// SaleBill 销售账单（包含完整关联数据）
	ObjectTypeSaleBill = "sale_bill"

	// BatchTag 分批标签
	ObjectTypeBatchTag = "batch_tag"

	// Desk 桌台
	ObjectTypeDesk = "desk"

	// Staff 员工
	ObjectTypeStaff = "staff"

	// ApiPermission API权限
	ObjectTypeApiPermission = "api_permission"

	// MultiLanguageName 多语言名称
	ObjectTypeMultiLanguageName = "multi_language_name"

	// ProductCategory 商品分类
	ObjectTypeProductCategory = "product_category"

	// ProductSauce 商品小料
	ObjectTypeProductSauce = "product_sauce"

	// ProductPackageAttribute 商品包属性
	ObjectTypeProductPackageAttribute = "product_package_attribute"

	// ProductMustPlanActive 商户是否有生效的必点方案（bool值）
	ObjectTypeProductMustPlanActive = "product_must_plan_active"

	// OrderItemRemark 订单商品备注
	ObjectTypeOrderItemRemark = "order_item_remark"

	// CompanySetting 商户设置
	ObjectTypeCompanySetting = "company_setting"

	// BuffetPackage 自助餐套餐
	ObjectTypeBuffetPackage = "buffet_package"

	// MarketingCoupon 营销优惠券
	ObjectTypeMarketingCoupon = "marketing_coupon"

	// MemberCoupon 会员优惠券
	ObjectTypeMemberCoupon = "member_coupon"

	// Tax 税种
	ObjectTypeTax = "tax"

	// ProductUnit 商品单位
	ObjectTypeProductUnit = "product_unit"

	// RelatedMaterial 关联材料
	ObjectTypeRelatedMaterial = "related_material"

	// Material 材料/物品
	ObjectTypeMaterial = "material"

	// WarehouseItem 仓库物品
	ObjectTypeWarehouseItem = "warehouse_item"

	// ProductBomCard 成本卡
	ObjectTypeProductBomCard = "product_bom_card"

	// OrderSource 订单来源
	ObjectTypeOrderSource = "order_source"

	// Nationality 国籍
	ObjectTypeNationality = "nationality"

	// PaymentMethod 支付方式
	ObjectTypePaymentMethod = "payment_method"

	// Member 会员
	ObjectTypeMember = "member"
)
