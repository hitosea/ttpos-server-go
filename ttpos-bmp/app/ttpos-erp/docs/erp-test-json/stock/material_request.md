## 品牌采购申请 stock.StockService / SaveMaterialRequest

```json

{
  "transaction_date": 1762153617,
  "company_abbr": "wfs-002",
  "branch": "重庆高老九火锅曼谷一号店",
  "required_by": 2082672000,
  "source_warehouse": "ttpos@qq.com-Normal-Default - CFG",
  "target_warehouse": "重庆高老九火锅曼谷一号店-Normal-Default - wfs-002",
  "supplier": "Headquarters - Supplier",
  "items": [
    {
      "item_code": "奶油-正常",
      "schedule_date": 2082672000,
      "qty": 100,
      "uom": "华莱士总部-g"
    },
    {
      "item_code": "奶油-正常",
      "schedule_date": 2082672000,
      "qty": 100,
      "uom": "华莱士总部-kg"
    }
  ]
}
```

```json
{
    "code": "0",
    "message": "保存物品申请单成功",
    "data": {
        "material_request_name": "MAT-MR-2025-00397",
        "purchase_order": "PUR-ORD-2025-00467",
        "@type": "type.googleapis.com/stock.SaveMaterialRequestResp"
    }
}
```


-----

## 收货 