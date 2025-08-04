package resp

import "ttpos-server-go/app/dto"

/***
  收银机"订单"中外送订单列表 接口返回结构体
***/

type OrderManageListMeta struct {
	dto.PageResponse
	TotalNum      int64 `json:"total_num"`      // 总数量
	UnpaidNum     int64 `json:"unpaid_num"`     // 待付款数量
	UnacceptNum   int64 `json:"unaccept_num"`   // 待接单数量
	AcceptNum     int64 `json:"accept_num"`     // 备餐中数量
	UndeliveryNum int64 `json:"undelivery_num"` // 待配送数量
	DeliveryNum   int64 `json:"delivery_num"`   // 配送中数量
	CompleteNum   int64 `json:"complete_num"`   // 已完成数量
	CancelNum     int64 `json:"cancel_num"`     // 已取消数量
}

type GetMemberOrderManageListResp struct {
	Meta OrderManageListMeta `json:"meta"`
	List []MemberOrderManage `json:"list"` // 订单列表
}

type MemberOrderManage struct {
	MemberSaleOrderUuid uint64          `json:"member_sale_order_uuid"` // 会员端销售订单UUID
	SerialNumber        string          `json:"serial_number"`          // 订单流水号
	OrderNo             string          `json:"order_no"`               // 订单号
	Status              uint            `json:"status"`                 // 订单状态.0-选购中 1-待付款 2-待商家接单 3-商家备餐中 4-待骑手接单 5-骑手正在赶往商家 6-骑手配送中 7-已完成 8-已取消
	StatusGroup         string          `json:"status_group"`           // 订单状态分组. "unpaid" 待付款， "unaccept" 待接单, "accept" 备餐中, "undelivery" 待配送, "delivery" 配送中, "completed" 已完成, "cancel" 已取消
	CreateTime          int64           `json:"create_time"`            // 创建时间,下单时间
	PayTime             int64           `json:"pay_time"`               // 支付时间
	OriginAmount        float64         `json:"origin_amount"`          // 订单原始金额。商品金额+配送费
	PayAmount           float64         `json:"pay_amount"`             // 实付金额。商品金额+配送费-会员折扣
	DeliveryFee         float64         `json:"delivery_fee"`           // 配送费
	PayType             string          `json:"pay_type"`               // 支付方式
	Contact             ContactInfo     `json:"contact"`                // 联系人信息
	Rider               RiderInfo       `json:"rider"`                  // 骑手信息
	Extra               OrderListsExtra `json:"extra"`                  // 通过当前数据控制按钮是否显示
}

type OrderListsExtra struct { // 通过当前数据控制按钮是否显示
	IsCellReject       bool `json:"is_cell_reject"`        // 是否可拒单
	IsCellRefund       bool `json:"is_cell_refund"`        // 是否可退款
	IsCellCancel       bool `json:"is_cell_cancel"`        // 是否可取消
	IsCellContactRider bool `json:"is_cell_contact_rider"` // 是否可联系骑手
}

type ContactInfo struct {
	Name  string `json:"name"`  // 联系人名称
	Phone string `json:"phone"` // 联系人电话
}

type CachierInfo struct {
	Uuid uint64 `json:"uuid"` // 收银员UUID
	Name string `json:"name"` // 收银员名称
}

type OrderMember struct {
	ID   uint   `json:"id"`   // 会员ID
	Name string `json:"name"` // 会员名称
}

type MemberProductManageList struct {
	List          []MemberOrderManageProduct `json:"list"`           // 订单商品列表
	ProductAmount float64                    `json:"product_amount"` // 商品金额. 所有商品金额总和，如商品A金额为2，商品B金额为3，则总金额为5
}

type MemberOrderManageProduct struct {
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 商品名称
	LocaleAttributeName dto.LocaleResponse `json:"locale_attribute_name"` // 商品属性
	ImageUrl            string             `json:"image_url"`             // 商品图片
	OriginUnitPrice     float64            `json:"origin_unit_price"`     // 商品原单价（折前）
	UnitPrice           float64            `json:"unit_price"`            // 商品单价（折后）
	Num                 float64            `json:"num"`                   // 数量
	TotalPrice          float64            `json:"total_price"`           // 总价. 总价=单价*数量
	OriginTotalPrice    float64            `json:"origin_total_price"`    // 商品原总价（折前）
	RefundAmount        float64            `json:"refund_amount"`         // 退菜金额
}

// 收银端“外送”订单管理页面订单详情
type GetMemberOrderManageDetailResp struct {
	MemberSaleOrderUuid uint64      `json:"member_sale_order_uuid"` // 会员端销售订单UUID
	BillType            uint        `json:"bill_type"`              // 订单类型	0:桌台订单 1:点餐订单 2:外送订单
	SerialNo            string      `json:"serial_no"`              // 订单流水号
	OrderNo             string      `json:"order_no"`               // 订单号
	Status              uint        `json:"status"`                 // 订单状态.0-选购中 1-待付款 2-待商家接单 3-商家备餐中 4-待骑手接单 5-骑手正在赶往商家 6-骑手配送中 7-已完成 8-已取消
	OriginAmount        float64     `json:"origin_amount"`          // 订单原始金额。商品金额+配送费
	PayAmount           float64     `json:"pay_amount"`             // 实付金额。商品金额+配送费-会员折扣
	RefundAmount        float64     `json:"refund_amount"`          // 退款金额
	MemberDiscount      float64     `json:"member_discount"`        // 会员折扣金额
	DeliveryFee         float64     `json:"delivery_fee"`           // 配送费
	PayType             string      `json:"pay_type"`               // 支付方式
	PayTime             int64       `json:"pay_time"`               // 支付时间
	CreateTime          int64       `json:"create_time"`            // 创建时间,下单时间
	FinshTime           int64       `json:"finish_time"`            // 完成时间
	CancelReason        string      `json:"cancel_reason"`          // 取消原因
	CancelTime          int64       `json:"cancel_time"`            // 取消时间
	Remark              string      `json:"remark"`                 // 订单备注
	Member              OrderMember `json:"member"`                 // 会员信息
	Cachier             CachierInfo `json:"cachier"`                // 收银员信息

	ProductList  MemberProductManageList `json:"product_list"`  // 订单商品列表
	OperationLog OperationLog            `json:"operation_log"` // 操作日志
}

// 收银端“外送”接单页面订单搜索
type GetMemberCashierOrderSearchResp struct {
	List []MemberCashierOrder `json:"list"` // 订单列表
}
