# Logistics 实体模型说明

## 基本信息

- **实体名称**: Logistics
- **表名**: logistics
- **所属模块**: ttpos-erp
- **描述**: 物流供应商实体，用于存储物流供应商的配置信息

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | int64 | 主键 | 自增 |
| Uuid | string | UUID | 物流供应商唯一标识 |
| Vendor | string | 供应商 | 如 "JT:极兔" |
| VendorUserId | string | 供应商用户ID | 如极兔的货主编码 |
| InfConf | string | 接口连接信息 | 如 ak/sk，根据不同供应商有所不同 |
| Remarks | string | 备注信息 | |
| Reserve1 | string | 保留字段1 | |
| Reserve2 | string | 保留字段2 | |

## 关联关系

### 关联实体
- 无直接关联实体

### 被引用
- **WarehouseLogistics.LogisticsId** → Logistics.Id（仓库物流关联）

## 数据流分析

### 数据来源
- 物流供应商配置
- 系统初始化脚本

### 数据流向

1. **物流配置流程**:
   - 配置物流供应商信息
   - 设置接口连接信息（API Key/Secret 等）
   - 关联到仓库

2. **物流使用流程**:
   - 根据仓库查询关联的物流供应商
   - 使用物流供应商的接口信息调用物流 API

### 业务场景
- 物流供应商管理
- 仓库物流关联
- 物流接口调用

## 供应商类型

| 供应商 | Vendor 值 | 说明 |
|-------|----------|------|
| 极兔 | JT | 极兔物流 |

## 索引建议

- 主键索引: Id
- 唯一索引: Uuid
- 普通索引: Vendor（供应商查询）

## 注意事项

1. **接口配置**: InfConf 字段存储 JSON 格式的接口配置信息
2. **供应商扩展**: 通过 Vendor 字段区分不同供应商
3. **保留字段**: Reserve1/Reserve2 用于未来扩展

## 使用场景

### 查询物流供应商

```go
// 根据 UUID 查询物流供应商
logistics, err := dao.Logistics.Ctx(ctx).
    Where(dao.Logistics.Columns().Uuid, uuid).
    One()
```

### 解析接口配置

```go
// 解析接口配置信息
var infConf map[string]interface{}
json.Unmarshal([]byte(logistics.InfConf), &infConf)
apiKey := infConf["api_key"].(string)
apiSecret := infConf["api_secret"].(string)
```

