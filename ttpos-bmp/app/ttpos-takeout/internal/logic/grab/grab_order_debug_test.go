package grab

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	grabfood "github.com/grab/grabfood-api-sdk-go"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2" // MySQL 驱动
	_ "github.com/gogf/gf/contrib/nosql/redis/v2"   // Redis 驱动
)

// TestSaveOrderFromSDK_WithItems 测试带多个商品和配料的订单
func TestSaveOrderFromSDK_WithItems(t *testing.T) {
	ctx := context.Background()
	s := &sGrab{}

	orderID := "G-ITEMS-8609817471094786"
	partnerMerchantID := "8609817471094784"

	req := &grabfood.SubmitOrderRequest{}
	req.SetOrderID(orderID)
	req.SetShortOrderNumber("M002")
	req.SetMerchantID("GFSBPOS-512-084")
	req.SetPartnerMerchantID(partnerMerchantID)
	req.SetPaymentType("CASHLESS")
	req.SetOrderState("")
	req.SetOrderTime(gtime.Now().Layout("2006-01-02T15:04:05Z"))

	// 货币
	currency := grabfood.Currency{}
	currency.SetCode("THB")
	currency.SetExponent(2)
	req.SetCurrency(currency)

	// 多个商品
	items := []grabfood.OrderItem{
		func() grabfood.OrderItem {
			item := grabfood.OrderItem{}
			item.SetId("TTPOS-ITEM-592")
			item.SetQuantity(4)
			item.SetPrice(8000) // 80.00 THB
			item.SetSpecifications("绿咖喱鸡饭")
			// 修饰符/配料
			mod := grabfood.OrderItemModifier{}
			mod.SetId("TTPOS-FLAVOR-592")
			mod.SetQuantity(1)
			mod.SetPrice(0) // 5.00 THB
			item.SetModifiers([]grabfood.OrderItemModifier{mod})
			return item
		}(),
		func() grabfood.OrderItem {
			item := grabfood.OrderItem{}
			item.SetId("TTPOS-ITEM-590")
			item.SetQuantity(5)
			item.SetPrice(16500) // 50.00 THB
			item.SetSpecifications("泰式奶茶3")

			// 修饰符/配料
			mod := grabfood.OrderItemModifier{}
			mod.SetId("TTPOS-FLAVOR-590")
			mod.SetQuantity(1)
			mod.SetPrice(0) // 5.00 THB
			item.SetModifiers([]grabfood.OrderItemModifier{mod})
			return item
		}(),
	}
	req.SetItems(items)

	// 价格汇总
	price := grabfood.OrderPrice{}
	price.SetSubtotal(21500) // (8000*2 + 500) + 5000
	price.SetTax(1505)
	price.SetEaterPayment(23005)
	req.SetPrice(price)

	// 功能标识
	featureFlags := grabfood.OrderFeatureFlags{}
	featureFlags.SetOrderAcceptedType("AUTO")
	featureFlags.SetOrderType("DeliveredByGrab")
	req.SetFeatureFlags(featureFlags)

	// 执行
	g.Log().Infof(ctx, "开始调试 HandleSubmitOrder: orderID=%s", orderID)

	err := s.HandleSubmitOrder(ctx, req)
	if err != nil {
		t.Fatalf("HandleSubmitOrder failed: %v", err)
	}

	g.Log().Infof(ctx, "✅ HandleSubmitOrder 成功: orderID=%s", orderID)
}
