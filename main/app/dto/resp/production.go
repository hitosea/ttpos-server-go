package resp

import "ttpos-server-go/app/dto"

type ProductionGroup struct {
	LocaleName     *dto.LocaleResponse `json:"locale_name"`  // 序列号
	ProductionList ProductionList      `json:"product_list"` // 送厨商品列表
}

type ProductionList struct {
	List []ProductionItem `json:"list"` // 送厨商品
}

type ProductionItem struct {
	SerialNo              string             `json:"serial_no"`               // 序列号
	Uuid                  uint64             `json:"uuid"`                    // 送厨商品Uuid
	LocaleName            dto.LocaleResponse `json:"locale_name"`             // 送厨商品名称
	Num                   uint               `json:"num"`                     // 送厨商品数量
	CreateTime            int64              `json:"create_time"`             // 送厨时间
	ProductAttributeNames string             `json:"product_attribute_names"` // 商品属性
}

// ProductionListWithPagination 商品列表响应
type ProductionListWithPagination struct {
	List         []ProductionGroup `json:"list"`          // 分组列表
	FinishedList ProductionList    `json:"finished_list"` // 最近三个上菜历史
	Meta         dto.PageResponse  `json:"meta"`          // 分页信息
}

// ProductionHistory 上菜历史
type ProductionHistory struct {
	List []ProductionGroup `json:"list"`
}
