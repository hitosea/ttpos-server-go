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


----

## 品牌采购包含直送物品 异常流
  - WPR3693287586267137 有多个供应商 
  - WPR3690984957411329 无供应商

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
        },
        {
            "item_code": "WPR3693287586267137",
            "schedule_date": 2082672000,
            "qty": 1,
            "uom": "pcs"
        },
        {
            "item_code": "WPR3690984957411329",
            "qty": 1,
            "uom": "蓝象总店-lxzd-bottle"
        }
    ]
}

```


## 响应
```json
{
    "code": "1",
    "message": "创建内部销售订单失败: 调用erp接口返回异常:ValidationError;Row #4: Set Supplier for item WPR3690984957411329",
    "data": null
}
```


-----
## 品牌采购包含直送物品 正常流
  - WPR3693287586267137 有多个供应商 
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
        },
        {
            "item_code": "WPR3693287586267137",
            "schedule_date": 2082672000,
            "qty": 1,
            "uom": "pcs"
        }
    ]
}
```

## 响应

```json
{
    "code": "0",
    "message": "保存物品申请单成功",
    "data": {
        "material_request_name": "MAT-MR-2025-00401",
        "purchase_order": "PUR-ORD-2025-00529",
        "@type": "type.googleapis.com/stock.SaveMaterialRequestResp"
    }
}
```