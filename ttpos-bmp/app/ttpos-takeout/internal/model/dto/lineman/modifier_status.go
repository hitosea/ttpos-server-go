package lineman

// ModifierStatusUpdateReq Lineman 修饰符状态更新请求
// API: PUT /v1/partners/{partnerId}/stores/{storeId}/menu/property/values/status
type ModifierStatusUpdateReq struct {
	PropertyValues []ModifierPropertyValue `json:"propertyValues"`
}

// ModifierPropertyValue 修饰符属性值
type ModifierPropertyValue struct {
	ID     string `json:"id"`     // Partner Modifier ID
	Status int    `json:"status"` // 1=AVAILABLE, 2=SOLD_OUT_TODAY, 3=SUSPENDED
}

// ModifierStatusUpdateResp Lineman 修饰符状态更新响应
// 复用公用的 LinemanResponse
type ModifierStatusUpdateResp = LinemanResponse
