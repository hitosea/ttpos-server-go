package constant

const NotDeleted = 0   // 未删除。如："delete_time = ?", constant.NotDeleted
const OptionalUuid = 0 // 可选uuid。表示不使用这个字段

const (
	Yes = 1  // 是, true
	No  = 0  // 否, false
	All = -1 // 全选
)

// NoPaginationPageSize 不分页的 PageSize 阈值
// 当 PageSize >= 此值时，Repository 层跳过分页逻辑，返回所有数据
const NoPaginationPageSize = 10000
