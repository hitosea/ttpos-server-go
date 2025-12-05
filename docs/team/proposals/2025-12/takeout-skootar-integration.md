# 整合 Skootar 订单逻辑到现有订单模型 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | User (AI Agent 代填)   |
| **日期**   | 2025-12-05   |
| **目标版本** | v2.11.0 |
| **状态**   | 已通过   |
| **关联任务** | - |
| **关联 Spec** | [docs/shared/specs/active/task-takeout-skootar-integration/requirements.md](../../../shared/specs/active/task-takeout-skootar-integration/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

目前 `ttpos-takeout` 模块中的 Skootar 配送逻辑实现较早，存在以下问题：
1.  **数据库耦合严重**：`takeout_job` 表中直接包含了 `skootar_id`, `skootar_name`, `skootar_phone` 等特定供应商的字段。
2.  **扩展性差**：新增其他配送服务商（如 Lalamove）时需要修改表结构。
3.  **模型不统一**：随着 Grab 对接（`takeout_order` 模型）的引入，现有的 Skootar 订单模型（基于 `job` 表）与新的通用订单模型存在差异，导致维护成本增加。

### 业务价值

1.  **提升系统可扩展性**：通过通用化数据模型，支持快速接入新的配送服务商。
2.  **降低维护成本**：统一订单/任务模型，减少特殊逻辑分支。
3.  **数据标准化**：消除数据库中的冗余字段，使数据结构更清晰。

### 目标用户

- [x] 开发人员 (维护代码、接入新渠道)
- [x] 运维人员 (数据库管理)

---

## 💡 解决方案概述

### 方案描述

采用 **"主表 + 扩展表"** 的设计模式，将 Skootar 订单逻辑适配到新的通用订单模型 (`takeout_order`) 中。
1.  **通用化**：将原 `takeout_job` 表中的通用订单字段（如金额、状态、客户信息等）合并/迁移到 `takeout_order` 主表。
2.  **保留特殊字段**：保留 `takeout_job` 表（或重命名为 `takeout_order_skootar`），去除已迁移的通用字段，仅保留 Skootar 特有的配送员信息（如 `skootar_id`, `rating` 等）。
3.  **关联**：通过 `order_uuid` 将扩展表与主表关联。

### 核心功能点

1.  **数据库重构**
    - **迁移通用数据**：将 `takeout_job` 中的历史订单数据迁移至 `takeout_order`。
    - **瘦身旧表**：将 `takeout_job` 改造为 Skootar 扩展表，仅保留 `skootar_*` 字段及关联键。
    - **建立关联**：确保新旧数据通过 UUID 正确关联。

2.  **代码重构**
    - **写入逻辑**：Skootar 下单时，基础信息写入 `takeout_order`，配送员信息写入扩展表。
    - **读取逻辑**：Controller 层查询订单时，需同时查询主表和扩展表（Join 或多次查询），聚合后返回给前端，保持 API 响应结构不变。
    - **Entity 更新**：重新生成并调整 Entity 定义。

3.  **兼容性保障**
    - **API 兼容**：对外 gRPC/HTTP 接口保持字段不变，由 Service 层负责数据聚合。
    - **数据迁移**：提供平滑迁移脚本，保证历史数据的完整性。

### 架构设计

```go
// 1. 通用订单表 (takeout_order)
type Order struct {
    Uuid           string  // 唯一标识
    ProviderName   string  // "skootar"
    Status         string  // 统一状态
    // ... 其他通用字段
}

// 2. Skootar 扩展表 (takeout_order_skootar / 原 takeout_job 瘦身)
type SkootarOrderInfo struct {
    OrderUuid     string  // 关联主表
    SkootarId     string  // 骑手ID
    SkootarName   string
    SkootarPhone  string
    SkootarRating float64
    // ... 其他特有字段
}
```

### 影响范围

**涉及模块**：
- [x] `ttpos-takeout`
    - `internal/logic/skootar`: 核心业务逻辑
    - `internal/model`: 数据模型
    - `manifest/sql`: 数据库变更

**涉及终端**：
- 后端逻辑重构，原则上不影响前端 API 契约，但需回归测试 Skootar 下单流程。

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**
- [x] **中**：涉及数据库字段迁移和核心逻辑修改，需确保数据兼容性。
- [ ] **高**

### 工作量预估

- **预计天数**: 3-5 天
- **预估 SP**: 5 SP

### 风险识别

1.  **数据丢失风险**：在清理旧字段前，必须确保数据已完整迁移到新字段。
2.  **兼容性问题**：如果前端或其他模块依赖了 `skootar_*` 字段（虽然大概率是后端封装），需要确认 API 返回结构是否需要保持兼容（通过 DTO 转换层适配）。

---

## 🔗 相关资源

### 相关文档

- `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/create_order.go`
- `docs/team/proposals/2025-12/takeout-grab-integration.md` (新订单模型参考)

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     |        |           |
| 技术负责人   |        |           |
| 开发代表     |        |           |

### 评审结论

- [ ] ✅ **批准**
- [ ] 🔄 **修改后批准**
- [ ] ❌ **拒绝**

