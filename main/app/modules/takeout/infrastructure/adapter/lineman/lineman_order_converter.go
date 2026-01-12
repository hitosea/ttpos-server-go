package lineman

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"ttpos-server-go/app/errors"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	valueobject "ttpos-server-go/app/modules/takeout/domain/value_object"
)

// LineManConverter LINE MAN 平台转换器实现
type LineManConverter struct {
	amountConversionFactor int64 // 金额转换因子，LINE MAN 使用 1 (泰铢单位本身就是最小单位)
}

// NewLineManConverter 创建 LINE MAN 转换器
func NewLineManConverter() *LineManConverter {
	return &LineManConverter{
		amountConversionFactor: 1, // LINE MAN 使用泰铢作为最小单位
	}
}

// LineManPlaceOrderRequest LINE MAN Place Order API 请求结构
type LineManPlaceOrderRequest struct {
	OrderID           string                  `json:"orderId"`            // 订单唯一ID，格式：LMF-yyMMdd-{number}
	OrderShortCode    string                  `json:"orderShortCode"`     // 订单短码（orderId 的最后 4 位数字）
	RestaurantRevenue float64                 `json:"restaurantRevenue"`  // 餐厅收入（已扣除平台补贴）
	OrderAcceptedTime string                  `json:"orderAcceptedTime"`  // 订单接受时间 (ISO 8601 格式)
	Items             []LineManOrderItem      `json:"items"`              // 订单商品列表
	AdditionalItems   []LineManAdditionalItem `json:"additionalItems"`    // 附加项目列表
	MemberID          string                  `json:"memberId,omitempty"` // LINE MAN 会员 ID
	CustomerType      string                  `json:"customerType"`       // 订单类型: DELIVERY 或 PICKUP
}

// LineManOrderItem LINE MAN 订单商品
type LineManOrderItem struct {
	ID          string                `json:"id"`                    // 菜品 ID
	Quantity    int                   `json:"quantity"`              // 数量
	UnitPrice   float64               `json:"unitPrice"`             // 单价（含选项加价和折扣）
	Memo        string                `json:"memo,omitempty"`        // 备注
	PromotionID string                `json:"promotionId,omitempty"` // 促销活动 ID
	Discount    float64               `json:"discount,omitempty"`    // 折扣金额
	Properties  []LineManItemProperty `json:"properties,omitempty"`  // 商品选项（如规格、口味等）
}

// LineManItemProperty LINE MAN 商品选项
type LineManItemProperty struct {
	ID     string                 `json:"id"`     // 选项 ID
	Values []LineManPropertyValue `json:"values"` // 选中的选项值列表
}

// LineManPropertyValue LINE MAN 选项值
type LineManPropertyValue struct {
	ID    string  `json:"id"`    // 选项值 ID
	Price float64 `json:"price"` // 选项价格 (THB)
}

// LineManAdditionalItem LINE MAN 附加项目（如服务费）
type LineManAdditionalItem struct {
	Name string `json:"name"` // 附加项目名称
}

// ConvertLineManStateToOrderState 将 LINE MAN 订单状态转换为内部订单状态
// LINE MAN 的订单通过 Webhook 推送时，默认状态为已接单（商家需要接受）
func ConvertLineManStateToOrderState(platformState string) int {
	// LINE MAN Webhook 推送的订单默认为待接单状态
	state := strings.ToUpper(strings.TrimSpace(platformState))

	switch state {
	case "NEW", "PENDING", "RECEIVED":
		return valueobject.TakeoutOrderStatePending // 0 - 待接单

	case "ACCEPTED", "PREPARING", "READY":
		return valueobject.TakeoutOrderStateAccepted // 1 - 已接单配餐中

	case "COLLECTED", "PICKED_UP":
		return valueobject.TakeoutOrderStateRiderProcessing // 3 - 骑手配送中

	case "DELIVERED", "COMPLETED":
		return valueobject.TakeoutOrderStateCompleted // 4 - 已完成

	case "REJECTED":
		return valueobject.TakeoutOrderStateRejected // 5 - 已拒单

	case "CANCELLED", "CANCELED":
		return valueobject.TakeoutOrderStateCanceled // 6 - 已取消

	default:
		// 未知状态，默认为待接单
		return valueobject.TakeoutOrderStatePending
	}
}

// ParseOrderWebhook 解析 LINE MAN 订单数据
func (c *LineManConverter) ParseOrderWebhook(rawData []byte) (interface{}, error) {
	var placeOrderReq LineManPlaceOrderRequest
	if err := json.Unmarshal(rawData, &placeOrderReq); err != nil {
		return nil, fmt.Errorf("解析 LINE MAN 订单数据失败: %w", err)
	}

	// 验证必填字段
	if placeOrderReq.OrderID == "" {
		return nil, errors.New("LINE MAN 订单数据缺少 orderId 字段")
	}
	if placeOrderReq.OrderShortCode == "" {
		return nil, errors.New("LINE MAN 订单数据缺少 orderShortCode 字段")
	}
	if placeOrderReq.CustomerType == "" {
		return nil, errors.New("LINE MAN 订单数据缺少 customerType 字段")
	}

	return &placeOrderReq, nil
}

// ConvertOrderToTakeoutOrder 将 LINE MAN 订单转换为通用外卖订单格式
func (c *LineManConverter) ConvertOrderToTakeoutOrder(
	orderUuid uint64,
	platform string,
	platformOrderId string,
	platformOrder interface{},
	rawDataJSON []byte,
	currentTime int64,
) (*takeoutModel.TakeoutOrder, error) {
	// 类型断言
	placeOrderReq, ok := platformOrder.(*LineManPlaceOrderRequest)
	if !ok {
		return nil, errors.New("平台订单数据类型错误，期望 *LineManPlaceOrderRequest")
	}

	order := &takeoutModel.TakeoutOrder{
		BaseModel: takeoutModel.BaseModel{
			Uuid:       orderUuid,
			CreateTime: currentTime,
			UpdateTime: currentTime,
		},
		Platform:           platform,
		PlatformOrderId:    platformOrderId,
		PlatformOrderState: "RECEIVED",                                 // LINE MAN Webhook 推送的订单默认为已接收
		OrderState:         ConvertLineManStateToOrderState("PENDING"), // 待接单
		IsAbnormal:         0,
		StockStatus:        0, // 库存充足
		RawData:            string(rawDataJSON),
	}

	// 基础字段映射
	order.ShortOrderNumber = placeOrderReq.OrderShortCode
	order.MerchantId = "" // LINE MAN 不提供 merchantId，可能需要从其他地方获取

	// 订单类型转换
	order.OrderType = c.convertCustomerTypeToOrderType(placeOrderReq.CustomerType)
	order.OrderAcceptedType = valueobject.TakeoutOrderAcceptedTypeManual // LINE MAN 默认手动接单

	// 会员信息
	if placeOrderReq.MemberID != "" {
		order.MembershipId = placeOrderReq.MemberID
	}

	// 价格信息映射（LINE MAN 使用泰铢，不需要转换）
	order.Subtotal = placeOrderReq.RestaurantRevenue
	order.EaterPayment = placeOrderReq.RestaurantRevenue // LINE MAN 只提供餐厅收入

	// 货币信息（泰铢）
	order.CurrencyCode = "THB"
	order.CurrencySymbol = "฿"
	order.CurrencyExponent = 0 // LINE MAN 使用泰铢，无小数位

	// 时间字段映射
	order.OrderTime = c.parseISO8601Time(placeOrderReq.OrderAcceptedTime)
	order.SubmitTime = order.OrderTime

	// 解析商品数据
	if len(placeOrderReq.Items) > 0 {
		order.TakeoutOrderItems = make([]takeoutModel.TakeoutOrderItem, 0, len(placeOrderReq.Items))
		for _, item := range placeOrderReq.Items {
			// 根据 item.ID 前缀判断商品类型
			ttposProductType := 0
			if strings.HasPrefix(item.ID, "TTPOS-PACKAGE-") {
				ttposProductType = 1
			}

			orderItem := takeoutModel.TakeoutOrderItem{
				Platform:         platform,
				PlatformItemId:   item.ID,
				TtposProductType: ttposProductType,
				Quantity:         item.Quantity,
				Price:            item.UnitPrice,
				Tax:              0, // LINE MAN 不提供税费信息
				Specifications:   item.Memo,
			}

			// 解析选项（Properties）为修饰符
			if len(item.Properties) > 0 {
				orderItem.TakeoutOrderItemModifiers = make([]takeoutModel.TakeoutOrderItemModifier, 0)
				for _, prop := range item.Properties {
					for _, val := range prop.Values {
						orderItemModifier := takeoutModel.TakeoutOrderItemModifier{
							Platform:           platform,
							PlatformModifierId: fmt.Sprintf("%s-%s", prop.ID, val.ID),
							Quantity:           1,
							Price:              val.Price,
							Tax:                0,
						}
						orderItem.TakeoutOrderItemModifiers = append(orderItem.TakeoutOrderItemModifiers, orderItemModifier)
					}
				}
			}

			order.TakeoutOrderItems = append(order.TakeoutOrderItems, orderItem)
		}
	}

	// LINE MAN 不提供收货人信息（通过 LINE MAN 配送，收货人信息在平台端）
	// 促销信息映射
	if len(placeOrderReq.Items) > 0 {
		promos := c.convertLineManPromos(orderUuid, platform, placeOrderReq, currentTime)
		order.SetTakeoutOrderPromos(promos)
	}

	return order, nil
}

// convertCustomerTypeToOrderType 转换 LINE MAN 客户类型为 TTPOS 订单类型
func (c *LineManConverter) convertCustomerTypeToOrderType(customerType string) string {
	switch strings.ToUpper(customerType) {
	case "DELIVERY":
		return valueobject.TakeoutOrderTypeDelivery // 平台配送
	case "PICKUP":
		return valueobject.TakeoutOrderTypeTakeaway // 自提
	default:
		return valueobject.TakeoutOrderTypeDelivery
	}
}

// parseISO8601Time 解析 ISO 8601 格式时间为 Unix 时间戳
// 支持格式: 2022-11-01T13:08:06+07:00
func (c *LineManConverter) parseISO8601Time(timeStr string) int64 {
	if timeStr == "" {
		return 0
	}

	// 尝试多种时间格式
	formats := []string{
		time.RFC3339,                    // 2022-11-01T13:08:06+07:00
		"2006-01-02T15:04:05Z",          // 2022-11-01T13:08:06Z
		"2006-01-02T15:04:05.000Z",      // 2022-11-01T13:08:06.000Z
		"2006-01-02T15:04:05-07:00",     // 2022-11-01T13:08:06-07:00
		"2006-01-02T15:04:05.000-07:00", // 2022-11-01T13:08:06.000-07:00
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t.Unix()
		}
	}

	return 0
}

// convertLineManPromos 将 LINE MAN 订单商品促销转换为 TTPOS 订单促销格式
func (c *LineManConverter) convertLineManPromos(
	orderUuid uint64,
	platform string,
	placeOrderReq *LineManPlaceOrderRequest,
	currentTime int64,
) []*takeoutModel.TakeoutOrderPromo {
	promos := make([]*takeoutModel.TakeoutOrderPromo, 0)

	// 遍历所有商品，提取促销信息
	for _, item := range placeOrderReq.Items {
		if item.PromotionID != "" && item.Discount > 0 {
			promo := &takeoutModel.TakeoutOrderPromo{
				BaseModel: takeoutModel.BaseModel{
					CreateTime: currentTime,
					UpdateTime: currentTime,
				},
				TakeoutOrderUuid: fmt.Sprintf("%d", orderUuid),
				Platform:         platform,
				PromoCode:        item.PromotionID,
				PromoName:        fmt.Sprintf("商品促销 - %s", item.ID),
				PromoAmount:      item.Discount,
				PromoAmountInMin: item.Discount,
			}
			promos = append(promos, promo)
		}
	}

	return promos
}
