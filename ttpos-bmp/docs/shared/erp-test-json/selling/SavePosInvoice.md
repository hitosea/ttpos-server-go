# 保存发票

```json

{
    "open_pos_entry_name": "POS-OPE-2025-00326",
    "company_abbr": "myj02d",
    "posting_datetime":  "{{$date.timestamp}}",
    "update_stock": 1,
    "currency": "THB",
    "price_list_currency": "THB",
    "branch": "美宜佳-2号店",
    "customer_uuid": "7222333311111",
    "order_no": "202509292821017621",
    "items": [
        {
            "item_code": "SP3685878339862529_01",
            "qty": 20,
            "rate": 25,
            "amount": 500
        }
    ],
    "payments": [
        {
            "mode_of_payment": "Cash",
            "amount": 500
        }
    ],
    "remark":"batch redo"
}
```

```json

{
    "code": "0",
    "message": "保存发票成功",
    "data": {
        "products_invoice_name": "",
        "material_invoice_name": "",
        "async_record_id": "18",
        "@type": "type.googleapis.com/selling.SavePosInvoiceResp"
    }
}
```