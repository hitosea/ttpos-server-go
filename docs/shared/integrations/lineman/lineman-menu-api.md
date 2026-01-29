# LINE MAN Menu API 文档

> TTPOS 系统与 LINE MAN 平台的菜单同步接口

---

## 概述

本文档描述 TTPOS 系统接收 LINE MAN 平台菜单相关 Webhook 回调的 API 接口规范。LINE MAN 通过 Webhook 通知 TTPOS 菜单状态变更或请求菜单数据同步。

### 功能特性

- **菜单同步通知**：接收 LINE MAN 菜单启用/禁用/更新通知
- **触发菜单同步**：响应 LINE MAN 的菜单数据拉取请求
- **多门店支持**：按门店粒度管理菜单同步

---

## 认证机制

所有菜单相关 Webhook 请求需要包含有效的 Bearer Token：

```http
Authorization: Bearer {access_token}
```

Token 获取方式参见：[OAuth API 文档](./lineman-oauth-api.md)

---

## 接口定义

### 1. 菜单同步通知

接收 LINE MAN 发送的菜单状态变更通知，包括启用、禁用、更新等操作。

#### 基本信息

| 项目 | 值 |
|------|-----|
| **端点** | `POST /partners/{partnerId}/stores/{storeId}/menus/notification` |
| **方法** | POST |
| **认证** | Bearer Token |
| **Content-Type** | application/json |

#### 路径参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `partnerId` | string | Y | 合作伙伴唯一 ID |
| `storeId` | string | Y | 门店唯一 ID（LINE MAN 门店 ID） |

#### 请求参数

| 参数 | 类型 | 必填 | 说明 | 示例 |
|------|------|------|------|------|
| `notificationType` | string | Y | 通知类型 | `ENABLE` |
| `details` | string | N | 通知详情描述 | `Menu enabled by admin` |

**notificationType 可选值**：

| 值 | 说明 | TTPOS 处理逻辑 |
|-----|------|---------------|
| `ENABLE` | 菜单启用 | 激活门店菜单同步，开始接收订单 |
| `DISABLE` | 菜单禁用 | 停止门店菜单同步，暂停接收订单 |
| `UPDATE` | 菜单更新 | 标记菜单需要重新同步 |

#### 请求示例

```http
POST /partners/partner-123/stores/store-456/menus/notification HTTP/1.1
Host: api.ttpos.example.com
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "notificationType": "ENABLE",
  "details": "Menu enabled by merchant"
}
```

#### 响应格式

##### 成功响应（200 OK）

```json
{
  "status": "ok",
  "code": "200",
  "message": "Notification received successfully"
}
```

##### 失败响应

```json
{
  "status": "fail",
  "code": "500",
  "message": "Failed to process notification"
}
```

#### 业务逻辑

```mermaid
flowchart TD
    A[接收通知] --> B{通知类型?}
    B -->|ENABLE| C[激活门店菜单]
    B -->|DISABLE| D[禁用门店菜单]
    B -->|UPDATE| E[标记菜单待同步]
    C --> F[更新门店状态]
    D --> F
    E --> F
    F --> G[返回成功响应]
```

---

### 2. 触发菜单同步

LINE MAN 请求 TTPOS 主动推送最新菜单数据。

#### 基本信息

| 项目 | 值 |
|------|-----|
| **端点** | `POST /partners/{partnerId}/stores/{storeId}/menus/trigger-sync` |
| **方法** | POST |
| **认证** | Bearer Token |
| **Content-Type** | application/json |

#### 路径参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `partnerId` | string | Y | 合作伙伴唯一 ID |
| `storeId` | string | Y | 门店唯一 ID（LINE MAN 门店 ID） |

#### 请求参数

无请求体参数，仅需路径参数。

#### 请求示例

```http
POST /partners/partner-123/stores/store-456/menus/trigger-sync HTTP/1.1
Host: api.ttpos.example.com
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### 响应格式

##### 成功响应（200 OK）

```json
{
  "status": "ok",
  "code": "200",
  "message": "Menu sync triggered successfully"
}
```

##### 失败响应

```json
{
  "status": "fail",
  "code": "500",
  "message": "Store not found or menu sync failed"
}
```

#### 业务逻辑

```mermaid
sequenceDiagram
    participant LM as LINE MAN
    participant BMP as TTPOS BMP
    participant Main as TTPOS Main
    participant LMAPI as LINE MAN Menu API

    LM->>BMP: POST /menus/trigger-sync
    BMP->>BMP: 查找门店配置
    BMP->>Main: 请求菜单数据
    Main-->>BMP: 返回菜单数据
    BMP->>LMAPI: 推送菜单到 LINE MAN
    LMAPI-->>BMP: 确认接收
    BMP-->>LM: {"status": "ok"}
```

---

## 菜单同步流程

### 完整同步流程

```mermaid
sequenceDiagram
    participant Merchant as 商户（TTPOS）
    participant Main as TTPOS Main
    participant BMP as TTPOS BMP
    participant LM as LINE MAN Platform

    Note over Merchant,LM: 1. 菜单启用
    LM->>BMP: POST /menus/notification (ENABLE)
    BMP->>BMP: 激活门店 LINE MAN 配置
    BMP-->>LM: {"status": "ok"}

    Note over Merchant,LM: 2. 触发菜单同步
    LM->>BMP: POST /menus/trigger-sync
    BMP->>Main: 获取门店菜单数据
    Main-->>BMP: 菜单数据（商品、分类、价格等）
    BMP->>LM: 推送菜单数据
    LM-->>BMP: 确认接收
    BMP-->>LM: {"status": "ok"}

    Note over Merchant,LM: 3. 商户更新菜单
    Merchant->>Main: 修改菜单（添加/编辑商品）
    Main->>BMP: 通知菜单变更
    BMP->>LM: 推送更新的菜单数据

    Note over Merchant,LM: 4. LINE MAN 通知更新
    LM->>BMP: POST /menus/notification (UPDATE)
    BMP->>BMP: 标记菜单待同步
    BMP-->>LM: {"status": "ok"}
```

---

## 错误处理

### 错误码说明

| HTTP 状态码 | status | code | 说明 |
|------------|--------|------|------|
| 200 | ok | 200 | 处理成功 |
| 200 | fail | 400 | 参数错误（notificationType 无效） |
| 200 | fail | 404 | 门店不存在或未配置 LINE MAN |
| 200 | fail | 500 | 服务器内部错误 |
| 401 | - | - | Token 无效或过期 |

### 常见问题

#### 问题 1: 门店未找到

**症状**：返回 `{"status": "fail", "code": "404"}`

**原因**：
- `storeId` 在 TTPOS 中未配置
- 门店未绑定 LINE MAN 平台

**解决**：
1. 确认门店已在 TTPOS 中创建
2. 检查门店是否已配置 LINE MAN 集成
3. 验证 `storeId` 是否正确

#### 问题 2: 菜单同步失败

**症状**：trigger-sync 返回失败

**原因**：
- TTPOS Main 模块无法访问
- 菜单数据格式不兼容
- LINE MAN API 调用失败

**解决**：
1. 检查 Main 模块服务状态
2. 查看 BMP 日志获取详细错误
3. 验证 LINE MAN API 凭证是否有效

---

## 代码参考

- **API 定义**: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/menu.go`
- **Controller**: `ttpos-bmp/app/ttpos-takeout/internal/controller/lineman/`
- **Logic**: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/`

---

## 版本历史

| 版本 | 日期 | 作者 | 变更内容 |
|------|------|------|----------|
| v1.0.0 | 2026-01-14 | Claude | 初始版本 |

---

**维护者**: TTPOS 后端开发组
**最后更新**: 2026-01-14
