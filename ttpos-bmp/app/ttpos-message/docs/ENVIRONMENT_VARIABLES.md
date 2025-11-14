# 环境变量配置指南

## 📋 概述

ttpos-message 服务使用环境变量进行配置管理，支持从 `.env` 文件或系统环境变量中读取配置。

## 🔧 配置方式

### 方式一：使用 .env 文件（推荐）

在项目根目录（`ttpos-bmp/`）创建 `.env` 文件：

```bash
cd /path/to/ttpos-bmp
touch .env
```

编辑 `.env` 文件，添加以下内容：

```bash
# ==================== 数据库配置 ====================
DB_USERNAME=root
DB_PASSWORD=your_password
DB_HOST=127.0.0.1
DB_PORT=3306
DB_PORT_OPEN=3306
DB_NAME=messages

# ==================== Nacos 配置 ====================
NACOS_SERVER_IP=127.0.0.1
NACOS_SERVER_PORT=8848

# ==================== RocketMQ 配置 ====================
ROCKETMQ_ENDPOINT=127.0.0.1:8081
ROCKETMQ_NAME_SRV_ADDR=127.0.0.1:9876
ROCKETMQ_ACCESS_KEY=your_access_key      # 可选
ROCKETMQ_SECRET_KEY=your_secret_key      # 可选
ROCKETMQ_BROKER_ADDR=127.0.0.1:10911     # 可选

# ==================== Mailgun 配置 ====================
MAILGUN_DOMAIN=your-domain.mailgun.org
MAILGUN_API_KEY=key-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
MAILGUN_FROM_EMAIL=noreply@yourdomain.com

# ==================== gRPC 配置（可选）====================
GRPC_ENDPOINTS=192.168.1.100:14032,192.168.1.101:14032
```

### 方式二：系统环境变量

```bash
# 临时设置（当前 shell 会话）
export MAILGUN_DOMAIN="your-domain.mailgun.org"
export MAILGUN_API_KEY="key-xxxx"
export MAILGUN_FROM_EMAIL="noreply@yourdomain.com"

# 永久设置（添加到 ~/.bashrc 或 ~/.zshrc）
echo 'export MAILGUN_DOMAIN="your-domain.mailgun.org"' >> ~/.bashrc
source ~/.bashrc
```

### 方式三：Docker 环境变量

```bash
docker run -d \
  -e MAILGUN_DOMAIN="your-domain.mailgun.org" \
  -e MAILGUN_API_KEY="key-xxxx" \
  -e MAILGUN_FROM_EMAIL="noreply@yourdomain.com" \
  ttpos-message:latest
```

## 📝 必需环境变量

### Mailgun 配置（必需）

| 变量名 | 说明 | 示例 | 获取方式 |
|--------|------|------|----------|
| `MAILGUN_DOMAIN` | Mailgun 域名 | `sandbox12345.mailgun.org` | [Mailgun 控制台](https://app.mailgun.com/) → Sending → Domains |
| `MAILGUN_API_KEY` | Mailgun API Key | `key-1234567890abcdef` | Mailgun 控制台 → Settings → API Keys → Private API key |
| `MAILGUN_FROM_EMAIL` | 发件人邮箱 | `noreply@yourdomain.com` | 必须是验证域名下的邮箱 |

### 数据库配置（必需）

| 变量名 | 说明 | 示例 |
|--------|------|------|
| `DB_USERNAME` | 数据库用户名 | `root` |
| `DB_PASSWORD` | 数据库密码 | `your_password` |
| `DB_HOST` | 数据库主机 | `127.0.0.1` |
| `DB_PORT` | 数据库端口 | `3306` |
| `DB_NAME` | 数据库名称 | `messages` |

### Nacos 配置（必需）

| 变量名 | 说明 | 示例 |
|--------|------|------|
| `NACOS_SERVER_IP` | Nacos 服务器 IP | `127.0.0.1` |
| `NACOS_SERVER_PORT` | Nacos 服务器端口 | `8848` |

### RocketMQ 配置（必需）

| 变量名 | 说明 | 示例 | 是否必需 |
|--------|------|------|----------|
| `ROCKETMQ_ENDPOINT` | RocketMQ 代理接入点 | `127.0.0.1:8081` | 是 |
| `ROCKETMQ_NAME_SRV_ADDR` | NameServer 地址 | `127.0.0.1:9876` | 是 |
| `ROCKETMQ_ACCESS_KEY` | 访问密钥 | `your_key` | 否（启用 ACL 时必需） |
| `ROCKETMQ_SECRET_KEY` | 密钥 | `your_secret` | 否（启用 ACL 时必需） |

## 🎯 Mailgun 配置详细步骤

### 第一步：注册 Mailgun 账号

1. 访问 [Mailgun 官网](https://www.mailgun.com/)
2. 点击 "Sign Up" 注册账号
3. 验证邮箱地址

### 第二步：获取 API Key

1. 登录 [Mailgun 控制台](https://app.mailgun.com/)
2. 点击左侧菜单 "Settings"
3. 选择 "API Keys"
4. 找到 "Private API key"
5. 点击眼睛图标显示完整的 Key
6. 复制 Key（以 `key-` 开头）
7. 设置到 `MAILGUN_API_KEY` 环境变量

### 第三步：配置域名

#### 使用 Sandbox 域名（测试用）

1. 在控制台左侧选择 "Sending" → "Domains"
2. 找到 Sandbox 域名（如 `sandbox12345.mailgun.org`）
3. 复制域名
4. 设置到 `MAILGUN_DOMAIN` 环境变量
5. 发件人邮箱设置为 `noreply@sandbox12345.mailgun.org`

**注意**：Sandbox 域名只能发送给预授权的邮箱地址。

#### 使用自定义域名（生产用）

1. 在 "Sending" → "Domains" 点击 "Add New Domain"
2. 输入你的域名（如 `mail.yourdomain.com`）
3. 按照提示添加 DNS 记录：
   - TXT 记录（SPF 验证）
   - TXT 记录（DKIM 验证）
   - CNAME 记录（追踪）
   - MX 记录（接收邮件，可选）
4. 等待 DNS 传播（通常 24-48 小时）
5. 验证域名状态为 "Verified"
6. 设置到 `MAILGUN_DOMAIN` 环境变量
7. 发件人邮箱设置为域名下的任意邮箱（如 `noreply@yourdomain.com`）

### 第四步：测试配置

```bash
# 1. 验证环境变量已设置
echo $MAILGUN_DOMAIN
echo $MAILGUN_API_KEY
echo $MAILGUN_FROM_EMAIL

# 2. 启动服务
make run

# 3. 查看配置是否正确加载
curl http://localhost:14031/debug/config

# 4. 发送测试邮件（通过 gRPC 调用）
```

## ⚙️ 配置优先级

配置读取优先级从高到低：

1. **环境变量** - 最高优先级
2. **配置文件** - `manifest/config/config.yaml`
3. **默认值** - 代码中定义的默认值

**示例**：

配置文件中：
```yaml
mailgun:
  domain: "$MAILGUN_DOMAIN"
  apiKey: "$MAILGUN_API_KEY"
```

环境变量：
```bash
MAILGUN_DOMAIN=my-domain.mailgun.org
```

最终读取的值：`my-domain.mailgun.org`（来自环境变量）

## 🔍 验证配置

### 方法一：使用调试接口

```bash
curl http://localhost:14031/debug/config
```

响应示例：
```json
{
  "mailgun": {
    "domain": "my-domain.mailgun.org",
    "fromEmail": "noreply@mydomain.com",
    "fromName": "TTPOS System",
    "timeout": "30"
  },
  "queue": {
    "enabled": true
  }
}
```

### 方法二：查看日志

启动服务时会输出配置信息：

```
2025-10-22 10:00:00 [INFO] 开始初始化 TTPOS 消息中心服务...
2025-10-22 10:00:00 [INFO] Mailgun 服务初始化成功 domain=my-domain.mailgun.org
2025-10-22 10:00:00 [INFO] 队列服务初始化成功
```

## ❌ 常见错误

### 错误 1：环境变量未生效

**症状**：配置了环境变量但服务报错"未配置"

**原因**：
- .env 文件位置不正确
- 环境变量格式错误（有多余空格）
- 未重启服务

**解决**：
```bash
# 检查 .env 文件位置（应该在 ttpos-bmp 根目录）
ls -la /path/to/ttpos-bmp/.env

# 检查环境变量格式（等号两边不要有空格）
# ✅ 正确：MAILGUN_API_KEY=key-xxxx
# ❌ 错误：MAILGUN_API_KEY = key-xxxx

# 重启服务
make restart
```

### 错误 2：Mailgun API Key 无效

**症状**：发送邮件时报 403 或 Forbidden 错误

**原因**：
- API Key 复制错误
- 使用了公钥而不是私钥
- API Key 已过期或被删除

**解决**：
1. 重新从 Mailgun 控制台复制 Private API Key
2. 确保 Key 以 `key-` 开头
3. 检查 Key 是否包含完整内容（通常很长）

### 错误 3：域名未验证

**症状**：发送邮件时报 "Domain not found" 错误

**原因**：
- 域名拼写错误
- 域名未在 Mailgun 中验证
- DNS 记录未正确配置

**解决**：
1. 检查域名拼写是否正确
2. 登录 Mailgun 控制台查看域名验证状态
3. 确认所有 DNS 记录已正确添加

## 🚀 生产环境最佳实践

### 1. 使用环境变量

**推荐**：
```bash
# 在部署脚本或 CI/CD 中设置
export MAILGUN_API_KEY="${SECRET_MAILGUN_API_KEY}"
```

**不推荐**：
```yaml
# config.yaml（不要在配置文件中写死敏感信息）
mailgun:
  apiKey: "key-1234567890"  # ❌ 不要这样做
```

### 2. 使用密钥管理服务

在生产环境，建议使用：
- Kubernetes Secrets
- HashiCorp Vault
- AWS Secrets Manager
- Azure Key Vault

### 3. 限制访问权限

- .env 文件权限设置为 600
- 不要将 .env 文件提交到 Git
- 定期轮换 API Key

```bash
# 设置 .env 文件权限
chmod 600 .env

# 确保 .env 在 .gitignore 中
echo ".env" >> .gitignore
```

## 📚 相关文档

- [Mailgun API 文档](https://documentation.mailgun.com/en/latest/api_reference.html)
- [GoFrame 配置管理](https://goframe.org/pages/viewpage.action?pageId=1114260)
- [项目 README](../README.MD)

---

**最后更新**: 2025-10-22  
**维护者**: TTPOS Team

