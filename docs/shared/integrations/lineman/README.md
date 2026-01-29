# LINE MAN 集成文档

> TTPOS 系统与 LINE MAN 外卖平台的集成文档索引

---

## 文档列表

| 文档 | 说明 | 更新日期 |
|------|------|----------|
| [OAuth API](./lineman-oauth-api.md) | OAuth 2.0 认证接口，获取访问令牌 | 2026-01-14 |
| [Menu API](./lineman-menu-api.md) | 菜单同步相关接口，包括同步通知和触发同步 | 2026-01-14 |
| [Webhook API](./lineman-webhook-api.md) | 订单 Webhook 接口，包括创建、更新、状态变更 | 2026-01-14 |

---

## 接口概览

### 认证接口

| 端点 | 方法 | 说明 |
|------|------|------|
| `/oauth2/token` | POST | 获取 OAuth 访问令牌 |

### 菜单接口

| 端点 | 方法 | 说明 |
|------|------|------|
| `/partners/{partnerId}/stores/{storeId}/menus/notification` | POST | 接收菜单同步通知 |
| `/partners/{partnerId}/stores/{storeId}/menus/trigger-sync` | POST | 触发菜单同步 |

### 订单接口

| 端点 | 方法 | 说明 |
|------|------|------|
| `/v1/partners/{partnerId}/stores/{storeId}/orders` | POST | 接收新订单创建 |
| `/v1/partners/{partnerId}/stores/{storeId}/orders` | PUT | 接收订单内容更新 |
| `/v1/partners/{partnerId}/stores/{storeId}/order/status` | POST | 接收订单状态更新 |

---

## 技术架构

```
LINE MAN Platform
       │
       │ Webhook (HTTPS)
       ▼
┌─────────────────┐
│   TTPOS BMP     │ ◄── 接收 Webhook，处理业务逻辑
│ (ttpos-takeout) │
└────────┬────────┘
         │ RocketMQ
         ▼
┌─────────────────┐
│   TTPOS Main    │ ◄── 消费订单事件，更新本地状态
└─────────────────┘
```

---

## 代码位置

- **API 定义**: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/`
- **Controller**: `ttpos-bmp/app/ttpos-takeout/internal/controller/lineman/`
- **Logic**: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/`
- **常量**: `ttpos-bmp/app/ttpos-takeout/internal/consts/`

---

**维护者**: TTPOS 后端开发组
**最后更新**: 2026-01-14
