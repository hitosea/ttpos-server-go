# POS 收银机 - Grab 外卖订单集成 需求文档

> 本文档定义 Grab 外卖订单集成到 POS 收银系统的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.12.0-pos-grab-order-integration.md](../../../../team/proposals/2025-12/v2.12.0-pos-grab-order-integration.md) |
| **创建日期**      | 2025-12-22                                                                                                   |
| **负责人**        | weifashi                                                                                                     |
| **目标 Sprint**   | v2.12.0                                                                                                      |
| **关联任务**      | DooTask #37459                                                                                               |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                  |
| **职责范围**      | 后端对接 Grab RPC + 提供 HTTP API 给前端（前端实现由前端团队负责）                                           |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | ✅ 已通过                |
| **审核人**   | weifashi                 |
| **审核日期** | 2025-12-22               |
| **审核意见** | 后端职责明确，RPC接口待与BMP团队确认 |

---

## 📋 概述

将 Grab 外卖平台集成到 TTPOS 收银系统的**后端服务**中，实现：
1. **对接 Grab RPC 服务**：通过 RPC 调用获取订单数据、同步订单状态（RPC 接口待 ttpos-bmp 团队定义）
2. **提供 HTTP API**：为前端提供订单列表、订单详情、接单、拒单等接口
3. **业务逻辑处理**：商品关联检查、库存检查、自动接单判断、KDS 通知等

**职责边界**：
- ✅ 后端负责：Grab RPC 对接、数据存储、业务逻辑、HTTP API 提供
- ❌ 后端不负责：前端界面实现、用户交互、权限 UI 控制（由前端团队实现）

## 🎯 产品对齐

支持 TTPOS 在东南亚市场的拓展战略，帮助商户通过统一平台管理多渠道订单，提升订单处理速度和准确性，改善顾客体验，最终提高商户的营业效率和满意度。

## 📝 用户故事

**作为** 收银员  
**我想** 在 POS 收银机上统一接收和处理来自 Grab 外卖平台的订单  
**以便于** 提高订单处理效率，减少平台切换，避免缺货和订单异常

**作为** 商户管理员  
**我想** 配置自动接单规则（金额上限、库存检查）和语音提醒  
**以便于** 灵活控制接单策略，降低运营风险

**作为** 厨房人员  
**我想** 在 KDS 厨显上看到 Grab 外卖订单的制作任务  
**以便于** 及时制作菜品，确保按时出餐

---

## 功能需求

### Requirement 1: Grab RPC 对接 - 订单同步

**用户故事**: 作为后端系统，我想通过 RPC 调用从 Grab 服务获取新订单数据，以便于存储和处理。

#### Grab 订单数据结构示例

```json
{
  "currency": {
    "code": "THB",
    "exponent": 2,
    "symbol": "฿"
  },
  "cutlery": false,
  "featureFlags": {
    "isMexEditOrder": false,
    "orderAcceptedType": "AUTO",
    "orderType": "DeliveredByGrab"
  },
  "items": [
    {
      "grabItemID": "THITE20251219100406295285",
      "id": "TTPOS-ITEM-3701988558241793",
      "modifiers": [
        {
          "id": "TTPOS-SAUCE-3671416295522520",
          "price": 0,
          "quantity": 1,
          "tax": 0
        },
        {
          "id": "TTPOS-ATTR-4290647166976000",
          "price": 0,
          "quantity": 1,
          "tax": 0
        }
      ],
      "price": 43389,
      "quantity": 1,
      "specifications": "",
      "tax": 0
    }
  ],
  "merchantID": "GFSBPOS-822-571",
  "orderID": "123456789-C7WBHBVGE76GAA",
  "orderReadyEstimation": {
    "allowChange": true,
    "estimatedOrderReadyTime": "2025-12-19T10:14:06.277879574Z",
    "maxOrderReadyTime": "2025-12-19T11:04:06.277879651Z"
  },
  "orderState": "NEW",
  "orderTime": "2025-12-19T10:04:06Z",
  "partnerMerchantID": "8609817471094784",
  "paymentType": "CASH",
  "price": {
    "deliveryFee": 1000,
    "eaterPayment": 44389,
    "grabFundPromo": 0,
    "merchantFundPromo": 0,
    "subtotal": 43389,
    "tax": 0
  },
  "shortOrderNumber": "GF-5447"
}
```

#### 验收标准

1. **WHEN** 调用 Grab RPC 接口获取新订单 **THEN** 系统 **SHALL** 成功解析并存储订单数据
2. **WHEN** RPC 调用失败 **THEN** 系统 **SHALL** 将订单加入重试队列，最多重试 3 次，每次间隔 5 秒
3. **WHEN** 检测到重复订单（相同 orderID） **THEN** 系统 **SHALL** 识别并跳过，不创建重复订单
4. **WHEN** 订单同步成功 **THEN** 系统 **SHALL** 将订单状态设置为"待接单"（状态码：1）
5. **WHEN** 订单状态在 Grab 端更新 **THEN** 系统 **SHALL** 通过 RPC 同步到本地数据库
6. **WHEN** 订单状态在本地更新 **THEN** 系统 **SHALL** 通过 RPC 调用同步到 Grab 端

#### 具体要求

- [x] 1.1 定义 Grab RPC 接口（与 ttpos-bmp 团队协作）
  - `GetNewOrders()` - 获取新订单列表
  - `GetOrderDetail(orderID)` - 获取订单详情
  - `UpdateOrderStatus(orderID, status)` - 更新订单状态
  - `AcceptOrder(orderID)` - 接单
  - `RejectOrder(orderID, reason)` - 拒单
- [x] 1.2 实现 RPC 客户端调用逻辑（main/app/modules/takeout/infrastructure/adapter/grab/）
- [x] 1.3 实现订单数据解析和转换（DTO 转换）
- [x] 1.4 实现订单去重逻辑（基于 orderID + merchantID）
- [x] 1.5 实现失败重试队列（使用 Redis 队列）
- [x] 1.6 记录 RPC 调用日志（成功/失败/重试）
- [x] 1.7 解析并存储 Grab 订单完整数据（JSON 格式）
- [x] 1.8 处理货币信息（currency.code, currency.symbol）
- [x] 1.9 处理价格信息（price 字段中的各项金额，单位为分）
- [x] 1.10 处理商品信息（items 数组，包含 modifiers）

#### RPC 接口待定事项

**需要与 ttpos-bmp 团队确认**：
- RPC 服务名称和方法签名
- 请求/响应数据结构（Protobuf 定义）
- 错误码定义和错误处理
- 超时配置和重试策略
- 服务发现和负载均衡

---

### Requirement 2: 商品关联检查

**用户故事**: 作为系统，我想在订单同步后自动检查 Grab 商品是否关联到 TTPOS 商品，以便于标记异常订单。

#### 验收标准

1. **WHEN** 订单同步成功后 **THEN** 系统 **SHALL** 自动执行商品关联检查
2. **WHEN** 订单中所有商品和规格均已关联 **THEN** 系统 **SHALL** 标记订单为"正常"
3. **WHEN** 订单中存在未关联的商品 **THEN** 系统 **SHALL** 标记订单为"异常（商品未关联）"
4. **WHEN** 订单中存在未关联的规格/修饰符 **THEN** 系统 **SHALL** 标记订单为"异常（规格未关联）"
5. **IF** 订单标记为异常 **THEN** 系统 **SHALL** 不执行自动接单，即使开启了自动接单功能
6. **IF** 订单标记为异常 **THEN** 系统 **SHALL** 在订单列表中显示异常标识（❌ 图标）
7. **WHEN** 收银员点击异常订单 **THEN** 系统 **SHALL** 显示友好提示，说明具体哪些商品或规格未关联

#### 具体要求

- [x] 2.1 实现商品关联映射表（ttpos_grab_item_mapping）
- [x] 2.2 实现规格关联映射表（ttpos_grab_modifier_mapping）
- [x] 2.3 实现商品关联检查逻辑（Service 层）
- [x] 2.4 实现异常标记机制（订单表增加 `is_abnormal` 字段）
- [x] 2.5 实现异常详情记录（JSON 格式存储未关联的商品/规格列表）
- [x] 2.6 前端显示异常订单标识和详情
- [x] 2.7 提供商品关联配置界面（Shop 管理端）

---

### Requirement 3: 库存检查

**用户故事**: 作为系统，我想在接单前检查库存，以便于避免接单后无法履约的情况。

#### 验收标准

1. **WHEN** 订单通过商品关联检查后 **THEN** 系统 **SHALL** 自动执行库存检查
2. **IF** 开启自动接单 **AND** 库存充足 **THEN** 系统 **SHALL** 进入"待接单"列表，等待自动接单
3. **IF** 开启自动接单 **AND** 库存不足 **THEN** 系统 **SHALL** 标记订单为"需人工处理（库存不足）"，不自动接单
4. **IF** 未开启自动接单 **AND** 库存不足 **THEN** 系统 **SHALL** 在接单时显示库存不足提示，但允许手动接单
5. **WHEN** 收银员手动接单库存不足的订单 **THEN** 系统 **SHALL** 显示确认对话框，询问是否继续接单
6. **WHEN** 库存检查完成 **THEN** 系统 **SHALL** 记录检查结果（库存充足/不足，具体商品）

#### 具体要求

- [x] 3.1 实现库存检查逻辑（调用现有库存服务）
- [x] 3.2 实现库存不足标记机制（订单表增加 `stock_status` 字段）
- [x] 3.3 实现自动接单库存判断逻辑
- [x] 3.4 前端显示库存不足提示
- [x] 3.5 前端实现手动接单确认对话框
- [x] 3.6 记录库存检查日志

---

### Requirement 4: HTTP API - 订单列表和详情

**用户故事**: 作为后端系统，我想提供订单列表和详情接口给前端，以便于前端展示订单信息。

#### 验收标准

1. **WHEN** 前端调用订单列表接口 **THEN** 系统 **SHALL** 返回分页的订单列表（支持按状态、时间筛选）
2. **WHEN** 前端调用订单详情接口 **THEN** 系统 **SHALL** 返回完整的订单信息（包括商品、价格、配送信息）
3. **WHEN** 订单包含货币信息 **THEN** 系统 **SHALL** 在响应中返回货币代码和符号（如 THB, ฿）
4. **WHEN** 订单包含价格信息 **THEN** 系统 **SHALL** 以分为单位返回（前端负责格式化显示）
5. **WHEN** 前端调用搜索接口 **THEN** 系统 **SHALL** 支持按订单号、联系人搜索
6. **WHEN** 查询订单列表 **THEN** 系统 **SHALL** 包含订单状态、异常标记、库存状态等字段

#### 具体要求

- [x] 4.1 实现订单列表接口（GET `/api/v1/takeout/grab/orders`）
  - 支持分页（page, page_size）
  - 支持状态筛选（status: 1=待接单, 2=已接单, 3=制作中, 4=已完成, 5=已拒单）
  - 支持时间范围筛选（start_time, end_time）
  - 支持搜索（search: 订单号、联系人）
- [x] 4.2 实现订单详情接口（GET `/api/v1/takeout/grab/orders/:id`）
- [x] 4.3 实现订单统计接口（GET `/api/v1/takeout/grab/orders/stats`）
  - 返回待接单、已接单、制作中、已完成的订单数量
- [x] 4.4 响应数据包含完整的 Grab 订单信息
- [x] 4.5 响应数据包含异常标记（is_abnormal）和异常详情（abnormal_detail）
- [x] 4.6 响应数据包含库存状态（stock_status: 1=充足, 2=不足）
- [x] 4.7 响应格式遵循 API 规范（{code, message, data{}}）
- [x] 4.8 所有接口需要身份验证（Token）
- [x] 4.9 记录 API 访问日志

---

### Requirement 5: HTTP API - 接单和拒单

**用户故事**: 作为后端系统，我想提供接单和拒单接口给前端，以便于前端调用实现订单处理。

#### 验收标准

1. **WHEN** 前端调用接单接口 **THEN** 系统 **SHALL** 更新订单状态为"已接单"（状态码：2）
2. **WHEN** 订单接单成功 **THEN** 系统 **SHALL** 通过 RPC 调用同步状态到 Grab 端
3. **WHEN** 订单接单成功 **THEN** 系统 **SHALL** 自动通知 KDS 厨显系统
4. **WHEN** Grab RPC 调用失败 **THEN** 系统 **SHALL** 返回错误信息，并将任务加入重试队列
5. **WHEN** 前端调用拒单接口 **THEN** 系统 **SHALL** 通过 RPC 调用 Grab 拒单接口
6. **WHEN** 拒单成功 **THEN** 系统 **SHALL** 更新订单状态为"已拒单"（状态码：5）
7. **WHEN** 拒单失败 **THEN** 系统 **SHALL** 返回错误信息，订单保持"待接单"状态
8. **WHEN** 接单/拒单操作完成 **THEN** 系统 **SHALL** 记录操作日志（操作人、操作时间、操作结果）

#### 具体要求

- [x] 5.1 实现接单接口（POST `/api/v1/takeout/grab/orders/:id/accept`）
  - 请求参数：无（从 Token 中获取操作人信息）
  - 响应：订单状态、接单时间
- [x] 5.2 实现拒单接口（POST `/api/v1/takeout/grab/orders/:id/reject`）
  - 请求参数：reason_code（拒单原因代码）
  - 响应：拒单结果
- [x] 5.3 实现拒单原因列表接口（GET `/api/v1/takeout/grab/reject_reasons`）
  - 从 Grab RPC 获取拒单原因列表（前端下拉选择使用）
- [x] 5.4 调用 Grab RPC 同步状态
- [x] 5.5 实现 Grab RPC 调用失败重试机制
- [x] 5.6 实现 KDS 通知逻辑（接单成功后）
- [x] 5.7 记录接单/拒单操作日志
- [x] 5.8 接口需要身份验证和权限校验（外卖权限）

---

### Requirement 6: HTTP API - 自动接单配置

**用户故事**: 作为后端系统，我想提供接单配置接口给前端（Shop 管理端），以便于商户管理员配置自动接单规则。

#### 验收标准

1. **WHEN** 前端调用获取配置接口 **THEN** 系统 **SHALL** 返回当前的自动接单配置
2. **WHEN** 前端调用保存配置接口 **THEN** 系统 **SHALL** 保存配置并立即生效（更新 Redis 缓存）
3. **WHEN** 订单同步完成并通过检查 **AND** 开启自动接单 **AND** 订单金额 ≤ 金额上限 **THEN** 系统 **SHALL** 自动接单
4. **WHEN** 订单金额 > 金额上限 **THEN** 系统 **SHALL** 不自动接单，需要手动处理
5. **WHEN** 配置保存成功 **THEN** 系统 **SHALL** 返回成功响应

#### 具体要求

- [x] 6.1 实现获取接单配置接口（GET `/api/v1/shop/takeout/grab/settings`）
  - 返回字段：auto_accept（自动接单开关）、max_amount（金额上限，单位：分）
- [x] 6.2 实现保存接单配置接口（POST `/api/v1/shop/takeout/grab/settings`）
  - 请求参数：auto_accept、max_amount
- [x] 6.3 实现自动接单判断逻辑（在订单同步后执行）
  - 检查配置开关
  - 检查订单金额
  - 检查商品关联
  - 检查库存
- [x] 6.4 实现配置缓存（Redis）
  - Key: `shop:{shop_id}:grab:settings`
  - TTL: 1 小时（配置更新时刷新）
- [x] 6.5 配置变更日志记录

---

### Requirement 7: KDS 厨显联动

**用户故事**: 作为厨房人员，我想在 KDS 厨显上看到 Grab 外卖订单的制作任务，以便于及时制作菜品。

#### 验收标准

1. **WHEN** 订单接单成功 **THEN** 系统 **SHALL** 自动通知 KDS 厨显系统
2. **WHEN** KDS 接收到通知 **THEN** 系统 **SHALL** 在厨显屏幕上显示订单卡片
3. **WHEN** 订单卡片显示 **THEN** 系统 **SHALL** 包含订单号、商品列表、制作要求、预计出餐时间
4. **WHEN** 订单卡片显示 **THEN** 系统 **SHALL** 标识为"Grab外卖"订单（显示 Grab 图标）
5. **WHEN** 厨师完成制作并点击"完成" **THEN** 系统 **SHALL** 更新订单状态为"已完成"
6. **WHEN** 订单状态更新为"已完成" **THEN** 系统 **SHALL** 同步到 Grab 平台

#### 具体要求

- [x] 7.1 后端实现 KDS 通知接口（调用现有 KDS 服务）
- [x] 7.2 KDS 前端实现 Grab 订单卡片展示
- [x] 7.3 KDS 前端显示 Grab 图标和外卖标识
- [x] 7.4 KDS 前端实现订单完成操作
- [x] 7.5 后端实现订单状态同步（KDS → POS → Grab）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/api.mdc` - API 设计规范
  - `.cursor/rules/database.mdc` - 数据库开发规范
  - `.cursor/rules/vue.mdc` - Vue 前端规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/api/v1/takeout/grab/orders`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 分页信息统一放在 meta 中
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] 金额字段使用 decimal(20,8)
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

**新增表**：
- `ttpos_grab_orders` - Grab 订单表
- `ttpos_grab_order_items` - Grab 订单商品表
- `ttpos_grab_item_mapping` - 商品关联映射表
- `ttpos_grab_modifier_mapping` - 规格关联映射表
- `ttpos_grab_sync_logs` - 订单同步日志表

### 性能要求

- [x] 订单同步响应时间 < 5 秒
- [x] 订单列表接口响应时间 < 200ms
- [x] 数据库查询优化（使用索引：order_id, merchant_id, order_state）
- [x] 缓存策略（Redis 缓存配置信息、订单列表）
- [x] 并发处理（使用 UUID 锁防止重复接单）

### 浏览器兼容性（管理后台）

**注意**：前端实现由前端团队负责，后端只需确保 API 兼容性。

- [x] HTTP API 响应格式标准化（JSON）
- [x] CORS 跨域配置（如需要）
- [x] WebSocket 支持（用于订单实时推送，可选）

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] Repository 层测试覆盖率 ≥ 80%
- [x] **Order 相关模块测试覆盖率 100%**（高风险）
- [x] 集成测试覆盖核心流程（订单同步 → 接单 → KDS 通知）
- [x] API 测试覆盖所有接口（使用 Postman/自动化测试）
- [x] RPC 调用 Mock 测试（不依赖真实 Grab RPC 服务）
- [x] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言（中文、英文、泰语、越南语等）
- [x] 所有文案使用多语言实现
- [x] 货币格式化支持多币种
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证（Token）
- [x] Webhook 接口需要签名验证（Grab 签名）
- [x] 敏感数据加密存储（顾客信息）
- [x] SQL 注入防护（使用参数化查询）
- [x] XSS 防护（前端输入校验）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] Grab API 调用失败时优雅降级（显示错误提示，不影响其他订单）
- [x] 事务管理（接单操作保证数据一致性）
- [x] 错误日志记录（使用 Logger 记录所有异常）
- [x] 故障恢复机制（失败重试队列）

---

## 验收标准

### 功能验收

1. **订单同步**: Webhook 推送和轮询兜底机制工作正常，订单在 60 秒内同步
2. **商品关联检查**: 未关联商品的订单正确标记为异常，不自动接单
3. **库存检查**: 库存不足的订单正确标记，自动接单跳过
4. **接单管理**: 统一接单界面正常工作，权限控制生效
5. **接单/拒单**: 接单和拒单操作正常，状态正确同步到 Grab
6. **自动接单**: 自动接单规则（金额上限、商品、库存）正确执行
7. **KDS 联动**: 接单成功后 KDS 正确显示订单，完成后状态同步

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%, Repository ≥ 80%, Order 模块 100%）
2. **API 测试**: 所有接口测试通过（Postman/自动化测试）
3. **集成测试**: 端到端流程测试通过（订单同步 → 接单 → KDS → 完成）
4. **手动测试**: 浏览器兼容性测试通过，UI/UX 符合预期
5. **压力测试**: 高峰期订单同步（100 订单/分钟）系统正常运行

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: Grab API 集成文档完整
3. **数据库文档**: 迁移脚本和表结构文档完整
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- RPC 客户端调用需要设置超时（10s）和重试策略

#### RPC 对接约束

- RPC 接口定义需要与 ttpos-bmp 团队协作完成
- Protobuf 文件存放在 `ttpos-bmp/api/` 目录
- RPC 调用必须有超时控制和错误处理
- RPC 服务发现使用 Nacos
- RPC 调用失败需要记录详细日志

#### Vue 模块（前端团队负责）

**后端不涉及前端实现**，前端团队需遵循：
- Vue 3 + TypeScript + Vite
- Element Plus 组件库
- 参考: `.cursor/rules/vue.mdc`

### 业务约束

- Grab 订单数据格式必须严格按照 Grab RPC 接口规范
- 价格单位为分（price: 43389 表示 433.89 元）
- 不做汇率换算，直接使用 Grab 提供的价格和货币信息
- 订单时间使用 UTC 时间戳，存储时转换为 Unix 时间戳（int 类型）
- 拒单原因列表必须从 Grab RPC 获取，不能自定义
- **前端实现由前端团队负责**，后端只提供 HTTP API

### 协作约束

- **Grab RPC 接口**：需要与 ttpos-bmp 团队协作定义 Protobuf 和实现
- **前端接口**：需要与前端团队确认 API 响应格式和字段
- **KDS 通知**：需要调用现有 KDS 服务接口

### 资源约束

- 开发时间: 10-15 天
- Story Point: 13-21（待技术评审拆分为 ≤ 5 SP 的子任务）

---

## 依赖关系

### 技术依赖

- `github.com/gin-gonic/gin` - HTTP 框架
- `google.golang.org/grpc` - gRPC 客户端（调用 Grab RPC 服务）
- `gorm.io/gorm` - ORM 框架
- `github.com/go-redis/redis/v8` - Redis 客户端
- 现有库存服务（`product.Service`）
- 现有 KDS 服务（`kds.Service`）

### 服务依赖

- **Main → ttpos-bmp (Grab RPC)**：gRPC 调用（订单同步、状态更新、接单、拒单）
- **POS 前端 → Main API**：HTTP 调用（订单列表、详情、接单、拒单）
- **Shop 前端 → Main API**：HTTP 调用（配置管理）
- **Main → KDS 服务**：内部调用（订单通知）

### RPC 接口依赖（待定）

**需要与 ttpos-bmp 团队协作定义**：

```protobuf
// 示例 Protobuf 定义（待确认）
service GrabOrderService {
  // 获取新订单列表
  rpc GetNewOrders(GetNewOrdersRequest) returns (GetNewOrdersResponse);
  
  // 获取订单详情
  rpc GetOrderDetail(GetOrderDetailRequest) returns (GetOrderDetailResponse);
  
  // 接单
  rpc AcceptOrder(AcceptOrderRequest) returns (AcceptOrderResponse);
  
  // 拒单
  rpc RejectOrder(RejectOrderRequest) returns (RejectOrderResponse);
  
  // 获取拒单原因列表
  rpc GetRejectReasons(GetRejectReasonsRequest) returns (GetRejectReasonsResponse);
  
  // 更新订单状态
  rpc UpdateOrderStatus(UpdateOrderStatusRequest) returns (UpdateOrderStatusResponse);
}
```

### 业务依赖

- **Grab RPC 服务**：需要 ttpos-bmp 团队先实现 Grab RPC 服务（对接 Grab 官方 API）
- **商品关联映射表**：需要提前配置（Shop 管理端，前端团队实现）
- **外卖开关**：需要在云平台开启（现有功能）
- **收银员权限**：需要"外卖"权限（现有权限系统）

### 前置条件

1. ttpos-bmp 团队完成 Grab RPC 服务开发和部署
2. Protobuf 接口定义完成并生成 Go 代码
3. Grab RPC 服务注册到 Nacos
4. 数据库迁移脚本执行完成

---

## 风险和缓解

### 风险 1: Grab RPC 接口尚未定义

**影响**: 高  
**概率**: 高  
**缓解措施**:

- 尽快与 ttpos-bmp 团队开会确定 RPC 接口
- 先定义 Protobuf，再并行开发（Main 端和 BMP 端）
- 使用 Mock RPC 服务进行本地开发和测试
- 建立清晰的接口文档和变更通知机制

### 风险 2: 商品映射复杂度

**影响**: 高  
**概率**: 高  
**缓解措施**:

- 设计灵活的商品映射表（支持一对多）
- 提供 HTTP API 给前端实现商品关联配置界面
- 异常订单明确返回未关联的商品/规格信息
- 支持手动处理异常订单

### 风险 3: RPC 调用稳定性

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 实现失败重试队列（最多 3 次）
- 记录详细日志，便于排查问题
- 设置超时和熔断机制
- 监控 RPC 调用成功率和延迟

### 风险 4: 高峰期性能问题

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 使用 Redis 队列异步处理订单同步
- 使用 Redis 缓存配置信息和订单列表
- 优化数据库查询（索引、分页）
- 实现限流保护（API 层面）

### 风险 5: 跨团队协作风险

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 明确职责边界（后端 vs 前端 vs BMP 团队）
- 建立定期沟通机制（每日站会）
- 使用 API 文档和 Protobuf 作为协作契约
- 优先完成接口定义和 Mock 数据

---

## 时间表

- **Phase 1 - RPC 接口定义和数据库设计**: 2 天
  - 与 ttpos-bmp 团队确定 Protobuf 定义
  - 数据库表设计和迁移脚本
  - Repository 层实现
  - Service 接口定义
- **Phase 2 - 订单同步和检查逻辑**: 3 天
  - RPC 客户端实现
  - 商品关联检查逻辑
  - 库存检查逻辑
  - 失败重试队列
- **Phase 3 - HTTP API 实现**: 3 天
  - 订单列表/详情接口
  - 接单/拒单接口
  - 配置管理接口
  - 拒单原因列表接口
- **Phase 4 - 业务逻辑和 KDS 联动**: 2 天
  - 自动接单判断逻辑
  - KDS 通知逻辑
  - 状态同步逻辑
- **Phase 5 - 测试和优化**: 2 天
  - 单元测试（Service + Repository）
  - 集成测试（Mock RPC）
  - API 测试（Postman）
  - 性能测试和优化
- **总计**: 12 天（SP = 15，待拆分）

**注意**：
- 前端开发由前端团队并行进行，不计入后端工作量
- ttpos-bmp 团队的 Grab RPC 服务开发不计入本工作量
- 需要 2 天用于跨团队协作和接口联调

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/vue.mdc` - Vue 开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构
- `docs/human/architecture/database-design.md` - 数据库设计

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- Grab 官方 API 文档: [待补充]
- ttpos-bmp Grab RPC 服务文档: [待创建]
- 现有外卖集成实现: `main/app/modules/takeout/`

---

## 跨团队协作清单

### 与 ttpos-bmp 团队

- [ ] 确定 Grab RPC 接口 Protobuf 定义
- [ ] 确定 RPC 服务名称和注册方式
- [ ] 确定错误码和错误处理策略
- [ ] 确定超时和重试策略
- [ ] 联调和集成测试

### 与前端团队

- [ ] 确认 HTTP API 响应格式和字段
- [ ] 确认订单列表筛选和搜索需求
- [ ] 确认异常订单展示方式
- [ ] 确认配置界面字段和交互
- [ ] API 联调和测试

### 与产品团队

- [ ] 确认自动接单规则
- [ ] 确认异常订单处理流程
- [ ] 确认库存不足处理方式
- [ ] 确认拒单原因展示方式

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/weifashi/2025-12/2025-12-22.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-22  
**作者**: weifashi  
**审核者**: [待指定]

