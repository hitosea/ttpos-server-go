# ReceiveClosePos 实体模型说明

## 基本信息

- **实体名称**: ReceiveClosePos
- **表名**: receive_close_pos
- **所属模块**: ttpos-erp
- **描述**: 接收关账记录实体，用于记录从 POS 系统接收并发送到 ERPNext 的关账请求和响应

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | int64 | 主键 | 自增 |
| PosOpenEntryName | string | 开账名称 | 关联 POS Opening Entry |
| PeriodEndDate | int64 | 结账时间 | Unix 时间戳 |
| Docstatus | string | 文档状态 | 参考 erpnext，如 "Draft", "Submitted" |
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
- **PosOpenEntryName** → ERPNext POS Opening Entry（开账记录）

### 被引用
- 无直接引用（记录表）

## 数据流分析

### 数据来源
- POS 系统关账请求
- 异步消息队列消费

### 数据流向

1. **关账流程**:
   - 接收 POS 系统关账请求
   - 记录请求数据（ReqMessage/ReqBody）
   - 调用 ERPNext API 创建关账记录
   - 记录响应数据（RespMessage/RespBody）

2. **数据追踪流程**:
   - 通过 PosOpenEntryName 查询开账记录
   - 用于问题排查和审计

### 业务场景
- POS 关账记录追踪
- 请求/响应审计
- 问题排查
- 数据同步状态监控

## 文档状态说明

| 状态值 | 说明 |
|-------|------|
| Draft | 草稿，创建中 |
| Submitted | 已提交，创建成功 |
| Cancelled | 已取消 |

## 索引建议

- 主键索引: Id
- 普通索引: PosOpenEntryName（开账记录查询）
- 普通索引: SiteCode（站点查询）
- 普通索引: Docstatus（状态查询）
- 普通索引: CreatedAt（时间范围查询）
- 复合索引: (SiteCode, PosOpenEntryName)（站点+开账查询）

## 注意事项

1. **数据记录**: 记录完整的请求和响应数据，便于问题排查
2. **Base64 编码**: ReqMessage/RespMessage 使用 base64 编码存储二进制数据
3. **站点隔离**: 通过 SiteCode 区分不同租户的数据
4. **审计追踪**: 所有操作都有完整的请求/响应记录
5. **关账时间**: PeriodEndDate 记录关账的时间点

## 使用场景

### 查询关账记录

```go
// 根据开账名称查询关账记录
closePos, err := dao.ReceiveClosePos.Ctx(ctx).
    Where(dao.ReceiveClosePos.Columns().PosOpenEntryName, openEntryName).
    Where(dao.ReceiveClosePos.Columns().SiteCode, siteCode).
    One()
```

### 问题排查

```go
// 查看请求和响应数据
reqData, _ := base64.StdEncoding.DecodeString(closePos.ReqMessage)
respData, _ := base64.StdEncoding.DecodeString(closePos.RespMessage)
```

