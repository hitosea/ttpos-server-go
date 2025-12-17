package model

// ProductPackageTakeout 外卖商品表，存储商品的外卖专属信息 `ttpos_product_package_takeout`
type ProductPackageTakeout struct {
	BaseModel
	ProductPackageUuid            uint64 `gorm:"default:0;column:product_package_uuid;comment:'商品包UUID，关联 ttpos_product_package.uuid'"`
	MultiLanguageNameUuid         uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称ID'"`
	HeadquarterUuid               uint64 `gorm:"default:0;column:headquarter_uuid;comment:'总部UUID,0表示不是总部商品'"`
	Name                          string `gorm:"type:text;column:name;comment:'商品包名称'"`
	ProductType                   uint   `gorm:"default:0;column:product_type;comment:'商品类型, 0-商品 1-套餐'"`
	TakeoutType                   uint   `gorm:"default:1;column:takeout_type;comment:'外卖类型 1-Grab 2-FoodPanda 3-其他'"`
	Status                        uint   `gorm:"default:0;column:status;comment:'外卖状态 0-下架 1-上架'"`
	CategoryUuid                  uint64 `gorm:"default:0;column:category_uuid;comment:'外卖分类UUID'"`
	SpecialCategoryUuid           uint64 `gorm:"default:0;column:special_category_uuid;comment:'外卖特色分类UUID'"`
	ImageFileUuid                 uint64 `gorm:"default:0;column:image_file_uuid;comment:'外卖商品图片UUID'"`
	Describe                      string `gorm:"default:'';column:describe;comment:'卖点描述'"`
	DescribeMultiLanguageNameUuid uint64 `gorm:"default:0;column:describe_multi_language_name_uuid;comment:'商品卖点多语言UUID'"`
	Source                        string `gorm:"type:varchar(50);default:'';column:source;index:idx_source_product;comment:'来源平台(grab/foodpanda/lineman等)'"`
	SourceProductId               string `gorm:"type:varchar(500);default:'';column:source_product_id;index:idx_source_product;comment:'来源平台商品唯一ID'"`

	// 关联关系
	ProductPackage                  ProductPackage                   `gorm:"foreignKey:product_package_uuid;references:uuid" json:"-"`              // 商品包
	MultiLanguageName               MultiLanguageName                `gorm:"foreignKey:multi_language_name_uuid;references:uuid" json:"-"`          // 多语言名称
	DescribeMultiLanguageName       MultiLanguageName                `gorm:"foreignKey:describe_multi_language_name_uuid;references:uuid" json:"-"` // 卖点多语言
	ProductCategory                 ProductCategory                  `gorm:"foreignKey:category_uuid;references:uuid" json:"-"`                     // 外卖分类
	ProductSpecialCategory          ProductCategory                  `gorm:"foreignKey:special_category_uuid;references:uuid" json:"-"`             // 外卖特色分类
	ImageFile                       File                             `gorm:"foreignKey:image_file_uuid;references:uuid" json:"-"`                   // 图片
	ProductBomTakeouts              []ProductBomTakeout              `gorm:"foreignKey:product_package_takeout_uuid;references:uuid" json:"-"`      // 外卖规格价格列表
	ProductPackageAttributeTakeouts []ProductPackageAttributeTakeout `gorm:"foreignKey:product_package_takeout_uuid;references:uuid" json:"-"`      // 外卖属性价格列表
}

// TableName 指定表名
func (ProductPackageTakeout) TableName() string {
	return "ttpos_product_package_takeout"
}

// IsUp 判断外卖商品是否上架
func (model *ProductPackageTakeout) IsUp() bool {
	return model.Status == 1 && model.DeleteTime == 0
}

// IsDown 判断外卖商品是否下架
func (model *ProductPackageTakeout) IsDown() bool {
	return !model.IsUp()
}
