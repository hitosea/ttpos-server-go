## 保存盘点记录 stock.StockService / SaveStockReconciliation

```json

{
  "company_abbr": "CFG",
  "branch": "蓝象总店",
  "posting_date": "2025-09-16",
  "posting_time": "06:04:27",
  "purpose": "Stock Reconciliation",
  "warehouse": "蓝象总店-Normal-Default - CFG",
  "items": [
    {
      "item_code": "WPR3685914633175041",
      "qty": 99931
    }
  ]
}
```
```json

{
    "code": "0",
    "message": "保存库存盘点成功",
    "data": {
        "stock_reconciliation_name": "MAT-RECO-2025-00012",
        "@type": "type.googleapis.com/stock.SaveStockReconciliationResp"
    }
}
```


## 提交盘点记录 stock.StockService / SubmitStockReconciliation


```json
{
  "stock_reconciliation_name": "MAT-RECO-2025-00012"
}
```

```json
{
    "code": "0",
    "message": "库存盘点单据提交成功",
    "data": {
        "message": "库存盘点单据提交成功",
        "@type": "type.googleapis.com/stock.SubmitStockReconciliationResp"
    }
}
```

## 查询盘点记录 stock.StockService / GetStockReconciliationList

```json

{
    "company_abbr": "CFG",
    "limit": 3
}
```

```json
{
    "code": "0",
    "message": "获取库存盘点列表成功",
    "data": {
        "stock_reconciliation_list": [
            {
                "items": [
                    {
                        "item_code": "WPR3685914633175041",
                        "item_name": "chicken wings",
                        "qty": 99931,
                        "warehouse": "蓝象总店-Normal-Default - CFG",
                        "valuation_rate": 1
                    }
                ],
                "name": "MAT-RECO-2025-00012",
                "company": "华莱士泰国",
                "purpose": "Stock Reconciliation",
                "posting_date": "2025-09-16",
                "posting_time": "6:04:27",
                "set_warehouse": "蓝象总店-Normal-Default - CFG"
            }
        ],
        "@type": "type.googleapis.com/stock.GetStockReconciliationListResp"
    }
}
```