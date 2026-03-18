package req

// HqControlSettingUpdateReq 更新总部控制设置请求
type HqControlSettingUpdateReq struct {
	HqControlDineShelf     *int `json:"hq_control_dine_shelf" binding:"omitempty,oneof=0 1"`     // 店内上下架控制模式：0-分开控制，1-统一控制
	HqControlTakeoutShelf  *int `json:"hq_control_takeout_shelf" binding:"omitempty,oneof=0 1"`  // 外卖上下架控制模式：0-分开控制，1-统一控制
	HqControlTakeoutPrice  *int `json:"hq_control_takeout_price" binding:"omitempty,oneof=0 1"`  // 外卖价格控制模式：0-分开控制，1-统一控制
	HqControlNegativeStock *int `json:"hq_control_negative_stock" binding:"omitempty,oneof=0 1"` // 负库存控制模式：0-分开控制，1-统一控制
}

// HqBatchPushReq 批量推送请求
type HqBatchPushReq struct {
	FieldTypes  []string `json:"field_types" binding:"required,min=1"` // 推送类型列表：dine_shelf/takeout_shelf/takeout_price/negative_stock
	IsAllStores bool     `json:"is_all_stores"`                        // 是否推送到所有子店
	StoreUuids  []uint64 `json:"store_uuids"`                          // 指定子店UUID列表（is_all_stores=false 时生效）
}

// HqBatchPushStoreListReq 获取可推送门店列表请求
type HqBatchPushStoreListReq struct{}
