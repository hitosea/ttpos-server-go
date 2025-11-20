package req

import (
	"fmt"
	"strconv"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"

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
	BatchTagUuid  uint64     `json:"batch_tag_uuid"`  // 分批类型UUID, 可选（前置模式时使用）
}

// InstantOrderPaymentPointsReq 设置订单的抵扣积分数量请求
type InstantOrderPaymentPointsReq struct {
	SaleBillUuid  uint64  `json:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64  `json:"sale_order_uuid"` // 销售订单UUID, 必填
	Points        float64 `json:"points"`          // 抵扣积分数量, 必填
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
		return errors.WithMessage(err, "销售账单UUID格式错误")
	}

	saleOrderUuid, err := strconv.ParseUint(saleOrderUuidStr, 10, 64)
	if err != nil {
		return errors.WithMessage(err, "销售订单UUID格式错误")
	}

	r.SaleBillUuid = saleBillUuid
	r.SaleOrderUuid = saleOrderUuid

	return nil
}

type InstantOrderPaymentCouponReq struct {
	SaleBillUuid      uint64 `json:"sale_bill_uuid"`     // 销售账单UUID, 必填
	SaleOrderUuid     uint64 `json:"sale_order_uuid"`    // 销售订单UUID, 必填
	CouponUuid        uint64 `json:"coupon_uuid"`        // 优惠券UUID, 必填。通用优惠券或会员优惠券的uuid
	CouponRequirement string `json:"coupon_requirement"` // 优惠券类型, 必填。通用优惠券“none”或会员优惠券“marketing”
}

// InstantOrderPaymentCreateReq 创建一个支付单请求
type InstantOrderPaymentCreateReq struct {
	SaleBillUuid      uint64  `json:"sale_bill_uuid"`      // 销售账单UUID, 必填
	SaleOrderUuid     uint64  `json:"sale_order_uuid"`     // 销售订单UUID, 必填
	PaymentMethodUuid uint64  `json:"payment_method_uuid"` // 支付方式UUID, 必填
	PaymentAmount     float64 `json:"payment_amount"`      // 支付金额, 必填
	PaymentOrderUuid  uint64  `json:"payment_order_uuid"`  // 支付单UUID, 非必填
}

// InstantOrderPaymentCancelReq 撤销一个支付单请求
type InstantOrderPaymentCancelReq struct {
	SaleBillUuid     uint64 `json:"sale_bill_uuid"`     // 销售账单UUID, 必填
	SaleOrderUuid    uint64 `json:"sale_order_uuid"`    // 销售订单UUID, 必填
	PaymentOrderUuid uint64 `json:"payment_order_uuid"` // 支付单UUID, 必填
}

// InstantOrderPaymentFinishReq 完成销售订单的付款结账请求
type InstantOrderPaymentFinishReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID, 必填
}

// InstantOrderFreeReq 免单请求
type InstantOrderFreeReq struct {
	SaleBillUuid  uint64   `json:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64   `json:"sale_order_uuid"` // 销售订单UUID, 必填
	ReasonIds     []uint64 `json:"reason_ids"`      // 免单原因标签ids
	Reason        string   `json:"reason"`          // 原因
}

// InstantOrderPaymentZeroRuleReq 设置结账抹零规则请求
type InstantOrderPaymentZeroRuleReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID, 必填
	ZeroRule      int    `json:"zero_rule"`       // 结账抹零规则, 必填
}

func (r *InstantOrderPaymentZeroRuleReq) Validate() error {
	if r.SaleBillUuid == 0 || r.SaleOrderUuid == 0 {
		return errors.New("SaleBillUuid或SaleOrderUuid不能为0")
	}
	// 兼容处理，如果ZeroRule为3，则设置为5
	if r.ZeroRule == constant.SaleBillSettingCheckoutZeroingMethodYuanAbandon {
		r.ZeroRule = constant.SaleBillSettingCheckoutZeroingMethodYuan
	}
	if r.ZeroRule == constant.SaleBillSettingCheckoutZeroingMethodNone ||
		r.ZeroRule == constant.SaleBillSettingCheckoutZeroingMethodPercent ||
		r.ZeroRule == constant.SaleBillSettingCheckoutZeroingMethodFixed ||
		r.ZeroRule == constant.SaleBillSettingCheckoutZeroingMethodYuan {
		return nil
	}
	return errors.WithMessage(errors.New("结账抹零规则错误"), fmt.Sprintf("ZeroRule: %d", r.ZeroRule))
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
	Uuid uint64  `json:"uuid"` // 销售订单商品UUID, 必填. 也能是顾客uuid、加钟uuid
	Num  float64 `json:"num"`  // 移动数量, 必填
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
	MemberUuid   uint64 `json:"member_uuid"`    // 会员UUID, 可选
}

type HideSaleBillListReq struct {
	dto.PageReq // 分页参数
}

type OrderTakeoutReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID, 必填
	Takeout      bool   `json:"takeout"`        // 是否打包,true：打包，false：不打包，默认堂食。 必填
}

type OrderMemberCancelReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID, 必填
}

type OrderPrintReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID, 必填
	PrintLang     string `json:"print_lang"`      // 打印语言, 可选
	PayMethodUuid uint64 `json:"pay_method_uuid"` // 支付方式UUID, 可选 (打印码时用)
}

type OrderPrintInvoiceReq struct {
	SaleBillUuid  uint64 `json:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64 `json:"sale_order_uuid"` // 销售订单UUID, 必填
	PrintLang     string `json:"print_lang"`      // 打印语言, 可选
	// 发票信息字段
	CompanyName      string `json:"company_name"`       // 公司名称
	CompanyAddr      string `json:"company_addr"`       // 公司地址
	CompanyTaxNumber string `json:"company_tax_number"` // 公司税号
	CompanyPhone     string `json:"company_phone"`      // 公司电话
}

func (r *OrderPrintInvoiceReq) Validate() error {
	if r.SaleBillUuid == 0 {
		return errors.New("销售账单UUID不能为空")
	}
	return nil
}

type InstantOrderCheckReq struct {
	SaleBillUuid  uint64 `form:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64 `form:"sale_order_uuid"` // 销售订单UUID, 必填
	IgnoreMust    bool   `form:"ignore_must"`     // 是否忽略必点，可选
}

type OrderUnlockReq struct {
	SaleBillUuid uint64 `json:"sale_bill_uuid"` // 销售账单UUID, 必填
}

type OrderInvoiceInfoReq struct {
	SaleBillUuid  uint64 `form:"sale_bill_uuid"`  // 销售账单UUID, 必填
	SaleOrderUuid uint64 `form:"sale_order_uuid"` // 销售订单UUID, 必填
}

type OrderCartInfoReq struct {
	SaleBillUuid uint64 `form:"sale_bill_uuid" json:"sale_bill_uuid"` // 销售账单UUID, 必填
	H5OrderUuid  uint64 `form:"h5_order_uuid" json:"h5_order_uuid"`   // H5订单UUID, 可选。处理扫码接单进入桌台时使用
}

// InstantOrderPaymentQrcodeReq
type InstantOrderPaymentQrcodeReq struct {
	SaleBillUuid      uint64  `form:"sale_bill_uuid" json:"sale_bill_uuid"`           // 销售账单UUID, 必填
	SaleOrderUuid     uint64  `form:"sale_order_uuid" json:"sale_order_uuid"`         // 销售订单UUID, 必填
	PaymentMethodUuid uint64  `form:"payment_method_uuid" json:"payment_method_uuid"` // 支付方式UUID, 必填
	PaymentAmount     float64 `form:"payment_amount" json:"payment_amount"`           // 支付金额, 必填
}

func (r *InstantOrderPaymentQrcodeReq) Validate() error {
	if r.SaleBillUuid == 0 || r.SaleOrderUuid == 0 {
		return errors.New("SaleBillUuid或SaleOrderUuid不能为0")
	}
	if r.PaymentMethodUuid == 0 {
		return errors.New("PaymentMethodUuid不能为空")
	}
	if r.PaymentAmount <= 0 {
		return errors.New("支付金额错误")
	}
	if r.PaymentAmount > 200000 {
		return errors.New("最大支付金额为200000")
	}
	if r.PaymentAmount < 1 {
		return errors.New("最小支付金额为1")
	}
	return nil
}

type OrderGetOrderMemberListReq struct {
	SaleBillUuid uint64 `form:"sale_bill_uuid"` // 销售账单UUID, 必填
}

type GetOrderCartProductBatchCookingListReq struct {
	SaleBillUuid uint64 `form:"sale_bill_uuid"` // 销售账单UUID, 必填
}

type OrderCartProductBatchCookingReq struct {
	SaleBillUuid          uint64   `json:"sale_bill_uuid"`           // 销售账单UUID, 必填
	SaleOrderUuid         uint64   `json:"sale_order_uuid"`          // 销售订单UUID, 必填
	SaleOrderProductUuids []uint64 `json:"sale_order_product_uuids"` // 销售订单商品UUID列表, 必填
	BatchTagUuid          uint64   `json:"batch_tag_uuid"`           // 分批类型UUID, 必填
}

func (r *OrderCartProductBatchCookingReq) Validate() error {
	if r.SaleBillUuid == 0 {
		return errors.New("销售账单UUID不能为空")
	}
	if r.SaleOrderUuid == 0 {
		return errors.New("销售订单UUID不能为空")
	}
	if len(r.SaleOrderProductUuids) == 0 {
		return errors.New("销售订单商品UUID列表不能为空")
	}
	if r.BatchTagUuid == 0 {
		return errors.New("分批类型UUID不能为空")
	}
	return nil
}

// ChangeBatchTagReq 更换分批类型请求
type ChangeBatchTagReq struct {
	SaleBillUuid          uint64   `json:"sale_bill_uuid" binding:"required"`           // 销售账单UUID
	SaleOrderProductUuids []uint64 `json:"sale_order_product_uuids" binding:"required"` // 销售订单商品UUID列表
	BatchTagUuid          uint64   `json:"batch_tag_uuid" binding:"required"`          // 分批类型UUID
}

func (r *ChangeBatchTagReq) Validate() error {
	if r.SaleBillUuid == 0 {
		return errors.New("销售账单UUID不能为空")
	}
	if len(r.SaleOrderProductUuids) == 0 {
		return errors.New("销售订单商品UUID列表不能为空")
	}
	if r.BatchTagUuid == 0 {
		return errors.New("分批类型UUID不能为空")
	}
	return nil
}
