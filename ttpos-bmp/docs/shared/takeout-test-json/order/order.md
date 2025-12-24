## 查询订单数据

## order.OrderService / GetOrderInfo

```json
{
    "shop_uuid": "8609817471094784",
    "order_uuid": "1jdg21y7uk8df23wdj1n99n100g77f4i",
    "request_id": "{{$string.nanoid}}"
}
```

```json
{
    "code": "0",
    "message": "success",
    "data": {
        "shop_uuid": "8609817471094784",
        "order_status": "DELIVERED",
        "order_type": "DeliveryByProvider",
        "raw_data": "{\"currency\":{\"code\":\"THB\",\"exponent\":2,\"symbol\":\"฿\"},\"cutlery\":false,\"featureFlags\":{\"isMexEditOrder\":false,\"orderAcceptedType\":\"AUTO\",\"orderType\":\"DeliveredByGrab\"},\"items\":[{\"grabItemID\":\"THITE20251219100406295285\",\"id\":\"TTPOS-ITEM-3701988558241793\",\"modifiers\":[{\"id\":\"TTPOS-SAUCE-3671416295522520\",\"price\":0,\"quantity\":1,\"tax\":0},{\"id\":\"TTPOS-ATTR-4290647166976000\",\"price\":0,\"quantity\":1,\"tax\":0}],\"price\":43389,\"quantity\":1,\"specifications\":\"\",\"tax\":0}],\"merchantID\":\"GFSBPOS-822-571\",\"orderID\":\"123456789-C7WBHBVGE76GAA\",\"orderReadyEstimation\":{\"allowChange\":true,\"estimatedOrderReadyTime\":\"2025-12-19T10:14:06.277879574Z\",\"maxOrderReadyTime\":\"2025-12-19T11:04:06.277879651Z\"},\"orderState\":\"NEW\",\"orderTime\":\"2025-12-19T10:04:06Z\",\"partnerMerchantID\":\"8609817471094784\",\"paymentType\":\"CASH\",\"price\":{\"deliveryFee\":1000,\"eaterPayment\":44389,\"grabFundPromo\":0,\"merchantFundPromo\":0,\"subtotal\":43389,\"tax\":0},\"shortOrderNumber\":\"GF-5447\"}",
        "provider_name": "grab",
        "@type": "type.googleapis.com/order.GetOrderInfoResp"
    }
}
```


----
## order.OrderService / MarkOrderReady

```json
{
    "takeout_order_uuid": "10aftdg7000df63y6ffdno1470xvplo5",
    "request_id": "{{$string.nanoid}}"
}
```

```json

{
    "code": "0",
    "message": "success",
    "data": {
        "order_uuid": "10aftdg7000df63y6ffdno1470xvplo5",
        "@type": "type.googleapis.com/order.MarkOrderReadyResp"
    }
}
```



---
## 判断订单是否可以取消 order.OrderService / CheckOrderCancelable

```json

{
    "takeout_order_uuid": "1r52xq27000df6akf5g84azl00r4ocpa",
    "request_id": "{{$string.nanoid}}"
}
```

```json
{
    "code": "0",
    "message": "success",
    "data": {
        "order_uuid": "1r52xq27000df6akf5g84azl00r4ocpa",
        "can_cancel": true,
        "non_cancellation_reason": "",
        "raw_data": "{\"cancelAble\":true,\"cancelReasons\":[{\"code\":1001,\"reason\":\"Most or all items are unavailable\"},{\"code\":1002,\"reason\":\"We're too busy right now\"},{\"code\":1003,\"reason\":\"We're closed right now\"},{\"code\":1004,\"reason\":\"We're closing soon\"}],\"cancelable\":true,\"limitTimes\":0,\"limitType\":\"not approaching limit\",\"nonCancellationReason\":\"\"}",
        "@type": "type.googleapis.com/order.CheckOrderCancelableResp"
    }
}
```


----
## 取消订单 order.OrderService / CancelOrder

```json
{
    "cancel_code": "1001",
    "takeout_order_uuid": "1r52xq27000df6b41fqk9o8s003p05s2",
    "request_id": "{{$string.nanoid}}"
}
```

```json

{
    "code": "0",
    "message": "订单已成功取消",
    "data": {
        "order_uuid": "1r52xq27000df6b41fqk9o8s003p05s2",
        "@type": "type.googleapis.com/order.CancelOrderResp"
    }
}
```