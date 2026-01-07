# LINE MAN 环境变量配置示例

## 新增环境变量（v2.13.1）

本次更新新增了以下环境变量：

```bash
# ✨ 新增：LINE MAN API 端点地址
LINEMAN_PLATFORM_ENDPOINT="https://beta-partner-order.wndv.co"  # staging 环境
```

## Staging 环境完整配置

```bash
# LINE MAN Platform 配置
LINEMAN_PLATFORM_CLIENT_ID="your_client_id"
LINEMAN_PLATFORM_CLIENT_SECRET="your_client_secret"
LINEMAN_PLATFORM_SECRET_KEY="your_jwt_secret_key"
LINEMAN_PLATFORM_ENDPOINT="https://beta-partner-order.wndv.co"  # ✨ 新增
LINEMAN_ENV="staging"

# LINE MAN Partner 配置（如果需要）
LINEMAN_PARTNER_DEFAULT_CLIENT_ID="partner_client_id"
LINEMAN_PARTNER_DEFAULT_CLIENT_SECRET="partner_client_secret"
```

## Production 环境完整配置

```bash
# LINE MAN Platform 配置
LINEMAN_PLATFORM_CLIENT_ID="your_prod_client_id"
LINEMAN_PLATFORM_CLIENT_SECRET="your_prod_client_secret"
LINEMAN_PLATFORM_SECRET_KEY="your_prod_jwt_secret_key"
LINEMAN_PLATFORM_ENDPOINT="https://partner-order.lineman.com"  # ✨ 新增（待确认）
LINEMAN_ENV="production"

# LINE MAN Partner 配置（如果需要）
LINEMAN_PARTNER_DEFAULT_CLIENT_ID="partner_prod_client_id"
LINEMAN_PARTNER_DEFAULT_CLIENT_SECRET="partner_prod_client_secret"
```

## 配置说明

### Platform 配置（app.provider.lineman.platform）
用于系统主动调用 LINE MAN API 时的认证。

- `LINEMAN_PLATFORM_CLIENT_ID`: LINE MAN 分配的客户端 ID
- `LINEMAN_PLATFORM_CLIENT_SECRET`: LINE MAN 分配的客户端密钥
- `LINEMAN_PLATFORM_SECRET_KEY`: JWT 签名密钥（用于生成 Partner Token）
- `LINEMAN_PLATFORM_ENDPOINT`: LINE MAN API 端点地址
  - Staging: `https://beta-partner-order.wndv.co`
  - Production: 待确认
- `LINEMAN_ENV`: 环境标识（staging/production）

### Partner 配置（app.provider.lineman.partner）
用于接收 LINE MAN 回调时验证 Partner Token。

- `LINEMAN_PARTNER_DEFAULT_CLIENT_ID`: 默认 Partner 客户端 ID
- `LINEMAN_PARTNER_DEFAULT_CLIENT_SECRET`: 默认 Partner 客户端密钥

## 在 config.tpl.yaml 中的使用

环境变量通过 `$变量名` 的形式在配置模板中引用：

```yaml
# manifest/config/config.tpl.yaml

app:
  provider:
    lineman:
      platform:
        clientId: "$LINEMAN_PLATFORM_CLIENT_ID"          # 引用环境变量
        clientSecret: "$LINEMAN_PLATFORM_CLIENT_SECRET"  # 引用环境变量
        secretKey: "$LINEMAN_PLATFORM_SECRET_KEY"        # 引用环境变量
        endpoint: "$LINEMAN_PLATFORM_ENDPOINT"           # ✨ 新增：引用环境变量
        environment: "$LINEMAN_ENV"                      # 引用环境变量
        timeout: "30s"
      partner:
        $LINEMAN_PARTNER_DEFAULT_CLIENT_ID:              # 动态 key，引用环境变量
          clientId: "$LINEMAN_PARTNER_DEFAULT_CLIENT_ID"
          clientSecret: "$LINEMAN_PARTNER_DEFAULT_CLIENT_SECRET"
          environment: "$LINEMAN_ENV"
          timeout: "60s"
```

## 部署时的注入方式

### Docker Compose

```yaml
# docker-compose.yml
services:
  ttpos-takeout:
    environment:
      - LINEMAN_PLATFORM_ENDPOINT=https://beta-partner-order.wndv.co
      - LINEMAN_PLATFORM_CLIENT_ID=${LINEMAN_PLATFORM_CLIENT_ID}
      - LINEMAN_PLATFORM_CLIENT_SECRET=${LINEMAN_PLATFORM_CLIENT_SECRET}
      - LINEMAN_PLATFORM_SECRET_KEY=${LINEMAN_PLATFORM_SECRET_KEY}
      - LINEMAN_ENV=staging
```

### Kubernetes

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
spec:
  template:
    spec:
      containers:
      - name: ttpos-takeout
        env:
        - name: LINEMAN_PLATFORM_ENDPOINT
          value: "https://beta-partner-order.wndv.co"
        - name: LINEMAN_PLATFORM_CLIENT_ID
          valueFrom:
            secretKeyRef:
              name: lineman-credentials
              key: client-id
        - name: LINEMAN_PLATFORM_CLIENT_SECRET
          valueFrom:
            secretKeyRef:
              name: lineman-credentials
              key: client-secret
        - name: LINEMAN_PLATFORM_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: lineman-credentials
              key: secret-key
        - name: LINEMAN_ENV
          value: "staging"
```

### 直接运行（开发环境）

```bash
# 设置环境变量
export LINEMAN_PLATFORM_ENDPOINT="https://beta-partner-order.wndv.co"
export LINEMAN_PLATFORM_CLIENT_ID="your_client_id"
export LINEMAN_PLATFORM_CLIENT_SECRET="your_client_secret"
export LINEMAN_PLATFORM_SECRET_KEY="your_jwt_secret_key"
export LINEMAN_ENV="staging"

# 运行服务
./ttpos-takeout
```

## 相关文件

- 配置模板: `manifest/config/config.tpl.yaml` - 定义环境变量占位符
- 配置结构: `internal/model/conf/provider.go` - 定义配置数据结构
- Token 服务: `internal/logic/lineman_token/lineman_token.go` - 使用配置

## 变更历史

### v2.13.1 (2026-01-07)
- ✨ 新增 `LINEMAN_PLATFORM_ENDPOINT` 环境变量
- 📝 更新配置模板和数据结构以支持 endpoint 配置

## 更新日期

2026-01-07

