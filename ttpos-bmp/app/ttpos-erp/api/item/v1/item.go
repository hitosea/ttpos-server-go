package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ItemSearchReq Item查询请求参数
type ItemSearchReq struct {
	g.Meta   `path:"/item/search" tags:"Item" method:"get" summary:"根据名称搜索商品"`
	Name     string `json:"name" v:"required#商品名称不能为空" dc:"商品名称"`
	Page     int    `json:"page" v:"min:1#页码必须大于0" d:"1" dc:"页码"`
	PageSize int    `json:"page_size" v:"min:1|max:100#每页数量必须在1-100之间" d:"10" dc:"每页数量"`
}

// ItemSearchRes Item查询响应结构
type ItemSearchRes struct {
	g.Meta     `mime:"application/json"`
	List       []ItemInfo `json:"list" dc:"商品列表"`
	Total      int64      `json:"total" dc:"总数量"`
	Page       int        `json:"page" dc:"当前页码"`
	PageSize   int        `json:"page_size" dc:"每页数量"`
	TotalPages int        `json:"total_pages" dc:"总页数"`
}

// ItemInfo 商品信息结构
type ItemInfo struct {
	ID           uint    `json:"id" dc:"主键ID"`
	Uuid         uint64  `json:"uuid" dc:"UUID"`
	Name         string  `json:"name" dc:"商品名称"`
	ImageName    string  `json:"image_name" dc:"图片名称"`
	Status       uint    `json:"status" dc:"状态, 0-上架 1-下架"`
	CreateTime   int64   `json:"create_time" dc:"创建时间"`
	UpdateTime   int64   `json:"update_time" dc:"更新时间"`
	Price        float64 `json:"price,omitempty" dc:"最低价格"`
	StockNum     int     `json:"stock_num,omitempty" dc:"库存数量"`
	CategoryName string  `json:"category_name,omitempty" dc:"分类名称"`
}
