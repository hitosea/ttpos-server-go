package resp

import "ttpos-server-go/app/dto"

// TransferOrderAnnotationItem 调拨单批注项（用于详情接口）
type TransferOrderAnnotationItem struct {
	AnnotationType int                `json:"annotation_type"` // 批注类型: 1-重新提交 2-门店通过 3-门店驳回 4-上级门店通过 5-上级门店驳回 6-发货门店通过 7-发货门店驳回 8-收货门店通过 9-收货门店驳回
	LocaleName     dto.LocaleResponse `json:"locale_name"`     // 批注类型名称（多语言）
	Content        string             `json:"content"`         // 批注内容
	CreateTime     int64              `json:"create_time"`     // 创建时间
}
