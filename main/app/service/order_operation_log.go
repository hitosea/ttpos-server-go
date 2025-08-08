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
	IsAuto          bool    `json:"is_auto"`          // 是否自动抹零
}

type ActionDescription struct {
	Desc               string `json:"desc"`                  // 行为描述
	SplitMessage       string `json:"split_message"`         // 拆单前缀
	HideLog            bool   `json:"hide_log"`              // 是否隐藏日志
	IsAutoCheckoutZero bool   `json:"is_auto_checkout_zero"` // 是否自动结账抹零
}

func getSplitMessage(ctx context.Context, log model.SaleOrderOperationRecord, language string) ActionDescription {
	var splitMessage string
	// 获取订单拆单序号。只有折扣、取消折扣、免单、结账手动抹零需要获取订单拆单序号
	if log.Action == constant.OrderDiscount || log.Action == constant.OrderCancelDiscount || log.Action == constant.OrderFreeSale || log.Action == constant.OrderCheckoutDiscount {
		// // 获取订单拆单序号。只有改价、折扣、取消折扣、免单、结账手动抹零需要获取订单拆单序号
		// if log.Action == constant.OrderChangePrice || log.Action == constant.OrderDiscount || log.Action == constant.OrderCancelDiscount || log.Action == constant.OrderFreeSale || log.Action == constant.OrderCheckoutDiscount {
		orderIndex, err := repository.NewSaleBillRepo(ctx.GetDB()).GetSaleOrderIndexByUuid(log.SaleBillUuid, log.SaleOrderUuid)
		if err != nil {
			ctx.Log().Info(fmt.Sprintf("GetSaleOrderIndexByUuid 获取订单拆单序号失败.err:%v", err))
			return ActionDescription{Desc: "", SplitMessage: ""}
		}
		if orderIndex != 0 {
			if orderIndex == -1 {
				// // 如果订单拆单序号为-1，且这条日志是改价日志，不隐藏日志也不显示拆单前缀
				// if log.Action == constant.OrderChangePrice {
				// 	return ActionDescription{Desc: "", SplitMessage: ""}
				// }
				// 如果订单拆单序号为-1，则隐藏日志
				return ActionDescription{HideLog: true}
			}
			splitMessage = "(" + i18n.Translate(language, "拆单") + strconv.Itoa(orderIndex) + ")"
		}
	}
	return ActionDescription{Desc: "", SplitMessage: splitMessage}
}

func (s *orderSrv) getActionDescription(ctx context.Context, log model.SaleOrderOperationRecord, language string) ActionDescription {
	var splitMessage string
	res := getSplitMessage(ctx, log, language)
	if res.HideLog {
		return res
	}
	splitMessage = res.SplitMessage
	switch log.Action {
	case constant.OrderOpenTable: // 开台
		var openTable event.OpenDeskPayload
		err := json.Unmarshal([]byte(log.Data), &openTable)
		if err == nil {
			desc := openTable.TableNo + ", " + i18n.Translate(language, "人数") + ": " + strconv.Itoa(int(openTable.MealNum))
			return ActionDescription{Desc: desc, SplitMessage: ""}
		}
	case constant.OrderSendKitchen: // 送厨
		var sendKitchen event.SentCookingPayload
		if err := json.Unmarshal([]byte(log.Data), &sendKitchen); err == nil {
			var desc []string
			for _, product := range sendKitchen.Products {
				attrString := ""
				if product.ProductAttr.GetLocale(language) != "" {
					attrString = " (" + product.ProductAttr.GetLocale(language) + ")"
				}
				desc = append(desc, product.ProductName.GetLocale(language)+attrString+" *"+utils.FormatFloat(product.TotalNum))
			}
			descStr := strings.Join(desc, "、")
			return ActionDescription{Desc: descStr, SplitMessage: ""}
		}
	case constant.OrderRefundProduct: // 退菜
		var returnProduct event.CancelSaleOrderProductPayload
		if err := json.Unmarshal([]byte(log.Data), &returnProduct); err == nil {
			attrString := ""
			if returnProduct.ProductAttr.GetLocale(language) != "" {
				attrString = " (" + returnProduct.ProductAttr.GetLocale(language) + ")"
			}
			desc := returnProduct.ProductName.GetLocale(language) + attrString + "*" +
				utils.FormatFloat(returnProduct.TotalNum) +
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
			return ActionDescription{Desc: desc, SplitMessage: ""}
		}
	case constant.OrderCancelRefundProduct: // 取消退菜
		var cancelRefundProduct event.CancelReturnSaleOrderProductPayload
		if err := json.Unmarshal([]byte(log.Data), &cancelRefundProduct); err == nil {
			attrString := ""
			if cancelRefundProduct.ProductAttr.GetLocale(language) != "" {
				attrString = " (" + cancelRefundProduct.ProductAttr.GetLocale(language) + ")"
			}
			desc := cancelRefundProduct.ProductName.GetLocale(language) + attrString + " *" +
				utils.FormatFloat(cancelRefundProduct.Num)
			return ActionDescription{Desc: desc, SplitMessage: ""}
		}
	case constant.OrderChangeTable: // 转台
		var changeTable event.ChangeDeskPayload
		if err := json.Unmarshal([]byte(log.Data), &changeTable); err == nil {
			desc := changeTable.Old.TableNo + "->" + changeTable.New.TableNo
			return ActionDescription{Desc: desc, SplitMessage: ""}

		}
	case constant.OrderChangePrice: // 改价
		var changePrice event.ChangeSaleOrderProductPricePayload
		if err := json.Unmarshal([]byte(log.Data), &changePrice); err == nil {
			attrString := ""
			if changePrice.ProductAttr.GetLocale(language) != "" {
				attrString = " (" + changePrice.ProductAttr.GetLocale(language) + ")"
			}
			desc := changePrice.ProductName.GetLocale(language) + attrString + " *" +
				utils.FormatFloat(changePrice.TotalNum) + " (" + s.settingSrv.SymbolPosition(ctx, changePrice.Price) + ")"
			return ActionDescription{Desc: desc, SplitMessage: splitMessage}
		}
	case constant.OrderUpdateMealNum: // 修改桌台就餐人数
		var updateMealNum event.ChangeMealNumSaleBillPayload
		if err := json.Unmarshal([]byte(log.Data), &updateMealNum); err == nil {
			desc := strconv.Itoa(int(updateMealNum.NewMealNum))
			return ActionDescription{Desc: desc, SplitMessage: ""}
		}
	case constant.OrderStayOrder: // 挂单 不用解析data
	case constant.OrderPickOrder: // 取单 不用解析data
	case constant.OrderWrapSaleBill: // 整单打包 不用解析data
	case constant.OrderUnwrapSaleBill: // 取消整单打包 不用解析data
	case constant.OrderProductFree: // 赠菜
		var productFree event.GiftSaleOrderProductPayload
		if err := json.Unmarshal([]byte(log.Data), &productFree); err == nil {
			attrString := ""
			if productFree.ProductAttr.GetLocale(language) != "" {
				attrString = " (" + productFree.ProductAttr.GetLocale(language) + ")"
			}
			desc := productFree.ProductName.GetLocale(language) + attrString + " *" + utils.FormatFloat(productFree.TotalNum) +
				" (" + s.settingSrv.SymbolPosition(ctx, productFree.TotalPrice) + ")"
			return ActionDescription{Desc: desc, SplitMessage: ""}
		}
	case constant.OrderCancelProductFree: // 取消赠菜
		var cancelProductFree event.CancelGiftSaleOrderProductPayload
		if err := json.Unmarshal([]byte(log.Data), &cancelProductFree); err == nil {
			attrString := ""
			if cancelProductFree.ProductAttr.GetLocale(language) != "" {
				attrString = " (" + cancelProductFree.ProductAttr.GetLocale(language) + ")"
			}
			desc := cancelProductFree.ProductName.GetLocale(language) + attrString + " *" + utils.FormatFloat(cancelProductFree.TotalNum) +
				" (" + s.settingSrv.SymbolPosition(ctx, cancelProductFree.TotalPrice) + ")"
			return ActionDescription{Desc: desc, SplitMessage: ""}
		}
	case constant.OrderProductWrap: // 打包
		var productWrap event.WrapSaleOrderProductPayload
		if err := json.Unmarshal([]byte(log.Data), &productWrap); err == nil {
			attrString := ""
			if productWrap.ProductAttr.GetLocale(language) != "" {
				attrString = " (" + productWrap.ProductAttr.GetLocale(language) + ")"
			}
			desc := productWrap.ProductName.GetLocale(language) + attrString + " *" + utils.FormatFloat(productWrap.Num)
			return ActionDescription{Desc: desc, SplitMessage: ""}
		}
	case constant.OrderProductUnwrap: // 取消打包
		var productUnwrap event.UnwrapSaleOrderProductPayload
		if err := json.Unmarshal([]byte(log.Data), &productUnwrap); err == nil {
			attrString := ""
			if productUnwrap.ProductAttr.GetLocale(language) != "" {
				attrString = " (" + productUnwrap.ProductAttr.GetLocale(language) + ")"
			}
			desc := productUnwrap.ProductName.GetLocale(language) + attrString + " *" + utils.FormatFloat(productUnwrap.Num)
			return ActionDescription{Desc: desc, SplitMessage: ""}
		}
	case constant.OrderProductMove: // 转菜
		var productMove event.ChangeDeskSaleOrderProductPayload
		if err := json.Unmarshal([]byte(log.Data), &productMove); err == nil {
			attrString := ""
			if productMove.ProductAttr.GetLocale(language) != "" {
				attrString = " (" + productMove.ProductAttr.GetLocale(language) + ")"
			}
			desc := fmt.Sprintf("%s %s *%f(%s%s)", productMove.ProductName.GetLocale(language), attrString, productMove.TotalNum, i18n.Translate(language, "转至"), productMove.ToTableNo)
			return ActionDescription{Desc: desc, SplitMessage: ""}
		}
	case constant.OrderDiscount: // 优惠折扣
		var discount DiscountPayload
		if err := json.Unmarshal([]byte(log.Data), &discount); err == nil {
			var desc string
			switch discount.DiscountType {
			case constant.DiscountOperationLogTypeChangePriceSaleOrder: // 改价
				desc = utils.FormatFloat(discount.OldPrice) + i18n.Translate(language, "改价") +
					utils.FormatFloat(discount.NewPrice) + " (" + s.settingSrv.SymbolPosition(ctx, discount.SpecialDiscount) + ")"
			case constant.DiscountOperationLogTypeDiscountSaleOrder: // 折扣
				desc = i18n.Translate(language, "折扣") + "-" + utils.FormatFloat(discount.RoundingRate) +
					"% (" + s.settingSrv.SymbolPosition(ctx, discount.SpecialDiscount) + ")"
			case constant.DiscountOperationLogTypeZeroSaleOrder: // 抹零
				roundingTypeMap := map[int]string{
					constant.DiscountZeroRulePercent: "抹分",
					constant.DiscountZeroRuleFixed:   "抹角",
					constant.DiscountZeroRuleRound:   "四舍五入保留一位小数",
					constant.DiscountZeroRuleInteger: "四舍五入到整数",
				}
				zeroName := i18n.Translate(language, "抹零")
				if discount.IsAuto {
					zeroName = i18n.Translate(language, "自动抹零")
				}
				desc = zeroName + "-" + i18n.Translate(language, roundingTypeMap[discount.RoundingType]) +
					" (" + s.settingSrv.SymbolPosition(ctx, discount.SpecialDiscount) + ")"
			}
			return ActionDescription{Desc: desc, SplitMessage: splitMessage}
		}
	case constant.OrderCancelDiscount: // 撤销优惠折扣 不需要解析data
		return ActionDescription{Desc: "", SplitMessage: splitMessage}
	case constant.OrderFreeSale: // 免单
		var freeSale event.FreeSaleOrderPayload
		if err := json.Unmarshal([]byte(log.Data), &freeSale); err == nil {
			desc := i18n.Translate(language, "免单") + " (" + s.settingSrv.SymbolPosition(ctx, freeSale.DiscountMoney) + ")"
			return ActionDescription{Desc: desc, SplitMessage: splitMessage}
		}
	case constant.OrderSettle: // 结账
		var settle event.CheckoutSaleOrderPayload
		if err := json.Unmarshal([]byte(log.Data), &settle); err == nil {
			var payTypeList []string
			if settle.IsFree {
				payTypeList = append(payTypeList, i18n.Translate(language, "免单")+": "+s.settingSrv.SymbolPosition(ctx, settle.DiscountMoney))
			}
			for _, payType := range settle.PayType {
				payTypeName := payType.Name
				if payType.Value == constant.PaymentMethodCodeCash {
					payType.Price = utils.DecimalSub(payType.Price, settle.ChangeDue)
				}
				payTypeList = append(payTypeList, payTypeName+": "+s.settingSrv.SymbolPosition(ctx, payType.Price))
			}
			desc := i18n.Translate(language, "订单金额") + " " + s.settingSrv.SymbolPosition(ctx, settle.OrderPrice) + ", " +
				i18n.Translate(language, "实付金额") + " " + s.settingSrv.SymbolPosition(ctx, settle.ActualPrice)
			if len(payTypeList) > 0 {
				desc = desc + " (" + strings.Join(payTypeList, "、") + ")"
			}
			return ActionDescription{Desc: desc, SplitMessage: ""}
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
				payTypeList = append(payTypeList, payTypeName+": "+s.settingSrv.SymbolPosition(ctx, payType.Price))
			}
			desc := strings.Join(payTypeList, "、")
			return ActionDescription{Desc: desc, SplitMessage: ""}
		}
	case constant.OrderRefund: // 退款
		var refundPayload event.ReturnOrderPayload
		if err := json.Unmarshal([]byte(log.Data), &refundPayload); err == nil {
			if refundPayload.RefundType == constant.ReturnOrderRefundTypeTotal { // 整单退款不显示商品
				return ActionDescription{Desc: "", SplitMessage: ""}
			}
			var desc []string
			for _, product := range refundPayload.Products {
				item := product.ProductName.GetLocale(language) + " (" + product.ProductAttr.GetLocale(language) + ") *" +
					utils.FormatFloat(product.TotalNum)
				// 如果商品属性为空，则去掉括弧（）
				if product.ProductAttr.IsNull() {
					item = product.ProductName.GetLocale(language) + " *" +
						utils.FormatFloat(product.TotalNum)
				}
				desc = append(desc, item)
			}
			descStr := strings.Join(desc, "、")
			return ActionDescription{Desc: descStr, SplitMessage: ""}
		}
	case constant.OrderOrderTaking: // 接单 不需要解析data
	case constant.OrderOrderReject: // 拒单 不需要解析data
	case constant.OrderMergeTable:
		var mergeTable event.MergeDeskPayload
		if err := json.Unmarshal([]byte(log.Data), &mergeTable); err == nil {
			desc := strings.Join(mergeTable.DeskNos, "、")
			return ActionDescription{Desc: desc, SplitMessage: ""}
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
					" (" + s.settingSrv.SymbolPosition(ctx, orderCheckoutDiscount.SpecialDiscount) + ")"
			}
			return ActionDescription{Desc: desc, SplitMessage: splitMessage, IsAutoCheckoutZero: orderCheckoutDiscount.IsAuto}
		}
	case constant.OrderSplitOrder: // 拆单
		var splitOrder event.SplitOrderPayload
		if err := json.Unmarshal([]byte(log.Data), &splitOrder); err == nil {
			var orderDetails []string
			for i, order := range splitOrder.Orders {
				orderDetails = append(orderDetails, strconv.Itoa(i+1)+"("+i18n.Translate(language, "订单金额")+"："+s.settingSrv.SymbolPosition(ctx, order.Amount)+")")
			}
			desc := strings.Join(orderDetails, ", ")
			return ActionDescription{Desc: desc, SplitMessage: ""}
		}
	case constant.OrderCancelSplitOrder: // 撤销拆单 不需要解析data
		return ActionDescription{Desc: "", SplitMessage: ""}

	// 会员端订单操作
	case constant.OrderCancelMemberSaleOrder: // 订单取消
		var cancelMemberSaleOrder event.CancelMemberOrderPayload
		if err := json.Unmarshal([]byte(log.Data), &cancelMemberSaleOrder.Data); err == nil {
			desc := ""
			if cancelMemberSaleOrder.Data.Type == "user_cancel" {
				desc = desc + " (" + i18n.Translate(language, "用户取消") + ")"
			}
			if cancelMemberSaleOrder.Data.Type == "timeout_cancel" {
				desc = desc + " (" + i18n.Translate(language, "超时取消") + ")"
			}
			if cancelMemberSaleOrder.Data.Type == "reject_order" {
				desc = desc + " (" + i18n.Translate(language, "商家拒单") + ")"
			}
			if cancelMemberSaleOrder.Data.Type == "shop_cancel" {
				desc = desc + " (" + i18n.Translate(language, "商家取消") + ")"
			}
			if cancelMemberSaleOrder.Data.Type == "rider_pickup_timeout" {
				desc = desc + " (" + i18n.Translate(language, "骑手超时未接单") + ")"
			}
			// 退款信息
			if len(cancelMemberSaleOrder.Data.Refunds) > 0 {
				desc = desc + "，" + i18n.Translate(language, "订单已退款") + " ("
				for i, refund := range cancelMemberSaleOrder.Data.Refunds {
					desc = desc + " " + refund.Name + "：" + s.settingSrv.SymbolPosition(ctx, refund.Amount)
					if i < len(cancelMemberSaleOrder.Data.Refunds)-1 {
						desc = desc + "、"
					}
				}
				desc = desc + ")"
			}

			return ActionDescription{Desc: desc}
		}
	}
	return ActionDescription{Desc: "", SplitMessage: ""}
}

func (s *orderSrv) getActionText(log model.SaleOrderOperationRecord, language string) string {

	actionTextMap := map[string]string{
		// 用餐订单 操作
		constant.OrderOpenTable:           i18n.Translate(language, "开台"),
		constant.OrderSendKitchen:         i18n.Translate(language, "送厨"),
		constant.OrderRefundProduct:       i18n.Translate(language, "退菜"),
		constant.OrderCancelRefundProduct: i18n.Translate(language, "取消退菜"),
		constant.OrderChangeTable:         i18n.Translate(language, "转台"),
		constant.OrderChangePrice:         i18n.Translate(language, "改价"),
		constant.OrderUpdateMealNum:       i18n.Translate(language, "人数"),
		constant.OrderStayOrder:           i18n.Translate(language, "挂单"),
		constant.OrderPickOrder:           i18n.Translate(language, "取单"),
		constant.OrderProductWrap:         i18n.Translate(language, "打包"),
		constant.OrderProductUnwrap:       i18n.Translate(language, "取消打包"),
		constant.OrderWrapSaleBill:        i18n.Translate(language, "整单打包"),
		constant.OrderUnwrapSaleBill:      i18n.Translate(language, "取消整单打包"),
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
		// 外送订单 操作
		constant.OrderCreateMemberSaleOrder:    i18n.Translate(language, "创建订单"),
		constant.OrderPayFinishMemberSaleOrder: i18n.Translate(language, "订单支付成功"),
		constant.OrderCancelMemberSaleOrder:    i18n.Translate(language, "订单取消"),
		constant.OrderAcceptMemberSaleOrder:    i18n.Translate(language, "商家接单"),
		constant.OrderPickMemberSaleOrder:      i18n.Translate(language, "出餐完成,呼叫骑手"),
		constant.OrderPickUpMemberSaleOrder:    i18n.Translate(language, "骑手已接单，正在赶往商家"),
		constant.OrderDeliveryMemberSaleOrder:  i18n.Translate(language, "骑手取货完成，开始配送"),
		constant.OrderFinishMemberSaleOrder:    i18n.Translate(language, "配送完成，订单完成"),
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
