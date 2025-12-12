package valueobject

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
)

// CampaignInfo 营销活动信息
type CampaignInfo struct {
	OriginalPrice int64  // 原价
	DiscountType  string // 折扣类型
	DiscountValue int64  // 折扣值
}

// MenuItem 菜品商品值对象（平台通用）
type MenuItem struct {
	ID                     string                 // 商品 ID
	Name                   string                 // 商品名称
	Sequence               int                    // 排序序号
	AvailableStatus        AvailableStatus        // 可用状态
	Price                  int64                  // 价格（单位：分）
	CampaignInfo           *CampaignInfo          // 营销活动信息
	Description            string                 // 商品描述
	DescriptionTranslation map[string]string      // 商品描述多语言
	Photos                 []string               // 商品图片 URL 列表
	ModifierGroups         []*ModifierGroup       // 修饰符组列表
	SellingTimeID          string                 // 关联的售卖时段 ID
	NameTranslation        map[string]string      // 多语言名称
	Extra                  map[string]interface{} // 平台特定扩展字段
}

// NewMenuItem 创建菜品商品值对象
func NewMenuItem(id, name string, sequence int, status AvailableStatus, price int64) (*MenuItem, error) {
	if id == "" {
		return nil, errors.New("商品 ID 不能为空")
	}
	if name == "" {
		return nil, errors.New("商品名称不能为空")
	}
	if price < 0 {
		return nil, errors.New("商品价格不能为负数")
	}

	return &MenuItem{
		ID:              id,
		Name:            name,
		Sequence:        sequence,
		AvailableStatus: status,
		Price:           price,
		Photos:          make([]string, 0),
		ModifierGroups:  make([]*ModifierGroup, 0),
		NameTranslation: make(map[string]string),
		Extra:           make(map[string]interface{}),
		Description:     "",
	}, nil
}

// AddModifierGroup 添加修饰符组
func (i *MenuItem) AddModifierGroup(group *ModifierGroup) {
	i.ModifierGroups = append(i.ModifierGroups, group)
}

// AddPhoto 添加商品图片
func (i *MenuItem) AddPhoto(photoURL string) {
	i.Photos = append(i.Photos, photoURL)
}

// Validate 验证菜品商品值对象
func (i *MenuItem) Validate() error {
	if i.ID == "" {
		return errors.New("商品 ID 不能为空")
	}
	if i.Name == "" {
		return errors.New("商品名称不能为空")
	}
	if i.Price < 0 {
		return errors.New("商品价格不能为负数")
	}

	// 验证所有修饰符组
	for _, group := range i.ModifierGroups {
		if err := group.Validate(); err != nil {
			return errors.WithMessage(err, "修饰符组验证失败")
		}
	}

	return nil
}

// GetSellingPoint 构建卖点描述
func (i *MenuItem) GetSellingPoint() dto.LocaleResponse {
	sellingPoint := dto.LocaleResponse{}

	if i.Description != "" {
		sellingPoint.EN = i.Description
	}

	// 从描述多语言填充
	if i.DescriptionTranslation != nil {
		if zh, ok := i.DescriptionTranslation["zh"]; ok {
			sellingPoint.ZH = zh
		}
		if th, ok := i.DescriptionTranslation["th"]; ok {
			sellingPoint.TH = th
		}
		if en, ok := i.DescriptionTranslation["en"]; ok {
			sellingPoint.EN = en
		}
		if zhtw, ok := i.DescriptionTranslation["zh-TW"]; ok {
			sellingPoint.ZHTW = zhtw
		}
		if ja, ok := i.DescriptionTranslation["ja"]; ok {
			sellingPoint.JA = ja
		}
		if ko, ok := i.DescriptionTranslation["ko"]; ok {
			sellingPoint.KO = ko
		}
		if my, ok := i.DescriptionTranslation["my"]; ok {
			sellingPoint.MY = my
		}
		if tr, ok := i.DescriptionTranslation["tr"]; ok {
			sellingPoint.TR = tr
		}
		if sv, ok := i.DescriptionTranslation["sv"]; ok {
			sellingPoint.SV = sv
		}
	}

	return sellingPoint
}
