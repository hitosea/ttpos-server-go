# 配置变更日志 - v2.13.1

## 📋 变更概要

版本 v2.13.1 新增了 LINE MAN 平台 endpoint 配置支持，用于 OAuth Access Token 获取功能。

## ✨ 新增配置

### 环境变量

| 变量名 | 描述 | Staging 默认值 | Production 默认值 |
|--------|------|----------------|-------------------|
| `LINEMAN_PLATFORM_ENDPOINT` | LINE MAN API 端点地址 | `https://beta-partner-order.wndv.co` | 待确认 |

### config.tpl.yaml 变更

**位置**: `app.provider.lineman.platform`

```diff
  lineman:
    platform:
      clientId: "$LINEMAN_PLATFORM_CLIENT_ID"
      clientSecret: "$LINEMAN_PLATFORM_CLIENT_SECRET"
      secretKey: "$LINEMAN_PLATFORM_SECRET_KEY"
+     endpoint: "$LINEMAN_PLATFORM_ENDPOINT"  # ✨ 新增
      environment: "$LINEMAN_ENV"
      timeout: "30s"
```

### Go 数据结构变更

**文件**: `internal/model/conf/provider.go`

```diff
  type Lineman struct {
      ClientID     string        `json:"clientId"`
      ClientSecret string        `json:"clientSecret"`
      SecretKey    string        `json:"secretKey"`
+     Endpoint     string        `json:"endpoint"`     // ✨ 新增
      Environment  string        `json:"environment"`
      Timeout      time.Duration `json:"timeout"`
  }
```

## 🔧 部署清单

### 1. 更新环境变量

**Staging 环境**：
```bash
LINEMAN_PLATFORM_ENDPOINT="https://beta-partner-order.wndv.co"
```

**Production 环境**：
```bash
LINEMAN_PLATFORM_ENDPOINT="https://partner-order.lineman.com"  # 待确认实际地址
```

### 2. 更新配置文件

确保 `config.tpl.yaml` 已更新为最新版本（包含 endpoint 配置项）。

### 3. 重新部署服务

```bash
# 拉取最新代码
git pull origin main

# 重新构建
make build

# 重启服务
docker-compose restart ttpos-takeout
# 或
kubectl rollout restart deployment/ttpos-takeout
```

## ⚠️ 注意事项

1. **必须配置**：`LINEMAN_PLATFORM_ENDPOINT` 是新增的必需环境变量，未配置会导致 OAuth Token 请求失败
2. **环境区分**：Staging 和 Production 使用不同的 endpoint 地址
3. **向后兼容**：如果暂时不使用 OAuth Access Token 功能，可以不立即更新此配置
4. **Production 地址确认**：部署到生产环境前，需要向 LINE MAN 技术支持确认正确的 API 端点地址

## 📚 相关文档

- 完整配置示例：`manifest/config/lineman-env-example.md`
- 需求提案：`docs/team/proposals/2026-01/v2.13.1-lineman-access-token.md`
- 配置模板：`manifest/config/config.tpl.yaml`

## 🔗 相关需求

- 提案：LINE MAN OAuth Access Token 缓存功能
- 参考实现：Grab OAuth Token 获取与缓存机制

## 📅 发布日期

2026-01-07

---

**维护者**: rikugun  
**版本**: v2.13.1  
**状态**: 待部署

