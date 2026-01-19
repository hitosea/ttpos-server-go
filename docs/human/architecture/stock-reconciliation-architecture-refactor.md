# 盘点单架构调整方案：ERPNext 作为数据承载，TTPOS 作为前端接口

> 将盘点单的数据承载从 TTPOS 调整为 ERPNext，实现双向实时同步

---

## 一、架构调整目标

### 1.1 调整前架构（当前）

```
TTPOS（主数据源）
    ↓ 单向同步（提交/审核时）
ERPNext（数据接收方）
```

**特点**：
- TTPOS 存储完整的盘点单数据
- TTPOS 操作时同步到 ERPNext
- ERPNext 操作不会回写到 TTPOS
- 数据不一致风险：TTPOS 和 ERPNext 可能不同步

### 1.2 调整后架构（目标）

```
ERPNext（主数据源）
    ↕ 双向实时同步
TTPOS（前端操作接口 + 数据缓存）
```

**特点**：
- ERPNext 存储完整的盘点单数据（主数据源）
- TTPOS 作为前端操作接口，数据从 ERPNext 同步
- ERPNext 操作时实时同步到 TTPOS
- TTPOS 操作时同步到 ERPNext
- 数据一致性：TTPOS 数据始终与 ERPNext 保持一致

---

## 二、核心调整点

### 2.1 数据存储调整

| 调整项 | 调整前 | 调整后 |
|--------|--------|--------|
| **主数据源** | TTPOS 数据库 | ERPNext 数据库 |
| **TTPOS 数据库角色** | 主存储 | 缓存/镜像 |
| **数据同步方向** | TTPOS → ERPNext（单向） | ERPNext ↔ TTPOS（双向） |
| **数据一致性** | 最终一致性 | 实时一致性 |

### 2.2 操作流程调整

| 操作 | 调整前 | 调整后 |
|------|--------|--------|
| **创建盘点单** | TTPOS 创建 → 同步到 ERPNext | TTPOS 创建 → ERPNext 创建 → 同步回 TTPOS |
| **编辑盘点单** | TTPOS 编辑 → 同步到 ERPNext | TTPOS 编辑 → ERPNext 更新 → 同步回 TTPOS |
| **提交盘点单** | TTPOS 提交 → ERPNext 创建 | TTPOS 提交 → ERPNext 创建 → 同步回 TTPOS |
| **审核盘点单** | TTPOS 审核 → ERPNext 提交 | TTPOS 审核 → ERPNext 提交 → 同步回 TTPOS |
| **ERPNext 操作** | 不同步 | ERPNext 操作 → 实时同步到 TTPOS |

---

## 三、用户映射与操作记录

### 3.1 问题说明

TTPOS 和 ERPNext 使用不同的用户体系：
- **TTPOS**：使用 `Staff`（员工）表，标识为 `uuid`、`username`、`real_name`
- **ERPNext**：使用 `User`（用户）表，标识为 `email`（用户名）、`full_name`

需要在两个系统中都能正确记录和显示操作人员。

### 3.2 解决方案

**核心方案**：
1. **用户映射表**：建立 TTPOS 员工与 ERPNext 用户的映射关系
2. **双重记录**：在盘点单表中同时记录 TTPOS 和 ERPNext 用户信息
3. **操作日志**：详细记录每次操作的用户信息和来源

**详细方案请参考**：`docs/human/architecture/stock-reconciliation-user-mapping.md`

---

## 四、详细调整方案

### 3.1 数据模型调整

#### 3.1.1 TTPOS 数据库表结构调整

**新增字段**：

```sql
ALTER TABLE `ttpos_stock_reconciliation` 
ADD COLUMN `sync_status` tinyint NOT NULL DEFAULT 0 COMMENT '同步状态 0-未同步 1-同步中 2-已同步 3-同步失败' AFTER `status`,
ADD COLUMN `sync_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '最后同步时间(时间戳)' AFTER `sync_status`,
ADD COLUMN `sync_error` text COMMENT '同步错误信息' AFTER `sync_time`,
ADD COLUMN `is_from_erp` tinyint NOT NULL DEFAULT 0 COMMENT '是否来自ERP 0-否 1-是' AFTER `sync_error`,
ADD INDEX `idx_erp_code` (`erp_code`),
ADD INDEX `idx_sync_status` (`sync_status`);
```

**字段说明**：
- `sync_status`：同步状态，用于追踪同步进度
- `sync_time`：最后同步时间，用于判断是否需要重新同步
- `sync_error`：同步错误信息，用于问题排查
- `is_from_erp`：标识数据来源，区分是 TTPOS 创建还是 ERPNext 同步

#### 3.1.2 数据存储策略

**TTPOS 数据库角色**：
- **缓存层**：存储 ERPNext 数据的本地缓存
- **操作层**：支持前端快速查询和操作
- **同步层**：与 ERPNext 保持实时同步

**数据生命周期**：
- TTPOS 创建的数据：先创建本地记录，然后同步到 ERPNext，ERPNext 返回后更新本地记录
- ERPNext 创建的数据：通过同步机制拉取到 TTPOS，创建本地记录

---

### 3.2 同步机制设计

#### 3.2.1 同步方式选择

**方案对比**：

| 方案 | 优点 | 缺点 | 推荐度 |
|------|------|------|--------|
| **Webhook（推送）** | 实时性好，延迟低 | 需要 ERPNext 支持 Webhook，网络稳定性要求高 | ⭐⭐⭐⭐⭐ |
| **消息队列（MQ）** | 可靠性高，支持重试 | 需要额外的消息队列基础设施 | ⭐⭐⭐⭐ |
| **轮询（Polling）** | 实现简单，不依赖外部 | 实时性差，资源消耗大 | ⭐⭐ |
| **混合方案** | 结合多种方案优势 | 实现复杂 | ⭐⭐⭐⭐ |

**推荐方案：Webhook + 轮询兜底**

- **主要方式**：ERPNext Webhook 推送变更到 TTPOS
- **兜底方式**：定时轮询 ERPNext，确保数据一致性
- **容错机制**：同步失败时重试，超过阈值后告警

#### 3.2.2 Webhook 实现方案

**ERPNext 端配置**：
- 在 ERPNext 中配置 Webhook，监听盘点单的创建、更新、提交、审核事件
- Webhook URL：`https://ttpos-api.example.com/api/v1/webhook/erp/stock_reconciliation`
- 认证方式：使用 API Key 或 Token 认证

**TTPOS 端接收**：
```go
// 新增 Webhook 接收接口
POST /api/v1/webhook/erp/stock_reconciliation
{
  "event": "stock_reconciliation.created|updated|submitted|approved",
  "data": {
    "name": "MAT-RECO-2025-00001",
    "company": "Company A",
    "warehouse": "WH-001",
    "status": "Draft|Submitted",
    "items": [...]
  }
}
```

#### 3.2.3 轮询兜底方案

**定时任务**：
- 每 5 分钟轮询一次 ERPNext，获取最近更新的盘点单
- 对比 TTPOS 和 ERPNext 的数据，发现不一致时同步
- 同步失败时记录日志，超过阈值后告警

**轮询逻辑**：
```go
// 定时任务：同步盘点单数据
func SyncStockReconciliationFromERP() {
    // 1. 查询 ERPNext 最近更新的盘点单（最近 10 分钟）
    // 2. 对比 TTPOS 数据
    // 3. 发现不一致时同步
    // 4. 记录同步日志
}
```

---

### 3.3 业务流程调整

#### 3.3.1 创建盘点单流程

**调整前**：
```
用户操作 → TTPOS 创建 → 保存到 TTPOS 数据库 → 提交时同步到 ERPNext
```

**调整后**：
```
用户操作 → TTPOS 创建 → 调用 ERPNext API 创建 → ERPNext 返回盘点单号 → 
保存到 TTPOS 数据库（标记 is_from_erp=0） → Webhook 推送 → 更新 TTPOS 数据
```

**关键代码调整**：
```go
// 调整前：先保存到 TTPOS，再同步到 ERPNext
func SaveStockReconciliation(ctx context.Context, req req.StockReconciliationSaveReq) {
    // 1. 保存到 TTPOS 数据库
    // 2. 提交时同步到 ERPNext
}

// 调整后：先创建到 ERPNext，再保存到 TTPOS
func SaveStockReconciliation(ctx context.Context, req req.StockReconciliationSaveReq) {
    // 1. 调用 ERPNext API 创建盘点单
    erpResp, err := erpSrv.CreateStockReconciliation(ctx, req)
    if err != nil {
        return err
    }
    
    // 2. 保存到 TTPOS 数据库（使用 ERPNext 返回的数据）
    stockReconciliation := &model.StockReconciliation{
        ErpCode: erpResp.StockReconciliationName,
        OrderNo: generateOrderNo(), // 生成 TTPOS 单据编号
        IsFromErp: false,
        SyncStatus: constant.SyncStatusSyncing,
    }
    // ... 保存逻辑
    
    // 3. 等待 Webhook 推送或主动拉取数据
    // 4. 更新 TTPOS 数据
}
```

#### 3.3.2 编辑盘点单流程

**调整前**：
```
用户操作 → TTPOS 更新 → 保存到 TTPOS 数据库 → 提交时同步到 ERPNext
```

**调整后**：
```
用户操作 → TTPOS 更新 → 调用 ERPNext API 更新 → ERPNext 返回 → 
更新 TTPOS 数据库 → Webhook 推送 → 更新 TTPOS 数据
```

#### 3.3.3 提交盘点单流程

**调整前**：
```
用户操作 → TTPOS 提交 → 调用 ERPNext API 创建 → 更新 TTPOS 状态
```

**调整后**：
```
用户操作 → TTPOS 提交 → 调用 ERPNext API 创建（如果未创建）或提交 → 
ERPNext 返回 → 更新 TTPOS 状态 → Webhook 推送 → 更新 TTPOS 数据
```

#### 3.3.4 审核盘点单流程

**调整前**：
```
用户操作 → TTPOS 审核 → 更新 TTPOS 库存 → 调用 ERPNext API 提交 → 更新 TTPOS 状态
```

**调整后**：
```
用户操作 → TTPOS 审核 → 调用 ERPNext API 提交 → ERPNext 更新库存 → 
Webhook 推送 → 更新 TTPOS 库存和状态
```

**关键调整**：
- **库存更新**：不再在 TTPOS 端更新库存，而是从 ERPNext 同步
- **审核逻辑**：审核操作直接调用 ERPNext API，等待 ERPNext 处理完成后同步结果

#### 3.3.5 ERPNext 操作同步流程

**新增流程**：
```
ERPNext 用户操作 → ERPNext 更新数据 → Webhook 推送 → TTPOS 接收 → 
更新 TTPOS 数据库 → 通知前端（WebSocket/SSE）
```

**关键代码**：
```go
// Webhook 接收处理
func HandleERPWebhook(c *gin.Context) {
    var webhookReq ERPWebhookReq
    if err := c.ShouldBindJSON(&webhookReq); err != nil {
        return err
    }
    
    // 根据事件类型处理
    switch webhookReq.Event {
    case "stock_reconciliation.created":
        syncStockReconciliationFromERP(webhookReq.Data)
    case "stock_reconciliation.updated":
        syncStockReconciliationFromERP(webhookReq.Data)
    case "stock_reconciliation.submitted":
        syncStockReconciliationFromERP(webhookReq.Data)
    case "stock_reconciliation.approved":
        syncStockReconciliationFromERP(webhookReq.Data)
        // 同步库存数据
        syncStockFromERP(webhookReq.Data)
    }
}

// 同步盘点单数据
func syncStockReconciliationFromERP(erpData ERPStockReconciliationData) {
    // 1. 查询 TTPOS 是否存在（通过 erp_code）
    // 2. 如果不存在，创建新记录（标记 is_from_erp=1）
    // 3. 如果存在，更新记录
    // 4. 同步明细数据
    // 5. 更新同步状态和时间
}
```

---

### 3.4 代码调整清单

#### 3.4.1 TTPOS Main 模块调整

**文件清单**：

1. **数据模型调整**
   - `main/app/model/stock_reconciliation.go`
     - 新增字段：`SyncStatus`, `SyncTime`, `SyncError`, `IsFromErp`
     - 调整字段：`ErpCode` 改为必填（作为主键关联）

2. **服务层调整**
   - `main/app/service/stock_reconciliation.go`
     - `SaveStockReconciliation()`：先调用 ERPNext API，再保存到 TTPOS
     - `submitStockReconciliation()`：调整提交逻辑，等待 ERPNext 返回
     - `ApproveStockReconciliation()`：调整审核逻辑，库存更新从 ERPNext 同步
     - 新增：`SyncStockReconciliationFromERP()`：从 ERPNext 同步数据
     - 新增：`HandleERPWebhook()`：处理 ERPNext Webhook

3. **ERP 服务调整**
   - `main/app/service/rpc/erp/stock.go`
     - `SubmitStockReconciliation()`：调整为创建盘点单（而非提交）
     - `ApproveStockReconciliation()`：保持提交逻辑
     - 新增：`CreateStockReconciliation()`：创建盘点单
     - 新增：`UpdateStockReconciliation()`：更新盘点单
     - 新增：`GetStockReconciliationList()`：获取盘点单列表（用于轮询）

4. **API 接口调整**
   - `main/app/api/v1/shop/shop_stock_reconciliation.go`
     - 新增：`POST /api/v1/webhook/erp/stock_reconciliation`：Webhook 接收接口
     - 调整：所有操作接口，增加同步状态返回

5. **定时任务调整**
   - `main/app/job/stock_reconciliation_sync.go`（新增）
     - 定时轮询 ERPNext，同步盘点单数据
     - 处理同步失败的重试逻辑

#### 3.4.2 BMP 模块调整

**文件清单**：

1. **Protobuf 定义调整**
   - `ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto`
     - 新增：`CreateStockReconciliation` RPC（创建盘点单）
     - 新增：`UpdateStockReconciliation` RPC（更新盘点单）
     - 新增：`GetStockReconciliationList` RPC（获取盘点单列表，支持时间范围）

2. **逻辑层调整**
   - `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go`
     - `SaveStockReconciliation()`：保持现有逻辑（创建盘点单）
     - 新增：`UpdateStockReconciliation()`：更新盘点单
     - 新增：`GetStockReconciliationList()`：获取盘点单列表（支持时间范围查询）

3. **控制器调整**
   - `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go`
     - 新增：`CreateStockReconciliation` 控制器
     - 新增：`UpdateStockReconciliation` 控制器
     - 新增：`GetStockReconciliationList` 控制器

---

### 3.5 数据同步策略

#### 3.5.1 同步时机

| 场景 | 同步方式 | 延迟 |
|------|----------|------|
| **TTPOS 创建盘点单** | 立即调用 ERPNext API | < 1秒 |
| **TTPOS 更新盘点单** | 立即调用 ERPNext API | < 1秒 |
| **ERPNext 创建/更新** | Webhook 推送 | < 5秒 |
| **ERPNext 操作失败** | 轮询兜底 | 5分钟 |
| **数据不一致检测** | 定时轮询 | 5分钟 |

#### 3.5.2 冲突处理

**场景 1：TTPOS 和 ERPNext 同时修改**

**处理策略**：
- ERPNext 数据优先（ERPNext 是主数据源）
- TTPOS 修改时检查版本号或时间戳
- 如果 ERPNext 数据更新，拒绝 TTPOS 修改，提示用户刷新

**实现**：
```go
func UpdateStockReconciliation(ctx context.Context, req req.StockReconciliationSaveReq) error {
    // 1. 查询 TTPOS 数据
    localData := getStockReconciliation(req.Uuid)
    
    // 2. 查询 ERPNext 数据
    erpData := getStockReconciliationFromERP(localData.ErpCode)
    
    // 3. 对比更新时间
    if erpData.UpdateTime > localData.SyncTime {
        // ERPNext 数据更新，拒绝修改
        return errors.New("数据已被更新，请刷新后重试")
    }
    
    // 4. 执行更新
    // ...
}
```

**场景 2：同步失败**

**处理策略**：
- 记录同步错误
- 标记同步状态为失败
- 定时重试（指数退避）
- 超过重试次数后告警

**实现**：
```go
func syncWithRetry(data *model.StockReconciliation, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := syncToERP(data)
        if err == nil {
            return nil
        }
        
        // 指数退避
        waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
        time.Sleep(waitTime)
    }
    
    // 超过重试次数，告警
    alertSyncFailure(data)
    return errors.New("同步失败，已超过最大重试次数")
}
```

---

### 3.6 库存更新调整

#### 3.6.1 调整前

- TTPOS 审核盘点单时，直接更新 TTPOS 库存
- 然后同步到 ERPNext，ERPNext 再更新库存

#### 3.6.2 调整后

- TTPOS 审核盘点单时，调用 ERPNext API 提交盘点单
- ERPNext 审核后更新 ERPNext 库存
- TTPOS 通过 Webhook 或轮询获取 ERPNext 库存更新
- TTPOS 同步更新本地库存

**关键代码**：
```go
func ApproveStockReconciliation(ctx context.Context, req req.StockReconciliationApproveReq) error {
    // 1. 调用 ERPNext API 提交盘点单
    err := erpSrv.ApproveStockReconciliation(ctx, req)
    if err != nil {
        return err
    }
    
    // 2. 更新 TTPOS 状态为"已审核"（等待库存同步）
    updateStatus(req.Uuid, constant.StockReconciliationStatusApproved)
    
    // 3. 等待 Webhook 推送库存更新
    // 或者主动拉取库存数据
    go syncStockFromERP(req.Uuid)
    
    return nil
}

// 同步库存数据
func syncStockFromERP(stockReconciliationUuid uint64) {
    // 1. 查询 ERPNext 盘点单
    erpData := getStockReconciliationFromERP(erpCode)
    
    // 2. 更新 TTPOS 库存
    for _, item := range erpData.Items {
        updateWarehouseItemStock(item.MaterialCode, item.CountedQuantity)
    }
    
    // 3. 生成盘盈盘亏记录
    generateProfitLossLog(stockReconciliationUuid, erpData.Items)
}
```

---

### 3.7 前端调整

#### 3.7.1 数据展示调整

**新增字段展示**：
- 同步状态：显示"同步中"、"已同步"、"同步失败"
- 数据来源：显示"TTPOS 创建"或"ERPNext 同步"
- 最后同步时间：显示最后同步时间

**状态提示**：
- 同步中：显示加载动画
- 同步失败：显示错误提示，提供重试按钮

#### 3.7.2 操作调整

**创建/编辑操作**：
- 增加同步状态提示
- 同步失败时提示用户重试
- 数据冲突时提示用户刷新

**实时更新**：
- 使用 WebSocket 或 SSE 接收 ERPNext 数据变更
- 自动刷新列表和详情页

---

## 五、实施步骤

### 4.1 第一阶段：基础设施准备（1-2周）

1. **数据库迁移**
   - 添加同步相关字段
   - 创建索引
   - 数据迁移脚本

2. **Webhook 基础设施**
   - 配置 ERPNext Webhook
   - 实现 Webhook 接收接口
   - 实现认证和验证

3. **轮询任务框架**
   - 实现定时任务框架
   - 实现重试机制
   - 实现告警机制

### 4.2 第二阶段：核心功能调整（2-3周）

1. **数据模型调整**
   - 更新 Model 定义
   - 更新 Repository 层
   - 更新 Service 层

2. **业务流程调整**
   - 调整创建流程
   - 调整编辑流程
   - 调整提交流程
   - 调整审核流程

3. **ERP 服务调整**
   - 新增创建/更新接口
   - 调整现有接口
   - 实现数据同步逻辑

### 4.3 第三阶段：同步机制实现（2-3周）

1. **Webhook 实现**
   - 实现 Webhook 接收
   - 实现数据同步逻辑
   - 实现错误处理

2. **轮询实现**
   - 实现定时轮询
   - 实现数据对比
   - 实现同步逻辑

3. **冲突处理**
   - 实现版本控制
   - 实现冲突检测
   - 实现冲突解决

### 4.4 第四阶段：测试与优化（2-3周）

1. **功能测试**
   - 单元测试
   - 集成测试
   - 端到端测试

2. **性能测试**
   - 同步性能测试
   - 并发测试
   - 压力测试

3. **稳定性测试**
   - 网络异常测试
   - ERPNext 异常测试
   - 数据一致性测试

### 4.5 第五阶段：上线与监控（1周）

1. **灰度发布**
   - 选择部分商户灰度
   - 监控数据同步情况
   - 收集反馈

2. **全量上线**
   - 全量发布
   - 持续监控
   - 问题处理

3. **监控告警**
   - 同步状态监控
   - 错误率监控
   - 性能监控

---

## 六、风险评估与应对

### 5.1 技术风险

| 风险 | 影响 | 应对措施 |
|------|------|----------|
| **ERPNext API 不稳定** | 高 | 实现重试机制、降级方案、告警 |
| **Webhook 丢失** | 中 | 轮询兜底、消息队列备份 |
| **数据不一致** | 高 | 定时对比、自动修复、告警 |
| **性能问题** | 中 | 异步处理、批量同步、缓存优化 |

### 5.2 业务风险

| 风险 | 影响 | 应对措施 |
|------|------|----------|
| **用户体验下降** | 中 | 优化同步速度、增加状态提示 |
| **数据丢失** | 高 | 数据备份、事务保护、回滚机制 |
| **功能缺失** | 低 | 功能对比、逐步迁移、兼容处理 |

### 5.3 应对策略

1. **降级方案**：ERPNext 不可用时，允许 TTPOS 本地操作，后续同步
2. **回滚方案**：保留原有代码，支持快速回滚
3. **监控告警**：实时监控同步状态，及时发现问题
4. **数据备份**：定期备份数据，支持数据恢复

---

## 七、迁移方案

### 6.1 数据迁移

**现有数据迁移**：
1. 查询所有已创建的盘点单（`is_open_erp = true`）
2. 检查是否有 `erp_code`，如果没有，调用 ERPNext API 创建
3. 同步 ERPNext 数据到 TTPOS
4. 更新同步状态和时间

**迁移脚本**：
```sql
-- 1. 添加新字段
ALTER TABLE `ttpos_stock_reconciliation` 
ADD COLUMN `sync_status` tinyint NOT NULL DEFAULT 0,
ADD COLUMN `sync_time` int(10) unsigned NOT NULL DEFAULT 0,
ADD COLUMN `sync_error` text,
ADD COLUMN `is_from_erp` tinyint NOT NULL DEFAULT 0;

-- 2. 迁移现有数据
UPDATE `ttpos_stock_reconciliation` 
SET `sync_status` = 2, `sync_time` = UNIX_TIMESTAMP() 
WHERE `erp_code` != '';

-- 3. 创建索引
CREATE INDEX `idx_erp_code` ON `ttpos_stock_reconciliation` (`erp_code`);
CREATE INDEX `idx_sync_status` ON `ttpos_stock_reconciliation` (`sync_status`);
```

### 6.2 功能迁移

**兼容处理**：
- 保留原有接口，新增新接口
- 通过配置开关控制使用新接口还是旧接口
- 逐步迁移，降低风险

**配置开关**：
```go
// 配置项
const UseERPAsDataSource = true // 是否使用 ERPNext 作为数据源

func SaveStockReconciliation(ctx context.Context, req req.StockReconciliationSaveReq) {
    if UseERPAsDataSource {
        // 新逻辑：先创建到 ERPNext
        return saveStockReconciliationNew(ctx, req)
    } else {
        // 旧逻辑：先保存到 TTPOS
        return saveStockReconciliationOld(ctx, req)
    }
}
```

---

## 八、监控指标

### 7.1 同步指标

- **同步成功率**：成功同步次数 / 总同步次数
- **同步延迟**：从操作到同步完成的时间
- **同步错误率**：同步失败次数 / 总同步次数
- **数据一致性**：TTPOS 和 ERPNext 数据一致的盘点单比例

### 7.2 性能指标

- **API 响应时间**：ERPNext API 调用耗时
- **Webhook 处理时间**：Webhook 接收和处理耗时
- **轮询处理时间**：定时轮询处理耗时

### 7.3 业务指标

- **盘点单创建成功率**：成功创建的盘点单数 / 总创建请求数
- **盘点单审核成功率**：成功审核的盘点单数 / 总审核请求数
- **用户操作延迟**：用户操作到界面更新的时间

---

## 九、24小时营业店面特殊处理

### 9.1 问题说明

24小时营业的店面在盘点时，可能存在未结算的订单，这些订单可能已经预出库（库存已扣减），需要在盘点时正确处理。

### 9.2 解决方案

**核心方案**：账面库存包含未结算订单的预出库数量

**计算公式**：
```
账面库存 = 当前库存（warehouse_item.stock）+ 未结算订单预出库数量
```

**详细方案请参考**：`docs/human/business/stock-reconciliation-24hour-handling.md`

### 9.3 实现要点

1. **查询未结算订单预出库**：关联查询出库单和订单表
2. **性能优化**：添加索引优化查询性能
3. **数据一致性**：账面库存在保存时计算，保存后不再变化

---

## 十、总结

### 8.1 核心调整

1. **数据源切换**：ERPNext 作为主数据源，TTPOS 作为缓存
2. **双向同步**：实现 TTPOS ↔ ERPNext 双向实时同步
3. **库存更新**：库存更新从 ERPNext 同步，不再在 TTPOS 端直接更新
4. **冲突处理**：ERPNext 数据优先，实现版本控制

### 8.2 实施要点

1. **分阶段实施**：基础设施 → 核心功能 → 同步机制 → 测试优化 → 上线
2. **风险控制**：降级方案、回滚方案、监控告警
3. **数据迁移**：现有数据迁移、功能兼容处理
4. **持续监控**：同步指标、性能指标、业务指标

### 8.3 预期收益

1. **数据一致性**：TTPOS 和 ERPNext 数据完全一致
2. **操作灵活性**：支持在 ERPNext 和 TTPOS 两端操作
3. **系统解耦**：TTPOS 作为前端接口，ERPNext 作为数据承载
4. **可扩展性**：支持更多前端应用接入

---

**文档版本**：v1.0  
**创建时间**：2025-01-16  
**维护者**：TTPOS Team

