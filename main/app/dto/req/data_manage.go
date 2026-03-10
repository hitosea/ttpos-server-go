package req

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
// 支持三种交互模式：
//   - 全选：SelectAll=true, DeselectedUuids=[]
//   - 全选后取消部分：SelectAll=true, DeselectedUuids=[取消的UUID]
//   - 手动勾选部分：SelectAll=false, SelectedUuids=[选中的UUID]
type SubmitDataManageOrderReq struct {
	SelectAll       bool                       `json:"select_all"`        // 是否全选当前筛选范围
	SelectedUuids   []uint64                   `json:"selected_uuids"`    // 手动勾选的UUID（select_all=false时使用）
	DeselectedUuids []uint64                   `json:"deselected_uuids"`  // 取消勾选的UUID（select_all=true时使用）
	Filter          DataManageOrderSubmitFilter `json:"filter"`            // 筛选条件（与查询列表时相同的参数）
}

// GetDataManageOrderSelectStatsReq 获取可选订单统计预览（参数同提交，不持久化）
type GetDataManageOrderSelectStatsReq struct {
	SelectAll       bool                       `json:"select_all"`        // 是否全选当前筛选范围
	SelectedUuids   []uint64                   `json:"selected_uuids"`    // 手动勾选的UUID（select_all=false时使用）
	DeselectedUuids []uint64                   `json:"deselected_uuids"`  // 取消勾选的UUID（select_all=true时使用）
	Filter          DataManageOrderSubmitFilter `json:"filter"`            // 筛选条件
}

// DataManageOrderSubmitFilter 提交订单时的筛选条件
type DataManageOrderSubmitFilter struct {
	OrderNo        string `json:"order_no"`         // 订单编号搜索
	DateType       int    `json:"date_type"`        // 时间类型,-1=全部、0=今天、1=昨天、2=本周
	QueryStartDate string `json:"query_start_date"` // 查询开始日期
	QueryEndDate   string `json:"query_end_date"`   // 查询结束日期
	BillType       int    `json:"bill_type"`        // 订单类型,-1=全部、0=餐单、1=外卖
}
