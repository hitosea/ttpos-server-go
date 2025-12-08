package resp

import "ttpos-server-go/app/dto"

// SyncResp 同步响应
type SyncResp struct {
	TaskUuid uint64 `json:"task_uuid"` // 任务UUID
	Message  string `json:"message"`   // 提示消息
}

// SyncTaskItemResp 同步任务明细响应
type SyncTaskItemResp struct {
	Uuid         uint64 `json:"uuid"`          // UUID
	TaskType     string `json:"task_type"`     // 任务类型
	TaskName     string `json:"task_name"`     // 任务名称
	Status       uint8  `json:"status"`        // 任务状态: 0-待执行, 1-执行中, 2-已完成, 3-失败
	ErrorMessage string `json:"error_message"` // 错误消息
	StartTime    int64  `json:"start_time"`    // 开始时间
	EndTime      int64  `json:"end_time"`      // 结束时间
	Duration     int64  `json:"duration"`      // 执行时长（秒）
}

// SyncTaskDetailResp 同步任务详情响应
type SyncTaskDetailResp struct {
	Uuid         uint64             `json:"uuid"`          // UUID
	Status       uint8              `json:"status"`        // 同步状态: 0-进行中, 1-已完成, 2-失败
	TotalCount   uint32             `json:"total_count"`   // 总任务数
	SuccessCount uint32             `json:"success_count"` // 成功任务数
	FailCount    uint32             `json:"fail_count"`    // 失败任务数
	StartTime    int64              `json:"start_time"`    // 开始时间
	EndTime      int64              `json:"end_time"`      // 结束时间
	Duration     int64              `json:"duration"`      // 执行时长（秒）
	CreateTime   int64              `json:"create_time"`   // 创建时间
	Items        []SyncTaskItemResp `json:"items"`         // 任务明细
}

// SyncTaskListResp 同步任务列表响应
type SyncTaskListResp struct {
	Uuid         uint64 `json:"uuid"`          // UUID
	Status       uint8  `json:"status"`        // 同步状态: 0-进行中, 1-已完成, 2-失败
	TotalCount   uint32 `json:"total_count"`   // 总任务数
	SuccessCount uint32 `json:"success_count"` // 成功任务数
	FailCount    uint32 `json:"fail_count"`    // 失败任务数
	StartTime    int64  `json:"start_time"`    // 开始时间
	EndTime      int64  `json:"end_time"`      // 结束时间
	Duration     int64  `json:"duration"`      // 执行时长（秒）
	CreateTime   int64  `json:"create_time"`   // 创建时间
}

// SyncTaskListPaginationResp 同步任务列表分页响应
type SyncTaskListPaginationResp struct {
	List []SyncTaskListResp `json:"list"`
	Meta dto.PageResponse   `json:"meta"`
}

// HeadquartersDataListResp 总部可同步数据列表响应
type HeadquartersDataListResp struct {
	DataGroups []DataGroup `json:"data_groups"` // 按种类分组的数据
}

// DataGroup 数据分组
type DataGroup struct {
	Type        string     `json:"type"`         // 数据类型: product_category - 商品分类, material_category - 物品分类, tax - 税类, unit - 单位, material - 物品, flavor - 规格, attribute - 属性, sauce - 加料, product - 商品, product_stock - 商品库存, bom_card - 成本卡, supplier - 供应商, coupon - 优惠券, full_reduction - 满额减, product_label - 菜品标签, marketing_activity - 营销活动
	TypeName    string     `json:"type_name"`    // 类型名称（如：商品分类、单位、优惠券等）
	Items       []DataItem `json:"items"`        // 该类型的数据列表
	SyncedUuids []uint64   `json:"synced_uuids"` // 分店已同步的总部数据uuid列表
}

// DataItem 数据项
type DataItem struct {
	Uuid        uint64             `json:"uuid"`         // 数据uuid
	LocaleName  dto.LocaleResponse `json:"locale_name"`  // 数据名称
	RelatedData []RelatedData      `json:"related_data"` // 关联数据（明确类型和uuid列表）
}

// RelatedData 关联数据
type RelatedData struct {
	Type  string   `json:"type"`  // 关联数据的类型: product_category - 商品分类, material_category - 物品分类, tax - 税类, unit - 单位, material - 物品, flavor - 规格, attribute - 属性, sauce - 加料, product - 商品, product_stock - 商品库存, bom_card - 成本卡, supplier - 供应商, coupon - 优惠券, full_reduction - 满额减, product_label - 菜品标签, marketing_activity - 营销活动
	Uuids []uint64 `json:"uuids"` // 关联的uuid列表
}

// GranularSyncResp 颗粒化同步响应
type GranularSyncResp struct {
	TaskUuid uint64 `json:"task_uuid"` // 同步任务uuid
	Message  string `json:"message"`   // 提示信息
}
