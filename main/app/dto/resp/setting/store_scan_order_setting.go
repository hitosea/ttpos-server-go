package setting

import "ttpos-server-go/pkg/utils"

// StoreScanOrderSettingResp 门店点餐配置响应
type StoreScanOrderSettingResp struct {
	IsEnabled            int `json:"is_enabled"`              // 启用状态：0-关闭，1-开启
	EnableDelivery       int `json:"enable_delivery"`         // 外送服务：0-关闭，1-开启
	EnableSelfPickup     int `json:"enable_self_pickup"`      // 到店自取：0-关闭，1-开启
	DeliveryAvailable    int `json:"delivery_available"`      // 外送服务是否可用（云平台是否开启）：0-不可用，1-可用
	SelfPickupAvailable  int `json:"self_pickup_available"`   // 到店自取是否可用（云平台是否开启）：0-不可用，1-可用
}

// IsStoreResting 判断商家是否休息中
// 未启用门店点餐(IsEnabled!=1)视为休息中；已启用但不在营业时间内也视为休息中
func (s StoreScanOrderSettingResp) IsStoreResting(timezone string, openingHours string) bool {
	if s.IsEnabled != 1 {
		return true
	}
	return !utils.SetTimezone(timezone).IsWithinOpeningHours(openingHours)
}
