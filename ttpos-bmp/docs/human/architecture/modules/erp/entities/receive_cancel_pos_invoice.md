# ReceiveCancelPosInvoice 实体模型说明

## 基本信息

- **实体名称**: ReceiveCancelPosInvoice
- **表名**: receive_cancel_pos_invoice
- **所属模块**: ttpos-erp
- **描述**: 接收取消 POS 发票记录实体，用于记录从 POS 系统接收并发送到 ERPNext 的取消发票请求和响应

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | int64 | 主键 | 自增 |
| OrderNo | string | 退款订单号 | 来自 ttpos |
| OpenPosEntryName | string | POS 开账名称 | 关联 POS Opening Entry |
| Docstatus | string | 文档状态 | 参考 erpnext，如 "Cancelled" |
| CreatedAt | int | 创建时间 | Unix 时间戳 |
| UpdatedAt | int | 更新时间 | Unix 时间戳 |
| ReqMessage | string | 请求数据 | base64 编码 |
| RespMessage | string | 响应数据 | base64 编码 |
| SiteCode | string | 站点编码 | erp_site_code，用来区分调哪个租户 |
| ReqBody | string | 请求文本 | 如果能转换 |
| RespBody | string | 响应文本 | 如果能转换 |

## 关联关系

### 关联实体
- **SiteCode** → Site.SiteCode（站点配置）
- **OpenPosEntryName** → ERPNext POS Opening Entry（开账记录）

### 被引用
- 无直接引用（记录表）

## 数据流分析

### 数据来源
- POS 系统取消发票请求
- 异步消息队列消费

### 数据流向

1. **取消发票流程**:
   - 接收 POS 系统取消发票请求
   - 记录请求数据（ReqMessage/ReqBody）
   - 调用 ERPNext API 取消发票
   - 记录响应数据（RespMessage/RespBody）

2. **数据追踪流程**:
   - 通过 OrderNo 查询订单记录
   - 用于问题排查和审计

### 业务场景
- POS 取消发票记录追踪
- 请求/响应审计
- 问题排查
- 数据同步状态监控

## 文档状态说明

| 状态值 | 说明 |
|-------|------|
| Cancelled | 已取消 |

## 索引建议

- 主键索引: Id
- 普通索引: OrderNo（订单号查询）
- 普通索引: OpenPosEntryName（开账记录查询）
- 普通索引: SiteCode（站点查询）
- 普通索引: Docstatus（状态查询）
- 普通索引: CreatedAt（时间范围查询）
- 复合索引: (SiteCode, OrderNo)（站点+订单查询）

## 注意事项

1. **数据记录**: 记录完整的请求和响应数据，便于问题排查
2. **Base64 编码**: ReqMessage/RespMessage 使用 base64 编码存储二进制数据
3. **站点隔离**: 通过 SiteCode 区分不同租户的数据
4. **审计追踪**: 所有操作都有完整的请求/响应记录
5. **取消操作**: 取消发票是不可逆操作，需要谨慎处理

## 使用场景

### 查询取消发票记录

```go
// 根据订单号查询取消发票记录
cancelInvoice, err := dao.ReceiveCancelPosInvoice.Ctx(ctx).
    Where(dao.ReceiveCancelPosInvoice.Columns().OrderNo, orderNo).
    Where(dao.ReceiveCancelPosInvoice.Columns().SiteCode, siteCode).
    One()
```

### 问题排查

```go
// 查看请求和响应数据
reqData, _ := base64.StdEncoding.DecodeString(cancelInvoice.ReqMessage)
respData, _ := base64.StdEncoding.DecodeString(cancelInvoice.RespMessage)
```

