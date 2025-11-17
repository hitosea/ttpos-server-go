
## 采购收货单 buying.BuyingService / SavePurchaseReceipt


### 多规格单位收货
```json

{
    "purchase_order_name": "PUR-ORD-2025-00467",
    "items": [
        {
            "item_code": "奶油-正常",
            "schedule_date": 2082672000,
            "qty": 3,
            "uom": "华莱士总部-g"
        },
        {
            "item_code": "奶油-正常",
            "schedule_date": 2082672000,
            "qty": 3,
            "uom": "华莱士总部-kg"
        }
    ]
}
```

```json
{
    "code": "0",
    "message": "保存采购收货成功",
    "data": {
        "purchase_receipt": {
            "items": [
                {
                    "item_code": "奶油-正常",
                    "item_name": "Cream - Normal",
                    "stock_uom": "华莱士总部-g",
                    "uom": "华莱士总部-g",
                    "qty": 3
                },
                {
                    "item_code": "奶油-正常",
                    "item_name": "Cream - Normal",
                    "stock_uom": "华莱士总部-g",
                    "uom": "华莱士总部-kg",
                    "qty": 3
                }
            ],
            "purchase_receipt_name": "MAT-PRE-2025-00303"
        },
        "@type": "type.googleapis.com/buying.SavePurchaseReceiptResp"
    }
}
```