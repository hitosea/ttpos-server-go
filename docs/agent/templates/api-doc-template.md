# API 文档模板

> 🤖 Agent 视角模板：用于 `docs/shared/api/{module}_api.md`，补充接口列表、参数、示例、错误码，并附 Graphiti/活动日志提醒。

---

## 元信息

| 字段      | 内容                       |
| --------- | -------------------------- | ------ | ------------ |
| 模块      | `{order                    | member | payment...}` |
| 版本      | `v2.1.0`                   |
| 更新时间  | `{YYYY-MM-DD}`             |
| 负责人    | `@`                        |
| 关联 Spec | `story-{module}-{feature}` |

---

## 1. 概述

- **用途**：说明模块职责、终端（POS/Shop/Admin/微服务）、依赖服务。
- **接口数量**：`P0: n / P1: n / P2: n`
- **认证方式**：JWT / API Key / OAuth / 内部服务间 Token

---

## 2. 快速索引

| 级别 | 接口            | 方法 | 路径                          | 描述     |
| ---- | --------------- | ---- | ----------------------------- | -------- |
| P0   | `Quick Payment` | POST | `/api/v1/order/quick_payment` | 一键支付 |
| P1   | ...             | ...  | ...                           | ...      |

> 按优先级或业务分组（下单、查询、配置等）。

---

## 3. 接口详情

### 3.1 {接口名称}

- **Method & URL**
  ```http
  POST /api/v1/order/quick_payment
  ```
- **Headers**
  | 名称 | 必填 | 示例 | 说明 |
  | --- | --- | --- | --- |

- **Request**

  ```json
  {
    "order_uuid": 123456789,
    "payment_method": 1
  }
  ```

  | 字段 | 类型 | 必填 | 说明 | 校验 |
  | ---- | ---- | ---- | ---- | ---- |

- **Response**

  ```json
  {
    "code": 1,
    "message": "success",
    "data": {
      "order_uuid": 123456789,
      "payment_status": 1
    },
    "meta": {}
  }
  ```

- **错误码**
  | code | message | 场景 |
  | --- | --- | --- |

- **示例**

  ```bash
  curl -X POST https://{host}/api/v1/order/quick_payment \
    -H "Authorization: Bearer <token>" \
    -d '{"order_uuid":123456789}'
  ```

- **备注**
  - 分页规则、速率限制、并发约束等。

> 对同一模块的其他接口重复以上章节。

---

## 4. Webhook / 回调（如有）

- URL、签名算法、重试机制、示例 Payload。

---

## 5. 依赖与配置

- 外部服务：Nacos、Redis、MQ。
- 环境变量：`ORDER_API_BASE_URL`, `ORDER_API_TIMEOUT`。
- 本地 Mock/Fixture 路径。

---

## 6. 测试与验证

- Postman / Thunder Collection 链接。
- 自动化测试位置：`main/app/api/..._test.go`。
- 覆盖率要求。

---

## 7. 变更记录

| 日期       | 版本   | 说明     | 负责人 |
| ---------- | ------ | -------- | ------ |
| 2025-11-17 | v1.0.0 | 初始创建 | @      |

---

## 8. Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 模板引用：`docs/agent/templates/graphiti-episode.md`

---

**最后更新**：2025-11-17
