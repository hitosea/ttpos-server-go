package valueobject

import "ttpos-server-go/app/errors"

// SellingTime 售卖时段值对象（平台通用）
type SellingTime struct {
	ID        string // 售卖时段 ID
	Name      string // 售卖时段名称
	Sequence  int    // 排序序号
	StartTime string // 生效开始时间
	EndTime   string // 生效结束时间
	// 平台特定的营业时间配置由适配器处理
}

// NewSellingTime 创建售卖时段值对象
func NewSellingTime(id, name string, sequence int, startTime, endTime string) (*SellingTime, error) {
	if id == "" {
		return nil, errors.New("售卖时段 ID 不能为空")
	}
	if name == "" {
		return nil, errors.New("售卖时段名称不能为空")
	}

	return &SellingTime{
		ID:        id,
		Name:      name,
		Sequence:  sequence,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

// Validate 验证售卖时段值对象
func (s *SellingTime) Validate() error {
	if s.ID == "" {
		return errors.New("售卖时段 ID 不能为空")
	}
	if s.Name == "" {
		return errors.New("售卖时段名称不能为空")
	}
	return nil
}
