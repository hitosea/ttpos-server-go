package req

import "ttpos-server-go/app/dto"

// SyncReq 同步请求
type SyncReq struct {
	TaskUuid      uint64 `json:"task_uuid" form:"task_uuid"`             // 任务UUID，如果传递则重新执行该任务
	IsSyncExecute bool   `json:"is_sync_execute" form:"is_sync_execute"` // 是否同步执行
}

// SyncTaskListReq 同步任务列表请求
type SyncTaskListReq struct {
	dto.PageReq
	Status *uint8 `json:"status" form:"status"` // 同步状态: 0-进行中, 1-已完成, 2-失败
}

// SyncTaskDetailReq 同步任务详情请求
type SyncTaskDetailReq struct {
	TaskUuid uint64 `json:"task_uuid" form:"task_uuid" binding:"required"` // 任务UUID
}

// GetHeadquartersDataListReq 获取总部可同步数据列表请求
type GetHeadquartersDataListReq struct {
	DataTypes []string `json:"data_types"` // 可选，指定查询的数据类型，不传则查询所有
}

// GranularSyncReq 颗粒化同步请求
type GranularSyncReq struct {
	SyncData GranularSyncData `json:"sync_data" binding:"required"` // 要同步的数据
}

// GranularSyncData 要同步的数据（按种类分组）
type GranularSyncData struct {
	ProductCategory   []uint64 `json:"product_category"`   // 商品分类
	Unit              []uint64 `json:"unit"`               // 单位
	Flavor            []uint64 `json:"flavor"`             // 规格
	Attribute         []uint64 `json:"attribute"`          // 属性
	Sauce             []uint64 `json:"sauce"`              // 加料
	Product           []uint64 `json:"product"`            // 商品
	MaterialCategory  []uint64 `json:"material_category"`  // 物品分类
	Material          []uint64 `json:"material"`           // 物品
	BomCard           []uint64 `json:"bom_card"`           // 成本卡
	Supplier          []uint64 `json:"supplier"`           // 供应商
	Tax               []uint64 `json:"tax"`                // 税类
	Coupon            []uint64 `json:"coupon"`             // 优惠券
	FullReduction     []uint64 `json:"full_reduction"`     // 满额减
	ProductLabel      []uint64 `json:"product_label"`      // 菜品标签
	MarketingActivity []uint64 `json:"marketing_activity"` // 营销活动
	PaymentMethod     []uint64 `json:"payment_method"`     // 支付方式
}
