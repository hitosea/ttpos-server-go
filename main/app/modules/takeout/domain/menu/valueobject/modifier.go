package valueobject

import (
	"ttpos-server-go/app/errors"
)

// AvailableStatus 可用状态
type AvailableStatus string

const (
	AvailableStatusAvailable   AvailableStatus = "AVAILABLE"
	AvailableStatusUnavailable AvailableStatus = "UNAVAILABLE"
)

// ToInt 转换为整数
func (s AvailableStatus) ToInt() int {
	switch s {
	case AvailableStatusAvailable:
		return 1
	case AvailableStatusUnavailable:
		return 0
	}
	return 0
}

// Modifier 修饰符值对象（规格/加料，平台通用）
type Modifier struct {
	ID              string                 // 修饰符 ID
	Name            string                 // 修饰符名称
	Sequence        int                    // 排序序号
	AvailableStatus AvailableStatus        // 可用状态
	Price           int64                  // 价格（单位：分）
	NameTranslation map[string]string      // 多语言名称
	Extra           map[string]interface{} // 平台特定扩展字段
}

// NewModifier 创建修饰符值对象
func NewModifier(id, name string, sequence int, status AvailableStatus, price int64) (*Modifier, error) {
	if id == "" {
		return nil, errors.New("修饰符 ID 不能为空")
	}
	if name == "" {
		return nil, errors.New("修饰符名称不能为空")
	}

	return &Modifier{
		ID:              id,
		Name:            name,
		Sequence:        sequence,
		AvailableStatus: status,
		Price:           price,
		NameTranslation: make(map[string]string),
		Extra:           make(map[string]interface{}),
	}, nil
}

// Validate 验证修饰符值对象
func (m *Modifier) Validate() error {
	if m.ID == "" {
		return errors.New("修饰符 ID 不能为空")
	}
	if m.Name == "" {
		return errors.New("修饰符名称不能为空")
	}
	return nil
}

// ModifierGroup 修饰符组值对象
type ModifierGroup struct {
	ID                string                 // 修饰符组 ID
	Name              string                 // 修饰符组名称
	NameTranslation   map[string]string      // 多语言名称
	Sequence          int                    // 排序序号
	AvailableStatus   AvailableStatus        // 可用状态
	SelectionRangeMin int                    // 最小选择数量
	SelectionRangeMax int                    // 最大选择数量
	Modifiers         []*Modifier            // 修饰符列表
	Extra             map[string]interface{} // 平台特定扩展字段
}

// NewModifierGroup 创建修饰符组值对象
func NewModifierGroup(id, name string, nameTranslation map[string]string, sequence int, status AvailableStatus, min, max int) (*ModifierGroup, error) {
	if id == "" {
		return nil, errors.New("修饰符组 ID 不能为空")
	}
	if name == "" {
		return nil, errors.New("修饰符组名称不能为空")
	}
	if min < 0 {
		return nil, errors.New("最小选择数量不能为负数")
	}
	if max < min {
		return nil, errors.New("最大选择数量不能小于最小选择数量")
	}

	return &ModifierGroup{
		ID:                id,
		Name:              name,
		NameTranslation:   nameTranslation,
		Sequence:          sequence,
		AvailableStatus:   status,
		SelectionRangeMin: min,
		SelectionRangeMax: max,
		Modifiers:         make([]*Modifier, 0),
		Extra:             make(map[string]interface{}),
	}, nil
}

// AddModifier 添加修饰符
func (g *ModifierGroup) AddModifier(modifier *Modifier) {
	g.Modifiers = append(g.Modifiers, modifier)
}

// Validate 验证修饰符组值对象
func (g *ModifierGroup) Validate() error {
	if g.ID == "" {
		return errors.New("修饰符组 ID 不能为空")
	}
	if g.Name == "" {
		return errors.New("修饰符组名称不能为空")
	}
	if g.SelectionRangeMin < 0 {
		return errors.New("最小选择数量不能为负数")
	}
	if g.SelectionRangeMax < g.SelectionRangeMin {
		return errors.New("最大选择数量不能小于最小选择数量")
	}

	// 验证所有修饰符
	for _, modifier := range g.Modifiers {
		if err := modifier.Validate(); err != nil {
			return errors.WithMessage(err, "修饰符验证失败")
		}
	}

	return nil
}
