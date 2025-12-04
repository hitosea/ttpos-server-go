package valueobject

import (
	"strings"
	"ttpos-server-go/app/errors"
)

// WarehouseCode 仓库编码值对象
type WarehouseCode struct {
	value string
}

// NewWarehouseCode 创建仓库编码
func NewWarehouseCode(code string) (WarehouseCode, error) {
	if code == "" {
		return WarehouseCode{}, errors.New("仓库编码不能为空")
	}

	// 转换为大写
	upperCode := strings.ToUpper(code)

	// 验证编码格式（2-20个字符，只能包含字母、数字和下划线）
	if len(upperCode) < 2 || len(upperCode) > 20 {
		return WarehouseCode{}, errors.New("仓库编码长度必须在2-20个字符之间")
	}

	return WarehouseCode{value: upperCode}, nil
}

// Value 获取编码值
func (c WarehouseCode) Value() string {
	return c.value
}

// Equals 比较两个编码是否相等
func (c WarehouseCode) Equals(other WarehouseCode) bool {
	return c.value == other.value
}

// String 实现 Stringer 接口
func (c WarehouseCode) String() string {
	return c.value
}


