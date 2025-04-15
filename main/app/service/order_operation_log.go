package service

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/utils"
)

// DiscountPayload 优惠折扣
type DiscountPayload struct {
	OldPrice        float64 `json:"old_price"`        // 进行整单打折前的总金额
	NewPrice        float64 `json:"new_price"`        // 整单打折后的总金额
	DiscountType    int     `json:"discount_type"`    // 折扣类型。1: 订单改价 2: 折扣 3:抹零
	SpecialDiscount float64 `json:"special_discount"` // 优惠金额。整单打折后的优惠金额=会员折扣后的订单应收金额-订单应收金额
	RoundingRate    float64 `json:"rounding_rate"`    // 打折率。如八折，则打折率是20； 如30%off，则打折率是30。统一展示格式为“优惠折扣：折扣-80%（￥50）”，无论是百分比打折还是百分比减免，都统一展示为百分比减免。
	RoundingType    int     `json:"rounding_type"`    // 抹零规则 1:抹分 2:抹角 3:四舍五入保留一位小数 4:四舍五入到整数
}

func (s *orderSrv) getActionDescription(ctx context.Context, log model.SaleOrderOperationRecord, language string) string {

	currencySetting, _ := s.settingSrv.GetCurrencySetting(ctx)

	switch log.Action {
	case constant.OrderOpenTable: // 开台
		var openTable event.OpenDeskPayload
		err := json.Unmarshal([]byte(log.Data), &openTable)
		if err == nil {
			return openTable.TableNo + ", " + i18n.Translate(language, "人数") + ": " + strconv.Itoa(int(openTable.MealNum))
		}
	case constant.OrderSendKitchen: // 送厨
		var sendKitchen event.SentCookingPayload
		if err := json.Unmarshal([]byte(log.Data), &sendKitchen); err == nil {
			var desc []string
			for _, product := range sendKitchen.Products {
				desc = append(desc, product.ProductName.GetLocale(language)+" ("+product.ProductAttr.GetLocale(language)+") *"+
					strconv.Itoa(int(product.TotalNum)))
			}
			return strings.Join(desc, "、")
		}
	case constant.OrderRefundProduct: // 退菜
		var returnProduct event.CancelSaleOrderProductPayload
		if err := json.Unmarshal([]byte(log.Data), &returnProduct); err == nil {
			desc := returnProduct.ProductName.GetLocale(language) + " (" + returnProduct.ProductAttr.GetLocale(language) + ")*" +
				strconv.Itoa(int(returnProduct.TotalNum)) +
				" (" + i18n.Translate(language, "原因") + ": "
			reason := returnProduct.Reason.GetLocale(language)
			if reason != "" {
				if returnProduct.CustomReason != "" {
					desc = desc + reason + "、" + returnProduct.CustomReason
				} else {
					desc = desc + reason
				}
			} else if returnProduct.CustomReason != "" {
				desc = desc + returnProduct.CustomReason
			}
			desc = desc + ")"
			return desc
		}
	case constant.OrderCancelRefundProduct: // 取消退菜
		var cancelRefundProduct event.CancelReturnSaleOrderProductPayload
		if err := json.Unmarshal([]byte(log.Data), &cancelRefundProduct); err == nil {
			return cancelRefundProduct.ProductName.GetLocale(language) + " (" + cancelRefundProduct.ProductAttr.GetLocale(language) + ") *" +
				strconv.Itoa(int(cancelRefundProduct.Num))
		}
	case constant.OrderChangeTable: // 转台
		var changeTable event.ChangeDeskPayload
		if err := json.Unmarshal([]byte(log.Data), &changeTable); err == nil {
			return changeTable.Old.TableNo + "->" + changeTable.New.TableNo
		}
	case constant.OrderChangePrice: // 改价
		var changePrice event.ChangeSaleOrderProductPricePayload
		if err := json.Unmarshal([]byte(log.Data), &changePrice); err == nil {
			return changePrice.ProductName.GetLocale(language) + " (" + changePrice.ProductAttr.GetLocale(language) + ") *" +
				strconv.Itoa(int(changePrice.TotalNum)) + " (" + currencySetting.Unit + utils.FormatFloat(changePrice.Price) + ")"
		}
	case constant.OrderUpdateMealNum: // 修改桌台就餐人数
		var updateMealNum event.ChangeMealNumSaleBillPayload
		if err := json.Unmarshal([]byte(log.Data), &updateMealNum); err == nil {
			return strconv.Itoa(int(updateMealNum.NewMealNum))
		}
	case constant.OrderStayOrder: // 挂单 不用解析data
	case constant.OrderPickOrder: // 取单 不用解析data
	case constant.OrderProductFree: // 赠菜
		var productFree event.GiftSaleOrderProductPayload
		if err := json.Unmarshal([]byte(log.Data), &productFree); err == nil {
			return productFree.ProductName.GetLocale(language) + " (" + productFree.ProductAttr.GetLocale(language) + ") *" + strconv.Itoa(int(productFree.TotalNum)) +
				" (" + currencySetting.Unit + utils.FormatFloat(productFree.TotalPrice) + ")"
		}
	case constant.OrderCancelProductFree: // 取消赠菜
		var cancelProductFree event.CancelGiftSaleOrderProductPayload
		if err := json.Unmarshal([]byte(log.Data), &cancelProductFree); err == nil {
			return cancelProductFree.ProductName.GetLocale(language) + " (" + cancelProductFree.ProductAttr.GetLocale(language) + ") *" + strconv.Itoa(int(cancelProductFree.TotalNum)) +
				" (" + currencySetting.Unit + utils.FormatFloat(cancelProductFree.TotalPrice) + ")"
		}
	case constant.OrderProductMove: // 转菜
		var productMove event.ChangeDeskSaleOrderProductPayload
		if err := json.Unmarshal([]byte(log.Data), &productMove); err == nil {
			return fmt.Sprintf("%s (%s) *%d(%s%s)", productMove.ProductName.GetLocale(language), productMove.ProductAttr.GetLocale(language), productMove.TotalNum, i18n.Translate(language, "转至"), productMove.ToTableNo)
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
		var freeSale event.FreeSaleOrderPayload
		if err := json.Unmarshal([]byte(log.Data), &freeSale); err == nil {
			return i18n.Translate(language, "免单") + " (" + currencySetting.Unit + utils.FormatFloat(freeSale.DiscountMoney) + ")"
		}
	case constant.OrderSettle: // 结账
		var settle event.CheckoutSaleOrderPayload
		if err := json.Unmarshal([]byte(log.Data), &settle); err == nil {
			var payTypeList []string
			if settle.IsFree {
				payTypeList = append(payTypeList, i18n.Translate(language, "免单")+": "+currencySetting.Unit+utils.FormatFloat(settle.DiscountMoney))
			}
			for _, payType := range settle.PayType {
				payTypeName := payType.Name
				if payType.Value == constant.PaymentMethodCodeCash {
					payType.Price = utils.DecimalSub(payType.Price, settle.ChangeDue)
				}
				payTypeList = append(payTypeList, payTypeName+": "+currencySetting.Unit+utils.FormatFloat(payType.Price))
			}
			desc := i18n.Translate(language, "订单金额") + " " + currencySetting.Unit + utils.FormatFloat(settle.OrderPrice) + ", " +
				i18n.Translate(language, "实付金额") + " " + currencySetting.Unit + utils.FormatFloat(settle.ActualPrice)
			if len(payTypeList) > 0 {
				desc = desc + " (" + strings.Join(payTypeList, "、") + ")"
			}
			return desc
		}
	case constant.OrderReverseSettle: // 反结账
		var reverseSettle event.OrderReverseSettlePayload
		if err := json.Unmarshal([]byte(log.Data), &reverseSettle); err == nil {
			var payTypeList []string
			for _, payType := range reverseSettle.PayTypes {
				payTypeName := payType.Name
				if payType.Value == constant.PaymentMethodCodeFreePay {
					payTypeName = i18n.Translate(language, "免单")
				}
				payTypeList = append(payTypeList, payTypeName+": "+currencySetting.Unit+utils.FormatFloat(payType.Price))
			}
			return strings.Join(payTypeList, "、")
		}
	case constant.OrderRefund: // 退款
		var refundPayload event.ReturnOrderPayload
		if err := json.Unmarshal([]byte(log.Data), &refundPayload); err == nil {
			var desc []string
			for _, product := range refundPayload.Products {
				item := product.ProductName.GetLocale(language) + " (" + product.ProductAttr.GetLocale(language) + ") *" +
					strconv.Itoa(int(product.TotalNum))
				// 如果商品属性为空，则去掉括弧（）
				if product.ProductAttr.IsNull() {
					item = product.ProductName.GetLocale(language) + " *" +
						strconv.Itoa(int(product.TotalNum))
				}
				desc = append(desc, item)
			}
			return strings.Join(desc, "、")
		}
	case constant.OrderOrderTaking: // 接单 不需要解析data
	case constant.OrderOrderReject: // 拒单 不需要解析data
	case constant.OrderMergeTable:
		var mergeTable event.MergeDeskPayload
		if err := json.Unmarshal([]byte(log.Data), &mergeTable); err == nil {
			return strings.Join(mergeTable.DeskNos, "、")
		}
	case constant.OrderOrderCancel: // 整单取消 不需要解析data
	case constant.OrderCheckoutDiscount: // 结账手动抹零
		var orderCheckoutDiscount event.CheckoutZeroSaleOrderPayload
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
		var splitOrder event.SplitOrderPayload
		if err := json.Unmarshal([]byte(log.Data), &splitOrder); err == nil {
			var orderDetails []string
			for i, order := range splitOrder.Orders {
				orderDetails = append(orderDetails, strconv.Itoa(i+1)+"("+i18n.Translate(language, "订单金额")+"："+currencySetting.Unit+utils.FormatFloat(order.Amount)+")")
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
		prefix = "(" + i18n.Translate(language, "拆单") + strconv.Itoa(common.Index) + ")"
	}
	if log.Action == constant.OrderRefund && common.RefundType == 1 {
		return prefix + i18n.Translate(language, "整单退款")
	}
	if log.Action == constant.OrderOrderTaking {
		var orderTaking event.AcceptH5OrderPayload
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
	var refundPayload event.ReturnOrderPayload
	currencySetting, _ := s.settingSrv.GetCurrencySetting(ctx)
	if err := json.Unmarshal([]byte(log.Data), &refundPayload); err == nil {
		for _, payType := range refundPayload.PayTypes {
			// 支付方式名称
			payTypeName := payType.Name
			if slices.Contains([]int{
				constant.PaymentMethodCodeFreePay,
				constant.PaymentMethodCodeBalance,
				constant.PaymentMethodCodeCash,
			}, payType.Code) {
				payTypeName = i18n.Translate(language, payTypeName)
			}
			// 退款支付类型
			data := resp.OrderOperationLogPaymentMethod{
				Price:            utils.FormatFloat(payType.Amount),
				Code:             payType.Code,
				Name:             payTypeName,
				RefundMoney:      utils.FormatFloat(payType.Amount),
				RefundStatus:     1,
				ReturnOrderUuid:  payType.ReturnOrderUuid,
				ReturnAmountUuid: payType.ReturnAmountUuid,
				PaymentOrderUuid: payType.PaymentOrderUuid,
				Unit:             currencySetting.Unit,
			}
			// 银行支付
			if slices.Contains([]int{
				constant.PaymentMethodCodeLianLianWechatPay,
				constant.PaymentMethodCodeLianLianAliPay,
				constant.PaymentMethodCodeLianLianQRPromptPay,
			}, payType.Code) {
				returnOrderRepo := repository.NewReturnOrderRepo(ctx.GetDB())
				orderAmount, err := returnOrderRepo.GetReturnOrderAmount(returnOrderRepo.WithReturnOrder(), returnOrderRepo.WhereUuid(payType.ReturnAmountUuid))
				if err == nil {
					data.RefundStatus = utils.IfInt(orderAmount.RefundStatus == 2, 0, 1)
				}
				if err == nil && payType.Code == constant.PaymentMethodCodeLianLianQRPromptPay {
					data.BankCode = orderAmount.ReturnOrder.BankCode
					data.AccountNo = orderAmount.ReturnOrder.AccountNo
					data.AccountName = orderAmount.ReturnOrder.AccountName
				}
			}
			refundPayTypes = append(refundPayTypes, data)
		}
	}
	return refundPayTypes
}
