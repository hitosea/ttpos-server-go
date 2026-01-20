# 盘点单架构调整产品提案：ERPNext 作为数据承载，TTPOS 作为前端操作接口

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | xiehao   |
| **日期**   | 2025-12-18   |
| **目标版本** | v2.0.0 |
| **状态**   | 待评审   |
| **关联任务** | -      |
| **关联 Spec** | -      |

---

## 🎯 背景和动机

### 问题描述

当前盘点单在 TTPOS 中创建、编辑、提交、审核，然后同步到 ERPNext。这种架构存在以下问题：

**1. 数据一致性问题**
- TTPOS 和 ERPNext 数据可能不一致，单向同步导致数据不同步风险
- ERPNext 端操作无法回写到 TTPOS，造成数据割裂
- 据统计，数据不一致率约为 2-3%，影响库存管理准确性

**2. 操作灵活性限制**
- 只能在 TTPOS 端操作盘点单，ERPNext 端操作无法同步到 TTPOS
- 无法满足多端操作需求，限制了用户的使用场景
- 用户需要在两个系统间切换，操作体验不一致

**3. 24小时营业店面问题**
- 盘点时未结算订单的库存占用未考虑，盘点结果不准确
- 盘点差异率高达 10-15%，影响库存管理准确性
- 无法正确处理未结算订单的预出库数量

**4. 用户体系不一致**
- TTPOS 和 ERPNext 用户体系独立，操作人员记录不完整
- 审计追溯困难，无法完整记录操作来源
- 影响合规性和审计要求

### 业务价值

解决这些问题能带来以下业务价值：

**1. 数据准确性提升**
- 数据一致性从 97-98% 提升到 99.9% 以上
- 盘点差异率从 10-15% 降低到 5% 以下
- 库存管理更加可靠，减少库存损失

**2. 操作效率提升**
- 支持 TTPOS 和 ERPNext 两端操作，用户可选择最方便的方式
- 实时同步，无需等待，操作效率提升 30% 以上
- 减少用户系统切换，提升用户体验

**3. 系统解耦与可扩展性**
- TTPOS 作为前端操作接口，ERPNext 作为数据承载，职责清晰
- 支持更多前端应用接入，提升系统可扩展性
- 降低系统耦合度，提升可维护性

**4. 合规性保障**
- 完整的操作记录，支持审计追溯
- 符合财务审计要求，提升合规性

### 目标用户

- [x] 收银员
- [x] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: **店长/运营人员、仓库管理员、财务人员**

---

## 💡 解决方案概述

### 方案描述

将盘点单的数据承载从 **TTPOS 调整为 ERPNext**，实现 **ERPNext 作为主数据源，TTPOS 作为前端操作接口和数据缓存**的双向实时同步架构。

**核心调整**：
1. **数据源切换**：ERPNext 作为主数据源，存储完整的盘点单数据
2. **TTPOS 角色转变**：TTPOS 作为前端操作接口和数据缓存，提供快速查询和操作能力
3. **双向实时同步**：实现 TTPOS ↔ ERPNext 双向实时同步，确保数据一致性
4. **库存更新调整**：库存更新从 ERPNext 同步，TTPOS 侧同步更新库存并生成出入库明细记录
5. **字段来源统一**：TTPOS 侧所有字段数据均从 ERPNext 同步获取，确保数据一致性
6. **单据编号管理**：盘点单编号使用 ERPNext 的 Series 自动生成，TTPOS 不再独立生成

**同步机制**：
- **主要方式**：ERPNext Webhook 推送变更到 TTPOS（实时性好，延迟 < 5秒）
- **兜底方式**：定时轮询 ERPNext（每 5 分钟），确保数据一致性
- **冲突处理**：ERPNext 数据优先，实现版本控制

### 核心功能点

1. **双向实时同步机制**
   - TTPOS 操作时同步到 ERPNext（创建、编辑、提交、审核）
   - ERPNext 操作时实时同步到 TTPOS（Webhook 推送 + 轮询兜底）
   - 数据冲突自动处理（ERPNext 数据优先）

2. **ERPNext 工作流配置**
   - 在 ERPNext 中配置符合 TTPOS 盘点业务的工作流
   - 配置 Webhook 实现双向实时同步
   - 工作流状态与 TTPOS 业务流程匹配

3. **用户映射管理**
   - TTPOS 员工与 ERPNext 用户映射关系管理
   - 支持手动和自动映射
   - 完整的操作人员记录（TTPOS/ERPNext 来源标识）

4. **24小时营业处理**
   - 账面库存包含未结算订单预出库数量
   - 自动计算未结算订单占用
   - 盘点结果准确可靠

5. **库存更新与出入库明细**
   - ERPNext 审核盘点单后自动更新库存
   - TTPOS 侧同步更新库存数据（从 ERPNext 同步）
   - TTPOS 侧自动生成出入库明细记录（盘盈入库、盘亏出库）
   - 出入库明细记录关联盘点单编号，支持追溯

6. **字段数据来源管理**
   - TTPOS 侧所有字段数据均从 ERPNext 同步获取
   - 盘点门店默认使用当前门店（从 ERPNext 的 warehouse 字段映射）
   - 单据编号使用 ERPNext 的 Series 自动生成（格式：MAT-RECO-YYYY-XXXXX）
   - 物品明细、数量、金额等所有数据均从 ERPNext 同步

7. **ERP 取消单据处理**
   - ERPNext 侧可对已完成的单据进行取消（Cancel）
   - TTPOS 侧通过 Webhook 实时接收取消事件
   - 自动回滚库存更新（恢复盘点前库存）
   - 标记出入库明细记录为已取消状态
   - 更新盘点单状态为"已取消"

8. **同步状态监控与失败处理**
   - 同步状态展示（同步中、已同步、同步失败）
   - 同步错误提示和重试机制
   - 数据一致性监控和告警
   - Webhook 失败时的降级处理（轮询兜底）
   - API 调用失败时的用户提示和重试机制

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [x] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [x] UI 组件
- [x] API 接口
- [x] 数据模型
- [x] 业务逻辑
- [x] 第三方集成（ERPNext）
- [x] 其他: **同步机制、Webhook 接收、定时任务**

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [x] **高**：涉及架构调整、第三方集成、复杂算法

**复杂度说明**：
- 需要调整数据模型和业务流程
- 需要实现双向同步机制（Webhook + 轮询）
- 需要处理数据冲突和同步失败场景
- 需要与 ERPNext 深度集成

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 60-80 天（8-12周）
- **预估 SP**: 40-60 SP（待技术评审确认）

**阶段划分**：
1. **第一阶段：基础设施准备**（1-2周）
   - 数据库迁移、Webhook 基础设施、轮询任务框架
2. **第二阶段：核心功能调整**（2-3周）
   - 数据模型调整、业务流程调整、ERP 服务调整
3. **第三阶段：同步机制实现**（2-3周）
   - Webhook 实现、轮询实现、冲突处理
4. **第四阶段：测试与优化**（2-3周）
   - 功能测试、性能测试、稳定性测试
5. **第五阶段：上线与监控**（1周）
   - 灰度发布、全量上线、监控告警

### 风险识别

**潜在风险**：

1. **ERPNext API 不稳定**
   - **影响**：高
   - **概率**：中
   - **应对措施**：
     - 实现重试机制（指数退避）
     - 降级方案：ERPNext 不可用时，允许 TTPOS 本地操作，后续同步
     - 监控告警：实时监控 API 调用成功率，超过阈值后告警

2. **Webhook 丢失或延迟**
   - **影响**：中
   - **概率**：低
   - **应对措施**：
     - 轮询兜底：定时轮询 ERPNext，确保数据一致性
     - 消息队列备份（可选）：使用消息队列作为备份机制
     - 同步状态监控：实时监控同步状态，发现异常及时处理

3. **数据不一致**
   - **影响**：高
   - **概率**：低
   - **应对措施**：
     - 定时对比：定时对比 TTPOS 和 ERPNext 数据，发现不一致时自动修复
     - 版本控制：实现版本号或时间戳机制，检测数据冲突
     - 告警机制：数据不一致时及时告警，人工介入处理

4. **性能问题**
   - **影响**：中
   - **概率**：低
   - **应对措施**：
     - 异步处理：同步操作异步化，不阻塞用户操作
     - 批量同步：支持批量同步，提升同步效率
     - 缓存优化：优化数据查询和缓存策略

5. **用户体验下降**
   - **影响**：中
   - **概率**：低
   - **应对措施**：
     - 优化同步速度：Webhook 推送延迟 < 5秒
     - 增加状态提示：显示同步状态，让用户了解操作进度
     - 数据冲突提示：数据冲突时提示用户刷新，避免操作失败

6. **数据丢失**
   - **影响**：高
   - **概率**：低
   - **应对措施**：
     - 数据备份：定期备份数据，支持数据恢复
     - 事务保护：关键操作使用事务，确保数据一致性
     - 回滚机制：保留原有代码，支持快速回滚

7. **功能缺失**
   - **影响**：低
   - **概率**：低
   - **应对措施**：
     - 功能对比：对比 TTPOS 和 ERPNext 功能，确保功能完整性
     - 逐步迁移：分阶段迁移，降低风险
     - 兼容处理：保留原有接口，支持平滑过渡

**缓解措施**：

1. **降级方案**：ERPNext 不可用时，允许 TTPOS 本地操作，后续同步
2. **回滚方案**：保留原有代码，支持快速回滚
3. **监控告警**：实时监控同步状态，及时发现问题
4. **数据备份**：定期备份数据，支持数据恢复
5. **灰度发布**：选择部分商户灰度，逐步全量上线
6. **测试充分**：功能测试、性能测试、稳定性测试全覆盖

---

## 🔗 相关资源

### 参考需求

- 类似功能: 盘点单现有实现（`docs/human/business/stock-reconciliation-product-overview.md`）
- 竞品分析: ERPNext 盘点单功能

### 相关文档

- 产品需求文档 (PRD): `docs/human/business/stock-reconciliation-product-overview.md`
- 技术架构方案: `docs/human/architecture/stock-reconciliation-architecture-refactor.md`
- ERP 同步机制: `docs/human/business/stock-reconciliation-erp-sync.md`
- ERPNext API 文档: `docs/shared/api/stock-reconciliation-erpnext-api.md`
- 用户映射方案: `docs/human/architecture/stock-reconciliation-user-mapping.md`
- 24小时营业处理: `docs/human/business/stock-reconciliation-24hour-handling.md`

---

## 🔄 ERPNext 工作流配置与双向同步实现

### ERPNext 工作流配置

#### 1. 工作流设计原则

**匹配 TTPOS 业务流程**：
- TTPOS 盘点单状态：已保存（0）→ 已提交（1）→ 已审核（2）→ 已驳回（3）
- ERPNext 盘点单状态：Draft（0）→ Submitted（1）→ Cancelled（2）
- 工作流需要确保状态流转与 TTPOS 保持一致

#### 2. ERPNext 工作流配置步骤

**步骤 1：创建工作流（Workflow）**

在 ERPNext 中创建盘点单工作流：

```
工作流名称：Stock Reconciliation Workflow for TTPOS
文档类型：Stock Reconciliation
工作流状态：Active
```

**盘点门店与仓库映射规则**：
- TTPOS 创建盘点单时，盘点门店默认使用当前门店（从用户上下文获取）
- TTPOS 侧 `warehouse_uuid` 映射到 ERPNext 侧 `set_warehouse` 字段（通过仓库的 `erp_code` 关联）
- ERPNext 侧 `set_warehouse` 字段必须对应 TTPOS 当前门店的仓库编码
- 如果用户切换门店，盘点门店自动更新为新的当前门店

**工作流状态定义**：

| TTPOS 状态 | ERPNext 状态 | docstatus | 说明 |
|-----------|-------------|-----------|------|
| 已保存（0） | Draft | 0 | 草稿状态，可编辑 |
| 已提交（1） | Submitted | 1 | 已提交，等待审核 |
| 已审核（2） | Submitted | 1 | 已审核通过，库存已更新 |
| 已驳回（3） | Cancelled | 2 | 已驳回，可重新编辑 |

**步骤 2：配置工作流状态转换规则**

```
状态转换规则：
1. Draft → Submitted
   - 触发条件：用户点击"提交"
   - 权限：Stock Manager 或以上
   - 动作：更新 docstatus = 1

2. Submitted → Cancelled
   - 触发条件：用户点击"驳回"
   - 权限：Stock Manager 或以上
   - 动作：更新 docstatus = 2

3. Cancelled → Draft
   - 触发条件：用户点击"重新编辑"
   - 权限：Stock Manager 或以上
   - 动作：更新 docstatus = 0
```

**步骤 3：配置工作流动作（Workflow Actions）**

在 ERPNext 中配置工作流动作，确保状态转换时触发 Webhook：

```
动作 1：提交盘点单
- 触发时机：Draft → Submitted
- 动作类型：Update Field
- 字段：docstatus
- 值：1
- 触发 Webhook：是

动作 2：驳回盘点单
- 触发时机：Submitted → Cancelled
- 动作类型：Update Field
- 字段：docstatus
- 值：2
- 触发 Webhook：是

动作 3：重新编辑盘点单
- 触发时机：Cancelled → Draft
- 动作类型：Update Field
- 字段：docstatus
- 值：0
- 触发 Webhook：是
```

#### 3. ERPNext Webhook 配置

**步骤 1：创建 Webhook**

在 ERPNext 中创建 Webhook，监听盘点单的变更事件：

```
Webhook 名称：TTPOS Stock Reconciliation Sync
Webhook URL：https://ttpos-api.example.com/api/v1/webhook/erp/stock_reconciliation
请求方法：POST
请求头：
  - Content-Type: application/json
  - Authorization: Bearer {api_key}
启用条件：
  - 文档类型：Stock Reconciliation
  - 事件类型：After Save, After Submit, After Cancel
```

**步骤 2：配置 Webhook 数据格式**

ERPNext Webhook 推送的数据格式：

```json
{
  "event": "stock_reconciliation.after_save|after_submit|after_cancel",
  "data": {
    "name": "MAT-RECO-2025-00001",
    "doctype": "Stock Reconciliation",
    "docstatus": 0|1|2,
    "company": "Company A",
    "posting_date": "2025-12-18",
    "posting_time": "14:30:00",
    "set_warehouse": "WH-001",
    "purpose": "Stock Reconciliation",
    "items": [
      {
        "item_code": "MAT-001",
        "item_name": "大米",
        "qty": 100.000,
        "warehouse": "WH-001",
        "valuation_rate": 1.0
      }
    ],
    "modified": "2025-12-18 14:30:00",
    "modified_by": "user@example.com"
  }
}
```

**步骤 3：配置 Webhook 认证**

使用 API Key 或 Token 认证：

```
认证方式：Bearer Token
Token 来源：ERPNext API Key（从站点配置获取）
请求头：Authorization: Bearer {api_key}
```

#### 4. 双向同步实现机制

**4.1 TTPOS → ERPNext 同步**

**创建盘点单**：
```
1. TTPOS 用户创建盘点单
   - 盘点门店默认使用当前门店（从用户上下文获取）
   - 用户选择盘点类型、盘点目的等信息
   ↓
2. TTPOS 调用 ERPNext API 创建盘点单
   POST /api/v2/document/Stock Reconciliation
   {
     "naming_series": "MAT-RECO-.YYYY.-",  // 使用 ERPNext Series 自动生成编号
     "company": "Company A",
     "posting_date": "2025-12-18",
     "posting_time": "14:30:00",
     "set_warehouse": "WH-001",  // 当前门店对应的 ERPNext 仓库编码
     "items": [
       {
         "item_code": "MAT-001",
         "qty": 100.000,
         "warehouse": "WH-001"
       }
     ]
   }
   ↓
3. ERPNext 返回盘点单号（由 Series 自动生成）
   {
     "data": {
       "name": "MAT-RECO-2025-00001",  // ERPNext Series 生成的编号
       "status": "Draft"
     }
   }
   ↓
4. TTPOS 保存盘点单（标记 is_from_erp=0）
   - order_no = "MAT-RECO-2025-00001"  // 使用 ERPNext 生成的编号
   - erp_code = "MAT-RECO-2025-00001"
   - sync_status = 1（同步中）
   - warehouse_uuid = 当前门店仓库 UUID（从 ERPNext warehouse 映射）
   ↓
5. ERPNext Webhook 推送创建事件
   ↓
6. TTPOS 接收 Webhook，同步所有字段数据
   - 同步 items 明细数据
   - 同步所有字段（posting_date, posting_time, warehouse 等）
   - sync_status = 2（已同步）
   - sync_time = 当前时间
   
【同步失败处理】
- IF API 调用失败 THEN
  - 显示错误提示："创建盘点单失败，请重试"
  - 记录错误日志（sync_error 字段）
  - sync_status = 3（同步失败）
  - 提供"重试"按钮，用户可手动重试
- IF Webhook 推送失败 THEN
  - 轮询任务在 5 分钟内自动同步
  - 如果轮询也失败，标记 sync_status = 3，提示用户联系管理员
```

**提交盘点单**：
```
1. TTPOS 用户提交盘点单
   ↓
2. TTPOS 调用 ERPNext API 提交盘点单
   PUT /api/v2/document/Stock Reconciliation/{name}
   {
     "docstatus": 1
   }
   ↓
3. ERPNext 更新状态为 Submitted
   ↓
4. TTPOS 更新本地状态为"已提交"
   ↓
5. ERPNext Webhook 推送提交事件
   ↓
6. TTPOS 接收 Webhook，更新同步状态
```

**审核盘点单**：
```
1. TTPOS 用户审核盘点单
   ↓
2. TTPOS 调用 ERPNext API 提交盘点单（如果未提交）
   PUT /api/v2/document/Stock Reconciliation/{name}
   {
     "docstatus": 1
   }
   ↓
3. ERPNext 自动更新库存（ERPNext 侧库存已更新）
   ↓
4. ERPNext Webhook 推送审核事件（包含库存更新和明细数据）
   {
     "event": "stock_reconciliation.after_submit",
     "data": {
       "name": "MAT-RECO-2025-00001",
       "docstatus": 1,
       "items": [
         {
           "item_code": "MAT-001",
           "qty": 100.000,
           "book_qty": 95.000,  // 账面数量
           "counted_qty": 100.000  // 实盘数量
         }
       ]
     }
   }
   ↓
5. TTPOS 接收 Webhook，同步数据并更新库存
   a. 更新盘点单状态为"已审核"
   b. 同步所有字段数据（从 ERPNext 获取）
   c. 同步库存数据到 TTPOS
     - 更新 warehouse_item 表的 stock 字段
     - 根据盘点差异生成出入库明细记录
     - 盘盈：创建入库记录（scene=3, log_type=0）
     - 盘亏：创建出库记录（scene=4, log_type=1）
     - 记录关联盘点单编号（order_no）
   d. sync_status = 2（已同步）
   
【同步失败处理】
- IF API 调用失败 THEN
  - 回滚事务，保持盘点单为"已提交"状态
  - 显示错误提示："审核盘点单失败，请重试"
  - 记录错误日志
  - 提供"重试"按钮
- IF Webhook 推送失败 THEN
  - 轮询任务在 5 分钟内自动同步库存数据
  - 如果轮询也失败，标记 sync_status = 3，提示用户联系管理员
```

**4.2 ERPNext → TTPOS 同步**

**ERPNext 用户操作盘点单**：
```
1. ERPNext 用户创建/编辑/提交/审核/取消盘点单
   ↓
2. ERPNext 触发 Webhook
   POST https://ttpos-api.example.com/api/v1/webhook/erp/stock_reconciliation
   {
     "event": "stock_reconciliation.after_save|after_submit|after_cancel",
     "data": {
       "name": "MAT-RECO-2025-00001",
       "docstatus": 0|1|2,
       "company": "Company A",
       "posting_date": "2025-12-18",
       "posting_time": "14:30:00",
       "set_warehouse": "WH-001",
       "items": [
         {
           "item_code": "MAT-001",
           "qty": 100.000,
           "warehouse": "WH-001"
         }
       ],
       "modified": "2025-12-18 14:30:00",
       "modified_by": "user@example.com"
     }
   }
   ↓
3. TTPOS 接收 Webhook
   - 验证请求（API Key 认证）
   - 解析事件类型和数据
   ↓
4. TTPOS 同步所有字段数据（从 ERPNext 获取）
   - 查询 TTPOS 是否存在（通过 erp_code/order_no）
   - 如果不存在，创建新记录（标记 is_from_erp=1）
   - 如果存在，更新记录
   - 同步所有字段：
     * order_no = data.name（ERPNext 生成的编号）
     * erp_code = data.name
     * warehouse_uuid = 从 data.set_warehouse 映射到 TTPOS warehouse_uuid
     * status = 从 data.docstatus 映射（0=Draft, 1=Submitted, 2=Cancelled）
     * 同步 items 明细数据（从 ERPNext 获取）
   ↓
5. 根据事件类型处理业务逻辑
   a. IF event = "after_submit" AND docstatus = 1 THEN
      - 如果盘点单已审核，同步更新库存
      - 生成出入库明细记录（盘盈/盘亏）
   b. IF event = "after_cancel" AND docstatus = 2 THEN
      - 回滚库存更新（恢复盘点前库存）
      - 标记出入库明细记录为已取消
      - 更新盘点单状态为"已取消"
   ↓
6. TTPOS 更新同步状态
   - sync_status = 2（已同步）
   - sync_time = 当前时间
   ↓
7. TTPOS 通知前端（WebSocket/SSE）
   - 推送数据变更通知
   - 前端自动刷新列表和详情页
   - 显示操作来源提示："此盘点单由 ERPNext 操作"
   
【Webhook 失败处理】
- IF Webhook 接收失败 THEN
  - 记录错误日志
  - 返回 HTTP 500，触发 ERPNext 重试机制（指数退避）
  - 轮询任务在 5 分钟内自动同步（兜底机制）
- IF 数据同步失败 THEN
  - 记录错误日志（sync_error 字段）
  - sync_status = 3（同步失败）
  - 发送告警通知（超过阈值后）
```

**4.3 轮询兜底机制**

**定时轮询任务**：
```
1. 每 5 分钟执行一次轮询任务
   ↓
2. 查询 ERPNext 最近更新的盘点单（最近 10 分钟）
   GET /api/v2/document/Stock Reconciliation
   ?filters=[["modified", ">", "2025-12-18 14:20:00"]]
   ↓
3. 对比 TTPOS 和 ERPNext 数据
   - 通过 erp_code 匹配盘点单
   - 对比 modified 时间戳
   - 发现不一致时同步
   ↓
4. 同步数据
   - 如果 ERPNext 数据更新，更新 TTPOS
   - 如果 TTPOS 数据更新，更新 ERPNext
   - 记录同步日志
   ↓
5. 同步失败处理
   - 记录错误信息
   - 标记 sync_status = 3（同步失败）
   - 记录 sync_error
   - 超过重试次数后告警
```

#### 5. 工作流与业务流程匹配

**TTPOS 业务流程**：
```
创建盘点单（已保存）
    ↓ 用户操作：提交
提交盘点单（已提交）
    ↓ 用户操作：审核通过
审核盘点单（已审核）
    ↓ 自动更新库存
更新库存
```

**ERPNext 工作流**：
```
创建盘点单（Draft, docstatus=0）
    ↓ 工作流动作：提交
提交盘点单（Submitted, docstatus=1）
    ↓ ERPNext 自动处理
更新库存
```

**状态映射关系**：
- TTPOS "已保存"（status=0） ↔ ERPNext "Draft"（docstatus=0）
- TTPOS "已提交"（status=1） ↔ ERPNext "Submitted"（docstatus=1）
- TTPOS "已审核"（status=2） ↔ ERPNext "Submitted"（docstatus=1，库存已更新）
- TTPOS "已驳回"（status=3） ↔ ERPNext "Cancelled"（docstatus=2）
- TTPOS "已取消"（status=4，新增） ↔ ERPNext "Cancelled"（docstatus=2，已审核后取消）

**字段数据来源映射**：
- `order_no`：从 ERPNext `name` 字段同步（ERPNext Series 生成的编号）
- `erp_code`：从 ERPNext `name` 字段同步
- `warehouse_uuid`：从 ERPNext `set_warehouse` 字段映射（通过 erp_code 关联）
- `status`：从 ERPNext `docstatus` 字段映射
- `posting_date`：从 ERPNext `posting_date` 字段同步
- `posting_time`：从 ERPNext `posting_time` 字段同步
- `items`：从 ERPNext `items` 数组同步（包括 item_code, qty, warehouse 等）
- 所有字段数据均从 ERPNext 同步获取，TTPOS 不独立生成或修改

#### 6. 关键配置项

**ERPNext 工作流配置**：
- 工作流名称：`Stock Reconciliation Workflow for TTPOS`
- 文档类型：`Stock Reconciliation`
- 状态字段：`docstatus`
- 触发事件：`After Save`, `After Submit`, `After Cancel`

**ERPNext Webhook 配置**：
- Webhook URL：`https://ttpos-api.example.com/api/v1/webhook/erp/stock_reconciliation`
- 请求方法：`POST`
- 认证方式：`Bearer Token`
- 事件类型：`After Save`, `After Submit`, `After Cancel`

**TTPOS Webhook 接收接口**：
- 路由：`POST /api/v1/webhook/erp/stock_reconciliation`
- 认证：验证 API Key
- 处理：解析事件类型，同步数据

**单据编号管理规则**：
- **ERPNext 侧**：使用 Series 自动生成盘点单编号
  - Series 格式：`MAT-RECO-.YYYY.-`
  - 生成格式：`MAT-RECO-YYYY-XXXXX`（如：MAT-RECO-2025-00001）
  - 序列号从 00001 开始，每年重置
- **TTPOS 侧**：不再独立生成单据编号
  - `order_no` 字段使用 ERPNext 生成的编号（从 `name` 字段同步）
  - `erp_code` 字段也使用 ERPNext 生成的编号（从 `name` 字段同步）
  - 确保两个系统使用相同的单据编号，便于关联和追溯

#### 7. ERP 操作在 TTPOS 侧的体现

**7.1 ERPNext 操作类型与 TTPOS 侧体现**

| ERPNext 操作 | Webhook 事件 | TTPOS 侧体现 | 用户提示 |
|-------------|-------------|-------------|---------|
| 创建盘点单 | `after_save` (docstatus=0) | 创建新记录，同步所有字段 | "此盘点单由 ERPNext 创建" |
| 编辑盘点单 | `after_save` (docstatus=0) | 更新记录，同步所有字段 | "此盘点单已被 ERPNext 更新" |
| 提交盘点单 | `after_submit` (docstatus=1) | 更新状态为"已提交" | "此盘点单已被 ERPNext 提交" |
| 审核盘点单 | `after_submit` (docstatus=1) | 更新状态为"已审核"，同步库存，生成出入库明细 | "此盘点单已被 ERPNext 审核，库存已更新" |
| 取消盘点单 | `after_cancel` (docstatus=2) | 更新状态为"已取消"，回滚库存 | "此盘点单已被 ERPNext 取消，库存已回滚" |

**7.2 前端展示规则**

**列表页展示**：
- 显示操作来源标识：
  - `is_from_erp=1`：显示"ERPNext 创建"标签
  - `is_from_erp=0`：显示"TTPOS 创建"标签
- 显示同步状态：
  - `sync_status=1`：显示"同步中"状态
  - `sync_status=2`：显示"已同步"状态
  - `sync_status=3`：显示"同步失败"状态（红色警告）

**详情页展示**：
- 显示操作来源和操作人：
  - "创建人：ERPNext 用户（user@example.com）"
  - "最后更新：ERPNext 用户（user@example.com）于 2025-12-18 14:30:00"
- 显示同步状态和时间：
  - "同步状态：已同步"
  - "最后同步时间：2025-12-18 14:30:00"
- 如果同步失败，显示错误信息：
  - "同步失败：{错误详情}"
  - 提供"重试同步"按钮

**7.3 操作权限控制**

**ERPNext 创建的盘点单**：
- TTPOS 侧可以查看、导出
- TTPOS 侧可以编辑（但会同步到 ERPNext）
- TTPOS 侧可以提交、审核（但会同步到 ERPNext）
- 如果 ERPNext 已提交/审核，TTPOS 侧不允许编辑（只读）

**TTPOS 创建的盘点单**：
- TTPOS 侧可以正常操作（创建、编辑、提交、审核）
- 所有操作都会同步到 ERPNext

**7.4 实时通知机制**

**WebSocket/SSE 推送**：
- 当 ERPNext 操作盘点单时，TTPOS 通过 Webhook 接收事件
- TTPOS 后端通过 WebSocket/SSE 推送变更通知到前端
- 前端自动刷新列表和详情页
- 显示通知："盘点单 {order_no} 已被 ERPNext 更新"

#### 8. 库存更新与出入库明细处理机制

**8.1 库存更新流程（ERPNext 为数据承载）**

**审核盘点单时的库存更新**：
```
1. ERPNext 审核盘点单（docstatus=1）
   ↓
2. ERPNext 自动更新库存（ERPNext 侧库存已更新）
   ↓
3. ERPNext Webhook 推送审核事件（包含库存更新数据）
   ↓
4. TTPOS 接收 Webhook，同步库存数据：
   a. 从 ERPNext 同步库存数据
   b. 更新 TTPOS 侧 warehouse_item 表的 stock 字段
   c. 确保 TTPOS 侧库存与 ERPNext 侧一致
   ↓
5. TTPOS 生成出入库明细记录：
   a. 计算盘点差异：diff = counted_qty - booked_qty
   b. IF diff > 0 THEN 盘盈入库
      - 创建入库记录（log_type=0, scene=3）
      - 记录数量、单价、金额
   c. IF diff < 0 THEN 盘亏出库
      - 创建出库记录（log_type=1, scene=4）
      - 记录数量、单价、金额
   d. 关联盘点单编号（order_no）
   ↓
6. 完成库存更新和出入库明细记录
```

**8.2 出入库明细记录结构**

**记录字段**：
- `log_type`：日志类型（0-入库，1-出库）
- `scene`：场景（3-盘盈入库，4-盘亏出库）
- `warehouse_uuid`：仓库 UUID
- `material_uuid`：物品 UUID
- `material_name`：物品名称
- `num`：数量（差异绝对值）
- `price`：单价（物品基准单位单价）
- `amount`：金额（单价 × 数量）
- `order_no`：关联盘点单编号
- `is_cancelled`：是否已取消（0-正常，1-已取消）

**8.3 取消单据时的库存回滚**

**ERPNext 取消已完成单据**：
```
1. ERPNext 取消盘点单（docstatus=2）
   ↓
2. ERPNext 自动回滚库存（ERPNext 侧库存已回滚）
   ↓
3. ERPNext Webhook 推送取消事件
   ↓
4. TTPOS 接收 Webhook，回滚库存：
   a. 从 ERPNext 同步库存数据（已回滚后的库存）
   b. 更新 TTPOS 侧 warehouse_item 表的 stock 字段
   c. 计算回滚数量：
      - 盘盈回滚：减少库存（出库）
      - 盘亏回滚：增加库存（入库）
   ↓
5. 标记出入库明细记录为已取消：
   a. 更新 warehouse_in_out_log 表的 is_cancelled 字段
   b. is_cancelled = 1（已取消）
   c. 保留记录用于审计追溯（不删除）
   ↓
6. 更新盘点单状态为"已取消"（status=4）
```

**8.4 数据一致性保障**

**库存数据一致性**：
- TTPOS 侧库存数据始终从 ERPNext 同步获取
- 审核时：TTPOS 同步 ERPNext 的库存更新
- 取消时：TTPOS 同步 ERPNext 的库存回滚
- 确保 TTPOS 侧库存与 ERPNext 侧完全一致

**出入库明细记录**：
- TTPOS 侧自动生成出入库明细记录（用于查询和报表）
- 记录关联盘点单编号，支持追溯
- 取消时标记为已取消，保留记录用于审计

#### 9. 同步失败处理与错误提示机制

**7.1 TTPOS → ERPNext 同步失败处理**

**场景 1：API 调用失败（网络错误、ERPNext 服务不可用）**
```
处理流程：
1. 捕获错误，记录错误日志（sync_error 字段）
2. 标记 sync_status = 3（同步失败）
3. 前端显示错误提示：
   - 创建/编辑："创建盘点单失败，请检查网络连接或联系管理员"
   - 提交："提交盘点单失败，请重试或联系管理员"
   - 审核："审核盘点单失败，请重试或联系管理员"
4. 提供"重试"按钮，用户可手动重试
5. 自动重试机制（最多 3 次，指数退避）
6. 超过重试次数后，发送告警通知
```

**场景 2：ERPNext 业务校验失败（仓库禁用、物品禁用等）**
```
处理流程：
1. 解析 ERPNext 返回的错误信息
2. 前端显示具体错误提示：
   - "仓库状态已关闭，请修改仓库状态"
   - "物品XXX状态已关闭，请修改物品状态"
   - "盘点单数据校验失败：{具体错误信息}"
3. 保持 TTPOS 状态不变（不更新状态）
4. 记录错误日志
```

**场景 3：Webhook 推送失败（ERPNext 侧）**
```
处理流程：
1. ERPNext 自动重试（指数退避，最多 5 次）
2. TTPOS 轮询任务兜底（每 5 分钟轮询一次）
3. 如果轮询也失败，标记 sync_status = 3
4. 前端显示提示："数据同步异常，系统将在后台自动同步"
```

**7.2 ERPNext → TTPOS 同步失败处理**

**场景 1：Webhook 接收失败（TTPOS 服务不可用）**
```
处理流程：
1. TTPOS 返回 HTTP 500，触发 ERPNext 重试机制
2. 记录错误日志
3. 轮询任务兜底（每 5 分钟轮询一次）
4. 如果持续失败，发送告警通知
```

**场景 2：数据同步失败（字段映射错误、数据格式错误等）**
```
处理流程：
1. 记录详细错误日志（包含原始数据）
2. 标记 sync_status = 3（同步失败）
3. 记录 sync_error 字段（错误详情）
4. 发送告警通知（包含错误数据和堆栈信息）
5. 管理员可在后台查看错误详情并手动修复
```

**场景 3：库存更新失败（审核时）**
```
处理流程：
1. 回滚事务，保持盘点单为"已提交"状态
2. 记录错误日志
3. 前端显示错误提示："库存更新失败，请重试或联系管理员"
4. 提供"重试"按钮
5. 自动重试机制（最多 3 次）
```

**7.3 数据冲突处理**

**冲突检测机制**：
- 使用 `modified` 时间戳对比 TTPOS 和 ERPNext 数据
- 如果 ERPNext 数据更新，拒绝 TTPOS 修改

**冲突处理流程**：
```
1. TTPOS 检测到数据冲突（ERPNext 数据更新）
2. 前端显示提示："数据已被 ERPNext 更新，请刷新页面"
3. 自动刷新数据（从 ERPNext 同步最新数据）
4. 用户需要重新操作
```

**7.4 ERPNext 取消单据处理**

**ERPNext 侧取消已完成单据**：
```
1. ERPNext 用户取消盘点单（docstatus = 2）
   ↓
2. ERPNext 触发 Webhook（after_cancel 事件）
   ↓
3. TTPOS 接收 Webhook，处理取消逻辑：
   a. 更新盘点单状态为"已取消"（status = 4，新增状态）
   b. 回滚库存更新：
      - 恢复盘点前的库存数量
      - 计算回滚数量 = 实盘数量 - 账面数量
      - 盘盈回滚：减少库存（出库）
      - 盘亏回滚：增加库存（入库）
   c. 标记出入库明细记录为已取消：
      - 在 warehouse_in_out_log 表添加 is_cancelled 字段
      - 标记为已取消（is_cancelled = 1）
      - 保留记录用于审计追溯
   d. 同步所有字段数据（从 ERPNext 获取）
   ↓
4. 前端显示提示："此盘点单已被 ERPNext 取消，库存已回滚"
   - 显示取消原因（如果有）
   - 显示取消时间和操作人
```

**7.5 同步状态监控与告警**

**同步状态定义**：
- `sync_status = 0`：未同步（初始状态）
- `sync_status = 1`：同步中（API 调用中）
- `sync_status = 2`：已同步（同步成功）
- `sync_status = 3`：同步失败（需要人工介入）

**监控指标**：
- 同步成功率（目标：> 99%）
- 同步延迟（目标：< 5 秒）
- 同步失败率（阈值：> 1% 时告警）

**告警机制**：
- 同步失败率超过 1% 时发送告警
- 连续 3 次同步失败时发送告警
- 同步延迟超过 10 秒时发送告警
- 告警通知方式：邮件、企业微信、钉钉等

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-stock-reconciliation-erpnext-refactor`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 店长/运营人员  
**我想** 在 TTPOS 和 ERPNext 两端都能操作盘点单，并且操作实时同步  
**以便于** 我可以选择最方便的操作方式，提升操作效率

**作为** 仓库管理员  
**我想** 盘点结果准确可靠，24小时营业店面的未结算订单正确处理  
**以便于** 我可以准确掌握库存情况，减少库存损失

**作为** 财务人员  
**我想** 盘点数据准确，操作记录完整，支持审计追溯  
**以便于** 我可以进行准确的财务核算，符合审计要求

### AC 验收标准（初稿）

1. **WHEN** 用户在 TTPOS 创建盘点单 **THEN** 系统 **SHALL** 立即同步到 ERPNext，并在 5 秒内完成同步
   - 盘点门店默认使用当前门店（从用户上下文获取）
   - 单据编号使用 ERPNext 的 Series 自动生成（格式：MAT-RECO-YYYY-XXXXX）
   - TTPOS 的 order_no 字段使用 ERPNext 生成的编号

2. **WHEN** 用户在 ERPNext 操作盘点单 **THEN** 系统 **SHALL** 通过 Webhook 实时同步到 TTPOS，延迟 < 5秒
   - TTPOS 侧所有字段数据均从 ERPNext 同步获取
   - 前端显示操作来源提示："此盘点单由 ERPNext 操作"

3. **WHEN** Webhook 推送失败 **THEN** 系统 **SHALL** 通过轮询兜底，在 5 分钟内完成同步
   - 轮询任务每 5 分钟执行一次
   - 对比 TTPOS 和 ERPNext 数据，发现不一致时自动同步

4. **IF** TTPOS 和 ERPNext 数据冲突 **THEN** 系统 **SHALL** 以 ERPNext 数据为准，提示用户刷新
   - 使用 modified 时间戳检测冲突
   - 前端显示提示："数据已被 ERPNext 更新，请刷新页面"
   - 自动刷新数据

5. **WHEN** 盘点单审核通过 **THEN** 系统 **SHALL** 从 ERPNext 同步库存更新，确保数据一致性
   - TTPOS 侧同步更新库存数据（warehouse_item 表）
   - TTPOS 侧自动生成出入库明细记录（warehouse_in_out_log 表）
   - 盘盈生成入库记录（scene=3, log_type=0）
   - 盘亏生成出库记录（scene=4, log_type=1）
   - 出入库明细记录关联盘点单编号（order_no）

6. **WHEN** 24小时营业店面盘点 **THEN** 系统 **SHALL** 正确计算未结算订单的库存占用，盘点差异率 < 5%
   - 账面库存包含未结算订单预出库数量
   - 自动计算未结算订单占用

7. **WHEN** 同步失败 **THEN** 系统 **SHALL** 记录错误信息，支持重试，超过阈值后告警
   - API 调用失败时显示错误提示，提供"重试"按钮
   - Webhook 接收失败时轮询兜底
   - 记录详细错误日志（sync_error 字段）
   - 同步失败率超过 1% 时发送告警

8. **WHEN** ERPNext 取消已完成单据 **THEN** 系统 **SHALL** 实时同步取消状态，回滚库存更新
   - TTPOS 接收 Webhook 取消事件
   - 更新盘点单状态为"已取消"（status=4）
   - 回滚库存更新（恢复盘点前库存）
   - 标记出入库明细记录为已取消（is_cancelled=1）
   - 前端显示提示："此盘点单已被 ERPNext 取消，库存已回滚"

### 线框图/原型（可选）

[附加 UI 线框图或原型链接]

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-18  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

