package service

import (
	"encoding/json"
	"strconv"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/utils"
)

// OpenDeskPayload 开桌
type OpenDeskPayload struct {
	MealNum  uint   `json:"meal_num"`  // 用餐人数
	IsBuffet bool   `json:"is_buffet"` // 是否是自助餐
	TableId  uint64 `json:"table_id"`  // 桌台ID
	TableNo  string `json:"table_no"`  // 桌台号
}

// SendKitchenPayload 送厨
type SendKitchenPayload struct {
	Products []OrderProduct `json:"products"` // 用餐人数
}
type OrderProduct struct {
	OrderProductId  uint64               `json:"order_product_id"` // 订单商品ID
	ProductId       uint64               `json:"product_id"`       // 商品ID
	ProductName     dto.LocaleResponse   `json:"product_name"`     // 商品名称
	ProductAttr     dto.LocaleResponse   `json:"product_attr"`     // 商品属性, 包含规格、属性、小料
	ProductAttrList []dto.LocaleResponse `json:"product_attrs"`    // 商品属性, 包含规格、属性、小料
	TotalNum        uint                 `json:"total_num"`        // 总数量
	IsBuffet        bool                 `json:"is_buffet"`        // 是否自助餐
	Remark          string               `json:"remark"`           // 备注
}

// RefundProductPayload 退菜
type RefundProductPayload struct {
	OrderProductId  uint64               `json:"order_product_id"` // 订单商品ID
	ProductId       uint64               `json:"product_id"`       // 商品ID
	ProductName     dto.LocaleResponse   `json:"product_name"`     // 商品名称
	ProductAttr     dto.LocaleResponse   `json:"product_attr"`     // 商品属性
	ProductAttrList []dto.LocaleResponse `json:"product_attrs"`    // 商品属性, 包含规格、属性、小料
	TotalNum        uint                 `json:"total_num"`        // 总数量。退菜的数量
	IsBuffet        bool                 `json:"is_buffet"`        // 是否自助餐
	Remark          string               `json:"remark"`           // 备注
	Reason          dto.LocaleResponse   `json:"reason"`           // 退菜原因
	CustomReason    string               `json:"custom_reason"`    // 自定义退菜原因
}

// CancelRefundProductPayload 取消退菜
type CancelRefundProductPayload struct {
	OrderProductId uint64             `json:"order_product_id"` // 订单商品ID
	ProductId      uint64             `json:"product_id"`       // 商品ID
	ProductName    dto.LocaleResponse `json:"product_name"`     // 商品名称
	ProductAttr    dto.LocaleResponse `json:"product_attr"`     // 商品属性
	Num            uint               `json:"num"`              // 退菜数量
	ParentId       uint64             `json:"parent_id"`        // 父订单ID
	OrderName      uint64             `json:"order_name"`       // 订单名称
}

// ChangeTablePayload 转台
type ChangeTablePayload struct {
	Old Table `json:"old"`
	New Table `json:"new"`
}
type Table struct {
	TableId uint64 `json:"table_id"`
	TableNo string `json:"table_no"`
}

// ChangePricePayload 改价
type ChangePricePayload struct {
	OrderProductId uint64             `json:"order_product_id"` // 销售订单产品ID
	ProductId      uint64             `json:"product_id"`       // 产品ID
	ProductName    dto.LocaleResponse `json:"product_name"`     // 产品名称
	ProductAttr    dto.LocaleResponse `json:"product_attr"`     // 产品属性
	TotalNum       uint               `json:"total_num"`        // 数量
	Price          float64            `json:"price"`            // 价格，单价
}

// UpdateMealNumPayload 修改桌台就餐人数
type UpdateMealNumPayload struct {
	OldMealNum uint `json:"old_meal_num"` // 旧桌台就餐人数
	NewMealNum uint `json:"new_meal_num"` // 新桌台就餐人数
}

// ProductFreePayload 赠菜
type ProductFreePayload struct {
	OrderProductId uint64             `json:"order_product_id"` // 订单商品ID
	ProductId      uint64             `json:"product_id"`       // 商品ID
	ProductName    dto.LocaleResponse `json:"product_name"`     // 商品名称
	ProductAttr    dto.LocaleResponse `json:"product_attr"`     // 商品属性
	ProductPrice   float64            `json:"product_price"`    // 商品价格
	TotalNum       uint               `json:"total_num"`        // 总数量
	TotalPrice     float64            `json:"total_price"`      // 总价格
	FreeTagIds     []uint64           `json:"free_tag_ids"`     // 赠菜原因ID
	FreeRemark     string             `json:"free_remark"`      // 赠菜自定义原因
}

// CancelProductFreePayload 取消赠菜
type CancelProductFreePayload struct {
	OrderProductId uint64             `json:"order_product_id"` // 订单商品ID
	ProductId      uint64             `json:"product_id"`       // 商品ID
	ProductName    dto.LocaleResponse `json:"product_name"`     // 商品名称
	ProductAttr    dto.LocaleResponse `json:"product_attr"`     // 商品属性
	ProductPrice   float64            `json:"product_price"`    // 商品价格
	TotalNum       uint               `json:"total_num"`        // 总数量
	TotalPrice     float64            `json:"total_price"`      // 总价格
	ParentId       uint64             `json:"parent_id"`        // 销售账单uuid
	OrderName      uint64             `json:"order_name"`       // 销售订单uuid
}

// ProductMovePayload 转菜
type ProductMovePayload struct {
	OrderProductId uint64             `json:"order_product_id"` // 订单商品ID
	ProductId      uint64             `json:"product_id"`       // 商品ID
	ProductName    dto.LocaleResponse `json:"product_name"`     // 商品名称
	ProductAttr    dto.LocaleResponse `json:"product_attr"`     // 商品属性
	TotalNum       uint               `json:"total_num"`        // 总数量
	ToOrderId      uint64             `json:"to_order_id"`      // 目标订单ID
	ToTableId      uint64             `json:"to_table_id"`      // 目标桌台ID
	ToTableNo      string             `json:"to_table_no"`      // 目标桌台编号
}

// DiscountPayload 优惠折扣
type DiscountPayload struct {
	OldPrice        float64 `json:"old_price"`        // 进行整单打折前的总金额
	NewPrice        float64 `json:"new_price"`        // 整单打折后的总金额
	DiscountType    int     `json:"discount_type"`    // 折扣类型。1: 订单改价 2: 折扣 3:抹零
	SpecialDiscount float64 `json:"special_discount"` // 优惠金额。整单打折后的优惠金额=会员折扣后的订单应收金额-订单应收金额
	RoundingRate    float64 `json:"rounding_rate"`    // 打折率。如八折，则打折率是20； 如30%off，则打折率是30。统一展示格式为“优惠折扣：折扣-80%（￥50）”，无论是百分比打折还是百分比减免，都统一展示为百分比减免。
	RoundingType    int     `json:"rounding_type"`    // 抹零规则 1:抹分 2:抹角 3:四舍五入保留一位小数 4:四舍五入到整数
}

type FreeSalePayload struct {
	OrderPrice float64 `json:"order_price"` // 订单金额
}

// SettlePayload 结账
type SettlePayload struct {
	OrderPrice  float64         `json:"order_price"`  // 订单应付金额
	PayPrice    float64         `json:"pay_price"`    // 最终应付金额。最终应付金额=订单应付金额+手续费（手续费=每笔付款单的手续费之和）
	ActualPrice float64         `json:"actual_price"` // 最终实付金额。最终实付金额=最终应付金额+找零金额。如果没有找零，则最终实付金额=最终应付金额。最终实付金额=每笔付款单的付款金额之和（含手续费）
	ChangeDue   float64         `json:"change_due"`   // 找零金额
	PayType     []SettlePayType `json:"pay_type"`     // 支付方式
	IsFree      bool            `json:"is_free"`      // 是否免单
}

type SettlePayType struct {
	Name           string  `json:"name"`            // 支付方式名称
	Value          int     `json:"value"`           // 支付方式值
	DisabledCancel uint    `json:"disabled_cancel"` // 是否禁用取消
	Price          float64 `json:"price"`           // 支付金额（含手续费）
	FeeMoney       float64 `json:"fee_money"`       // 手续费
}

// ReverseSettlePayload 反结账
type ReverseSettlePayload struct {
	PayType []SettlePayType `json:"pay_type"` // 支付方式
}

type ReturnOrderProduct struct {
	OrderProductId  uint64               `json:"order_product_id"` // 订单商品ID
	ProductId       uint64               `json:"product_id"`       // 商品ID
	ProductName     dto.LocaleResponse   `json:"product_name"`     // 商品名称
	ProductAttr     dto.LocaleResponse   `json:"product_attr"`     // 商品属性, 包含规格、属性、小料
	ProductAttrList []dto.LocaleResponse `json:"product_attrs"`    // 商品属性, 包含规格、属性、小料
	TotalNum        uint                 `json:"total_num"`        // 总数量
	IsBuffet        bool                 `json:"is_buffet"`        // 是否自助餐
	Remark          string               `json:"remark"`           // 备注
}

type RefundPayType struct {
	Name          string  `json:"name"`           // 退款支付方式名称
	Code          int     `json:"code"`           // 退款支付方式代号
	Amount        float64 `json:"amount"`         // 退款金额
	PaymentStatus int     `json:"payment_status"` // 支付状态
}

// RefundPayload 退款
type RefundPayload struct {
	Products     []ReturnOrderProduct `json:"products"`       // 退款商品
	PayTypes     []RefundPayType      `json:"pay_type"`       // 支付方式
	RefundType   int                  `json:"refund_type"`    // 退款方式：1-整单退款；2-部分退款
	IsSplitOrder bool                 `json:"is_split_order"` // 是否拆单
	Index        int                  `json:"index"`          // 子单索引
}

// MergeTablePayload 并台
type MergeTablePayload struct {
	DeskNos []string
}

// OrderCheckoutDiscountPayload 结账手动抹零
type OrderCheckoutDiscountPayload struct {
	Operation       string  `json:"operation"`        // 操作类型。add: 设置结账抹零，cancel: 撤销结账抹零
	RoundingType    int     `json:"rounding_type"`    // 抹零规则。0-实款实收 1-抹分 2-抹角 5-抹元
	SpecialDiscount float64 `json:"special_discount"` // 优惠金额
	Reason          string  `json:"reason"`           // 原因(撤销时使用)
}

// OrderTakingPayload 接单
type OrderTakingPayload struct {
	IsAutoOrder bool   `json:"is_auto_order"` // 是否自动接单
	H5OrderUuid uint64 `json:"h5_order_uuid"` // h5订单Uuid
}

// SplitOrderPayload 拆单
type SplitOrderPayload struct {
	Orders []Order `json:"orders"`
}

type Order struct {
	SaleOrderUuid uint64  `json:"sale_order_uuid"` // 销售订单Uuid
	OrderName     string  `json:"order_name"`      // 订单名称，顺序
	Amount        float64 `json:"amount"`          // 订单金额
}

func (s *orderSrv) getActionDescription(ctx context.Context, log model.SaleOrderOperationRecord, language string) string {

	currencySetting, _ := s.settingSrv.GetCurrencySetting(ctx)

	switch log.Action {
	case constant.OrderOpenTable: // 开台
		var openTable OpenDeskPayload
		err := json.Unmarshal([]byte(log.Data), &openTable)
		if err == nil {
			return openTable.TableNo + ", " + i18n.Translate(language, "人数") + ": " + strconv.Itoa(int(openTable.MealNum))
		}
	case constant.OrderSendKitchen: // 送厨
		var sendKitchen SendKitchenPayload
		if err := json.Unmarshal([]byte(log.Data), &sendKitchen); err == nil {
			var desc []string
			for _, product := range sendKitchen.Products {
				desc = append(desc, product.ProductName.GetLocale(language)+" ("+product.ProductAttr.GetLocale(language)+") *"+
					strconv.Itoa(int(product.TotalNum)))
			}
			return strings.Join(desc, "、")
		}
	case constant.OrderRefundProduct: // 退菜
		var returnProduct RefundProductPayload
		if err := json.Unmarshal([]byte(log.Data), &returnProduct); err == nil {
			desc := returnProduct.ProductName.GetLocale(language) + " (" + returnProduct.ProductAttr.GetLocale(language) + ")*" +
				strconv.Itoa(int(returnProduct.TotalNum)) +
				" (" + i18n.Translate(language, "原因") + ": " + returnProduct.Reason.GetLocale(language)

			if returnProduct.CustomReason != "" {
				desc = desc + "、" + returnProduct.CustomReason
			}
			desc = desc + ")"
		}
	case constant.OrderCancelRefundProduct: // 取消退菜
		var cancelRefundProduct CancelRefundProductPayload
		if err := json.Unmarshal([]byte(log.Data), &cancelRefundProduct); err == nil {
			return cancelRefundProduct.ProductName.GetLocale(language) + " (" + cancelRefundProduct.ProductAttr.GetLocale(language) + ") *" +
				strconv.Itoa(int(cancelRefundProduct.Num))
		}
	case constant.OrderChangeTable: // 转台
		var changeTable ChangeTablePayload
		if err := json.Unmarshal([]byte(log.Data), &changeTable); err == nil {
			return changeTable.Old.TableNo + "->" + changeTable.New.TableNo
		}
	case constant.OrderChangePrice: // 改价
		var changePrice ChangePricePayload
		if err := json.Unmarshal([]byte(log.Data), &changePrice); err == nil {
			return changePrice.ProductName.GetLocale(language) + " (" + changePrice.ProductAttr.GetLocale(language) + ") *" +
				strconv.Itoa(int(changePrice.TotalNum)) + " (" + currencySetting.Unit + utils.FormatFloat(changePrice.Price) + ")"
		}
	case constant.OrderUpdateMealNum: // 修改桌台就餐人数
		var updateMealNum UpdateMealNumPayload
		if err := json.Unmarshal([]byte(log.Data), &updateMealNum); err == nil {
			return strconv.Itoa(int(updateMealNum.NewMealNum))
		}
	case constant.OrderStayOrder: // 挂单 不用解析data
	case constant.OrderPickOrder: // 取单 不用解析data
	case constant.OrderProductFree: // 赠菜
		var productFree ProductFreePayload
		if err := json.Unmarshal([]byte(log.Data), &productFree); err == nil {
			return productFree.ProductName.GetLocale(language) + " (" + productFree.ProductAttr.GetLocale(language) + ") *" + strconv.Itoa(int(productFree.TotalNum)) +
				" (" + currencySetting.Unit + utils.FormatFloat(productFree.TotalPrice) + ")"
		}
	case constant.OrderCancelProductFree: // 取消赠菜
		var cancelProductFree CancelProductFreePayload
		if err := json.Unmarshal([]byte(log.Data), &cancelProductFree); err == nil {
			return cancelProductFree.ProductName.GetLocale(language) + " (" + cancelProductFree.ProductAttr.GetLocale(language) + ") *" + strconv.Itoa(int(cancelProductFree.TotalNum)) +
				" (" + currencySetting.Unit + utils.FormatFloat(cancelProductFree.TotalPrice) + ")"
		}
	case constant.OrderProductMove: // 转菜
		var productMove ProductMovePayload
		if err := json.Unmarshal([]byte(log.Data), &productMove); err == nil {
			return productMove.ProductName.GetLocale(language) + " (" + productMove.ProductAttr.GetLocale(language) + ") (" + i18n.Translate(language, "转至") +
				productMove.ToTableNo + ")"
		}
	case constant.OrderDiscount: // 优惠折扣
		var discount DiscountPayload
		if err := json.Unmarshal([]byte(log.Data), &discount); err == nil {
			var desc string
			switch discount.DiscountType {
			case constant.DiscountOperationLogTypeChangePriceSaleOrder: // 改价
				desc = utils.FormatFloat(discount.OldPrice) + i18n.Translate(language, "改价") +
					utils.FormatFloat(discount.NewPrice) + " (" + currencySetting.Unit + utils.FormatFloat(discount.SpecialDiscount) + ")"
			case constant.DiscountOperationLogTypeDiscountSaleOrder: // 折扣
				desc = i18n.Translate(language, "折扣") + "-" + utils.FormatFloat(discount.RoundingRate) +
					"% (" + currencySetting.Unit + utils.FormatFloat(discount.SpecialDiscount) + ")"
			case constant.DiscountOperationLogTypeZeroSaleOrder: // 抹零
				roundingTypeMap := map[int]string{
					constant.DiscountZeroRulePercent: "抹分",
					constant.DiscountZeroRuleFixed:   "抹角",
					constant.DiscountZeroRuleRound:   "四舍五入保留一位小数",
					constant.DiscountZeroRuleInteger: "四舍五入到整数",
				}
				desc = i18n.Translate(language, "抹零") + "-" + i18n.Translate(language, roundingTypeMap[discount.RoundingType]) +
					" (" + currencySetting.Unit + utils.FormatFloat(discount.SpecialDiscount) + ")"
			}
			return desc
		}
	case constant.OrderCancelDiscount: // 撤销优惠折扣 不需要解析data
	case constant.OrderFreeSale: // 免单
		var freeSale FreeSalePayload
		if err := json.Unmarshal([]byte(log.Data), &freeSale); err == nil {
			return i18n.Translate(language, "免单") + " (" + currencySetting.Unit + utils.FormatFloat(freeSale.OrderPrice) + ")"
		}
	case constant.OrderSettle: // 结账
		var settle SettlePayload
		if err := json.Unmarshal([]byte(log.Data), &settle); err == nil {
			var payTypeList []string
			if settle.IsFree {
				payTypeList = append(payTypeList, i18n.Translate(language, "免单")+": "+currencySetting.Unit+utils.FormatFloat(settle.OrderPrice))
			}
			for _, payType := range settle.PayType {
				payTypeName := payType.Name
				if payType.Value == constant.PaymentMethodCodeCash {
					payType.Price = utils.DecimalSub(payType.Price, settle.ChangeDue)
				}
				payTypeList = append(payTypeList, payTypeName+": "+currencySetting.Unit+utils.FormatFloat(payType.Price))
			}
			desc := i18n.Translate(language, "订单金额") + " " + currencySetting.Unit + utils.FormatFloat(settle.OrderPrice) + "，" +
				i18n.Translate(language, "实付金额") + " " + currencySetting.Unit + utils.FormatFloat(settle.ActualPrice)
			if len(payTypeList) > 0 {
				desc = desc + " (" + strings.Join(payTypeList, "、") + ")"
			}
			return desc
		}
	case constant.OrderReverseSettle: // 反结账
		var reverseSettle ReverseSettlePayload
		if err := json.Unmarshal([]byte(log.Data), &reverseSettle); err == nil {
			var payTypeList []string
			for _, payType := range reverseSettle.PayType {
				payTypeName := payType.Name
				if payType.Value == constant.PaymentMethodCodeFreePay {
					payTypeName = i18n.Translate(language, "免单")
				}
				payTypeList = append(payTypeList, payTypeName+": "+currencySetting.Unit+utils.FormatFloat(payType.Price))
			}
			return strings.Join(payTypeList, "、")
		}
	case constant.OrderRefund: // 退款
		var refundPayload RefundPayload
		if err := json.Unmarshal([]byte(log.Data), &refundPayload); err == nil {
			var desc []string
			for _, product := range refundPayload.Products {
				desc = append(desc, product.ProductName.GetLocale(language)+" ("+product.ProductAttr.GetLocale(language)+") *"+
					strconv.Itoa(int(product.TotalNum)))
			}
			return strings.Join(desc, "、")
		}
	case constant.OrderOrderTaking: // 接单 不需要解析data
	case constant.OrderOrderReject: // 拒单 不需要解析data
	case constant.OrderMergeTable:
		var mergeTable MergeTablePayload
		if err := json.Unmarshal([]byte(log.Data), &mergeTable); err == nil {
			return strings.Join(mergeTable.DeskNos, "、")
		}
	case constant.OrderOrderCancel: // 整单取消 不需要解析data
	case constant.OrderCheckoutDiscount: // 结账手动抹零
		var orderCheckoutDiscount OrderCheckoutDiscountPayload
		if err := json.Unmarshal([]byte(log.Data), &orderCheckoutDiscount); err == nil {
			var desc string
			if orderCheckoutDiscount.Operation == constant.OrderCheckoutDiscountCancel {
				desc = i18n.Translate(language, "自动取消")
				if orderCheckoutDiscount.Reason != "" {
					desc = desc + " (" + i18n.Translate(language, "原因") + "：" + i18n.Translate(language, orderCheckoutDiscount.Reason) + ")"
				}
			} else {
				discountTypeMap := map[int]string{
					constant.SaleBillSettingCheckoutZeroingMethodNone:    "实款实收",
					constant.SaleBillSettingCheckoutZeroingMethodPercent: "抹分",
					constant.SaleBillSettingCheckoutZeroingMethodFixed:   "抹角",
					constant.SaleBillSettingCheckoutZeroingMethodYuan:    "抹元",
				}
				desc = i18n.Translate(language, discountTypeMap[orderCheckoutDiscount.RoundingType]) +
					" (" + currencySetting.Unit + utils.FormatFloat(orderCheckoutDiscount.SpecialDiscount) + ")"
			}
			return desc
		}
	case constant.OrderSplitOrder: // 拆单
		var splitOrder SplitOrderPayload
		if err := json.Unmarshal([]byte(log.Data), &splitOrder); err == nil {
			var orderDetails []string
			for i, order := range splitOrder.Orders {
				orderDetails = append(orderDetails, strconv.Itoa(i+1)+"（"+i18n.Translate(language, "订单金额")+"："+currencySetting.Unit+utils.FormatFloat(order.Amount)+"）")
			}
			return strings.Join(orderDetails, ", ")
		}
	case constant.OrderCancelSplitOrder: // 撤销拆单 不需要解析data
	}
	return ""
}

func (s *orderSrv) getActionText(log model.SaleOrderOperationRecord, language string) string {

	actionTextMap := map[string]string{
		constant.OrderOpenTable:           i18n.Translate(language, "开台"),
		constant.OrderSendKitchen:         i18n.Translate(language, "送厨"),
		constant.OrderRefundProduct:       i18n.Translate(language, "退菜"),
		constant.OrderCancelRefundProduct: i18n.Translate(language, "取消退菜"),
		constant.OrderChangeTable:         i18n.Translate(language, "转台"),
		constant.OrderChangePrice:         i18n.Translate(language, "改价"),
		constant.OrderUpdateMealNum:       i18n.Translate(language, "人数"),
		constant.OrderStayOrder:           i18n.Translate(language, "挂单"),
		constant.OrderPickOrder:           i18n.Translate(language, "取单"),
		constant.OrderProductFree:         i18n.Translate(language, "赠菜"),
		constant.OrderCancelProductFree:   i18n.Translate(language, "取消赠菜"),
		constant.OrderProductMove:         i18n.Translate(language, "转菜"),
		constant.OrderDiscount:            i18n.Translate(language, "优惠折扣"),
		constant.OrderFreeSale:            i18n.Translate(language, "优惠折扣"),
		constant.OrderCancelDiscount:      i18n.Translate(language, "撤销优惠折扣"),
		constant.OrderSettle:              i18n.Translate(language, "结账"),
		constant.OrderReverseSettle:       i18n.Translate(language, "反结账"),
		constant.OrderRefund:              i18n.Translate(language, "部分退款"), // 默认部分退款
		constant.OrderOrderTaking:         i18n.Translate(language, "接单"),   // 默认非自动接单
		constant.OrderOrderReject:         i18n.Translate(language, "拒单"),
		constant.OrderMergeTable:          i18n.Translate(language, "并台"),
		constant.OrderOrderCancel:         i18n.Translate(language, "整单取消"),
		constant.OrderCheckoutDiscount:    i18n.Translate(language, "结账手动抹零"),
		constant.OrderSplitOrder:          i18n.Translate(language, "拆单"),
		constant.OrderCancelSplitOrder:    i18n.Translate(language, "撤销拆单"),
	}

	var text, prefix string
	var ok bool
	if text, ok = actionTextMap[log.Action]; !ok {
		return i18n.Translate(language, "未知操作")
	}
	// 拆单前缀、退款类型
	type Common struct {
		IsSplitOrder bool `json:"is_split_order"` // 是否免单
		Index        int  `json:"index"`          // 子单索引
		RefundType   int  `json:"refund_type"`    // 退款类型：1-整单退款、2-部分退款
	}
	var common Common
	json.Unmarshal([]byte(log.Data), &common)
	if common.IsSplitOrder && common.Index > 0 {
		prefix = "（" + i18n.Translate(language, "拆单") + strconv.Itoa(common.Index) + "）"
	}
	if log.Action == constant.OrderRefund && common.RefundType == 1 {
		return prefix + i18n.Translate(language, "整单退款")
	}
	if log.Action == constant.OrderOrderTaking {
		var orderTaking OrderTakingPayload
		json.Unmarshal([]byte(log.Data), &orderTaking)
		if orderTaking.IsAutoOrder {
			return i18n.Translate(language, "系统自动接单")
		}
	}

	return prefix + text
}

func (s *orderSrv) getRefundPayType(ctx context.Context, log model.SaleOrderOperationRecord, language string) []resp.OrderOperationLogPaymentMethod {
	refundPayTypes := make([]resp.OrderOperationLogPaymentMethod, 0)
	if log.Action != constant.OrderRefund {
		return refundPayTypes
	}
	var refundPayload RefundPayload
	currencySetting, _ := s.settingSrv.GetCurrencySetting(ctx)
	if err := json.Unmarshal([]byte(log.Data), &refundPayload); err == nil {
		for _, payType := range refundPayload.PayTypes {
			refundPayTypes = append(refundPayTypes, resp.OrderOperationLogPaymentMethod{
				Code:          payType.Code,
				Name:          payType.Name,
				RefundMoney:   utils.FormatFloat(payType.Amount),
				PaymentStatus: payType.PaymentStatus,
				Unit:          currencySetting.Unit,
			})
		}
	}
	return refundPayTypes
}
