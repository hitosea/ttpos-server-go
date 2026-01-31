# story-shop-transfer-resubmit-annotation 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 5 |
| 总任务数 | 8 |
| 已完成 | 8 |
| 完成率 | 100% |

---

## Phase 1: 数据层 + 核心逻辑

### 1.1 创建调拨单批注常量

| 项目 | 内容 |
|------|------|
| File | `main/app/constant/transfer_order_annotation.go` |
| Purpose | 定义调拨单批注操作类型常量和多语言映射 |
| Requirements | Req 2 - 审核批注功能 |

**常量定义**:
- TransferOrderAnnotationTypeResubmit (1) - 重新提交
- TransferOrderAnnotationTypeShopApprove (2) - 门店通过
- TransferOrderAnnotationTypeShopReject (3) - 门店驳回
- TransferOrderAnnotationTypeParentApprove (4) - 上级门店通过
- TransferOrderAnnotationTypeParentReject (5) - 上级门店驳回
- TransferOrderAnnotationTypeShipperApprove (6) - 发货门店通过
- TransferOrderAnnotationTypeShipperReject (7) - 发货门店驳回
- TransferOrderAnnotationTypeReceiverApprove (8) - 收货门店通过
- TransferOrderAnnotationTypeReceiverReject (9) - 收货门店驳回

- [x] 完成

---

### 1.2 扩展调拨单模型

| 项目 | 内容 |
|------|------|
| File | `main/app/model/transfer_order.go` |
| Purpose | 新增 annotations JSON 字段存储批注列表 |
| Requirements | Req 2, Req 3 |

**新增字段**:
- Annotations string (JSON TEXT) - 批注列表

**新增结构**:
- TransferOrderAnnotationJSON - 批注 JSON 序列化结构

- [x] 完成

---

### 1.3 扩展批注操作辅助方法

| 项目 | 内容 |
|------|------|
| File | `main/app/service/transfer_order/helper.go` |
| Purpose | 批注 JSON 操作和跨库同步 |
| Requirements | Req 2, Req 3 |

**新增方法**:
- AppendAnnotation - 追加批注到 JSON
- GetAnnotationList - 解析批注 JSON 列表
- TruncateAnnotation - 截断批注内容用于摘要
- GetApproveAnnotationType - 获取通过批注类型
- GetRejectAnnotationType - 获取驳回批注类型

**注意**: 跨库同步通过现有的 `CopyDataToHeadquarter` 方法实现，无需新增同步方法

- [x] 完成

---

### 1.4 扩展 TransferOrderSrv 服务

| 项目 | 内容 |
|------|------|
| File | `main/app/service/transfer_order/transfer_order.go` |
| Purpose | 实现重新提交和批注核心逻辑 |
| Requirements | Req 1, Req 2, Req 3 |

**修改方法**:
- UpdateTransferOrder - 扩展支持驳回状态重新提交，追加批注，同步到 SAAS 库
- GetTransferOrderDetail - 从 JSON 解析批注列表和最新批注摘要
- ApproveTransferOrder - 追加批注（通过现有 CopyDataToHeadquarter 同步）
- RejectTransferOrder - 追加批注（通过现有 CopyDataToHeadquarter 同步）

- [x] 完成

---

## Phase 2: API 层集成

### 2.1 创建批注响应 DTO

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/resp/transfer_order_annotation.go` |
| Purpose | 定义批注响应结构（多语言） |
| Requirements | Req 3 |

**响应 DTO**:
- TransferOrderAnnotationItem（用于详情接口中的批注列表，含 LocaleAnnotationTypeName）

- [x] 完成

---

### 2.2 扩展现有请求 DTO

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/req/transfer_order.go` |
| Purpose | 为 approve/reject 添加 annotation 字段，扩展 update 支持重新提交 |
| Requirements | Req 1, Req 2 |

**修改**:
- TransferOrderApproveReq 新增 Annotation string 字段
- TransferOrderRejectReq 新增 Annotation string 字段
- TransferOrderUpdateReq 新增 IsSubmit、IsConfirm、Annotation 字段和 isResubmit 私有字段（带 getter/setter）

- [x] 完成

---

### 2.3 新增 API Handler

| 项目 | 内容 |
|------|------|
| File | `main/app/api/v1/shop/shop_transfer.go` |
| Purpose | 新增重新提交接口 |
| Requirements | Req 1 |

**新增 Handler**:
- ResubmitTransferOrder - POST /shop/transfer/order/resubmit
  - 内部调用 `UpdateTransferOrder` 并设置 `SetIsResubmit(true)`

- [x] 完成

---

### 2.4 创建数据库迁移

| 项目 | 内容 |
|------|------|
| File | `admin/database/migrations/20260127160000_add_annotations_to_transfer_order.php` |
| Purpose | 为调拨单表添加 annotations 字段 |
| Requirements | 数据模型支撑 |

**迁移内容**:
- 为 ttpos_transfer_order 表添加 annotations TEXT 字段
- **TARGET = 'all'** - 同时应用到 SAAS 库和所有门店库

**同步更新**:
- [x] 创建迁移文件
- [x] 更新 `admin/database/seeds/shop_01.sql`

- [x] 完成

---

## 提交清单

### 代码质量
- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过（注：business_data_resp 有预存问题，与本功能无关）
- [x] 无废弃文件需删除（未创建独立批注表）

### 功能完整性
- [x] AC 1-5: 重新提交审核功能
- [x] AC 1-8: 审核批注功能（所有节点）
- [x] AC 1-3: 批注历史展示（发起者+审批者可见）
- [x] 跨库同步正确（SAAS ↔ 门店）
- [x] 批注内容按原文存储

### 迁移同步
- [x] 迁移文件已创建（TARGET = 'all'）
- [x] shop_01.sql 已更新

---

## 文件变更汇总

### 新建文件
| 文件 | 说明 |
|------|------|
| `main/app/constant/transfer_order_annotation.go` | 批注常量和多语言映射 |
| `main/app/dto/resp/transfer_order_annotation.go` | 批注响应 DTO |
| `admin/database/migrations/20260127160000_add_annotations_to_transfer_order.php` | 迁移文件 |

### 修改文件
| 文件 | 变更 |
|------|------|
| `main/app/model/transfer_order.go` | 新增 Annotations 字段和 JSON 结构 |
| `main/app/service/transfer_order/transfer_order.go` | 扩展 Update/Detail/Approve/Reject 方法，支持批注和跨库同步 |
| `main/app/service/transfer_order/helper.go` | 新增批注 JSON 操作辅助方法 |
| `main/app/api/v1/shop/shop_transfer.go` | 新增 Resubmit Handler 和路由 |
| `main/app/dto/req/transfer_order.go` | 扩展 approve/reject/update 请求结构 |
| `main/app/dto/resp/transfer_order.go` | 扩展 detail 响应结构（Annotations, LatestAnnotation） |
| `admin/database/seeds/shop_01.sql` | 新增 annotations 字段 |

---

**版本**: v2.0.0
**创建日期**: 2026-01-27
**更新日期**: 2026-01-27

### 变更记录

| 版本 | 日期 | 变更内容 |
|------|------|----------|
| v2.0.0 | 2026-01-27 | 架构重构：废弃独立批注表，改用调拨单主表 JSON 字段，支持跨库同步 |
| v1.0.0 | 2026-01-27 | 初始版本 |
