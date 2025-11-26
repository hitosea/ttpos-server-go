# ShopCashier 实体模型说明

## 基本信息

- **实体名称**: ShopCashier
- **表名**: shop_cashier
- **所属模块**: ttpos-erp
- **描述**: 收银员配置实体，用于存储收银员的 ERPNext 授权信息和关联关系

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | int64 | 主键 | 自增 |
| ShopUuid | string | 商店UUID | 关联商店 |
| AdminUuid | string | 商店管理员UUID | 关联管理员 |
| CashierEmail | string | 收银员邮箱 | ERPNext 用户邮箱 |
| ApiKey | string | API 密钥 | ERPNext API Key |
| ApiSecret | string | API 密钥 | ERPNext API Secret |
| CompanyAbbr | string | 公司缩写 | 如 "CFG" |
| Branch | string | 分支 | 如 "Wallace Burger (CFG)" |
| SiteCode | string | 站点编码 | 关联 erp_site.site_code |

## 关联关系

### 关联实体
- **ShopUuid** → Shop（商店）
- **AdminUuid** → Admin（管理员）
- **SiteCode** → Site.SiteCode（站点配置）
- **CashierEmail** → ERPNext User（ERPNext 用户）

### 被引用
- **身份模拟** - 通过 CashierEmail 获取授权信息

## 数据流分析

### 数据来源
- 收银员配置
- 管理员分配收银员

### 数据流向

1. **收银员授权流程**:
   - 查询收银员配置
   - 获取 API Key/Secret
   - 生成 Authorization Header
   - 以收银员身份执行 ERPNext 操作

2. **POS 发票创建流程**:
   - 使用收银员身份创建发票
   - 确保发票创建者正确

### 业务场景
- 收银员身份模拟
- POS 发票创建
- 权限隔离

## 授权机制

### 收银员授权
ERPNext 使用 Token 授权方式：

```go
authorization := fmt.Sprintf("token %s:%s", cashier.ApiKey, cashier.ApiSecret)
```

## 索引建议

- 主键索引: Id
- 唯一索引: (ShopUuid, CashierEmail)（商店+收银员唯一）
- 普通索引: CashierEmail（邮箱查询）
- 普通索引: SiteCode（站点查询）
- 普通索引: CompanyAbbr（公司查询）

## 注意事项

1. **安全存储**: API Key/Secret 应加密存储
2. **身份模拟**: 用于以收银员身份创建 POS 发票
3. **权限控制**: 收银员只能操作其关联的商店数据
4. **站点关联**: 必须关联正确的站点编码

