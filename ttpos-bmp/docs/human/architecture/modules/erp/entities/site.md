# Site 实体模型说明

## 基本信息

- **实体名称**: Site
- **表名**: site
- **所属模块**: ttpos-erp
- **描述**: ERP 站点配置实体，用于存储多租户 ERPNext 站点的连接信息

## 字段说明

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| Id | int64 | 主键 | 自增 |
| Uuid | string | UUID | 站点唯一标识 |
| SiteName | string | 站点名称 | 显示名称 |
| SiteUrl | string | 站点地址 | ERPNext 访问地址 |
| Remark | string | 备注 | |
| SiteCode | string | 站点编码 | 与 ttpos 映射，如 "0", "1", "4" |
| ApiKey | string | API 密钥 | ERPNext API Key |
| ApiSecret | string | API 密钥 | ERPNext API Secret |

## 关联关系

### 关联实体
- 无直接关联实体

### 被引用
- **WarehouseLogistics.SiteCode** → Site.SiteCode（仓库物流关联）
- **gRPC 上下文** - 通过 Metadata 传递站点编码
- **HTTP 客户端** - 根据站点编码获取授权信息

## 数据流分析

### 数据来源
- 运维人员配置
- 系统初始化脚本

### 数据流向

1. **站点授权流程**:
   - gRPC 请求携带站点编码（Metadata）
   - 从 site 表查询站点配置
   - 生成 Authorization Header
   - 设置 HTTP 客户端 Prefix

2. **多租户路由流程**:
   - 根据站点编码查询配置
   - 动态切换 ERPNext 站点
   - 隔离不同租户数据

### 业务场景
- 多租户 ERPNext 对接
- 站点级别权限隔离
- API 授权管理

## 站点编码说明

| 编码 | 名称 | 说明 |
|-----|------|------|
| 0 | UAT | 测试环境 |
| 1 | TTPOS | TTPOS 正式站点 |
| 4 | Wallace | Wallace 站点 |

## 授权机制

### API Key/Secret 授权
ERPNext 使用 Token 授权方式：

```
Authorization: token {api_key}:{api_secret}
```

## 索引建议

- 主键索引: Id
- 唯一索引: SiteCode（站点编码查询）
- 唯一索引: Uuid（UUID 查询）

## 注意事项

1. **安全存储**: API Key/Secret 应加密存储
2. **权限控制**: 不同站点的数据严格隔离
3. **默认站点**: 未指定站点时使用 UAT 或配置默认
4. **缓存策略**: 站点配置可缓存提升性能

## 配置示例

```sql
INSERT INTO site (uuid, site_name, site_url, site_code, api_key, api_secret, remark) VALUES
(UUID(), 'UAT', 'https://uat.erp.example.com', '0', 'uat_api_key', 'uat_api_secret', '测试环境'),
(UUID(), 'TTPOS', 'https://ttpos.erp.example.com', '1', 'ttpos_api_key', 'ttpos_api_secret', 'TTPOS正式站点'),
(UUID(), 'Wallace', 'https://wallace.erp.example.com', '4', 'wallace_api_key', 'wallace_api_secret', 'Wallace站点');
```
