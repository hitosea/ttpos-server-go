package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/model"
	"ttpos-server-go/i18n"
)

func (s *orderSrv) getActionDescription(log model.SaleOrderOperationRecord, language string) string {

	// 开桌
	type OpenDeskPayload struct {
		MealNum  uint   `json:"meal_num"`  // 用餐人数
		IsBuffet bool   `json:"is_buffet"` // 是否是自助餐
		TableId  uint64 `json:"table_id"`  // 桌台ID
		TableNo  string `json:"table_no"`  // 桌台号
	}

	// 送厨
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
	type SendKitchenPayload struct {
		Products []OrderProduct `json:"products"` // 用餐人数
	}

	// 退菜
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

	// 取消退菜
	type CancelRefundProductPayload struct {
		OrderProductId uint64             `json:"order_product_id"` // 订单商品ID
		ProductId      uint64             `json:"product_id"`       // 商品ID
		ProductName    dto.LocaleResponse `json:"product_name"`     // 商品名称
		ProductAttr    dto.LocaleResponse `json:"product_attr"`     // 商品属性
		Num            uint               `json:"num"`              // 退菜数量
		ParentId       uint64             `json:"parent_id"`        // 父订单ID
		OrderName      uint64             `json:"order_name"`       // 订单名称
	}

	// 转台
	type Table struct {
		TableId uint64 `json:"table_id"`
		TableNo string `json:"table_no"`
	}
	type ChangeTablePayload struct {
		Old Table `json:"old"`
		New Table `json:"new"`
	}

	// 改价
	type ChangePricePayload struct {
		OrderProductId uint64             `json:"order_product_id"` // 销售订单产品ID
		ProductId      uint64             `json:"product_id"`       // 产品ID
		ProductName    dto.LocaleResponse `json:"product_name"`     // 产品名称
		ProductAttr    dto.LocaleResponse `json:"product_attr"`     // 产品属性
		TotalNum       uint               `json:"total_num"`        // 数量
		Price          float64            `json:"price"`            // 价格，单价
	}

	// 修改桌台就餐人数
	type UpdateMealNumPayload struct {
		OldMealNum uint `json:"old_meal_num"` // 旧桌台就餐人数
		NewMealNum uint `json:"new_meal_num"` // 新桌台就餐人数
	}

	// 赠菜
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

	// 取消赠菜
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

	// 转菜
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

	// 优惠折扣
	type DiscountPayload struct {
		OldPrice        float64 `json:"old_price"`        // 原价。订单改价前的价格
		NewPrice        float64 `json:"new_price"`        // 新价格。订单改价后的价格
		DiscountType    int     `json:"discount_type"`    // 折扣类型。1: 订单改价
		SpecialDiscount float64 `json:"special_discount"` // 优惠金额。订单改价后的优惠金额
	}

	// 结账
	type PayType struct {
		Name           string  `json:"name"`            // 支付方式名称
		Value          int     `json:"value"`           // 支付方式值
		DisabledCancel uint    `json:"disabled_cancel"` // 是否禁用取消
		Price          float64 `json:"price"`           // 支付金额（含手续费）
		FeeMoney       float64 `json:"fee_money"`       // 手续费
	}
	type SettlePayload struct {
		SaleBill    *model.SaleBill
		OrderPrice  float64   `json:"order_price"`  // 订单应付金额
		PayPrice    float64   `json:"pay_price"`    // 最终应付金额。最终应付金额=订单应付金额+手续费（手续费=每笔付款单的手续费之和）
		ActualPrice float64   `json:"actual_price"` // 最终实付金额。最终实付金额=最终应付金额+找零金额。如果没有找零，则最终实付金额=最终应付金额。最终实付金额=每笔付款单的付款金额之和（含手续费）
		ChangeDue   float64   `json:"change_due"`   // 找零金额
		PayType     []PayType `json:"pay_type"`     // 支付方式
	}

	// 并台
	type MergeTablePayload struct {
		DeskNos []string
	}

	// 结账手动抹零
	type OrderCheckoutDiscountPayload struct {
		Operation string `json:"operation"` // 操作类型。add: 设置结账抹零，cancel: 撤销结账抹零
		Remark    string `json:"remark"`    // 备注
	}

	switch log.Action {
	case constant.OrderOpenTable:
	case constant.OrderSendKitchen:
	case constant.OrderRefundProduct:
	case constant.OrderCancelRefundProduct:
	case constant.OrderChangeTable:
	case constant.OrderChangePrice:
	case constant.OrderUpdateMealNum:
	case constant.OrderStayOrder: // 不用解析data
	case constant.OrderPickOrder: // 不用解析data
	case constant.OrderProductFree:
	case constant.OrderCancelProductFree:
	case constant.OrderProductMove:
	case constant.OrderDiscount:
	case constant.OrderCancelDiscount: // 不需要解析data
	case constant.OrderSettle:
	//case constant.OrderReverseSettle:
	//case constant.OrderRefund:
	case constant.OrderOrderTaking: // 不需要解析data
	case constant.OrderOrderReject: // 不需要解析data
	case constant.OrderMergeTable:
	case constant.OrderOrderCancel: // 不需要解析data
	case constant.OrderCheckoutDiscount:
	case constant.OrderSplitOrder:
	case constant.OrderCancelSplitOrder: // 不需要解析data
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

	var text string
	var ok bool
	if text, ok = actionTextMap[log.Action]; !ok {
		return i18n.Translate(language, "未知操作")
	}

	//if log.Action == constant.OrderRefund { // 退款
	//
	//}

	return text

	// $texts = [
	//            self::ACTION_OPEN_TABLE => __('开台'),
	//            self::ACTION_SEND_KITCHEN => __('送厨'),
	//            self::ACTION_REFUND_PRODUCT => __('退菜'),
	//            self::ACTION_CANCEL_REFUND_PRODUCT => __('取消退菜'),
	//            self::ACTION_CHANGE_TABLE => __('转台'),
	//            self::ACTION_CHANGE_PRICE => __('改价'),
	//            self::ACTION_UPDATE_MEAL_NUM => __('人数'),
	//            self::ACTION_STAY_ORDER => __('挂单'),
	//            self::ACTION_PICK_ORDER => __('取单'),
	//            self::ACTION_PRODUCT_FREE => __('赠菜'),
	//            self::ACTION_CANCEL_PRODUCT_FREE => __('取消赠菜'),
	//            self::ACTION_PRODUCT_MOVE => __('转菜'),
	//            self::ACTION_DISCOUNT => __('优惠折扣'),
	//            self::ACTION_CANCEL_DISCOUNT => __('撤销优惠折扣'),
	//            self::ACTION_SETTLE => __('结账'),
	//            self::ACTION_REVERSE_SETTLE => __('反结账'),
	//            self::ACTION_REFUND => __('部分退款'),
	//            self::ACTION_ORDER_TAKING => __('接单'),
	//            self::ACTION_ORDER_REJECT => __('拒单'),
	//            self::ACTION_MERGE_TABLE => __('并台'),
	//            self::ACTION_ORDER_CANCEL => __('整单取消'),
	//            self::ACTION_CHECKOUT_DISCOUNT => __('结账手动抹零'),
	//            self::ACTION_SPLIT_ORDER => __('拆单'),
	//            self::ACTION_CANCEL_SPLIT_ORDER => __('撤销拆单'),
	//        ];
	//        // 拆单信息拼接
	//        $splitStr = '';
	//        if (($data['parent_id'] ?? 0) > 0) {
	//            $splitStr = '（' . __('拆单') . ($data['order_name'] ?? '') . '）';
	//        }
	//        //
	//        if ($action == self::ACTION_REFUND && ($data['refund_type'] ?? 1) == 1) {
	//            $content = __('整单退款');
	//            return $splitStr ? $splitStr . $content : $content;
	//        }
	//        //
	//        if ($action == self::ACTION_ORDER_TAKING && ($data['is_auto_order'] ?? false) == true) {
	//            return __('系统自动接单');
	//        }
	//        //
	//        $content = isset($texts[$action]) ? $texts[$action] : __('未知操作');
	//        return $splitStr ? $splitStr . $content : $content;
	return ""
}
