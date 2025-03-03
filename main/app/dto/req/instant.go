package req

import (
	"errors"
	"strconv"
	"ttpos-server-go/app/dto"

	"github.com/gin-gonic/gin"
)

// InstantOrderGetInfoReq 获取点餐订单详情请求
type InstantOrderGetInfoReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID
}

// AddProduct 添加商品请求
type AddProduct struct {
	Uuid       uint64                `json:"uuid"`        // 商品UUID, 必填
	FlavorUuid uint64                `json:"flavor_uuid"` // 规格UUID, 必填
	SauceUuids []uint64              `json:"sauce_uuids"` // 小料UUID, 非必填
	Attributes []AddProductAttribute `json:"attributes"`  // 商品属性, 非必填
}

// AddProductAttribute 添加商品属性
type AddProductAttribute struct {
	GroupUuid  uint64   `json:"group_uuid"`  // 属性组UUID
	ValueUuids []uint64 `json:"value_uuids"` // 属性值UUID
}

// InstantOrderAddProductReq 添加商品请求
type InstantOrderAddProductReq struct {
	SaleBillUuid  uint64     `json:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64     `json:"sale_order_uuid"` // 销售订单UUID, 必填
	Product       AddProduct `json:"product"`         // 商品, 必填
}

// InstantOrderPaymentInfoReq 结账页面信息请求
type InstantOrderPaymentInfoReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID, 必填
}

func (r *InstantOrderPaymentInfoReq) Parse(c *gin.Context) error {
	saleBillUuidStr := c.Query("sale_bill_uuid")
	saleOrderUuidStr := c.Query("sale_order_uuid")
	if saleBillUuidStr == "" || saleOrderUuidStr == "" {
		return errors.New("销售账单UUID和销售订单UUID不能为空")
	}

	saleBillUuid, err := strconv.ParseUint(saleBillUuidStr, 10, 64)
	if err != nil {
		return errors.New("销售账单UUID格式错误")
	}

	saleOrderUuid, err := strconv.ParseUint(saleOrderUuidStr, 10, 64)
	if err != nil {
		return errors.New("销售订单UUID格式错误")
	}

	r.SaleBillUuid = saleBillUuid
	r.SaleOrderUuid = saleOrderUuid

	return nil
}

// InstantOrderPaymentCreateReq 创建一个支付单请求
type InstantOrderPaymentCreateReq struct {
	SaleBillUuid      uint64  `json:"sale_bill_uuid"`      // 销售账单UUID, 必填
	SaleOrderUuid     uint64  `json:"sale_order_uuid"`     // 销售订单UUID, 必填
	PaymentMethodUuid uint64  `json:"payment_method_uuid"` // 支付方式UUID, 必填
	PaymentAmount     float64 `json:"payment_amount"`      // 支付金额, 必填
}

// InstantOrderPaymentFinishReq 完成销售订单的付款结账请求
type InstantOrderPaymentFinishReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID, 必填
}

type InstantOrderSaleOrderCreateReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID, 必填
}

// InstantOrderSaleOrderMoveProductReq 从一个销售订单移动商品到另一个销售订单请求
type InstantOrderSaleOrderMoveProductReq struct {
	SaleBillUuid uint64        `json:"sale_bill_uuid"` // 销售账单UUID, 必填
	From         uint64        `json:"from"`           // 来源销售订单UUID, 必填
	To           uint64        `json:"to"`             // 目标销售订单UUID, 必填
	Products     []MoveProduct `json:"products"`       // 移动商品, 必填
}

// MoveProduct 移动商品
type MoveProduct struct {
	Uuid uint64 `json:"uuid"` // 销售订单商品UUID, 必填
	Num  uint   `json:"num"`  // 移动数量, 必填
}

type InstantOrderMustPlanConfirmReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID, 必填
}

type InstantOrderSaleOrderDeleteReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID, 必填
}

type InstantOrderSaleOrderDeleteAllReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID, 必填
}

type HideSaleBillListReq struct {
	dto.PageReq // 分页参数
}
