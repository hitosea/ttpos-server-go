package resp

// SpecialCategory 表示特殊分类
type SpecialCategory struct {
	Name string `json:"name"` // 分类名称
	Id   uint   `json:"id"`   // 分类ID
}

// ChildCategory 表示子分类
type ChildCategory struct {
	Name string `json:"name"` // 子分类名称
	Key  string `json:"key"`  // 子分类键
}

// Category 表示分类，包含子分类
type Category struct {
	Name     string          `json:"name"`     // 分类名称
	Key      string          `json:"key"`      // 分类键
	Children []ChildCategory `json:"children"` // 子分类列表
}

// ProductCategory 表示产品分类，包含特殊分类和普通分类
type ProductCategory struct {
	SpecialCategoryList []SpecialCategory `json:"specialCategoryList"` // 特殊分类列表
	CategoryList        []Category        `json:"categoryList"`        // 分类列表
}
