package model

import "ttpos-server-go/app/dto"

// ReturnFoodReason 退菜原因表 ttpos_return_food_reason
type ReturnFoodReason struct {
	BaseModel
	Name                  string `gorm:"default:'';column:name;comment:'名称'"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称ID'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}

// FreeReason 赠品或免费订单原因表 ttpos_free_reason
type FreeReason struct {
	BaseModel
	Name                  string `gorm:"default:'';column:name;comment:'名称'"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称ID'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}

// OrderRemark 整单备注表 ttpos_order_remark
type OrderRemark struct {
	BaseModel
	Name                  string `gorm:"default:'';column:name;comment:'名称'"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称ID'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}

// OrderItemRemark 单品备注原因表 ttpos_order_item_remark
type OrderItemRemark struct {
	BaseModel
	Name                  string `gorm:"default:'';column:name;comment:'名称'"`
	MultiLanguageNameUuid uint64 `gorm:"default:0;column:multi_language_name_uuid;comment:'多语言名称ID'"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}

func (*OrderItemRemark) TableName() string {
	return "ttpos_order_item_remark"
}

func GetReturnFoodReasonNames(list []*ReturnFoodReason) dto.LocaleResponse {
	nameList := make([]dto.LocaleResponse, 0)
	for _, reason := range list {
		nameList = append(nameList, reason.MultiLanguageName.GetNames())
	}
	return getLocaleResponse(nameList, "、")
}
