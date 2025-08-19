package erp

import "github.com/gogf/gf/v2/container/gmap"

// FrappeDoc 定义泛型结构体，类似 TypeScript 的 FrappeDoc<T>
type FrappeDoc struct {
	// Owner 创建文档的用户
	Owner string `json:"owner,omitempty"`
	// Creation 文档创建的日期和时间，ISO 格式
	Creation string `json:"creation,omitempty"`
	// Modified 文档最后修改的日期和时间，ISO 格式
	Modified string `json:"modified,omitempty"`
	// ModifiedBy 最后修改文档的用户
	ModifiedBy string `json:"modified_by,omitempty"`
	// Idx 文档的索引位置
	Idx int `json:"idx,omitempty"`
	// Docstatus 0 - 已保存, 1 - 已提交, 2 - 已取消
	Docstatus int `json:"docstatus,omitempty"`
	// Parent 文档的父文档
	Parent any `json:"parent,omitempty"`
	// Parentfield 父文档字段
	Parentfield any `json:"parentfield,omitempty"`
	// Parenttype 父文档类型
	Parenttype any `json:"parenttype,omitempty"`
	// DocType 表的主键
	Name string `json:"name,omitempty"`
	// DocType 文档类型
	Doctype string `json:"doctype,omitempty"`
}

// RequestParams 定义请求参数结构体
type RequestParams struct {
	// Fields 要返回的字段列表
	Fields []string `json:"fields,omitempty"`
	// Filters 筛选条件列表
	Filters [][]string `json:"filters,omitempty"`
	// OrFilters 筛选条件列表
	OrFilters [][]string `json:"or_filters,omitempty"`
	// LimitStart 分页起始位置
	LimitStart int `json:"limit_start,omitempty"`
	// Limit 分页返回的数量
	Limit int `json:"limit,omitempty"`
	// OrderBy 排序参数
	OrderBy OrderBy `json:"order_by,omitempty"`
	// GroupBy 分组参数
	GroupBy string `json:"group_by,omitempty"`
	// AsDict 作为字典返回
	AsDict bool `json:"as_dict,omitempty"`
}

// OrderBy 定义排序参数结构体
type OrderBy struct {
	// Field 要排序的字段
	Field string `json:"field,omitempty"`
	// Order 排序顺序，可选值为 "asc"（升序）或 "desc"（降序）
	Order string `json:"order,omitempty"`
}

type ApiResp struct {
	Data   *gmap.Map  `json:"data,omitempty"`
	Errors []ApiError `json:"errors,omitempty"`
}

type ApiError struct {
	TypeStr   string `json:"type,omitempty"`
	Message   string `json:"message,omitempty"`
	Title     string `json:"title,omitempty"`
	Indicator string `json:"indicator,omitempty"`
}

type ErpReq struct {
	DocType string `json:"docType,omitempty"`
	Name    string `json:"name,omitempty"`
	Method  string `json:"method,omitempty"`
}

type ReportParams struct {
	ReportName           string `json:"report_name,omitempty"`                           // 报表名称
	Filters              string `json:"filters,omitempty"`                               //json 格式的筛选条件
	IgnorePreparedReport bool   `default:"true" json:"ignore_prepared_report,omitempty"` // 是否忽略已准备好的报表
}
