package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/utils"
)

type PayType struct {
	Name  string  `json:"name"`
	Value int     `json:"value"`
	Price float64 `json:"price"`
}

type RechargeLog struct {
	OrderPrice    float64   `json:"order_price"`    //  订单金额
	RechargeMoney float64   `json:"recharge_money"` //  充值金额
	PayPrice      float64   `json:"pay_price"`      //  订单应付
	GiftMoney     float64   `json:"gift_money"`     //  赠送金额
	GiftPoint     float64   `json:"gift_point"`     //  赠送积分
	PayFee        float64   `json:"pay_fee"`        //  支付手续费
	ChangeDue     float64   `json:"change_due"`     //  找零
	PayType       []PayType `json:"pay_type"`       //  支付方式
}

// 确认充值日志data
func (s *rechargeOrderSrv) getRechargeLogData(order model.MemberRechargeOrder) string {
	var payTypes []PayType
	for _, paymentOrder := range order.PaymentOrders {
		payTypes = append(payTypes, PayType{
			Name:  paymentOrder.PaymentMethodName,
			Value: paymentOrder.PaymentMethod.Code,
			Price: paymentOrder.Amount,
		})
	}
	operationData, _ := json.Marshal(RechargeLog{
		OrderPrice:    order.RechargeAmount,
		RechargeMoney: order.RechargeAmount,
		PayPrice:      s.getRechargeOrderAmount(order.PaymentOrders),
		GiftMoney:     order.GiftAmount,
		GiftPoint:     order.GiftPoint,
		PayFee:        s.getPayFee(order.PaymentOrders),
		ChangeDue:     s.getChargeDue(order.PaymentOrders),
		PayType:       payTypes,
	})
	return string(operationData)
}

type ChangeAmountLog struct {
	RechargeMoney    float64 `json:"recharge_money"`
	OldRechargeMoney float64 `json:"old_recharge_money"`
}

// 订单金额变化日志data
func (s *rechargeOrderSrv) getChangeAmountLogData(recharge, oldRecharge float64) string {
	operationData, _ := json.Marshal(ChangeAmountLog{
		RechargeMoney:    recharge,
		OldRechargeMoney: oldRecharge,
	})
	return string(operationData)
}

type ReverseSettleLog struct {
	PayPrice  float64   `json:"pay_price"`  //  应收金额
	ChangeDue float64   `json:"change_due"` //  找零
	PayType   []PayType `json:"pay_type"`
}

// 订单反结账日志data
func (s *rechargeOrderSrv) getReverseSettleLogData(order model.MemberRechargeOrder) string {
	var payTypes []PayType
	for _, paymentOrder := range order.PaymentOrders {
		payTypes = append(payTypes, PayType{
			Name:  paymentOrder.PaymentMethodName,
			Value: paymentOrder.PaymentMethod.Code,
			Price: paymentOrder.Amount,
		})
	}
	operationData, _ := json.Marshal(ReverseSettleLog{
		PayPrice:  order.Amount,
		ChangeDue: s.getChargeDue(order.PaymentOrders),
		PayType:   payTypes,
	})
	return string(operationData)
}

type RefundLog struct {
	RefundType     uint            `json:"refund_type"`
	RefundMoney    float64         `json:"refund_money"`
	RefundPayTypes []RefundPayType `json:"refund_pay_types"`
}

// 订单退款日志data
func (s *rechargeOrderSrv) getRefundData(refundType uint, refundMoney float64, refundPayTypes []RefundPayType) string {
	operationData, _ := json.Marshal(RefundLog{
		RefundType:     refundType,
		RefundMoney:    refundMoney,
		RefundPayTypes: refundPayTypes,
	})
	return string(operationData)
}

func (s *rechargeOrderSrv) getActionDescription(ctx context.Context, log model.MemberRechargeOrderOperationLog, language string) string {
	currencySetting, _ := s.settingSrv.GetCurrencySetting(ctx)
	switch log.Action {
	case constant.RechargeOrderActionChangeAmount:
		var changeAmount ChangeAmountLog
		json.Unmarshal([]byte(log.Data), &changeAmount)
		return fmt.Sprintf("%.2f %s %.2f", changeAmount.OldRechargeMoney, i18n.Translate(language, "变更为"), changeAmount.RechargeMoney)
	case constant.RechargeOrderActionRecharge:
		var recharge RechargeLog
		json.Unmarshal([]byte(log.Data), &recharge)
		var payTypeList []string
		for _, payType := range recharge.PayType {
			price := payType.Price
			if payType.Value == constant.PaymentMethodCodeCash {
				price = utils.DecimalSub(price, recharge.ChangeDue)
			}
			payTypeList = append(payTypeList, fmt.Sprintf("%s: %s%.2f", payType.Name, currencySetting.Unit, price))
		}
		desc := fmt.Sprintf("%s %s %.2f，%s %s %.2f",
			i18n.Translate(language, "订单金额"), currencySetting.Unit, recharge.RechargeMoney,
			i18n.Translate(language, "实付金额"), currencySetting.Unit, recharge.PayPrice)
		if len(payTypeList) > 0 {
			return desc + "(" + strings.Join(payTypeList, "、") + ")"
		}
		return desc
	case constant.RechargeOrderActionReverseSettle:
		var reverseSettle ReverseSettleLog
		json.Unmarshal([]byte(log.Data), &reverseSettle)
		var payTypeList []string
		for _, payType := range reverseSettle.PayType {
			price := payType.Price
			if payType.Value == constant.PaymentMethodCodeCash {
				price = utils.DecimalSub(price, reverseSettle.ChangeDue)
			}
			payTypeList = append(payTypeList, fmt.Sprintf("%s: %s%.2f", payType.Name, currencySetting.Unit, price))
		}
		return strings.Join(payTypeList, "、")
	case constant.RechargeOrderActionRefund:
		var refundLog RefundLog
		json.Unmarshal([]byte(log.Data), &refundLog)
		var payTypeList []string
		for _, payType := range refundLog.RefundPayTypes {
			payTypeList = append(payTypeList, fmt.Sprintf("%s: %s%.2f", payType.Name, currencySetting.Unit, payType.Amount))
		}
		return strings.Join(payTypeList, "、")
	}
	return ""
}

func (s *rechargeOrderSrv) getActionText(log model.MemberRechargeOrderOperationLog, language string) string {
	if log.Action == constant.RechargeOrderActionRefund {
		var refundLog RefundLog
		json.Unmarshal([]byte(log.Data), &refundLog)
		if refundLog.RefundType == constant.ReturnOrderRefundTypeTotal {
			return i18n.Translate(language, "整单退款")
		}
	}
	texts := map[string]string{
		constant.RechargeOrderActionGenerateOrder: i18n.Translate(language, "生成订单"),
		constant.RechargeOrderActionChangeAmount:  i18n.Translate(language, "变更充值金额"),
		constant.RechargeOrderActionOrderCancel:   i18n.Translate(language, "取消"),
		constant.RechargeOrderActionRecharge:      i18n.Translate(language, "充值"),
		constant.RechargeOrderActionReverseSettle: i18n.Translate(language, "反结账"),
		constant.RechargeOrderActionRefund:        i18n.Translate(language, "部分退款"),
	}

	if text, ok := texts[log.Action]; ok {
		return text
	} else {
		return i18n.Translate(language, "未知操作")
	}
}
