package value_object

import (
	"ttpos-server-go/app/errors"
)

// ModifierType 修饰符类型
type ModifierType string

const (
	ModifierTypeFlavor    ModifierType = "flavor"    // 规格（如大杯、中杯、小杯）
	ModifierTypeSauce     ModifierType = "sauce"     // 加料（如珍珠、椰果、布丁）
	ModifierTypeAttr      ModifierType = "attr"      // 属性（如冰度、糖度）
	ModifierTypeCommodity ModifierType = "commodity" // 套餐商品
)

// AvailableStatus 可用状态
type AvailableStatus string

const (
	AvailableStatusAvailable    AvailableStatus = "AVAILABLE"        // 可用
	AvailableStatusUnavailable  AvailableStatus = "UNAVAILABLE"      // 不可用
	AvailableStatusUnavailToday AvailableStatus = "UNAVAILABLETODAY" // 不可用今天
	AvailableStatusHide         AvailableStatus = "HIDE"             // 隐藏	无库存
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
