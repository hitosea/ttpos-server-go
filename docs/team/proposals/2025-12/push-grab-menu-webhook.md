# Push Grab Menu Webhook 实现提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2025-12-09   |
| **目标版本** | v2.11.0 |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | `task-takeout-grab-oauth-partner-webhook-simple` |

---

## 🎯 背景和动机

### 问题描述

在 GrabFood Self-Serve Activation 流程中，当商户选择"导出当前 Grab 门店菜单到 POS"时，GrabFood 会调用 Partner 的 `Push Grab Menu Webhook` 端点，将现有菜单数据推送给 POS 系统。

当前 `ttpos-takeout` 项目中已定义了 `PushGrabMenuWebhookReq` 和 `PushGrabMenuWebhookRes` 的 API 结构，但缺少完整的请求体定义和业务逻辑实现。

**参考文档**: [GrabFood Push Grab Menu Webhook](https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/push-grab-menu-webhook)

### 业务价值

- 支持 GrabFood Self-Serve Activation 流程的完整实现
- 允许商户在激活集成时导出现有 Grab 菜单到 POS 系统
- 减少商户手动录入菜单的工作量
- 支持 Grab 菜单结构的标准化 JSON 格式

### 目标用户

- [x] 商户管理员（通过 GrabFood 激活集成）
- [x] 系统集成（POS 系统接收菜单数据）

---

## 💡 解决方案概述

### 方案描述

实现 `Push Grab Menu Webhook` 端点，接收 GrabFood 推送的菜单数据。需要基于 Grab 官方 SDK (`github.com/grab/grabfood-api-sdk-go`) 的类型定义，确保与 GrabFood API 规范完全一致。

**关键点**：
1. **使用 SDK 类型**: 直接复用 `grabfood.GetMenuNewResponse` 的结构，该结构与 Push Grab Menu Webhook 请求体完全一致
2. **字段对齐**: 包含 `merchantID`, `partnerMerchantID`, `currency`, `sellingTimes`, `categories`
3. **业务处理**: 接收后存储菜单数据，供后续菜单同步使用

### SDK 类型参考

```go
// Grab SDK: GetMenuNewResponse (可复用于 Push Grab Menu Webhook)
type GetMenuNewResponse struct {
    MerchantID        *string        `json:"merchantID,omitempty"`
    PartnerMerchantID *string        `json:"partnerMerchantID,omitempty"`
    Currency          Currency       `json:"currency"`
    SellingTimes      []SellingTime  `json:"sellingTimes"`
    Categories        []MenuCategory `json:"categories"`
}

// Currency 货币信息
type Currency struct {
    Code     string `json:"code"`     // 货币代码: SGD, MYR, THB
    Symbol   string `json:"symbol"`   // 货币符号: S$, RM, ฿
    Exponent int32  `json:"exponent"` // 指数: 2 (除VN为0)
}

// SellingTime 销售时间
type SellingTime struct {
    StartTime    *string       `json:"startTime,omitempty"`    // UTC 格式
    EndTime      *string       `json:"endTime,omitempty"`      // UTC 格式
    Id           *string       `json:"id,omitempty"`           // 唯一 ID
    Name         *string       `json:"name,omitempty"`         // 名称
    ServiceHours *ServiceHours `json:"serviceHours,omitempty"` // 服务时间
}

// MenuCategory 菜单分类
type MenuCategory struct {
    Id              string             `json:"id"`
    Name            string             `json:"name"`
    NameTranslation *map[string]string `json:"nameTranslation,omitempty"`
    AvailableStatus string             `json:"availableStatus"`
    SellingTimeID   string             `json:"sellingTimeID"`
    Sequence        *int32             `json:"sequence,omitempty"`
    Items           []MenuItem         `json:"items"`
}
```

### 核心功能点

1. **API 定义**: 补充 `PushGrabMenuWebhookReq` 的请求体字段，与 Grab SDK 类型对齐
2. **Controller 实现**: 实现 webhook 处理逻辑，解析请求体
3. **业务逻辑**: 存储菜单数据（可选：触发菜单同步流程）
4. **响应处理**: 成功返回 HTTP 204 No Content

### 影响范围

**涉及终端**：
- [x] 外卖集成服务 (ttpos-takeout)

**涉及模块**：
- [x] API 接口定义
- [x] Controller 实现
- [x] 业务逻辑
- [x] 第三方集成 (GrabFood)

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：主要是类型定义和基础 webhook 处理

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1（待技术评审确认）

### 风险识别

**潜在风险**：
1. 菜单数据结构复杂，需确保与 SDK 类型完全一致
2. 需要处理 Authorization Header 的 Partner Token 验证

**缓解措施**：
1. 直接复用 Grab SDK 类型定义，避免手动定义出错
2. 参考现有 Partner OAuth Token Webhook 的鉴权实现

---

## 🔗 相关资源

### 参考需求

- GrabFood API 文档: https://developer.grab.com/docs/grabfood/api/v1-1-3/#tag/push-grab-menu-webhook
- Grab SDK: `github.com/grab/grabfood-api-sdk-go@v1.0.2`

### 相关文档

- 现有实现: `ttpos-bmp/app/ttpos-takeout/api/grab/v1/push_grab_menu_webhook.go`
- SDK Wrapper: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/sdk_wrapper.go`
- 菜单 DTO: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu.go`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | -      |           |
| 技术负责人   | -      |           |
| 开发代表     | rikugun |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

---

## 📝 附录

### API 规范 (来自 GrabFood 文档)

**Endpoint**: `POST /partner/menu` (或自定义路径)

**Headers**:
- `Authorization`: Bearer <ACCESS_TOKEN_HERE>
- `Content-Type`: application/json

**Request Body Example**:
```json
{
  "merchantID": "1-CYNGRUNGSBCCC",
  "partnerMerchantID": "Partner-ABECU",
  "currency": {
    "code": "SGD",
    "symbol": "S$",
    "exponent": 2
  },
  "sellingTimes": [
    {
      "startTime": "2022-03-01 10:00:00",
      "endTime": "2025-01-21 22:00:00",
      "id": "partner-sellingTimeID-1",
      "name": "Lunch deal",
      "serviceHours": {
        "mon": {
          "openPeriodType": "OpenPeriod",
          "periods": [{ "startTime": "11:30", "endTime": "21:30" }]
        }
      }
    }
  ],
  "categories": [
    {
      "id": "category_id",
      "name": "Value set",
      "availableStatus": "AVAILABLE",
      "sellingTimeID": "partner-sellingTimeID-1",
      "items": [...]
    }
  ]
}
```

**Response**: `204 No Content`

### 实现要点

1. **类型定义选择**:
   - 方案 A: 直接使用 `grabfood.GetMenuNewResponse` 作为请求体类型
   - 方案 B: 创建本地类型 `PushGrabMenuRequest`，字段与 SDK 对齐

2. **推荐方案**: 方案 A，直接复用 SDK 类型，确保与 GrabFood API 100% 兼容

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**维护者**: rikugun  
**相关规范**: `.cursor/rules/go-bmp.mdc`
