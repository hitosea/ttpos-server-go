package resp

// HqControlSettingResp 总部控制设置响应
type HqControlSettingResp struct {
	HqControlDineShelf     int `json:"hq_control_dine_shelf"`     // 店内上下架控制模式：0-分开控制，1-统一控制
	HqControlTakeoutShelf  int `json:"hq_control_takeout_shelf"`  // 外卖上下架控制模式：0-分开控制，1-统一控制
	HqControlTakeoutPrice  int `json:"hq_control_takeout_price"`  // 外卖价格控制模式：0-分开控制，1-统一控制
	HqControlNegativeStock int `json:"hq_control_negative_stock"` // 负库存控制模式：0-分开控制，1-统一控制
}

// HqBatchPushStoreListResp 可推送门店列表响应
type HqBatchPushStoreListResp struct {
	List []HqBatchPushStoreItem `json:"list"` // 子店列表
}

// HqBatchPushStoreItem 可推送门店项
type HqBatchPushStoreItem struct {
	CompanyUuid uint64 `json:"company_uuid"` // 子店UUID
	Name        string `json:"name"`         // 子店名称
	Status      int    `json:"status"`       // 子店状态
}

// HqBatchPushResp 批量推送响应
type HqBatchPushResp struct {
	Message string `json:"message"` // 推送结果消息
}
