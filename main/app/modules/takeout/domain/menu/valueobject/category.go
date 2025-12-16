package valueobject

import "ttpos-server-go/app/errors"

// Category 菜单分类值对象（平台通用）
type Category struct {
	ID              string                 // 分类 ID
	Name            string                 // 分类名称
	Sequence        int                    // 排序序号
	AvailableStatus AvailableStatus        // 可用状态
	Items           []*MenuItem            // 商品列表
	SellingTimeID   string                 // 关联的售卖时段 ID
	NameTranslation map[string]string      // 多语言名称
	Extra           map[string]interface{} // 平台特定扩展字段
}

// NewCategory 创建菜单分类值对象
func NewCategory(id, name string, sequence int, status AvailableStatus) (*Category, error) {
	if id == "" {
		return nil, errors.New("分类 ID 不能为空")
	}
	if name == "" {
		return nil, errors.New("分类名称不能为空")
	}

	return &Category{
		ID:              id,
		Name:            name,
		Sequence:        sequence,
		AvailableStatus: status,
		Items:           make([]*MenuItem, 0),
		NameTranslation: make(map[string]string),
		Extra:           make(map[string]interface{}),
	}, nil
}

// AddItem 添加商品
func (c *Category) AddItem(item *MenuItem) {
	c.Items = append(c.Items, item)
}

// Validate 验证菜单分类值对象
func (c *Category) Validate() error {
	if c.ID == "" {
		return errors.New("分类 ID 不能为空")
	}
	if c.Name == "" {
		return errors.New("分类名称不能为空")
	}

	// 验证所有商品
	for _, item := range c.Items {
		if err := item.Validate(); err != nil {
			return errors.WithMessage(err, "商品验证失败")
		}
	}

	return nil
}
