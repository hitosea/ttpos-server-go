package req

import "encoding/json"

// SaveDataManageStatusReq 保存数据管理开关状态
type SaveDataManageStatusReq struct {
	IsEnableDataManage bool `json:"is_enable_data_manage"` // 是否启用数据管理
}

// SetDataManageStaffReq 设置数据管理操作人员
type SetDataManageStaffReq struct {
	StaffUuids []uint64 `json:"staff_uuids"` // 员工UUID列表
}

// GetDataManageOrderListReq 获取已选订单列表
type GetDataManageOrderListReq struct {
	PageNo         int    `form:"page_no" binding:"required,min=1"`           // 页码
	PageSize       int    `form:"page_size" binding:"required,min=1,max=100"` // 页面大小
	OrderNo        string `form:"order_no"`                                   // 订单编号搜索
	DateType       int    `form:"date_type"`                                  // 时间类型,-1=全部、0=今天、1=昨天、2=本周
	QueryStartDate string `form:"query_start_date"`                           // 查询开始日期（格式：YYYY-MM-DD）
	QueryEndDate   string `form:"query_end_date"`                             // 查询结束日期（格式：YYYY-MM-DD）
	BillType       int    `form:"bill_type"`                                  // 订单类型,-1=全部、0=餐单、1=外卖
}

// RestoreDataManageOrderReq 恢复（移除）单条已选订单
type RestoreDataManageOrderReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid" binding:"required"` // 销售单UUID
}

// GetDataManageOrderSelectReq 获取可选订单列表
type GetDataManageOrderSelectReq struct {
	PageNo         int    `form:"page_no" binding:"required,min=1"`           // 页码
	PageSize       int    `form:"page_size" binding:"required,min=1,max=100"` // 页面大小
	OrderNo        string `form:"order_no"`                                   // 订单编号搜索
	DateType       int    `form:"date_type"`                                  // 时间类型,-1=全部、0=今天、1=昨天、2=本周
	QueryStartDate string `form:"query_start_date"`                           // 查询开始日期（格式：YYYY-MM-DD）
	QueryEndDate   string `form:"query_end_date"`                             // 查询结束日期（格式：YYYY-MM-DD）
	BillType       int    `form:"bill_type"`                                  // 订单类型,-1=全部、0=餐单、1=外卖
}

// SubmitDataManageOrderReq 提交订单选择
// Filters 数组中每个元素独立处理：
//   - SelectAll=true：当前筛选条件范围内全部加入数据管理（DeselectedUuids 可排除部分）
//   - SelectAll=false + SelectedUuids：将指定订单新增到数据管理
//   - SelectAll=false + DeselectedUuids：将指定订单从数据管理移除
type SubmitDataManageOrderReq struct {
	Filters []DataManageOrderSubmitFilter `json:"filters"` // 筛选条件数组，每组独立处理
}

// GetDataManageOrderSelectStatsReq 获取可选订单统计预览（不持久化）
// 仅支持单个筛选条件：
//   - 无 filter 时默认近7天
//   - 有 filter 时按该筛选条件预览
type GetDataManageOrderSelectStatsReq struct {
	Filter *DataManageOrderSubmitFilter `json:"filter"` // 单个筛选条件
}

// UnmarshalJSON 兼容两种请求体格式：
// 1) 新格式：{"filter": {...}}
// 2) 旧格式：{...DataManageOrderSubmitFilter 字段...}
func (r *GetDataManageOrderSelectStatsReq) UnmarshalJSON(data []byte) error {
	type alias GetDataManageOrderSelectStatsReq
	var wrapped alias
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	if wrapped.Filter != nil {
		*r = GetDataManageOrderSelectStatsReq(wrapped)
		return nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if _, ok := raw["filter"]; ok {
		*r = GetDataManageOrderSelectStatsReq(wrapped)
		return nil
	}

	legacyKeys := []string{
		"select_all", "selected_uuids", "deselected_uuids",
		"order_no", "date_type", "query_start_date", "query_end_date", "bill_type",
	}
	hasLegacyKey := false
	for _, key := range legacyKeys {
		if _, ok := raw[key]; ok {
			hasLegacyKey = true
			break
		}
	}
	if !hasLegacyKey {
		*r = GetDataManageOrderSelectStatsReq(wrapped)
		return nil
	}

	var legacy DataManageOrderSubmitFilter
	if err := json.Unmarshal(data, &legacy); err != nil {
		return err
	}
	r.Filter = &legacy
	return nil
}

// DataManageOrderSubmitFilter 提交/预览时的筛选条件（含选择操作）
type DataManageOrderSubmitFilter struct {
	SelectAll       bool     `json:"select_all"`       // 是否全选当前筛选范围
	SelectedUuids   []uint64 `json:"selected_uuids"`   // 手动新增的UUID（select_all=false时使用）
	DeselectedUuids []uint64 `json:"deselected_uuids"` // 取消/移除的UUID（select_all=true时排除，select_all=false时移除）
	OrderNo         string   `json:"order_no"`         // 订单编号搜索
	DateType        int      `json:"date_type"`        // 时间类型,-1=全部、0=今天、1=昨天、2=本周
	QueryStartDate  string   `json:"query_start_date"` // 查询开始日期
	QueryEndDate    string   `json:"query_end_date"`   // 查询结束日期
	BillType        int      `json:"bill_type"`        // 订单类型,-1=全部、0=餐单、1=外卖
}
