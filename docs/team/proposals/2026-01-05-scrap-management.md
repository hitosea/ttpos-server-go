# 报废管理功能需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | {提案人姓名}   |
| **日期**   | 2026-01-05   |
| **目标版本** | v2.0.0 |
| **状态**   | 待评审   |
| **关联任务** | -      |
| **关联 Spec** | -      |

---

## 🎯 背景和动机

### 问题描述

当前 TTPOS 系统缺少报废品管理功能，门店在遇到损坏、过期、变质等无法使用的物品时，无法通过系统化的流程进行报废处理。这导致以下问题：

**1. 库存管理不准确**
- 报废品仍占用库存，导致账面库存与实际库存不符
- 无法准确追踪报废品的处理过程
- 影响库存盘点准确性，盘点差异率增加

**2. 操作流程不规范**
- 报废操作依赖手工记录，容易遗漏或错误
- 缺乏统一的报废单据管理
- 无法追溯报废原因和处理结果

**3. 与 ERP 系统数据不一致**
- TTPOS 侧报废操作无法同步到 ERP 系统
- ERP 系统中的 Stock Entry（Material Issue）单据需要手动创建
- 数据割裂导致财务核算不准确

**4. 审计追溯困难**
- 缺乏完整的报废记录和附件
- 无法追溯报废操作的历史记录
- 影响合规性和审计要求

### 业务价值

解决这些问题能带来以下业务价值：

**1. 库存管理准确性提升**
- 报废品及时扣减库存，库存准确率提升 5-10%
- 减少库存积压和资金占用
- 提升库存盘点准确性

**2. 操作流程规范化**
- 统一的报废单据管理流程
- 完整的报废记录和附件管理
- 提升操作效率和规范性

**3. 数据一致性保障**
- TTPOS 与 ERP 系统数据实时同步
- 自动生成 ERP Stock Entry 单据
- 减少人工操作错误

**4. 合规性保障**
- 完整的操作记录和审计追溯
- 符合财务审计要求
- 支持附件上传，保留证据

### 目标用户

- [ ] 收银员
- [x] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: **店长、仓库管理员、财务人员**

---

## 💡 解决方案概述

### 方案描述

在 TTPOS 侧设计报废管理板块，门店可通过提交报废单进行物品处理。报废单提交后自动同步到 ERP 系统，生成对应的 Stock Entry 单据（Stock Entry Type 为 Material Issue），并自动扣减库存。

**核心流程**：
1. **TTPOS 侧操作**：选择报损仓库 → 选择仓库中的物品 → 上传附件 → 提交报损单
2. **ERP 同步**：提交后自动同步到 ERP，生成 Stock Entry 单据（Material Issue 类型）
3. **库存扣减**：TTPOS 和 ERP 两侧同时扣减库存，记录出入库明细
4. **状态同步**：ERP 侧操作实时同步到 TTPOS，保持数据一致性

**同步机制**：
- **主要方式**：TTPOS 提交后立即调用 ERP API 同步（同步调用）
- **失败处理**：同步失败时记录错误，支持手动重试
- **ERP 操作同步**：通过 Webhook 或轮询机制，将 ERP 侧操作同步到 TTPOS

### 核心功能点

1. **报废单管理**
   - 创建报废单（选择仓库、物品、数量）
   - 上传附件（支持多文件上传）
   - 提交报废单（提交后等待审核）
   - 审核通过报废单（审核通过后扣减库存并同步到ERP）
   - 查看报废单列表和详情
   - 删除草稿报废单（仅能删除草稿状态，已同步需先取消ERP单据）

2. **ERP 集成与同步机制**
   
   **TTPOS 侧操作与 ERP 侧操作的实时对应关系**：
   
   | TTPOS 侧操作 | ERP 侧对应操作 | ERP 单据状态 | ERP Workflow | 说明 |
   |------------|--------------|------------|------------|------|
   | 保存报废单 | 创建 Stock Entry 单据并保存 | docstatus=0（Dratf） | - | 实时同步，创建ERP单据 |
   | 更新报废单 | 更新 Stock Entry 单据 | docstatus=0（Dratf） | - | 实时同步，更新ERP单据 |
| 提交报废单 | 提交审核 Stock Entry 单据 | docstatus=0（Pending review状态） | 进入审核流程 | 实时同步，提交ERP单据，触发Workflow审核，docstatus仍为0 |
| 审核通过报废单 | 触发 Workflow 审核通过 | docstatus=0 → docstatus=1（已完成） | 审核通过 | 实时同步，触发ERP Workflow审核，审核通过后docstatus变为1并扣减库存 |
   | 驳回报废单 | 触发 Workflow 驳回 | docstatus=0（Pending review状态，驳回） | 审核驳回 | 实时同步，触发ERP Workflow驳回，不扣减库存 |
   | 删除草稿报废单 | 删除 Stock Entry 单据 | docstatus=0 → 删除 | - | 实时同步，删除ERP草稿单据 |
   
   **ERP → TTPOS 同步操作（ERP 侧操作，TTPOS 被动接收）**：
   
   | ERP 侧操作 | TTPOS 侧对应操作 | ERP 单据状态 | 说明 |
   |-----------|----------------|------------|------|
   | 取消已完成报废单 | 接收 Webhook 更新状态 | docstatus=1, is_submitted=1 → docstatus=2（已取消） | ERP侧取消后，通过Webhook同步到TTPOS，回滚库存 |
   
   **TTPOS → ERP 同步流程**：
   
   **1. 保存报废单时（创建/更新）**：
   - **同步时机**：用户保存报废单时立即同步（同步调用，阻塞等待）
   - **ERP 操作**：创建或更新 Stock Entry 单据，状态为草稿（docstatus=0）
   - **同步方式**：通过 ttpos-bmp 服务的 gRPC 接口调用 ERP API
   - **ERP 单据类型**：Stock Entry 单据，Stock Entry Type 为 "Material Issue"
   - **单据参数**：
     - 报损门店默认为当前门店（从 CompanySetting 获取）
     - 单据编号使用 ERP Series（ERP 自动生成，首次创建时返回）
     - 过账日期和时间为当前时间
   - **同步流程**：
     1. 用户点击"保存"按钮（创建或更新报废单）
     2. 调用 ERP API 创建或更新 Stock Entry 单据（docstatus=0，Dratf状态）
     3. 同步成功：更新 `erp_code`（如果是创建）、`sync_status=2`（同步成功）、`sync_time`
     4. 同步失败：更新 `sync_status=3`（同步失败）、`sync_error_message`，记录错误日志
     5. **注意**：此时不扣减库存，ERP 侧单据为草稿状态
   
   **2. 提交报废单时**：
   - **同步时机**：用户提交报废单时立即同步（同步调用，阻塞等待）
   - **ERP 操作**：提交 Stock Entry 单据（触发 ERP Workflow 审核流程）
   - **同步流程**：
     1. 用户点击"提交"按钮
     2. 校验报废单状态（必须是已保存状态，status=0）
     3. 校验物品库存是否充足（仅校验，不扣减）
     4. 调用 ERP API 提交 Stock Entry 单据（docstatus=0，触发 ERP Workflow 审核流程，Pending review状态）
     5. 同步成功：更新报废单状态为"已提交"（status=1）、`sync_status=2`、`sync_time`
     6. 同步失败：更新 `sync_status=3`（同步失败）、`sync_error_message`，状态保持"已保存"（status=0），记录错误日志
     7. **注意**：此时不扣减库存，ERP 侧单据进入Workflow审核流程，docstatus仍为0（Pending review状态），等待审核
   
   **3. 审核通过报废单时**：
   - **同步时机**：审核人员审核通过报废单时立即同步（同步调用，阻塞等待）
   - **ERP 操作**：触发 Workflow 审核通过，单据状态从 docstatus=0 变为 docstatus=1
   - **ERP Workflow 审核机制**：
     - ERP 侧 Stock Entry 单据配置了 Workflow，增加了审核环节
     - Workflow 审核环节对应 TTPOS 侧的审核操作
     - 当 TTPOS 侧审核通过时，调用 ERP API 触发 Workflow 审核流程
     - ERP 侧 Workflow 审核通过后，单据状态从 docstatus=0（Pending review）变为 docstatus=1（已完成）
     - ERP 侧 Workflow 审核通过时，自动扣减 ERP 侧库存
   - **同步流程**：
     1. 审核人员点击"审核通过"按钮
     2. 校验报废单状态（必须是已提交状态，status=0）
     3. 再次校验物品库存是否充足（防止审核期间库存被消耗）
     4. 调用 ERP API 审核通过 Stock Entry 单据（触发 ERP Workflow 审核流程，Approved状态）
     5. 同步成功：
        - ERP 侧 Workflow 审核通过，单据状态变为已完成
        - 更新报废单状态为"已审核通过"（status=1）、`sync_status=1`、`sync_time`
        - **扣减 TTPOS 侧库存**：
          - 扣减 `ttpos_warehouse_item.stock`
          - 扣减 `ttpos_material.stock_num`
        - **记录出入库明细**（`ttpos_warehouse_in_out_log`）：
          - `log_type` = 1（出库）
          - `scene` = 7（报废出库，`WarehouseInOutLogSceneScrapOut`）
          - 记录完整的物品、数量、金额等信息
     6. 同步失败：更新 `sync_status=3`（同步失败）、`sync_error_message`，状态保持"已提交"（status=1），**不扣减库存**，记录错误日志
     7. **注意**：
        - 只有 ERP Workflow 审核通过后，TTPOS 侧才扣减库存
        - ERP 侧 Workflow 审核通过时，ERP 侧自动扣减库存
        - TTPOS 侧和 ERP 侧库存扣减时机一致（都在审核通过后）
   
   **4. 驳回报废单时**：
   - **同步时机**：审核人员驳回报废单时立即同步（同步调用，阻塞等待）
   - **ERP 操作**：触发 Workflow 驳回 Stock Entry 单据
   - **ERP Workflow 驳回机制**：
     - ERP 侧 Workflow 配置了驳回操作，审核人员可以驳回单据
     - Workflow 驳回对应 TTPOS 侧的驳回操作
     - 当 TTPOS 侧驳回时，调用 ERP API 触发 Workflow 驳回流程
     - ERP 侧 Workflow 驳回后，单据保持在 Pending review 状态，但标记为驳回
     - ERP 侧 Workflow 驳回时，不扣减库存
   - **同步流程**：
     1. 审核人员点击"驳回"按钮
     2. 校验报废单状态（必须是已提交状态，status=0，Pending review）
     3. 调用 ERP API 驳回 Stock Entry 单据（触发 ERP Workflow 审批流程，Rejected状态）
     4. 同步成功：
        - ERP 侧 Workflow 驳回，单据状态标记为已驳回
        - 更新报废单状态为"已驳回"（status=3）、`sync_status=2`、`sync_time`
        - **不扣减库存**（驳回时不需要扣减库存）
     5. 同步失败：更新 `sync_status=3`（同步失败）、`sync_error_message`，状态保持"已提交"（status=1），记录错误日志
     6. **注意**：
        - 驳回时不会扣减库存
        - 驳回后可以重新提交或修改报废单
        - 驳回原因会同步到 ERP 侧
   
   **5. 删除草稿报废单时**：
   - **同步时机**：用户删除草稿报废单时立即同步（同步调用，阻塞等待）
   - **ERP 操作**：删除 Stock Entry 单据
   - **删除场景**：
     - 只能删除状态为草稿（已保存，status=0）的报废单
     - 已提交（status=1）、已审核通过（status=2）、已驳回（status=3）的报废单不允许删除
   - **同步流程**：
     1. 用户点击"删除"按钮
     2. 校验报废单状态（只能是已保存状态，status=0，草稿状态）
     3. 校验是否已同步到 ERP（通过 `erp_code` 判断）
     4. **如果已同步到 ERP**：
        - 调用 ERP API 删除 Stock Entry 单据
        - 同步成功：软删除报废单（`delete_time` 设置为当前时间）
        - 同步失败：更新 `sync_status=3`、`sync_error_message`，不删除报废单，记录错误日志
     5. **如果未同步到 ERP**（`erp_code` 为空）：
        - 直接软删除报废单（`delete_time` 设置为当前时间）
        - 无需调用 ERP API
     6. **注意**：
        - 删除草稿时不扣减库存（草稿状态未扣减库存）
        - 已删除的报废单不会在列表中显示，但可以通过查询已删除数据查看
        - ERP 侧单据被删除后，无法恢复，需要重新创建报废单
   
   **ERP → TTPOS 同步（ERP 侧操作）**：
   
   - **同步方式**：采用 Webhook（推送）+ 轮询兜底的混合方案
   - **Webhook 机制（主要方式）**：
     - ERP 侧配置 Webhook，监听 Stock Entry 的状态变更事件
     - Webhook URL：`POST /api/v1/webhook/erp/stock_entry`
     - 监听事件：
       - `stock_entry.cancelled`：ERP 侧取消单据
       - `stock_entry.reverted`：ERP 侧删除单据
       - `stock_entry.submitted`：ERP 侧提交单据
       - `stock_entry.approved`：ERP 侧审核通过单据
     - 实时性好，延迟低（< 5秒）
     - 认证方式：使用 API Key 或 Token 认证（与 ERP 侧约定）
   - **轮询兜底机制（备用方式）**：
     - 定时任务每 5 分钟执行一次
     - 查询 ERP 中最近更新的 Stock Entry 单据（根据 `erp_code` 匹配）
     - 对比 TTPOS 和 ERP 的数据状态，发现不一致时同步更新
     - 作为 Webhook 的兜底方案，确保数据最终一致性
   - **同步流程**：
     1. ERP 侧操作（如取消）触发 Webhook 推送
     2. TTPOS 接收 Webhook 请求，验证签名和参数
     3. 根据 `erp_code` 查询 TTPOS 报废单
     4. 根据事件类型和单据状态，执行相应的同步操作（详细处理逻辑见下方各事件说明）
     5. 如果 Webhook 推送失败，由轮询任务兜底处理

   **同步失败后的具体处理方式**：
   
   - **失败场景分类**：
     
     1. **网络超时失败**
        - 触发条件：ERP API 调用超时（默认超时时间 30 秒）
        - 失败表现：HTTP 请求超时或 gRPC 调用超时
        - 处理方式：
          - 记录错误信息到 `sync_error_message` 字段（如："ERP API 调用超时：连接超时 30 秒"）
          - 更新 `sync_status=3`（同步失败）
          - 报废单状态保持为"已提交"（status=1），允许重试
          - 记录错误日志（包含报废单UUID、错误信息、时间戳）
          - 前端显示"同步失败：网络超时，请检查网络连接后重试"
        - 重试策略：支持手动重试，无次数限制（网络问题通常是暂时性的）
     
     2. **数据校验失败**
        - 触发条件：ERP 侧数据校验失败（如物品编码不存在、仓库编码不存在、数量无效等）
        - 失败表现：ERP 返回业务错误码（如：Item Code not found、Warehouse not found 等）
        - 处理方式：
          - 解析 ERP 返回的校验错误信息，详细记录到 `sync_error_message` 字段
          - 更新 `sync_status=3`（同步失败）
          - 报废单状态保持为"已提交"（status=1）
          - 记录错误日志（包含具体的失败字段和原因）
          - 前端显示"同步失败：数据校验失败 - [具体错误]，请检查报废单数据后修改并重试"
          - 在报废单详情页高亮显示错误的字段（如：物品编码不存在）
        - 重试策略：用户需要先修改报废单数据，然后才能重试同步
        - 特别注意：如果是因为物品或仓库在 ERP 中不存在，需要先同步基础数据
     
     3. **库存不足失败**
        - 触发条件：ERP 侧校验库存时发现库存不足（虽然 TTPOS 侧已扣减，但 ERP 侧库存可能已被其他操作消耗）
        - 失败表现：ERP 返回库存不足错误
        - 处理方式：
          - 记录错误信息到 `sync_error_message` 字段（如："物品 XXX 在 ERP 侧库存不足，当前库存：10，需要：20"）
          - 更新 `sync_status=3`（同步失败）
          - 报废单状态保持为"已提交"（status=1）
          - **重要**：TTPOS 侧库存已经扣减，但 ERP 侧未同步，存在数据不一致
          - 记录错误日志并标记为"需要人工处理"
          - 前端显示"同步失败：ERP 侧库存不足，请联系管理员处理（TTPOS 侧库存已扣减）"
        - 处理方案：
          - 方案A：管理员在 ERP 侧手动调整库存，然后在 TTPOS 侧重试同步
          - 方案B：TTPOS 侧回滚库存扣减，让用户重新提交（但需要确认 ERP 侧库存已恢复）
          - 方案C：记录不一致状态，等待定时同步任务处理

   - **重试机制详细说明**：
     
     **手动重试**：
     - 触发方式：用户在报废单详情页点击"重试同步"按钮
     - 重试条件：
       - 报废单状态为"已提交"（status=1）
       - 同步状态为"同步失败"（sync_status=3）
       - 重试次数未超过限制（记录在 `retry_count` 字段，新增字段）
     - 重试流程：
       1. 校验重试条件
       2. 更新 `retry_count` = `retry_count` + 1
       3. 更新 `sync_status=1`（同步中）
       4. 清空 `sync_error_message`（可选，保留最后一次错误）
       5. 重新调用 ERP API 同步
       6. 根据结果更新状态
     - 重试限制：
       - 网络超时：无限制（网络问题是暂时性的）
       - 数据校验失败：需要先修改数据，然后重试（不限制次数）
       - 库存不足：需要人工处理，不允许自动重试
     
     **自动重试（可选功能）**：
     - 触发方式：定时任务（每 5 分钟执行一次）
     - 重试条件：
       - 同步状态为"同步失败"（sync_status=3）
       - 失败原因属于可自动重试的类型（网络超时）
       - 失败时间超过重试间隔（首次失败后 5 分钟）
       - 自动重试次数未超过限制（最多 3 次）
     - 重试策略：指数退避
       - 第1次：失败后 5 分钟
       - 第2次：失败后 10 分钟
       - 第3次：失败后 20 分钟
     - 重试流程：
       1. 查询符合条件的报废单（`sync_status=3` 且 `auto_retry_count < 3`）
       2. 计算重试间隔，如果已达到重试时间，执行重试
       3. 更新 `auto_retry_count`（新增字段）
       4. 重新调用 ERP API
       5. 记录重试结果
     - 注意：自动重试仅在系统空闲时执行，避免影响正常业务
     
     **批量重试**：
     - 触发方式：管理员在报废单列表页选择多个失败的单据，点击"批量重试"
     - 重试流程：逐个执行重试，记录每个单据的重试结果
     - 使用场景：系统故障恢复后，批量处理失败的报废单

   - **数据一致性保障**：
     
     **TTPOS 侧已扣减但 ERP 未同步的处理**：
     - 场景：TTPOS 侧提交报废单时先扣减库存，但 ERP 同步失败
     - 问题：TTPOS 库存已扣减，但 ERP 未扣减，数据不一致
     - 处理方案：
       - 方案A（推荐）：保持当前状态，等待同步成功
         - TTPOS 侧库存已扣减（符合业务逻辑，物品已报废）
         - ERP 侧同步成功后会自动扣减库存
         - 如果同步一直失败，需要管理员手动处理
       - 方案B：回滚库存扣减（需要用户确认）
         - 如果用户确认不再报废，可以回滚 TTPOS 侧库存
         - 需要记录回滚操作日志
         - 报废单状态更新为"已取消"
     
     **ERP 侧取消但 TTPOS 未同步的处理**：
     - 场景：ERP 侧取消已完成单据，但 Webhook 推送失败，轮询任务还未执行
     - 问题：ERP 侧已取消，但 TTPOS 未取消，数据不一致
     - 处理方案：
       - 轮询任务会定期检查并同步状态
       - 如果发现不一致，自动回滚 TTPOS 侧库存
       - 更新报废单状态为"已取消"（status=4）
       - 记录同步日志，确保可追溯
     
     **最终一致性保障**：
     - 通过轮询兜底机制，确保最终数据一致
     - 定期对比 TTPOS 和 ERP 的数据，发现不一致时自动修复
     - 记录所有同步操作日志，支持审计追溯

3. **库存扣减**
   - TTPOS 侧扣减库存
   - ERP 侧扣减库存
   - 记录出入库明细（WarehouseInOutLog）

4. **数据一致性保障**
   - ERP 侧取消操作同步到 TTPOS
   - 出入库流水延迟处理机制
   - 同步失败重试机制

5. **字段数据来源**
   - 所有字段数据均从 ERP 获取
   - 仓库、物品、单位等信息同步自 ERP
   - **字段对齐说明**：
     - `scrap_reason`（报损原因）：TTPOS 侧必填，同步到 ERP 侧对应字段，字段名称对齐
     - `reject_reason`（驳回原因）：TTPOS 侧驳回时必填，同步到 ERP 侧对应字段，字段名称对齐
     - `remark`（备注）：TTPOS 侧可选，同步到 ERP 侧对应字段，字段名称对齐

## 📖 用户故事

### 用户故事1：仓库管理员

**作为** 仓库管理员  
**我希望** 能够通过系统管理报废单的完整流程  
**以便** 及时处理损坏、过期、变质等无法使用的物品，规范报废流程，确保库存准确性

**主要功能场景**：
- 创建报废单：选择仓库、物品，填写报废数量，上传附件，记录报废原因
- 修改和删除草稿报废单：更正错误信息或删除不需要的报废单
- 提交报废单：提交报废单进入审核流程，等待审核人员审核
- 查看报废单列表和详情：了解报废单的状态、同步状态、明细信息
- 重试同步：当同步失败时，能够重试同步到 ERP

**期望体验**：
- 操作简单直观，能够快速完成报废单的创建和提交
- 能够实时看到同步状态，了解与 ERP 的同步情况
- 同步失败时能够清楚看到错误原因并支持重试

### 用户故事2：店长

**作为** 店长  
**我希望** 能够审核和管理报废单  
**以便** 控制报废流程，确保报废操作的合理性和准确性

**主要功能场景**：
- 查看报废单列表：了解门店内所有的报废单情况
- 审核通过报废单：审核合理的报废申请，确认后自动扣减库存并同步到 ERP
- 驳回报废单：拒绝不合理的报废申请，可以要求重新提交
- 查看报废单详情和附件：了解报废的具体原因和证据
- 监控同步状态：确保报废单能够成功同步到 ERP

**期望体验**：
- 能够快速浏览和审核报废单，提高审核效率
- 能够清楚看到报废原因和附件，做出准确的审核决策
- 审核操作后能够及时看到结果和库存变化

### 用户故事3：财务人员

**作为** 财务人员  
**我希望** 能够查看报废单的完整信息和同步状态  
**以便** 进行财务核算和审计追溯

**主要功能场景**：
- 查看报废单列表：了解门店的报废情况，支持按时间、仓库、状态等筛选
- 查看报废单详情：查看完整的报废单信息、明细、金额、附件等
- 查看同步状态：确认报废单是否成功同步到 ERP，查看 ERP 单据编号
- 查看操作日志：了解报废单的操作历史和状态变化

**期望体验**：
- 能够方便地查询和筛选报废单，支持导出功能（未来）
- 能够看到完整的历史记录和审计信息
- 能够确认数据已正确同步到 ERP 系统

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
- [ ] UI 组件
- [x] API 接口
- [x] 数据模型
- [x] 业务逻辑
- [x] 第三方集成（ERP）
- [x] 其他: **出入库明细、库存管理、同步机制**

---

## 📊 详细技术方案

### 1. 数据库设计

#### 1.1 报废单主表（ttpos_scrap_order）

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| id | bigint unsigned | 主键ID | AUTO_INCREMENT |
| uuid | bigint unsigned | 唯一标识 | UNIQUE KEY |
| order_no | varchar(255) | 单据编号 | TTPOS侧生成 |
| erp_code | varchar(255) | ERP单据编号 | ERP Stock Entry 编号 |
| warehouse_uuid | bigint unsigned | 仓库ID | 关联 ttpos_warehouse |
| status | int | 状态 | 0-已保存 1-已审核通过（已完成） 2-已提交 3-已驳回 4-已取消 |
| sync_status | int | 同步状态 | 0-未同步 1-同步中 2-同步成功 3-同步失败 |
| sync_error_message | text | 同步错误信息 | 同步失败时记录详细错误信息 |
| sync_time | int | 同步时间 | 时间戳，最后成功同步的时间 |
| retry_count | int | 手动重试次数 | 默认为0，记录用户手动重试的次数 |
| auto_retry_count | int | 自动重试次数 | 默认为0，记录系统自动重试的次数 |
| last_retry_time | int | 最后重试时间 | 时间戳，最后一次重试的时间 |
| submit_time | int | 提交时间 | 时间戳 |
| scrap_reason | text | 报损原因 | 必填，说明报废原因 |
| reject_reason | text | 驳回原因 | 驳回时填写，从ERP同步或TTPOS侧填写 |
| attachment_urls | text | 附件URL列表 | JSON格式，存储多个附件URL |
| create_time | int | 创建时间 | 时间戳 |
| update_time | int | 更新时间 | 时间戳 |
| delete_time | int | 删除时间 | 时间戳 |

**索引**：
- PRIMARY KEY (`id`)
- UNIQUE KEY `uk_uuid` (`uuid`)
- KEY `idx_order_no` (`order_no`)
- KEY `idx_erp_code` (`erp_code`)
- KEY `idx_warehouse_uuid` (`warehouse_uuid`)
- KEY `idx_status` (`status`)
- KEY `idx_sync_status` (`sync_status`)

#### 1.2 报废单明细表（ttpos_scrap_order_item）

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| id | bigint unsigned | 主键ID | AUTO_INCREMENT |
| uuid | bigint unsigned | 唯一标识 | UNIQUE KEY |
| scrap_order_uuid | bigint unsigned | 报废单ID | 关联 ttpos_scrap_order |
| material_uuid | bigint unsigned | 物品ID | 关联 ttpos_material |
| material_name | text | 物品名称 | JSON格式，备份多语言 |
| material_code | varchar(255) | 物品编码 | ERP物品编码 |
| num | decimal(22,4) | 报废数量 | 基准单位数量 |
| price | decimal(22,4) | 单价 | 基准单位单价 |
| amount | decimal(22,4) | 金额 | 单价*数量 |
| material_base_unit_uuid | bigint unsigned | 基准单位ID | 关联 ttpos_material_unit |
| material_base_unit_name | text | 基准单位名称 | JSON格式 |
| create_time | int | 创建时间 | 时间戳 |
| update_time | int | 更新时间 | 时间戳 |
| delete_time | int | 删除时间 | 时间戳 |

**索引**：
- PRIMARY KEY (`id`)
- UNIQUE KEY `uk_uuid` (`uuid`)
- KEY `idx_scrap_order_uuid` (`scrap_order_uuid`)
- KEY `idx_material_uuid` (`material_uuid`)

#### 1.3 报废单单位明细表（ttpos_scrap_order_item_unit）

| 字段名 | 类型 | 说明 | 备注 |
|--------|------|------|------|
| id | bigint unsigned | 主键ID | AUTO_INCREMENT |
| uuid | bigint unsigned | 唯一标识 | UNIQUE KEY |
| scrap_order_item_uuid | bigint unsigned | 报废单明细ID | 关联 ttpos_scrap_order_item |
| material_unit_uuid | bigint unsigned | 单位ID | 关联 ttpos_material_unit |
| material_unit_name | text | 单位名称 | JSON格式，备份多语言 |
| quantity | decimal(22,4) | 单位数量 | 该单位的数量 |
| create_time | int | 创建时间 | 时间戳 |
| update_time | int | 更新时间 | 时间戳 |
| delete_time | int | 删除时间 | 时间戳 |

**索引**：
- PRIMARY KEY (`id`)
- UNIQUE KEY `uk_uuid` (`uuid`)
- KEY `idx_scrap_order_item_uuid` (`scrap_order_item_uuid`)
- KEY `idx_material_unit_uuid` (`material_unit_uuid`)

### 2. API 设计

#### 2.1 获取报废单列表

```
GET /api/v1/scrap_order/list
```

**请求参数**：
- `page_no` (int): 页码，默认1
- `page_size` (int): 每页数量，默认20
- `warehouse_uuids` ([]uint64): 仓库UUID列表，可选
- `status_in` ([]int): 状态列表，可选
- `keyword` (string): 关键字搜索（单据编号、ERP编号），可选
- `start_create_time` (int): 开始创建时间，时间戳，可选
- `end_create_time` (int): 结束创建时间，时间戳，可选

**响应格式**：
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "uuid": 1234567890,
        "order_no": "SCRAP-2026-00001",
        "erp_code": "MAT-STE-2026-00001",
        "warehouse_uuid": 9876543210,
        "warehouse_name": {"zh": "主仓库", "en": "Main Warehouse"},
        "status": 2,
        "sync_status": 2,
        "items_count": 3,
        "total_amount": 1000.00,
        "submit_time": 1704441600,
        "create_time": 1704441000
      }
    ],
    "meta": {
      "page_no": 1,
      "page_size": 20,
      "total": 100
    }
  }
}
```

#### 2.2 获取报废单详情

```
GET /api/v1/scrap_order/info
```

**请求参数**：
- `uuid` (uint64): 报废单UUID，必填

**响应格式**：
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "uuid": 1234567890,
    "order_no": "SCRAP-2026-00001",
    "erp_code": "MAT-STE-2026-00001",
    "warehouse_uuid": 9876543210,
    "warehouse_name": {"zh": "主仓库", "en": "Main Warehouse"},
    "status": 2,
    "sync_status": 2,
    "sync_error_message": "",
    "scrap_reason": "物品过期，需要报废",
    "reject_reason": "",
    "remark": "其他说明信息",
    "attachment_urls": ["https://example.com/attachment1.jpg", "https://example.com/attachment2.jpg"],
    "items": [
      {
        "uuid": 1111111111,
        "material_uuid": 2222222222,
        "material_name": {"zh": "苹果", "en": "Apple"},
        "material_code": "MAT-001",
        "num": 10.0000,
        "price": 5.0000,
        "amount": 50.0000,
        "material_base_unit_uuid": 3333333333,
        "material_base_unit_name": {"zh": "公斤", "en": "kg"},
        "units": [
          {
            "material_unit_uuid": 3333333333,
            "material_unit_name": {"zh": "公斤", "en": "kg"},
            "quantity": 10.0000
          }
        ]
      }
    ],
    "submit_time": 1704441600,
    "create_time": 1704441000,
    "update_time": 1704441600
  }
}
```

#### 2.3 创建/更新报废单

```
POST /api/v1/scrap_order/create
POST /api/v1/scrap_order/update
```

**请求参数**：
- `uuid` (uint64): 报废单UUID，更新时必填
- `warehouse_uuid` (uint64): 仓库UUID，必填
- `scrap_reason` (string): 报损原因，必填
- `attachment_urls` ([]string): 附件URL列表，可选
- `remark` (string): 备注，可选
- `items` ([]ScrapOrderItemReq): 报废单明细列表，必填
  - `material_uuid` (uint64): 物品UUID，必填
  - `num` (float64): 报废数量（基准单位），必填
  - `price` (float64): 单价，必填
  - `units` ([]ScrapOrderItemUnitReq): 单位明细列表，必填
    - `material_unit_uuid` (uint64): 单位UUID，必填
    - `quantity` (float64): 单位数量，必填

**响应格式**：
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "uuid": 1234567890
  }
}
```

**业务逻辑**：
1. 校验报损原因（必填，不能为空）
2. 创建或更新报废单数据（保存到数据库，包含报损原因）
3. **实时同步到 ERP**：
   - 如果是创建：调用 ERP API 创建 Stock Entry 单据（docstatus=0，草稿状态）
   - 如果是更新：调用 ERP API 更新 Stock Entry 单据（docstatus=0，草稿状态）
4. 同步成功：更新 `erp_code`（如果是创建）、`sync_status=2`、`sync_time`
5. 同步失败：更新 `sync_status=3`、`sync_error_message`，记录错误日志
6. **注意**：此时不扣减库存，ERP 侧单据为草稿状态，报损原因会同步到 ERP 侧（与TTPOS侧scrap_reason字段对齐）

#### 2.4 提交报废单

```
POST /api/v1/scrap_order/submit
```

**请求参数**：
- `uuid` (uint64): 报废单UUID，必填

**响应格式**：
```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

**业务逻辑**：
1. 校验报废单状态（必须是已保存状态，status=0）
2. 校验物品库存是否充足（仅校验，不扣减）
3. **调用 ERP API 提交 Stock Entry 单据**（docstatus=0，触发 ERP Workflow 审核流程，Pending review状态）
4. 同步成功：更新报废单状态为"已提交"（status=1）、`sync_status=2`、`sync_time`
5. 同步失败：更新 `sync_status=3`、`sync_error_message`，状态保持"已保存"（status=0），记录错误日志
6. **注意**：此时不扣减库存，ERP 侧单据为 Pending review 状态，进入 Workflow 审核流程，等待审核

#### 2.5 审核通过报废单

```
POST /api/v1/scrap_order/approve
```

**请求参数**：
- `uuid` (uint64): 报废单UUID，必填
- `remark` (string): 审核备注，可选

**响应格式**：
```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

**业务逻辑**：
1. 校验报废单状态（必须是已提交状态，status=1）
2. 再次校验物品库存是否充足（防止审核期间库存被消耗）
3. **调用 ERP API 审核通过 Stock Entry 单据**（触发 Workflow 审核通过，docstatus从0变为1）
4. 同步成功：
   - 更新报废单状态为"已审核通过"（status=2）、`sync_status=2`、`sync_time`
   - **扣减 TTPOS 侧库存**：
     - 扣减 `ttpos_warehouse_item.stock`
     - 扣减 `ttpos_material.stock_num`
   - **记录出入库明细**（`ttpos_warehouse_in_out_log`）：
     - `log_type` = 1（出库，`WarehouseInOutLogLogTypeOut`）
     - `scene` = 7（报废出库，`WarehouseInOutLogSceneScrapOut`）
     - 记录完整的物品、数量、金额等信息
5. 同步失败：更新 `sync_status=3`、`sync_error_message`，状态保持"已提交"（status=1），**不扣减库存**，记录错误日志
6. **注意**：只有 ERP 审核通过后，TTPOS 侧才扣减库存

#### 2.6 驳回报废单

```
POST /api/v1/scrap_order/reject
```

**请求参数**：
- `uuid` (uint64): 报废单UUID，必填


**响应格式**：
```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

**请求参数**：
- `uuid` (uint64): 报废单UUID，必填
- `reject_reason` (string): 驳回原因，必填

**业务逻辑**：
1. 校验报废单状态（必须是已提交状态，status=1，Pending review）
2. 校验驳回原因（必填，不能为空）
3. 校验操作权限（只有审核人员可以驳回）
4. **调用 ERP API 驳回 Stock Entry 单据**（触发 ERP Workflow 驳回流程，传递驳回原因）
5. 同步成功：
   - 更新报废单状态为"已驳回"（status=3）、`sync_status=2`、`sync_time`
   - 记录驳回原因到 `reject_reason` 字段
   - **不扣减库存**（驳回时不需要扣减库存）
6. 同步失败：更新 `sync_status=3`、`sync_error_message`，状态保持"已提交"（status=1），记录错误日志
7. **注意**：
   - 驳回时不会扣减库存
   - 驳回后可以重新提交或修改报废单
   - 驳回原因会同步到 ERP 侧，供后续参考

#### 2.7 删除草稿报废单

```
DELETE /api/v1/scrap_order/delete
```

**请求参数**：
- `uuid` (uint64): 报废单UUID，必填

**响应格式**：
```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

**业务逻辑**：

**删除场景**：
- **场景1：已同步到 ERP 的草稿报废单**
  - 条件：报废单状态为已保存（status=0），且 `erp_code` 不为空
  - 处理：
    1. 校验报废单状态（只能是已保存状态，status=0，草稿状态）
    2. 调用 ERP API 取消 Stock Entry 单据（docstatus=0 → docstatus=2，已取消）
    3. 同步成功：软删除报废单（`delete_time` 设置为当前时间）
    4. 同步失败：更新 `sync_status=3`、`sync_error_message`，不删除报废单，记录错误日志，允许重试
- **场景2：未同步到 ERP 的草稿报废单**
  - 条件：报废单状态为已保存（status=0），且 `erp_code` 为空
  - 处理：
    1. 校验报废单状态（只能是已保存状态，status=0，草稿状态）
    2. 直接删除报废单（`delete_time` 设置为当前时间）
    3. 无需调用 ERP API

**删除限制**：
- ✅ **允许删除**：已保存状态（status=0，草稿状态）的报废单
- ❌ **不允许删除**：
  - 已提交（status=1，Pending review）
  - 已审核通过（status=1，已完成）
  - 已驳回（status=3）
  - 已取消（status=4）

**特殊情况处理**：
- 如果 ERP 侧单据已提交或已审核，需要先在 ERP 侧取消，然后才能在 TTPOS 侧删除
- 删除操作不可恢复，删除前需要用户确认

**注意事项**：
- 删除草稿时不扣减库存（草稿状态未扣减库存）
- 删除是软删除，数据保留在数据库中，只是标记为已删除（`delete_time > 0`）
- 已删除的报废单不会在列表中显示，但可以通过查询已删除数据查看
- ERP 侧单据被取消后，无法恢复，需要重新创建报废单
- 如果同步到 ERP 失败，报废单不会被删除，用户可以重试删除

#### 2.7 删除草稿报废单

```
DELETE /api/v1/scrap_order/delete
```

**请求参数**：
- `uuid` (uint64): 报废单UUID，必填

**响应格式**：
```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

**业务逻辑**：

**删除场景**：
- **场景1：已同步到 ERP 的草稿报废单**
  - 条件：报废单状态为已保存（status=0），且 `erp_code` 不为空
  - 处理：
    1. 校验报废单状态（只能是已保存状态，status=0，草稿状态）
    2. 调用 ERP API 取消 Stock Entry 单据（docstatus=0 → docstatus=2，已取消）
    3. 同步成功：软删除报废单（`delete_time` 设置为当前时间）
    4. 同步失败：更新 `sync_status=3`、`sync_error_message`，不删除报废单，记录错误日志，允许重试
- **场景2：未同步到 ERP 的草稿报废单**
  - 条件：报废单状态为已保存（status=0），且 `erp_code` 为空
  - 处理：
    1. 校验报废单状态（只能是已保存状态，status=0，草稿状态）
    2. 直接软删除报废单（`delete_time` 设置为当前时间）
    3. 无需调用 ERP API

**删除限制**：
- ✅ **允许删除**：已保存状态（status=0，草稿状态）的报废单
- ❌ **不允许删除**：
  - 已提交（status=0，Pending review）
  - 已审核通过（status=1，已完成）
  - 已驳回（status=3）
  - 已取消（status=4）

**特殊情况处理**：
- 如果 ERP 侧单据已提交或已审核，需要先在 ERP 侧取消，然后才能在 TTPOS 侧删除
- 删除操作不可恢复，删除前需要用户确认

**注意事项**：
- 删除草稿时不扣减库存（草稿状态未扣减库存）
- 删除是软删除，数据保留在数据库中，只是标记为已删除（`delete_time > 0`）
- 已删除的报废单不会在列表中显示，但可以通过查询已删除数据查看
- ERP 侧单据被取消后，无法恢复，需要重新创建报废单
- 如果同步到 ERP 失败，报废单不会被删除，用户可以重试删除

#### 2.9 重试同步

```
POST /api/v1/scrap_order/retry_sync
```

**请求参数**：
- `uuid` (uint64): 报废单UUID，必填

**响应格式**：
```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

**业务逻辑**：
- 重新调用 ERP API 同步报废单
- 更新同步状态和错误信息

### 3. ERP 集成设计

#### 3.1 TTPOS → ERP 同步操作对应关系

**操作对应表**：

| TTPOS 操作 | ERP API 方法 | ERP 单据状态变化 | ERP Workflow | 说明 |
|-----------|------------|----------------|------------|------|
| 创建/更新报废单 | `SaveStockEntry` | 创建/更新，docstatus=0（草稿） | - | 保存时同步 |
| 提交报废单 | `SubmitStockEntry` | docstatus=0（保持为0，Pending review状态） | 进入审核流程 | 提交时同步，触发Workflow审核，docstatus仍为0 |
| 审核通过报废单 | `ApproveStockEntry` | docstatus=0 → docstatus=1（已完成） | 审核通过 | 审核时同步，触发Workflow审核通过，docstatus从0变为1 |
| 删除草稿报废单 | `DeleteStockEntry` | docstatus=0 → 删除 | - | 删除时同步，仅适用于草稿状态 |

**接口路径**：通过 ttpos-bmp 服务的 gRPC 接口调用 ERP

**通用调用流程**：
1. 准备请求参数
   - 从 `CompanySetting` 获取 `ErpnextCompanyAbbr` 和 `ErpnextBranchName`
   - 从报废单获取仓库、物品、数量等信息
   - 转换为 ERP 需要的格式
2. 调用 gRPC 接口
   - 服务：`ttpos-bmp/app/ttpos-erp`
   - 方法：根据操作类型调用对应方法
   - 超时时间：30 秒
3. 处理响应
   - 成功：解析返回的数据，更新报废单状态
   - 失败：解析错误信息，更新同步状态和错误消息

#### 3.2 创建/更新报废单（SaveStockEntry）

**调用时机**：用户保存报废单时立即调用（同步调用，阻塞等待）

**请求参数**：
```go
type SaveStockEntryReq struct {
    CompanyAbbr string // ERP公司简称，从 CompanySetting.ErpnextCompanyAbbr 获取
    Branch      string // ERP分支名称，从 CompanySetting.ErpnextBranchName 获取
    StockEntryType string // 固定为 "Material Issue"
    PostingDate int64  // 过账日期，时间戳（当前日期）
    PostingTime int64  // 过账时间，时间戳（当前时间）
    Items []StockEntryItemReq // 明细列表，从报废单明细转换
    ScrapReason string // 报损原因，从报废单 scrap_reason 字段获取，必填，与ERP侧字段对齐
    Remark string // 备注，从报废单 remark 字段获取，可选
    ErpCode string // ERP单据编号，更新时必填，创建时为空
}

type StockEntryItemReq struct {
    ItemCode string // 物品编码，从 material.code 获取
    Warehouse string // 仓库编码，从 warehouse.erp_code 获取
    Qty float64 // 数量，基准单位数量，从 item.num 获取
    BasicRate float64 // 单价，基准单位单价，从 item.price 获取
    SerialNo string // 序列号，可选，暂时为空
}
```

**响应处理**：
- 成功：返回 `erp_code`（如果是创建），更新 `sync_status=2`、`sync_time`
- 失败：更新 `sync_status=3`、`sync_error_message`，记录错误日志

#### 3.3 提交报废单（SubmitStockEntry）

**调用时机**：用户提交报废单时立即调用（同步调用，阻塞等待）

**请求参数**：
```go
type SubmitStockEntryReq struct {
    ErpCode string // ERP单据编号，必填
    CompanyAbbr string // ERP公司简称
    Branch string // ERP分支名称
    ScrapReason string // 报损原因，必填，与TTPOS侧scrap_reason字段对齐
}
```

**响应处理**：
- 成功：
  - 更新报废单状态为"已提交"（status=1）、`sync_status=2`、`sync_time`
  - 报损原因已同步到 ERP 侧（与TTPOS侧scrap_reason字段对齐）
- 失败：更新 `sync_status=3`、`sync_error_message`，状态保持"已保存"（status=0）

#### 3.4 审核通过报废单（ApproveStockEntry）

**调用时机**：审核人员审核通过报废单时立即调用（同步调用，阻塞等待）

**ERP Workflow 审核机制说明**：
- **ERP 侧配置**：Stock Entry 单据在 ERP 侧配置了 Workflow，增加了审核环节
- **Workflow 对应关系**：ERP 侧 Workflow 审核环节对应 TTPOS 侧的审核操作
- **审核流程**：
  1. TTPOS 侧审核人员点击"审核通过"按钮
  2. TTPOS 调用 ERP API `ApproveStockEntry`，触发 ERP Workflow 审核流程
  3. ERP 侧 Workflow 执行审核逻辑（可能包含多级审核、审批规则等）
  4. ERP 侧 Workflow 审核通过后：
     - Stock Entry 单据状态从 docstatus=0（Pending review）变为 docstatus=1（已完成）
     - ERP 侧自动扣减库存
  5. TTPOS 侧收到审核通过响应后，扣减 TTPOS 侧库存
- **审核权限**：ERP 侧 Workflow 配置了审核权限，只有具有审核权限的用户才能审核通过
- **审核规则**：ERP 侧 Workflow 可以配置审核规则（如金额限制、多级审核等）

**请求参数**：
```go
type ApproveStockEntryReq struct {
    ErpCode string // ERP单据编号，必填
    CompanyAbbr string // ERP公司简称
    Branch string // ERP分支名称
    Approver string // 审核人（可选，从 TTPOS 当前登录用户获取）
    ApproveRemark string // 审核备注（可选）
}
```

**响应处理**：
- 成功：
  - ERP 侧 Workflow 审核通过，单据状态从 docstatus=0（Pending review）变为 docstatus=1（已完成）
  - 更新报废单状态为"已审核通过"（status=2）、`sync_status=2`、`sync_time`
  - 扣减 TTPOS 侧库存
  - 记录出入库明细
- 失败：
  - 如果 ERP Workflow 审核被拒绝，返回审核拒绝原因
  - 更新 `sync_status=3`、`sync_error_message`，状态保持"已提交"（status=1），不扣减库存
  - 记录错误日志，包含审核拒绝原因

#### 3.5 驳回报废单（RejectStockEntry）

**调用时机**：审核人员驳回报废单时立即调用（同步调用，阻塞等待）

**ERP Workflow 驳回机制说明**：
- **ERP 侧配置**：Stock Entry 单据在 ERP 侧配置了 Workflow，支持驳回操作
- **Workflow 对应关系**：ERP 侧 Workflow 驳回环节对应 TTPOS 侧的驳回操作
- **驳回流程**：
  1. TTPOS 侧审核人员点击"驳回"按钮
  2. TTPOS 调用 ERP API `RejectStockEntry`，触发 ERP Workflow 驳回流程
  3. ERP 侧 Workflow 执行驳回逻辑（可能包含多级审核、驳回规则等）
  4. ERP 侧 Workflow 驳回后：
     - Stock Entry 单据状态保持 Pending review，但标记为驳回
     - ERP 侧不扣减库存
  5. TTPOS 侧收到驳回响应后，更新状态为"已驳回"
- **驳回权限**：ERP 侧 Workflow 配置了驳回权限，只有具有驳回权限的用户才能驳回
- **驳回规则**：ERP 侧 Workflow 可以配置驳回规则（如驳回后重新提交流程等）

**请求参数**：
```go
type RejectStockEntryReq struct {
    ErpCode string // ERP单据编号，必填
    CompanyAbbr string // ERP公司简称
    Branch string // ERP分支名称
    RejectReason string // 驳回原因，必填，与TTPOS侧reject_reason字段对齐
    Rejecter string // 驳回人（可选，从 TTPOS 当前登录用户获取）
}
```

**响应处理**：
- 成功：
  - ERP 侧 Workflow 驳回，单据状态保持 Pending review，标记为驳回
  - ERP 侧记录驳回原因（与TTPOS侧reject_reason字段对齐）
  - 更新报废单状态为"已驳回"（status=3）、`sync_status=2`、`sync_time`
  - 记录驳回原因到 `reject_reason` 字段
  - 不扣减库存
- 失败：
  - 更新 `sync_status=3`、`sync_error_message`，状态保持"已提交"（status=1）
  - 记录错误日志

**注意事项**：
- 驳回操作不会扣减库存
- 驳回后可以重新提交或修改报废单

#### 3.6 删除草稿报废单（DeleteStockEntry）

**调用时机**：用户删除草稿报废单时立即调用（同步调用，阻塞等待）

**删除场景说明**：
- **场景1：已同步到 ERP 的草稿报废单**（`erp_code` 不为空）
  - 需要调用 ERP API 取消 Stock Entry 单据
  - ERP 侧单据状态：docstatus=0（草稿）→ docstatus=2（已取消）
- **场景2：未同步到 ERP 的草稿报废单**（`erp_code` 为空）
  - 无需调用 ERP API
  - 直接删除 TTPOS 侧报废单

**请求参数**：
```go
type CancelStockEntryReq struct {
    ErpCode string // ERP单据编号，必填（如果已同步到ERP）
    CompanyAbbr string // ERP公司简称
    Branch string // ERP分支名称
}
```

**响应处理**：

**场景1：已同步到 ERP**：
- 成功：
  - ERP 侧单据被取消（docstatus=2）
  - TTPOS 侧报废单软删除（`delete_time` 设置为当前时间）
- 失败：
  - 更新 `sync_status=3`、`sync_error_message`
  - 不删除报废单，记录错误日志
  - 允许用户重试删除

**场景2：未同步到 ERP**：
- 直接软删除报废单，无需调用 ERP API

**注意事项**：
- 只能删除草稿状态（status=0）的报废单
- 删除草稿时不扣减库存（草稿状态未扣减库存）
- 删除是软删除，数据保留在数据库中
- ERP 侧单据被取消后无法恢复

**通用错误类型识别**：
- 网络超时：`err` 为 timeout 错误
- ERP 系统错误：`Code` 为 "500" 或 "系统错误"
- 数据校验失败：`Code` 为 "400" 或包含 "not found"、"invalid" 等关键字
- 库存不足：`Message` 包含 "stock"、"inventory"、"不足" 等关键字
- 权限不足：`Code` 为 "401" 或 "403"

#### 3.6 ERP → TTPOS 同步（ERP 侧操作，TTPOS 被动接收）

**说明**：以下操作由 ERP 侧主动发起，TTPOS 侧通过 Webhook 或轮询机制被动接收并更新状态。

**方案选择**：采用 **Webhook（推送）+ 轮询兜底** 的混合方案

**方案1：Webhook（主要方式）**

**ERP 侧配置**：
- 在 ERPNext 中配置 Webhook，监听 Stock Entry 的状态变更事件
- **Workflow 配置**：Stock Entry 单据配置了 Workflow，增加了审核环节
  - Workflow 审核环节对应 TTPOS 侧的审核操作
  - Workflow 可以配置多级审核、审核规则、审核权限等
  - Workflow 审核通过后，单据状态变为已审核，自动扣减库存
- Webhook URL：`POST /api/v1/webhook/erp/stock_entry`
   - 监听事件类型：
     - `stock_entry.cancelled`：ERP 侧取消单据（docstatus=2）
     - `stock_entry.submitted`：ERP 侧提交单据（docstatus=0，进入Workflow审核流程，Pending review状态）
     - `stock_entry.approved`：ERP 侧 Workflow 审核通过单据（docstatus=1, is_submitted=1，已完成）
     - `stock_entry.workflow_rejected`：ERP 侧 Workflow 审核驳回单据（docstatus=0，Pending review状态，标记为驳回）
- 认证方式：使用 API Key 或 Token 认证（与 ERP 侧约定签名算法）
- 实时性好，延迟低（< 5秒）

**TTPOS 侧 Webhook 接收接口**：
```
POST /api/v1/webhook/erp/stock_entry
```

**请求参数示例**：
```json
{
  "event": "stock_entry.cancelled",
  "timestamp": 1704441600,
  "signature": "sha256_signature_string",
  "data": {
    "name": "MAT-STE-2026-00001",
    "docstatus": 2,
    "company": "Company A",
    "branch": "Branch A",
    "warehouse": "WH-001",
    "posting_date": "2026-01-05",
    "posting_time": "10:30:00"
  }
}
```

**Webhook 处理流程**：
1. **验证请求**：
   - 验证签名（防止伪造请求）
   - 验证时间戳（防止重放攻击，时间差不超过5分钟）
   - 验证公司信息（确保是当前公司的单据）
2. **查询报废单**：
   - 根据 `data.name`（ERP 单据编号）查询 TTPOS 报废单
   - 如果找不到，记录日志并返回成功（避免 ERP 重复推送）
3. **处理不同事件**：
   
   **事件：stock_entry.cancelled（取消）**：
   - 判断：`docstatus == 2`（ERP 侧单据被取消）
   - 说明：ERP 侧取消已完成报废单后触发，TTPOS 侧被动接收并更新状态（TTPOS 侧无取消按钮，只能由 ERP 侧操作）
   - 操作：
     - **如果 TTPOS 侧报废单为已完成状态（status=1）**：
       1. 更新报废单状态为"已取消"（status=4）
       2. **回滚 TTPOS 侧库存**（增加库存，恢复之前扣减的数量）：
          - 增加 `ttpos_warehouse_item.stock`
          - 增加 `ttpos_material.stock_num`
       3. **记录出入库明细**（入库，Scene=报废取消入库，新增场景常量 `WarehouseInOutLogSceneScrapCancelIn = 8`）
       4. 更新 `sync_status=2`、`sync_time`
       5. 记录操作日志
       - **注意**：已完成报废单被ERP侧取消时，需要回滚已扣减的库存
     - **如果 TTPOS 侧报废单为草稿状态（status=0）**：
       1. 软删除报废单（`delete_time` 设置为当前时间）
       2. 更新 `sync_status=2`、`sync_time`
       3. 记录操作日志
       - **注意**：草稿状态未扣减库存，无需回滚库存
   
   **事件：stock_entry.submitted（提交）**：
   - 判断：`docstatus == 0` 且之前状态为草稿（进入Workflow审核流程）
   - 说明：ERP 侧单据提交后，进入 Workflow 审核流程，docstatus仍为0（Pending review状态）
   - 操作：
     1. 更新报废单状态为"已提交"（status=1）
     2. 更新 `sync_status=2`（同步成功）
     3. 更新 `sync_time`
     4. 记录操作日志
   
   **事件：stock_entry.workflow_rejected（Workflow 审核驳回）**：
   - 判断：Workflow 审核被驳回（docstatus=0，Pending review状态，标记为驳回）
   - 说明：ERP 侧 Workflow 审核驳回后触发（可能是 ERP 侧主动驳回，也可能是 TTPOS 侧驳回后 ERP 确认）
   - 操作：
     1. 更新报废单状态为"已驳回"（status=3）
     2. 更新 `sync_status=2`（同步成功，状态已同步）
     3. 记录驳回原因到 `reject_reason` 字段（从 Webhook 数据中获取，与ERP侧字段对齐）
     4. 记录操作日志
   - **注意**：
     - 审核驳回时，不扣减库存
     - 驳回后可以重新提交或修改报废单
     - 如果 TTPOS 侧已通过 API 驳回，此 Webhook 事件主要用于状态同步确认
   
   **事件：stock_entry.approved（审核通过）**：
   - 判断：`docstatus == 1`（ERP Workflow 审核通过，从docstatus=0变为docstatus=1）
   - 说明：此事件由 ERP 侧 Workflow 审核通过后触发，docstatus从0变为1
   - 操作：
     1. 更新报废单状态为"已审核通过"（status=2）
     2. 确认 ERP 侧已扣减库存（Workflow 审核通过时，docstatus从0变为1，自动扣减库存）
     3. 如果 TTPOS 侧还未扣减库存，则扣减 TTPOS 侧库存（防止重复扣减）
     4. 记录操作日志
   - **注意**：如果 TTPOS 侧已通过 API 审核通过并扣减库存，此 Webhook 事件主要用于状态同步确认
4. **返回响应**：
   - 成功：返回 `{"code": 1, "message": "success"}`
   - 失败：返回 `{"code": 0, "message": "error message"}`，ERP 侧会重试

**方案2：轮询兜底（备用方式）**

**执行时机**：定时任务，每 5 分钟执行一次

**轮询流程**：
1. **查询待同步的报废单**：
   - 查询状态为"已同步"（status=2）的报废单
   - 查询条件：`sync_status=2` 且 `erp_code != ''`
2. **批量查询 ERP 单据**：
   - 根据 `erp_code` 列表，批量查询 ERP Stock Entry 单据
   - 调用 ERP API：`GetStockEntryList`（需要 ERP 侧提供批量查询接口）
3. **对比状态**：
   - 对比 TTPOS 和 ERP 的 `docstatus` 状态
   - 如果 ERP 状态为取消（docstatus=2）但 TTPOS 状态为已完成（status=1），说明需要取消
   - 如果 ERP 状态为取消（docstatus=2）但 TTPOS 状态为已保存（status=0），说明需要删除
4. **执行同步**：
   - 调用与 Webhook 相同的处理逻辑
   - 如果是已完成报废单被取消，回滚库存、更新状态、记录明细
   - 如果是草稿报废单被取消，直接删除
5. **记录日志**：
   - 记录轮询结果和同步操作日志
   - 统计同步成功和失败的数量

**轮询兜底的优势**：
- 确保数据最终一致性
- 处理 Webhook 推送失败的情况
- 处理 ERP 侧 Webhook 配置错误的情况
- 作为 Webhook 的备份机制

**Webhook 和轮询的协调**：
- Webhook 处理优先，实时性高
- 轮询作为兜底，确保不遗漏
- 如果 Webhook 已处理，轮询会跳过（通过时间戳判断）
- 避免重复处理（通过状态判断）

#### 3.3 同步日志和审计

**同步日志记录**：
- 记录所有同步操作（成功和失败）
- 日志内容：
  - 报废单UUID和编号
  - 同步方向（TTPOS→ERP 或 ERP→TTPOS）
  - 同步时间
  - 同步结果（成功/失败）
  - 错误信息（如果失败）
  - 请求参数和响应内容（用于问题排查）

**审计追溯**：
- 支持查询报废单的完整同步历史
- 显示每次同步的详细信息
- 支持导出同步日志（用于审计）

### 4. 库存扣减设计

#### 4.1 TTPOS 侧库存扣减

**时机**：审核人员审核通过报废单时扣减

**扣减流程**：
1. **提交时**（不扣减库存）：
   - 仅校验库存是否充足（用于提示用户，不实际扣减）
   - 如果库存不足，提示用户并阻止提交
   - 如果库存充足，允许提交，状态更新为"待审核"（status=1）

2. **审核通过时**（扣减库存）：
   - 再次校验库存是否充足（防止审核期间库存被消耗）
   - 如果库存不足：
     - 审核失败，提示"审核期间库存不足，当前库存：[X]，需要：[Y]"
     - 报废单状态保持"待审核"（status=1）
     - 不允许审核通过
   - 如果库存充足：
     - 扣减 `ttpos_warehouse_item.stock`
     - 扣减 `ttpos_material.stock_num`
     - 记录出入库明细（见下方详细说明）
     - 调用 ERP API 同步报废单
     - 如果 ERP 同步失败，需要回滚库存扣减（增加库存）

**出入库明细记录**（`ttpos_warehouse_in_out_log`）：
- `log_type` = 1（出库，`WarehouseInOutLogLogTypeOut`）
- `scene` = 7（报废出库，新增场景常量 `WarehouseInOutLogSceneScrapOut`）
- `warehouse_uuid` = 报废单仓库
- `material_uuid` = 物品UUID
- `material_name` = 物品名称（JSON格式，备份多语言）
- `material_base_unit_uuid` = 物品基准单位UUID
- `material_base_unit_name` = 物品基准单位名称（JSON格式）
- `num` = 报废数量（基准单位数量）
- `price` = 单价（基准单位单价）
- `amount` = 金额（单价*数量）
- `order_no` = 报废单编号
- `create_time` = 当前时间戳（审核通过时间）

**同步失败时的库存回滚**：
- 如果 ERP 同步失败，需要回滚已扣减的库存：
  1. 增加 `ttpos_warehouse_item.stock`（恢复扣减的数量）
  2. 增加 `ttpos_material.stock_num`（恢复扣减的数量）
  3. 删除或标记已记录的出入库明细为无效（或记录一条回滚明细）
  4. 报废单状态保持"已提交待审核"（status=1），同步状态为"同步失败"（sync_status=3）
  5. 记录错误日志，支持后续重试

**新增场景常量定义**（需要在 `main/app/constant/warehouse.go` 中添加）：
```go
// WarehouseInOutLogScene 出入库日志场景
const (
    // ... 已有常量 ...
    WarehouseInOutLogSceneScrapOut = 7      // 报废出库（TTPOS→ERP 同步成功）
    WarehouseInOutLogSceneScrapCancelIn = 8 // 报废取消入库（已完成报废单取消后回滚库存）
)
```

#### 4.2 ERP 侧库存扣减

**时机**：TTPOS 审核通过后同步到 ERP，触发 ERP Workflow 审核通过，docstatus从0变为1，ERP 侧自动扣减库存

**逻辑**：
- TTPOS 审核通过后，调用 ERP API 触发 Workflow 审核通过
- ERP 侧 Workflow 审核通过后，单据状态从 docstatus=0（Pending review）变为 docstatus=1（已完成）
- ERP 侧 Workflow 审核通过时，ERP 系统自动扣减库存
- TTPOS 通过 Webhook 或轮询获取 ERP 审核状态和库存更新

#### 4.3 数据承载方案考虑

**当前方案**：TTPOS 和 ERP 两侧都扣减库存

**未来方案**：ERP 作为数据承载，TTPOS 仅作为操作入口

**调整建议**：
- 如果未来采用 ERP 作为数据承载，TTPOS 侧不直接扣减库存
- TTPOS 提交报废单后，等待 ERP 审核通过
- ERP 审核后通过 Webhook 通知 TTPOS，TTPOS 再扣减本地库存
- 出入库明细从 ERP 同步，或通过 Webhook 推送

### 5. 同步失败处理（补充说明）

> 注：详细的同步失败处理说明已在"核心功能点 - ERP 集成与同步机制"章节中说明，此处为补充说明。

#### 5.1 失败场景快速参考

| 失败场景 | 失败原因 | 重试策略 |
|---------|---------|---------|
| 网络超时 | ERP API 调用超时（30秒） | 支持手动重试，无次数限制 |
| ERP 系统错误 | ERP 返回 5xx 错误 | 支持手动重试，最多5次 |
| 数据校验失败 | 物品不存在、仓库不存在等 | 需先修改数据，然后重试 |
| 库存不足 | ERP 侧库存不足 | 需人工处理，不允许自动重试 |
| 重复提交 | 同一单据多次提交 | 检查并更新状态 |
| 权限不足 | API 权限不足或 Token 过期 | 需管理员处理，不允许重试 |

#### 5.2 重试限制说明

**重试次数限制**：
- `retry_count`：手动重试次数（用户点击重试按钮）
- `auto_retry_count`：自动重试次数（系统定时任务重试）

**不同失败类型的重试限制**：
- 网络超时：`retry_count` 无限制，`auto_retry_count` 最多3次
- ERP 系统错误：`retry_count` 最多5次，`auto_retry_count` 最多3次
- 数据校验失败：`retry_count` 无限制（需先修改数据），不允许自动重试
- 库存不足：不允许重试（需人工处理）
- 权限不足：不允许重试（需管理员处理）

#### 5.3 用户界面提示规范

**同步中状态（sync_status=1）**：
- 状态文本：显示"同步中..."
- 状态图标：加载动画（spinner）
- 操作限制：禁用编辑、删除、重试按钮
- 提示信息："正在同步到 ERP 系统，请稍候..."

**同步成功状态（sync_status=2）**：
- 状态文本：显示"同步成功"
- 状态图标：成功图标（绿色对勾）
- 显示信息：显示 ERP 单据编号（`erp_code`）
- 提示信息："已成功同步到 ERP，单据编号：[ERP编号]"

**同步失败状态（sync_status=3）**：
- 状态文本：显示"同步失败"
- 状态图标：失败图标（红色叉号）
- 显示信息：显示错误信息（`sync_error_message` 的前100个字符）
- 操作按钮：
  - 如果允许重试：显示"重试同步"按钮（根据失败类型判断）
  - 如果不允许重试：显示"联系管理员"按钮
- 提示信息：
  - 可重试："同步失败，请点击重试按钮重试（已重试 X 次）"
  - 不可重试："同步失败，请联系管理员处理：[错误信息]"

**详情页错误信息展示**：
- 完整显示 `sync_error_message`（无长度限制）
- 如果是数据校验失败，高亮显示错误的字段
- 提供错误原因分析和处理建议

- 持续时长：[已失败时长]
- 重试次数：手动 [retry_count] 次，自动 [auto_retry_count] 次
- 处理建议：[根据失败类型提供建议]
- 操作链接：[跳转到报废单详情页]
```

### 6. 出入库流水延迟处理

#### 6.1 延迟场景

**场景1：ERP 审核延迟**
- TTPOS 提交报废单后，ERP 侧需要人工审核
- 审核通过后才会扣减 ERP 库存
- 导致 TTPOS 和 ERP 库存不一致

**场景2：Webhook 推送延迟**
- ERP 审核后，Webhook 推送可能延迟
- 导致 TTPOS 侧状态更新延迟

**场景3：网络问题**
- 网络不稳定导致同步延迟

#### 6.2 处理方案

**方案1：轮询兜底**
- 定时任务（每5分钟）轮询 ERP，查询已审核的 Stock Entry
- 对比 TTPOS 数据，发现不一致时同步更新

**方案2：状态标记**
- 报废单增加 `erp_approved_time` 字段，记录 ERP 审核时间
- 出入库明细增加 `erp_synced` 字段，标记是否已同步到 ERP
- 通过对比时间戳，识别未同步的流水

**方案3：最终一致性**
- 接受短暂的数据不一致
- 通过定时同步任务，最终达到一致状态
- 在库存查询时，优先使用 ERP 数据（如果已同步）

### 7. 字段数据来源

#### 7.1 数据来源规则

**所有字段数据均从 ERP 获取**：

1. **仓库信息**
   - 来源：ERP Warehouse
   - TTPOS 侧：通过同步任务同步到 `ttpos_warehouse`
   - 字段：`erp_code`、`name`（多语言）

2. **物品信息**
   - 来源：ERP Item
   - TTPOS 侧：通过同步任务同步到 `ttpos_material`
   - 字段：`code`、`name`（多语言）、`base_unit`、`valuation`

3. **单位信息**
   - 来源：ERP UOM
   - TTPOS 侧：通过同步任务同步到 `ttpos_material_unit`
   - 字段：`name`（多语言）、`conversion_factor`

4. **单据编号**
   - 来源：ERP Series
   - TTPOS 侧：调用 ERP API 时，ERP 返回单据编号
   - 字段：`erp_code`

#### 7.2 数据同步时机

**仓库、物品、单位数据**：
- 通过现有的 ERP 同步任务同步
- 定时同步（每5分钟）或手动触发同步

**报废单数据**：
- 提交报废单时，实时调用 ERP API 获取单据编号
- ERP 侧操作通过 Webhook 或轮询同步到 TTPOS

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：
- 涉及 ERP 集成，需要调用 ERP API
- 需要处理同步失败、重试、数据一致性等复杂场景
- 需要新增数据库表、API 接口、业务逻辑

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 8-10 天
- **预估 SP**: 13-21 SP（待技术评审确认）

**任务分解**：
1. 数据库设计和迁移（1-2天）
2. 模型和 Repository 层（1-2天）
3. Service 层业务逻辑（2-3天）
4. API 接口开发（1-2天）
5. ERP 集成开发（2-3天）
6. 前端页面开发（2-3天）
7. 测试和联调（1-2天）

### 风险识别

**潜在风险**：

1. **ERP 同步失败率高**
   - 风险：网络不稳定、ERP 系统故障导致同步失败
   - 影响：报废单无法同步到 ERP，数据不一致
   - 缓解措施：
     - 实现重试机制（手动+自动）
     - 提供错误提示
     - 支持手动修复和重新同步

2. **库存扣减不一致**
   - 风险：TTPOS 和 ERP 库存扣减时机不同，导致数据不一致
   - 影响：库存数据不准确，影响盘点
   - 缓解措施：
     - 实现轮询兜底机制
     - 记录出入库明细，支持对账
     - 考虑未来采用 ERP 作为数据承载的方案

3. **ERP 取消同步延迟**
   - 风险：ERP 侧取消操作无法及时同步到 TTPOS
   - 影响：TTPOS 侧状态不准确，库存未回滚
   - 缓解措施：
     - 实现 Webhook 实时同步
     - 轮询兜底机制
     - 定时同步任务

4. **数据量过大**
   - 风险：报废单数据量大，影响查询性能
   - 影响：列表查询慢，用户体验差
   - 缓解措施：
     - 添加合适的索引
     - 分页查询
     - 考虑数据归档策略

5. **附件存储**
   - 风险：附件文件过大，存储成本高
   - 影响：存储空间不足，上传失败
   - 缓解措施：
     - 限制附件大小（如10MB）
     - 限制附件数量（如5个）
     - 使用对象存储服务（OSS）

---

## 🔗 相关资源

### 参考需求

- 类似功能: 盘点单功能（`docs/shared/specs/refactor-stock-reconciliation-erp-datasource/`）
- 参考文档: 盘点单架构调整提案（`docs/team/proposals/2025-12-18-stock-reconciliation-erpnext-refactor.md`）

### 相关文档

- ERP 集成文档: `docs/shared/integrations/`
- 库存管理架构: `docs/human/architecture/entities/warehouse.md`
- 同步机制设计: `docs/human/architecture/features/sync.md`

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

- [ ] 创建 Spec：`story-shop-scrap-management`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### 业务流程图

#### 报废单创建与审核流程

```
[用户：店长/仓库管理员]
    ↓
[创建报废单]
    ↓
选择报损仓库
    ↓
选择仓库中的物品
    ↓
填写报废数量
    ↓
上传附件（可选）
    ↓
[保存报废单] → 同步到ERP创建草稿单据 (docstatus=0)
    ↓
[提交报废单] → 同步到ERP提交审核 (docstatus=0, Pending review)
    ↓
[审核人员审核]
    ├─ [审核通过] → 同步到ERP触发Workflow审核通过
    │                  ↓
    │              ERP Workflow审核通过 (docstatus=1)
    │                  ↓
    │              扣减TTPOS库存
    │                  ↓
    │              扣减ERP库存
    │                  ↓
    │              记录出入库明细
    │                  ↓
    │              [完成]
    │
    └─ [驳回] → 同步到ERP触发Workflow驳回
                   ↓
               单据保持Pending review状态
                   ↓
               [可重新提交或修改]
```

#### ERP取消流程

```
[ERP侧取消已完成报废单]
    ↓
Webhook推送 → stock_entry.cancelled事件
    ↓
TTPOS接收Webhook
    ↓
验证签名和参数
    ↓
根据erp_code查询报废单
    ↓
更新报废单状态为"已取消" (status=4)
    ↓
回滚TTPOS库存
    ↓
回滚ERP库存（ERP侧自动处理）
    ↓
[完成]
```

### 数据流向图

#### TTPOS → ERP 数据流向

```
┌─────────────────────────────────────────────────┐
│          TTPOS 系统（报废管理模块）                │
│                                                  │
│  ┌──────────────────────────────────────────┐  │
│  │  报废单表 (ttpos_scrap_order)            │  │
│  │  - uuid                                  │  │
│  │  - erp_code (ERP单据编号)                │  │
│  │  - status (状态)                         │  │
│  │  - sync_status (同步状态)                │  │
│  └──────────────────────────────────────────┘  │
│                    │                            │
│                    │ gRPC调用                   │
│                    ↓                            │
│  ┌──────────────────────────────────────────┐  │
│  │      ttpos-bmp 服务                      │  │
│  │  (微服务网关，ERP集成层)                  │  │
│  └──────────────────────────────────────────┘  │
│                    │                            │
│                    │ HTTP API调用               │
│                    ↓                            │
└────────────────────┼────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────┐
│         ERPNext 系统 (ERP集成)                   │
│                                                  │
│  ┌──────────────────────────────────────────┐  │
│  │  Stock Entry 单据                        │  │
│  │  - name (单据编号)                       │  │
│  │  - stock_entry_type = "Material Issue"   │  │
│  │  - docstatus (0=Draft, 1=Submitted)      │  │
│  │  - workflow_state (审核状态)             │  │
│  └──────────────────────────────────────────┘  │
│                    │                            │
│                    ↓                            │
│  ┌──────────────────────────────────────────┐  │
│  │  Bin (库存记录)                          │  │
│  │  - 库存扣减                              │  │
│  └──────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

#### ERP → TTPOS 数据流向

```
┌─────────────────────────────────────────────────┐
│         ERPNext 系统 (ERP集成)                   │
│                                                  │
│  ┌──────────────────────────────────────────┐  │
│  │  Stock Entry 单据状态变更                 │  │
│  │  - stock_entry.cancelled                 │  │
│  │  - stock_entry.reverted                  │  │
│  │  - stock_entry.submitted                 │  │
│  │  - stock_entry.approved                  │  │
│  └──────────────────────────────────────────┘  │
│                    │                            │
│                    │ Webhook推送                │
│                    ↓                            │
└────────────────────┼────────────────────────────┘
                     │
                     ↓
┌─────────────────────────────────────────────────┐
│          TTPOS 系统（报废管理模块）                │
│                                                  │
│  ┌──────────────────────────────────────────┐  │
│  │  Webhook接收接口                         │  │
│  │  POST /api/v1/webhook/erp/stock_entry   │  │
│  └──────────────────────────────────────────┘  │
│                    │                            │
│                    │ 验证签名、查询报废单        │
│                    ↓                            │
│  ┌──────────────────────────────────────────┐  │
│  │  更新报废单状态                          │  │
│  │  - 状态同步                              │  │
│  │  - 库存回滚（如取消）                    │  │
│  └──────────────────────────────────────────┘  │
│                    │                            │
│                    ↓                            │
│  ┌──────────────────────────────────────────┐  │
│  │  TTPOS库存表                             │  │
│  │  - ttpos_warehouse_item.stock            │  │
│  │  - ttpos_material.stock_num              │  │
│  └──────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

#### 完整数据流向（包含轮询兜底）

```
┌──────────────────────────────────────────────────────────────┐
│                    数据同步机制                                │
│                                                              │
│  1. 实时同步（主要方式）                                      │
│     TTPOS操作 → gRPC → ttpos-bmp → ERP API → ERP系统         │
│                                                              │
│  2. ERP变更推送（主要方式）                                   │
│     ERP状态变更 → Webhook推送 → TTPOS Webhook接口 → 更新状态  │
│                                                              │
│  3. 轮询兜底（备用方式）                                      │
│     定时任务（5分钟）→ 查询ERP最近更新 → 对比状态 → 同步更新  │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### User Story（初稿）

**作为** 店长/仓库管理员  
**我想** 通过系统提交报废单，处理损坏、过期、变质的物品  
**以便于** 规范报废流程，准确扣减库存，与 ERP 系统数据保持一致

### AC 验收标准（初稿）

1. **WHEN** 用户选择仓库和物品，填写报废数量 **THEN** 系统 **SHALL** 允许创建报废单
2. **WHEN** 用户上传附件 **THEN** 系统 **SHALL** 保存附件URL到报废单
3. **WHEN** 用户提交报废单 **THEN** 系统 **SHALL** 扣减 TTPOS 库存，记录出入库明细，同步到 ERP
4. **IF** ERP 同步成功 **THEN** 系统 **SHALL** 更新报废单状态为"已同步"，显示 ERP 单据编号
5. **IF** ERP 同步失败 **THEN** 系统 **SHALL** 记录错误信息，允许用户重试同步
6. **WHEN** ERP 侧取消已完成报废单 **THEN** 系统 **SHALL** 同步更新 TTPOS 状态为已取消，回滚库存
7. **WHEN** 用户查询报废单列表 **THEN** 系统 **SHALL** 显示报废单基本信息、状态、同步状态
8. **WHEN** 用户查看报废单详情 **THEN** 系统 **SHALL** 显示完整的报废单信息、明细、附件

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
**创建日期**: 2026-01-05  
**维护者**: TTPOS Team  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`, `.cursor/rules/api.mdc`, `.cursor/rules/database.mdc`

