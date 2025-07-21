package dto

import "github.com/gogf/gf/v2/container/gmap"

// FrappeDoc 定义泛型结构体，类似 TypeScript 的 FrappeDoc<T>
type FrappeDoc struct {
	// Owner 创建文档的用户
	Owner string `json:"owner"`
	// Creation 文档创建的日期和时间，ISO 格式
	Creation string `json:"creation"`
	// Modified 文档最后修改的日期和时间，ISO 格式
	Modified string `json:"modified"`
	// ModifiedBy 最后修改文档的用户
	ModifiedBy string `json:"modified_by"`
	// Idx 文档的索引位置
	Idx int `json:"idx"`
	// Docstatus 0 - 已保存, 1 - 已提交, 2 - 已取消
	Docstatus int `json:"docstatus"`
	// Parent 文档的父文档
	Parent any `json:"parent,omitempty"`
	// Parentfield 父文档字段
	Parentfield any `json:"parentfield,omitempty"`
	// Parenttype 父文档类型
	Parenttype any `json:"parenttype,omitempty"`
	// DocType 表的主键
	Name string `json:"name"`
	// DocType 文档类型
	Doctype string `json:"doctype"`
}

// RequestParams 定义请求参数结构体
type RequestParams struct {
	// Fields 要返回的字段列表
	Fields []string
	// Filters 筛选条件列表
	Filters [][]string
	// OrFilters 筛选条件列表
	OrFilters [][]string
	// LimitStart 分页起始位置
	LimitStart int
	// Limit 分页返回的数量
	Limit int
	// OrderBy 排序参数
	OrderBy OrderBy
	// GroupBy 分组参数
	GroupBy string
	// AsDict 作为字典返回
	AsDict bool
}

// OrderBy 定义排序参数结构体
type OrderBy struct {
	// Field 要排序的字段
	Field string
	// Order 排序顺序，可选值为 "asc"（升序）或 "desc"（降序）
	Order string
}

type ApiResp struct {
	data   *gmap.Map
	errors []ApiError
}

type ApiError struct {
	typeStr   string `json:"type"`
	message   string `json:"message"`
	title     string `json:"title"`
	indicator string `json:"indicator"`
}
