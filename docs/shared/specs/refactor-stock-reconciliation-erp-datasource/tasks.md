# 盘点单架构调整任务清单

> 将盘点单数据承载从 TTPOS 调整为 ERPNext，实现双向实时同步

---

## 一、数据库调整

### 1.1 数据表结构调整

- [ ] **添加同步相关字段**
  - `sync_status`：同步状态（0-未同步，1-同步中，2-已同步，3-同步失败）
  - `sync_time`：最后同步时间（时间戳）
  - `sync_error`：同步错误信息（TEXT）
  - `is_from_erp`：是否来自 ERP（0-否，1-是）

- [ ] **添加索引**
  - `idx_erp_code`：ERP 盘点单号索引
  - `idx_sync_status`：同步状态索引

- [ ] **创建迁移脚本**
  - 文件：`admin/database/migrations/YYYYMMDD_add_stock_reconciliation_sync_fields.sql`
  - 包含：字段添加、索引创建、数据迁移

### 1.2 数据模型调整

- [ ] **更新 Model 定义**
  - 文件：`main/app/model/stock_reconciliation.go`
  - 添加新字段到 `StockReconciliation` 结构体

- [ ] **更新常量定义**
  - 文件：`main/app/constant/stock_reconciliation.go`
  - 添加同步状态常量

---

## 二、BMP 模块调整

### 2.1 Protobuf 定义

- [ ] **新增 RPC 接口**
  - 文件：`ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto`
  - `CreateStockReconciliation`：创建盘点单
  - `UpdateStockReconciliation`：更新盘点单
  - `GetStockReconciliationList`：获取盘点单列表（支持时间范围）

- [ ] **生成 Protobuf 代码**
  - 运行 `make proto` 生成 Go 代码

### 2.2 逻辑层实现

- [ ] **实现创建盘点单逻辑**
  - 文件：`ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go`
  - 方法：`CreateStockReconciliation()`
  - 功能：创建盘点单到 ERPNext

- [ ] **实现更新盘点单逻辑**
  - 文件：`ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go`
  - 方法：`UpdateStockReconciliation()`
  - 功能：更新 ERPNext 盘点单

- [ ] **实现获取盘点单列表逻辑**
  - 文件：`ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_reconciliation.go`
  - 方法：`GetStockReconciliationList()`
  - 功能：支持时间范围查询，用于轮询同步

### 2.3 控制器实现

- [ ] **实现创建盘点单控制器**
  - 文件：`ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go`
  - 方法：`CreateStockReconciliation()`

- [ ] **实现更新盘点单控制器**
  - 文件：`ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go`
  - 方法：`UpdateStockReconciliation()`

- [ ] **实现获取盘点单列表控制器**
  - 文件：`ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go`
  - 方法：`GetStockReconciliationList()`

---

## 三、TTPOS Main 模块调整

### 3.1 ERP 服务层

- [ ] **新增创建盘点单接口**
  - 文件：`main/app/service/rpc/erp/stock.go`
  - 方法：`CreateStockReconciliation()`
  - 功能：调用 BMP 模块创建盘点单

- [ ] **新增更新盘点单接口**
  - 文件：`main/app/service/rpc/erp/stock.go`
  - 方法：`UpdateStockReconciliation()`
  - 功能：调用 BMP 模块更新盘点单

- [ ] **新增获取盘点单列表接口**
  - 文件：`main/app/service/rpc/erp/stock.go`
  - 方法：`GetStockReconciliationList()`
  - 功能：调用 BMP 模块获取盘点单列表（用于轮询）

- [ ] **调整提交盘点单接口**
  - 文件：`main/app/service/rpc/erp/stock.go`
  - 方法：`SubmitStockReconciliation()`
  - 功能：如果盘点单未创建，先创建；如果已创建，直接提交

### 3.2 盘点单服务层

- [ ] **调整创建盘点单逻辑**
  - 文件：`main/app/service/stock_reconciliation.go`
  - 方法：`SaveStockReconciliation()`
  - 调整：先调用 ERPNext API 创建，再保存到 TTPOS

- [ ] **调整提交盘点单逻辑**
  - 文件：`main/app/service/stock_reconciliation.go`
  - 方法：`submitStockReconciliation()`
  - 调整：如果盘点单未创建，先创建；如果已创建，直接提交

- [ ] **调整审核盘点单逻辑**
  - 文件：`main/app/service/stock_reconciliation.go`
  - 方法：`ApproveStockReconciliation()`
  - 调整：不再直接更新库存，等待 ERPNext 同步库存

- [ ] **新增从 ERPNext 同步盘点单数据**
  - 文件：`main/app/service/stock_reconciliation.go`
  - 方法：`SyncStockReconciliationFromERP()`
  - 功能：从 ERPNext 同步盘点单数据到 TTPOS

- [ ] **新增同步库存数据**
  - 文件：`main/app/service/stock_reconciliation.go`
  - 方法：`SyncStockFromERP()`
  - 功能：从 ERPNext 同步库存数据

- [ ] **新增冲突检测**
  - 文件：`main/app/service/stock_reconciliation.go`
  - 方法：`CheckDataConflict()`
  - 功能：检测 TTPOS 和 ERPNext 数据是否冲突

### 3.3 API 接口层

- [ ] **新增 Webhook 接收接口**
  - 文件：`main/app/api/v1/webhook/erp_stock_reconciliation.go`（新建）
  - 路由：`POST /api/v1/webhook/erp/stock_reconciliation`
  - 功能：接收 ERPNext Webhook 推送

- [ ] **调整盘点单接口返回**
  - 文件：`main/app/api/v1/shop/shop_stock_reconciliation.go`
  - 调整：返回同步状态、数据来源等信息

### 3.4 定时任务

- [ ] **实现定时轮询任务**
  - 文件：`main/app/job/stock_reconciliation_sync.go`（新建）
  - 功能：定时轮询 ERPNext，同步盘点单数据

- [ ] **实现同步重试机制**
  - 文件：`main/app/job/stock_reconciliation_sync.go`
  - 功能：同步失败时重试（指数退避）

- [ ] **实现数据一致性检查**
  - 文件：`main/app/job/stock_reconciliation_sync.go`
  - 功能：对比 TTPOS 和 ERPNext 数据，发现不一致时同步

---

## 四、Webhook 配置

### 4.1 ERPNext 端配置

- [ ] **配置 Webhook**
  - 事件：`Stock Reconciliation` 的创建、更新、提交、审核事件
  - URL：`https://ttpos-api.example.com/api/v1/webhook/erp/stock_reconciliation`
  - 认证：API Key 或 Token

- [ ] **测试 Webhook**
  - 创建测试盘点单
  - 验证 Webhook 推送
  - 验证 TTPOS 接收

### 4.2 TTPOS 端实现

- [ ] **实现 Webhook 接收**
  - 文件：`main/app/api/v1/webhook/erp_stock_reconciliation.go`
  - 功能：接收 ERPNext Webhook 推送

- [ ] **实现 Webhook 验证**
  - 文件：`main/app/api/v1/webhook/erp_stock_reconciliation.go`
  - 功能：验证 Webhook 签名和来源

- [ ] **实现事件处理**
  - 文件：`main/app/api/v1/webhook/erp_stock_reconciliation.go`
  - 功能：根据事件类型处理数据同步

---

## 五、数据迁移

### 5.1 现有数据迁移

- [ ] **编写迁移脚本**
  - 文件：`admin/database/migrations/YYYYMMDD_migrate_stock_reconciliation_to_erp.sql`
  - 功能：迁移现有盘点单数据

- [ ] **迁移逻辑**
  - 查询所有已创建的盘点单（`is_open_erp = true`）
  - 检查是否有 `erp_code`，如果没有，调用 ERPNext API 创建
  - 同步 ERPNext 数据到 TTPOS
  - 更新同步状态和时间

- [ ] **测试迁移脚本**
  - 在测试环境执行
  - 验证数据完整性
  - 验证同步状态

### 5.2 功能兼容

- [ ] **添加配置开关**
  - 文件：`main/app/config/config.go`
  - 配置项：`UseERPAsDataSource`（是否使用 ERPNext 作为数据源）

- [ ] **实现兼容逻辑**
  - 文件：`main/app/service/stock_reconciliation.go`
  - 功能：根据配置开关选择新逻辑或旧逻辑

---

## 六、测试

### 6.1 单元测试

- [ ] **测试创建盘点单**
  - 文件：`main/app/service/stock_reconciliation_test.go`
  - 场景：创建、ERPNext API 调用、数据保存

- [ ] **测试更新盘点单**
  - 文件：`main/app/service/stock_reconciliation_test.go`
  - 场景：更新、ERPNext API 调用、数据更新

- [ ] **测试提交盘点单**
  - 文件：`main/app/service/stock_reconciliation_test.go`
  - 场景：提交、ERPNext API 调用、状态更新

- [ ] **测试审核盘点单**
  - 文件：`main/app/service/stock_reconciliation_test.go`
  - 场景：审核、ERPNext API 调用、库存同步

- [ ] **测试数据同步**
  - 文件：`main/app/service/stock_reconciliation_test.go`
  - 场景：从 ERPNext 同步数据、冲突检测、错误处理

### 6.2 集成测试

- [ ] **测试完整流程**
  - 创建 → 编辑 → 提交 → 审核
  - 验证数据同步
  - 验证状态流转

- [ ] **测试 Webhook 推送**
  - ERPNext 创建盘点单 → Webhook 推送 → TTPOS 接收
  - ERPNext 更新盘点单 → Webhook 推送 → TTPOS 接收
  - ERPNext 提交盘点单 → Webhook 推送 → TTPOS 接收
  - ERPNext 审核盘点单 → Webhook 推送 → TTPOS 接收

- [ ] **测试轮询兜底**
  - 模拟 Webhook 失败
  - 验证轮询同步
  - 验证数据一致性

### 6.3 端到端测试

- [ ] **测试用户操作流程**
  - 前端创建盘点单 → 后端处理 → ERPNext 创建 → 数据同步
  - 前端编辑盘点单 → 后端处理 → ERPNext 更新 → 数据同步
  - 前端提交盘点单 → 后端处理 → ERPNext 提交 → 数据同步
  - 前端审核盘点单 → 后端处理 → ERPNext 审核 → 库存同步

- [ ] **测试 ERPNext 操作流程**
  - ERPNext 创建盘点单 → Webhook 推送 → TTPOS 接收 → 前端更新
  - ERPNext 更新盘点单 → Webhook 推送 → TTPOS 接收 → 前端更新
  - ERPNext 提交盘点单 → Webhook 推送 → TTPOS 接收 → 前端更新
  - ERPNext 审核盘点单 → Webhook 推送 → TTPOS 接收 → 库存更新 → 前端更新

### 6.4 性能测试

- [ ] **测试同步性能**
  - 同步延迟测试
  - 并发同步测试
  - 批量同步测试

- [ ] **测试 API 性能**
  - ERPNext API 响应时间
  - Webhook 处理时间
  - 轮询处理时间

### 6.5 稳定性测试

- [ ] **测试网络异常**
  - ERPNext API 不可用
  - Webhook 推送失败
  - 轮询同步失败

- [ ] **测试数据一致性**
  - TTPOS 和 ERPNext 数据对比
  - 冲突检测和处理
  - 数据修复

---

## 七、监控与告警

### 7.1 监控指标

- [ ] **同步指标监控**
  - 同步成功率
  - 同步延迟
  - 同步错误率
  - 数据一致性

- [ ] **性能指标监控**
  - API 响应时间
  - Webhook 处理时间
  - 轮询处理时间

- [ ] **业务指标监控**
  - 盘点单创建成功率
  - 盘点单审核成功率
  - 用户操作延迟

### 7.2 告警配置

- [ ] **同步失败告警**
  - 同步失败率超过阈值
  - 同步延迟超过阈值
  - 数据不一致超过阈值

- [ ] **性能告警**
  - API 响应时间超过阈值
  - Webhook 处理时间超过阈值
  - 轮询处理时间超过阈值

---

## 八、文档更新

### 8.1 技术文档

- [ ] **更新架构文档**
  - 文件：`docs/human/architecture/stock-reconciliation-architecture-refactor.md`
  - 内容：架构调整说明

- [ ] **更新 API 文档**
  - 文件：`docs/shared/api/stock-reconciliation.md`
  - 内容：新增接口、调整接口

- [ ] **更新数据模型文档**
  - 文件：`docs/human/business/stock-reconciliation-product-overview.md`
  - 内容：数据模型调整

### 8.2 业务文档

- [ ] **更新产品文档**
  - 文件：`docs/human/business/stock-reconciliation-product-overview.md`
  - 内容：业务流程调整

- [ ] **更新同步机制文档**
  - 文件：`docs/human/business/stock-reconciliation-erp-sync.md`
  - 内容：双向同步机制

---

## 九、上线准备

### 9.1 灰度发布

- [ ] **选择灰度商户**
  - 选择部分商户进行灰度测试
  - 配置灰度开关

- [ ] **监控灰度数据**
  - 监控同步状态
  - 监控错误率
  - 收集用户反馈

### 9.2 全量上线

- [ ] **全量发布**
  - 全量开启新功能
  - 持续监控

- [ ] **问题处理**
  - 快速响应问题
  - 及时修复问题

### 9.3 回滚准备

- [ ] **回滚方案**
  - 保留原有代码
  - 配置开关回滚
  - 数据回滚脚本

---

## 十、验收标准

### 10.1 功能验收

- [ ] **创建盘点单**
  - TTPOS 创建 → ERPNext 创建 → 数据同步成功
  - ERPNext 创建 → TTPOS 同步 → 数据同步成功

- [ ] **更新盘点单**
  - TTPOS 更新 → ERPNext 更新 → 数据同步成功
  - ERPNext 更新 → TTPOS 同步 → 数据同步成功

- [ ] **提交盘点单**
  - TTPOS 提交 → ERPNext 提交 → 数据同步成功
  - ERPNext 提交 → TTPOS 同步 → 数据同步成功

- [ ] **审核盘点单**
  - TTPOS 审核 → ERPNext 审核 → 库存同步成功
  - ERPNext 审核 → TTPOS 同步 → 库存同步成功

### 10.2 性能验收

- [ ] **同步延迟**
  - Webhook 推送延迟 < 5秒
  - 轮询同步延迟 < 5分钟
  - API 调用延迟 < 1秒

- [ ] **同步成功率**
  - 同步成功率 > 99%
  - 同步错误率 < 1%

### 10.3 稳定性验收

- [ ] **数据一致性**
  - TTPOS 和 ERPNext 数据一致性 > 99.9%
  - 冲突检测和处理正常

- [ ] **错误处理**
  - 网络异常时降级处理正常
  - 同步失败时重试正常
  - 数据不一致时自动修复正常

---

**创建时间**：2025-01-16  
**维护者**：TTPOS Team
























