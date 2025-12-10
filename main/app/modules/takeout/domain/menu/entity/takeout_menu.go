package entity

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/menu/valueobject"
)

// TakeoutMenu 外卖菜单聚合根（平台通用）
type TakeoutMenu struct {
	Currency     *valueobject.Currency      // 货币配置
	SellingTimes []*valueobject.SellingTime // 售卖时段列表
	Categories   []*valueobject.Category    // 分类及商品树
	CompanyUuid  uint64                     // 公司 UUID
	Extra        map[string]interface{}     // 平台特定扩展字段
}

// NewTakeoutMenu 创建外卖菜单聚合根
func NewTakeoutMenu(companyUuid uint64) (*TakeoutMenu, error) {
	if companyUuid == 0 {
		return nil, errors.New("公司 UUID 不能为空")
	}

	return &TakeoutMenu{
		SellingTimes: make([]*valueobject.SellingTime, 0),
		Categories:   make([]*valueobject.Category, 0),
		CompanyUuid:  companyUuid,
		Extra:        make(map[string]interface{}),
	}, nil
}

// SetCurrency 设置货币
func (m *TakeoutMenu) SetCurrency(code, symbol string, exponent int) error {
	currency, err := valueobject.NewCurrency(code, symbol, exponent)
	if err != nil {
		return errors.WithMessage(err, "创建货币对象失败")
	}
	if currency == nil {
		return errors.New("货币配置不能为空")
	}
	m.Currency = currency
	return nil
}

// AddSellingTime 添加售卖时段
func (m *TakeoutMenu) AddSellingTime(sellingTime *valueobject.SellingTime) error {
	if sellingTime == nil {
		return errors.New("售卖时段不能为空")
	}
	if err := sellingTime.Validate(); err != nil {
		return errors.WithMessage(err, "售卖时段验证失败")
	}
	m.SellingTimes = append(m.SellingTimes, sellingTime)
	return nil
}

// AddCategory 添加分类
func (m *TakeoutMenu) AddCategory(category *valueobject.Category) error {
	if category == nil {
		return errors.New("分类不能为空")
	}
	if err := category.Validate(); err != nil {
		return errors.WithMessage(err, "分类验证失败")
	}
	m.Categories = append(m.Categories, category)
	return nil
}

// Validate 验证菜单数据完整性
func (m *TakeoutMenu) Validate() error {
	if m.CompanyUuid == 0 {
		return errors.New("公司 UUID 不能为空")
	}
	if m.Currency == nil {
		return errors.New("货币配置不能为空")
	}
	if err := m.Currency.Validate(); err != nil {
		return errors.WithMessage(err, "货币配置验证失败")
	}

	// 验证所有售卖时段
	for _, sellingTime := range m.SellingTimes {
		if err := sellingTime.Validate(); err != nil {
			return errors.WithMessage(err, "售卖时段验证失败")
		}
	}

	// 验证所有分类
	for _, category := range m.Categories {
		if err := category.Validate(); err != nil {
			return errors.WithMessage(err, "分类验证失败")
		}
	}

	return nil
}

// GetCategoryByID 根据 ID 获取分类
func (m *TakeoutMenu) GetCategoryByID(id string) *valueobject.Category {
	for _, category := range m.Categories {
		if category.ID == id {
			return category
		}
	}
	return nil
}

// GetSellingTimeByID 根据 ID 获取售卖时段
func (m *TakeoutMenu) GetSellingTimeByID(id string) *valueobject.SellingTime {
	for _, sellingTime := range m.SellingTimes {
		if sellingTime.ID == id {
			return sellingTime
		}
	}
	return nil
}
