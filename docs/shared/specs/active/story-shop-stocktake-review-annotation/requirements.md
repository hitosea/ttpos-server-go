# 盘点单审核批注与重新提交 需求文档

## 📋 基本信息

| 项目              | 内容                                                                 |
| ----------------- | -------------------------------------------------------------------- |
| **Spec ID**       | story-shop-stocktake-review-annotation                               |
| **来源 Proposal** | [shop-inventory-review-annotation](../../../../team/proposals/2026-01/shop-inventory-review-annotation.md) |
| **关联任务**      | [#39220](http://t.hitosea.com/manage/project/368/task/39220)         |
| **创建日期**      | 2026-01-26                                                           |
| **负责人**        | xiezhihuan                                                           |
| **目标版本**      | v2.16.0                                                              |

## 📋 审核状态

| 项目         | 内容   |
| ------------ | ------ |
| **审核状态** | 待审核 |
| **审核人**   |        |
| **审核日期** |        |

---

## 📝 用户故事

**作为** 店长/商户管理员、运营人员、总部审批人员
**我想** 在盘点单被驳回后能够修改并重新提交审核，同时在审核时能添加批注说明
**以便于** 减少重复录入工作，提高审批沟通效率，并满足操作留痕的合规要求

---

## 功能需求

### Requirement 1: 扩展盘点单状态机

**用户故事**: 作为单据发起人，我想在单据被驳回后能够重新提交审核，以便于不需要重新创建单据

#### 状态机扩展

```
现有状态流转：
  待审核 → 门店通过 → 总部通过（完成）
  待审核 → 门店驳回（终态）
  待审核 → 门店通过 → 总部驳回（终态）

扩展后状态流转：
  待审核 → 门店通过 → 总部通过（完成）
  待审核 → 门店驳回 → 重新提交 → 待审核（循环）
  待审核 → 门店通过 → 总部驳回 → 重新提交 → 待审核（循环）
```

#### 新增接口

**重新提交接口**（复用保存接口）`POST /shop/stock_reconciliation/save`

```go
type StockReconciliationSaveReq struct {
    // ... 现有字段
    IsResubmit bool   `json:"is_resubmit"`  // 是否重新提交（已驳回状态的盘点单重新提交）
    Annotation string `json:"annotation"`   // 批注内容（重新提交时使用）
}
```

#### 验收标准

1. **WHEN** 盘点单状态为"已驳回" **THEN** 系统 **SHALL** 允许发起人通过 save 接口（is_resubmit=true）重新提交
2. **WHEN** 非发起人尝试重新提交 **THEN** 系统 **SHALL** 返回错误"只有发起人才能重新提交"
3. **WHEN** 发起人重新提交 **THEN** 系统 **SHALL** 支持修改盘点单信息后提交
4. **WHEN** 重新提交成功 **THEN** 系统 **SHALL** 将单据状态变更为"待审核"
5. **WHEN** 重新提交时传入 annotation **THEN** 系统 **SHALL** 保存批注记录（类型=重新发起）
6. **WHEN** 创建盘点单 **THEN** 系统 **SHALL** 记录发起人员工UUID（creator_staff_uuid）
7. **WHEN** 提交盘点单 **THEN** 系统 **SHALL** 记录提交人员工UUID（submitter_staff_uuid）

---

### Requirement 2: 审核接口增加批注字段

**用户故事**: 作为审核人，我想在通过或驳回时添加批注说明，以便于发起人了解审核意见

#### 现有接口修改

**通过接口** `POST /shop/stock_reconciliation/approve`

```go
// 现有参数
type StockReconciliationApproveReq struct {
    Uuid uint64 `json:"uuid" binding:"required"`
}

// 增加批注字段
type StockReconciliationApproveReq struct {
    Uuid       uint64 `json:"uuid" binding:"required"`
    Annotation string `json:"annotation"`  // 批注内容（非必填）
}
```

**驳回接口** `POST /shop/stock_reconciliation/reject`

```go
// 现有参数
type StockReconciliationRejectReq struct {
    Uuid uint64 `json:"uuid" binding:"required"`
}

// 增加批注字段
type StockReconciliationRejectReq struct {
    Uuid       uint64 `json:"uuid" binding:"required"`
    Annotation string `json:"annotation"`  // 批注内容（非必填）
}
```

#### 验收标准

1. **WHEN** 审核人调用 approve 接口 **THEN** 系统 **SHALL** 支持传入 annotation 字段（非必填）
2. **WHEN** 审核人调用 reject 接口 **THEN** 系统 **SHALL** 支持传入 annotation 字段（非必填）
3. **WHEN** 审核人提交审核 **THEN** 系统 **SHALL** 将批注信息保存到批注表

---

### Requirement 3: 新增盘点单批注表

**用户故事**: 作为系统，我需要记录所有审核操作的批注信息，以便于追溯审批历史

#### 数据模型

**表名**: `ttpos_stock_reconciliation_annotation`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint | 主键 |
| uuid | bigint | UUID |
| stock_reconciliation_uuid | bigint | 关联盘点单 UUID |
| annotation_type | int | 批注类型：1=重新发起, 2=驳回, 3=通过 |
| content | text | 批注内容 |
| create_time | int | 创建时间（Unix 时间戳） |
| update_time | int | 更新时间 |
| delete_time | int | 删除时间（软删除） |

#### 批注类型常量

```go
const (
    AnnotationTypeResubmit = 1  // 重新发起
    AnnotationTypeReject   = 2  // 驳回
    AnnotationTypeApprove  = 3  // 通过
)
```

#### 验收标准

1. **WHEN** 发起人重新提交审核 **THEN** 系统 **SHALL** 在批注表新增一条记录，类型为"重新发起"
2. **WHEN** 审核人驳回单据 **THEN** 系统 **SHALL** 在批注表新增一条记录，类型为"驳回"
3. **WHEN** 审核人通过单据 **THEN** 系统 **SHALL** 在批注表新增一条记录，类型为"通过"
4. **WHEN** 批注内容为空 **THEN** 系统 **SHALL** 保存空字符串（不跳过记录）

---

### Requirement 4: 批注历史（集成到详情接口）

**用户故事**: 作为用户，我想在查看盘点单详情时看到所有批注历史，以便于了解审批过程

#### API 设计

批注列表集成到盘点单详情接口中：

```
GET /shop/stock_reconciliation/detail?uuid=xxx
```

#### 响应格式（新增 annotations 字段）

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "uuid": 8267304538112001,
    "order_no": "ST202601271234560001",
    "status": 3,
    "items": [...],
    "annotations": [
      {
        "uuid": 8267304538112500,
        "annotation_type": 2,
        "annotation_type_name": "驳回",
        "content": "数量有误，请核实",
        "create_time": 1737878400
      }
    ]
  }
}
```

#### 验收标准

1. **WHEN** 用户查询盘点单详情 **THEN** 系统 **SHALL** 在响应中包含批注列表（按创建时间倒序）
2. **WHEN** 单据有多次驳回和重新提交 **THEN** 系统 **SHALL** 展示完整的操作时间线
3. **WHEN** 单据无批注记录 **THEN** 系统 **SHALL** 返回空列表（不返回 null）

---

## 非功能需求

### 测试要求

- [ ] 单元测试覆盖率 ≥ 80%
- [ ] Service 层测试覆盖核心业务逻辑
- [ ] Repository 层测试覆盖数据访问

### 平台兼容性

- [x] Shop 商家管理端（Web）
- [x] 后端 API

---

## 技术约束

- Go 版本: 1.23+
- 框架: Gin + GORM
- 分层架构: API → Service → Repository → Model
- 必须遵循 CLAUDE.md 和 .cursor/rules/go-main.mdc 规范
- 数据库表名使用 `ttpos_` 前缀
- 迁移文件表名不带前缀
- 切片响应使用 `make` 初始化

### 资源约束

- Story Point: 5

---

## 风险和缓解

### 风险 1: 批注表设计需支持扩展

**影响**: 中
**缓解措施**: 预留 `source_type` 字段，便于后续扩展到采购单、调拨单等其他单据类型

### 风险 2: 状态机变更可能影响现有流程

**影响**: 中
**缓解措施**: 仅扩展驳回状态的后续流转，不修改现有正向流程

---

**版本**: v1.0.0
**创建日期**: 2026-01-26
