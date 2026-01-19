# WarehouseLogistics 实体模型说明

## 基本信息

- **实体名称**: WarehouseLogistics
- **表名**: warehouse_logistics
- **所属模块**: ttpos-erp
- **描述**: 仓库物流关联实体，用于关联仓库和物流供应商

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | int64 | 主键 | 自增 |
| SiteCode | string | 站点编码 | 关联 erp_site.site_code |
| ShopUuid | string | ttpos 商铺ID | 关联商店 |
| WarehouseCode | string | 仓库编码 | erpnext warehouse |
| LogisticsId | int64 | 物流ID | 关联 logistics.id |

## 关联关系

### 关联实体
- **SiteCode** → Site.SiteCode（站点配置）
- **ShopUuid** → Shop（商店）
- **WarehouseCode** → ERPNext Warehouse（ERPNext 仓库）
- **LogisticsId** → Logistics.Id（物流供应商）

### 被引用
- 无直接引用（关联表）

## 数据流分析

### 数据来源
- 仓库物流配置
- 系统初始化脚本

### 数据流向

1. **配置流程**:
   - 为仓库配置物流供应商
   - 建立仓库与物流的关联关系

2. **使用流程**:
   - 根据仓库查询关联的物流供应商
   - 使用物流供应商的接口信息调用物流 API

### 业务场景
- 仓库物流关联
- 发货物流选择
- 物流接口调用

## 索引建议

- 主键索引: Id
- 普通索引: SiteCode（站点查询）
- 普通索引: ShopUuid（商店查询）
- 普通索引: WarehouseCode（仓库查询）
- 普通索引: LogisticsId（物流查询）
- 唯一索引: (SiteCode, ShopUuid, WarehouseCode)（站点+商店+仓库唯一）

## 注意事项

1. **多对一关系**: 一个仓库可以关联一个物流供应商
2. **站点隔离**: 通过 SiteCode 区分不同租户的数据
3. **仓库编码**: WarehouseCode 对应 ERPNext 的 Warehouse DocType 名称

## 使用场景

### 查询仓库物流

```go
// 根据仓库编码查询物流供应商
warehouseLogistics, err := dao.WarehouseLogistics.Ctx(ctx).
    Where(dao.WarehouseLogistics.Columns().WarehouseCode, warehouseCode).
    Where(dao.WarehouseLogistics.Columns().SiteCode, siteCode).
    One()

// 查询物流供应商信息
logistics, err := dao.Logistics.Ctx(ctx).
    Where(dao.Logistics.Columns().Id, warehouseLogistics.LogisticsId).
    One()
```

