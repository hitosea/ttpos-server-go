## 保存支付方式 selling.SellingService / SaveModeOfPayment

### 请求
```json

{
    "company_abbr": "wallace-mg",
    "branch": "曼谷55533",
    "channel": "LineWO",
    "pay_type": "UnionPay"
}

```

### 响应
```json

{
    "code": "0",
    "message": "保存支付方式成功",
    "data": {
        "name": "LineWO-UnionPay-0001 - wallace-mg",
        "@type": "type.googleapis.com/selling.SaveModeOfPaymentResp"
    }
}
````


## 无渠道时

```json
{
    "company_abbr": "wallace-mg",
    "branch": "曼谷55533",
    "pay_type": "UnionPay"
}
```

```json

{
    "code": "0",
    "message": "保存支付方式成功",
    "data": {
        "name": "UnionPay-0001 - wallace-mg",
        "@type": "type.googleapis.com/selling.SaveModeOfPaymentResp"
    }
}
```


-----

## 启用或禁用
```json

{
    "company_abbr": "wallace-mg",
        "branch": "曼谷55533",
    "name":"LineW4-Alipay-0004 - wallace-mg",
    "enabled": true
}

```

```
{
    "code": "0",
    "message": "保存支付方式成功",
    "data": {
        "name": "LineW4-Alipay-0004 - wallace-mg",
        "@type": "type.googleapis.com/selling.SaveModeOfPaymentResp"
    }
}
```