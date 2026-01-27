# story-shop-stocktake-review-annotation 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 5 |
| 总任务数 | 10 |
| 已完成 | 0 |
| 完成率 | 0% |

---

## Phase 1: Model + Repository + 迁移

### 1.0 扩展盘点单 Model

| 项目 | 内容 |
|------|------|
| File | `main/app/model/stock_reconciliation.go` |
| Purpose | 添加发起人和提交人字段 |
| Requirements | R1 |
| Fields | CreatorStaffUuid（发起人）, SubmitterStaffUuid（提交人）|
| Note | 需要对应的数据库迁移添加这两个字段 |

- [ ] 完成

### 1.1 创建批注常量

| 项目 | 内容 |
|------|------|
| File | `main/app/constant/stock_reconciliation_annotation.go` |
| Purpose | 定义批注类型常量 |
| Requirements | R3: 批注类型 1=重新发起, 2=驳回, 3=通过 |

- [ ] 完成

### 1.2 创建批注 Model

| 项目 | 内容 |
|------|------|
| File | `main/app/model/stock_reconciliation_annotation.go` |
| Purpose | 批注数据模型 |
| Requirements | R3: 数据模型定义 |
| Fields | id, uuid, stock_reconciliation_uuid, annotation_type, content, create_time, update_time, delete_time |

- [ ] 完成

### 1.3 创建数据库迁移

| 项目 | 内容 |
|------|------|
| File | `admin/database/migrations/{timestamp}_create_stock_reconciliation_annotation.php` |
| Purpose | 创建批注表 |
| Requirements | R3: 数据模型 |
| Note | 表名不带 ttpos_ 前缀 |

- [ ] 完成

### 1.4 更新 shop_01.sql

| 项目 | 内容 |
|------|------|
| File | `admin/database/seeds/shop_01.sql` |
| Purpose | 同步种子文件 |
| Requirements | 项目规范 |

- [ ] 完成

### 1.5 创建批注 Repository

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/stock_reconciliation_annotation_repo.go` |
| Purpose | 批注数据访问层 |
| Requirements | R3, R4 |
| Methods | Create, GetListByStockReconciliationUuid |

- [ ] 完成

---

## Phase 2: Service 层业务逻辑

### 2.1 修改请求 DTO

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/req/stock_reconciliation.go` |
| Purpose | 修改保存请求和审核请求 |
| Requirements | R1, R2, R4 |
| Changes | SaveReq 增加 IsResubmit 和 Annotation 字段；ApproveReq/RejectReq 增加 Annotation 字段；新增 AnnotationListReq |

- [ ] 完成

### 2.2 创建响应 DTO

| 项目 | 内容 |
|------|------|
| File | `main/app/dto/resp/stock_reconciliation_annotation.go` |
| Purpose | 批注列表响应 |
| Requirements | R4 |

- [ ] 完成

### 2.3 扩展 Service 层

| 项目 | 内容 |
|------|------|
| File | `main/app/service/stock_reconciliation.go` |
| Purpose | 扩展保存和审核逻辑，新增批注查询 |
| Requirements | R1, R2, R3, R4 |
| Changes | 扩展 SaveStockReconciliation 支持重新提交；修改 ApproveStockReconciliation, RejectStockReconciliation；新增 GetAnnotationList |
| Leverage | 现有保存和审核逻辑 |
| Note | 重新提交复用 SaveStockReconciliation，通过 IsResubmit 标识区分；新增发起人验证逻辑 |

- [ ] 完成

---

## Phase 3: API 层 + 测试

### 3.1 扩展 API 层

| 项目 | 内容 |
|------|------|
| File | `main/app/api/v1/shop/shop_stock_reconciliation.go` |
| Purpose | 修改审核接口，新增重新提交和批注查询接口 |
| Requirements | R1, R2, R4 |
| Changes | 修改 ApproveStockReconciliation, RejectStockReconciliation；新增 ResubmitStockReconciliation, GetAnnotationList |
| Routes | POST /resubmit, GET /annotation |

- [ ] 完成

### 3.2 编写单元测试

| 项目 | 内容 |
|------|------|
| File | `main/app/service/stock_reconciliation_test.go` |
| Purpose | 单元测试覆盖 |
| Requirements | 覆盖率 ≥ 80% |
| Scenarios | 驳回后重新提交、批注保存、批注查询排序 |

- [ ] 完成

---

## 提交清单

### 代码质量
- [ ] `go mod tidy` 执行
- [ ] `go fmt ./...` 执行
- [ ] `go vet ./...` 通过
- [ ] 测试通过: `go test ./...`

### 功能完整性
- [ ] R1: 驳回后可重新提交审核
- [ ] R2: 审核接口支持批注字段
- [ ] R3: 批注记录正确保存
- [ ] R4: 批注历史按时间倒序返回

### 迁移同步
- [ ] 迁移文件已创建
- [ ] shop_01.sql 已更新

---

**版本**: v1.0.0
