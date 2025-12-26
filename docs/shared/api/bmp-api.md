# BMP 微服务 API 文档

> 本文档描述 BMP (Business Microservice Platform) 微服务模块的 API 接口。

## 📋 目录

- [库存服务 (Stock Service)](#库存服务-stock-service)

## 📦 库存服务 (Stock Service)

### GetItemStockByBin - 根据仓库和商品代码查询货位库存信息

#### 接口描述
根据仓库代码和可选的商品代码查询货位库存信息，返回按货位分组的库存数据。

#### 请求参数

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| warehouse | string | 是 | 仓库代码 |
| item_code | string | 否 | 商品代码，为空时返回该仓库所有商品 |

#### 响应数据

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "items": [
      {
        "item_code": "string",              // 商品代码
        "actual_qty": "string",             // 实际库存数量
        "projected_qty": "string",           // 预计库存数量
        "reserved_qty_for_pos": "string",    // POS预留数量
        "stock_uom": "string",               // 库存单位
        "valuation_rate": "string"           // 估价率
      }
    ]
  }
}
```

#### 字段说明

| 字段名 | 类型 | 说明 |
|--------|------|------|
| item_code | string | 商品代码 |
| actual_qty | string | 实际库存数量 |
| projected_qty | string | 预计库存数量 |
| reserved_qty_for_pos | string | POS预留数量 |
| stock_uom | string | 库存单位 |
| valuation_rate | string | 估价率 |

#### 错误响应

```json
{
  "code": 0,
  "message": "错误信息",
  "data": {}
}
```

#### 使用示例

```bash
# 查询指定仓库的所有商品库存
grpcurl -plaintext localhost:14022 ttpos_erp.StockService/GetItemStockByBin \
  -d '{"warehouse": "MAIN_WH"}'

# 查询指定仓库和商品的库存
grpcurl -plaintext localhost:14022 ttpos_erp.StockService/GetItemStockByBin \
  -d '{"warehouse": "MAIN_WH", "item_code": "ITEM001"}'
```

#### 服务信息

- **服务名**: StockService
- **方法名**: GetItemStockByBin
- **端口**: 14022 (gRPC)
- **模块**: ttpos-erp

---

**版本**: v1.0.0
**最后更新**: 2025-12-26
**维护者**: BMP 开发组
