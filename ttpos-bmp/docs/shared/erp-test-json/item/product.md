# 商品服务

## 更新商品 item.ProductService / UpdateProduct

### 更新物品单位
```json
{
    "item_code": "SP3691844810702849_00",
    "stock_uom": "Cup"
}
```

响应

```json
{
    "code": "0",
    "message": "更新商品信息成功",
    "data": {
        "attributes": [],
        "item_code": "SP3691844810702849_00",
        "not_for_sale": false,
        "internal_code": "",
        "disabled": false,
        "stock_uom": "Cup",
        "@type": "type.googleapis.com/item.UpdateProductResp"
    }
}
```