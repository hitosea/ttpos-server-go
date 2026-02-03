# story-shop-stock-loss 任务清单

## 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 5 |
| 总任务数 | 13 |
| 已完成 | 0 |
| 完成率 | 0% |

---

## Phase 1: 数据层

### 1.1 创建数据库迁移文件

| 项目 | 内容 |
|------|------|
| File | `admin/database/migrations/20260202000001_create_stock_loss_tables.php` |
| Purpose | 创建报损单相关数据表 |
| Requirements | Req 1, Req 2, Req 6 |
| Tables | `ttpos_stock_loss`, `ttpos_stock_loss_item`, `ttpos_stock_loss_annotation`, `ttpos_stock_loss_file` |

**表结构要点**:
- `ttpos_stock_loss`: 主表，包含状态、同步状态、时间戳字段
- `ttpos_stock_loss_item`: 明细表，包含单位换算字段
- `ttpos_stock_loss_annotation`: 批注表，记录操作历史
- `ttpos_stock_loss_file`: 附件关联表，关联 ttpos_file

- [ ] 完成

### 1.2 更新种子文件

| 项目 | 内容 |
|------|------|
| File | `admin/database/seeds/shop_01.sql` |
| Purpose | 同步表结构到种子文件 |
| Requirements | 数据库规范 |

- [ ] 完成

### 1.3 创建 Model 文件

| 项目 | 内容 |
|------|------|
| Files | `main/app/model/stock_loss.go`<br/>`main/app/model/stock_loss_item.go`<br/>`main/app/model/stock_loss_annotation.go`<br/>`main/app/model/stock_loss_file.go` |
| Purpose | 定义数据模型 |
| Leverage | 参考 `model/stock_reconciliation.go`、`model/transfer_order_file.go` |

- [ ] 完成

### 1.4 创建 Repository 文件

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/stock_loss_repo.go` |
| Purpose | 数据访问层，含 Option 模式 |
| Leverage | 参考 `repository/stock_reconciliation_repo.go` |
| Methods | `Create`, `Update`, `Delete`, `GetByUuid`, `GetList`, `CreateItems`, `GetItemsByStockLossUuid`, `CreateAnnotation`, `GetAnnotationsByStockLossUuid`, `CreateFiles`, `DeleteFilesByStockLossUuid`, `GetFilesByStockLossUuid` |

- [ ] 完成

---

## Phase 2: 业务层

### 2.1 创建 DTO 文件

| 项目 | 内容 |
|------|------|
| Files | `main/app/dto/req/stock_loss_req.go`<br/>`main/app/dto/resp/stock_loss_resp.go` |
| Purpose | 请求/响应对象定义 |
| Requirements | Req 1-6 |

**Request DTO**:
- `CreateStockLossReq`: 创建报损单
- `UpdateStockLossReq`: 更新报损单
- `DeleteStockLossReq`: 删除报损单
- `StockLossListReq`: 列表查询（含筛选）
- `GetStockLossReq`: 详情查询
- `SubmitStockLossReq`: 提交
- `ApproveStockLossReq`: 审核通过
- `RejectStockLossReq`: 驳回（含驳回原因）
- `ResubmitStockLossReq`: 重新提交

**Response DTO**:
- `StockLossListResp`: 列表响应
- `StockLossDetailResp`: 详情响应（含明细、附件）
- `StockLossAnnotationsResp`: 批注列表响应
- `StockLossFileResp`: 附件响应（file_uuid, file_url, file_name）

- [ ] 完成

### 2.2 创建 Service 文件 - CRUD

| 项目 | 内容 |
|------|------|
| File | `main/app/service/stock_loss.go` |
| Purpose | CRUD 业务逻辑 |
| Requirements | Req 1 |
| Leverage | `WarehouseService`, `MaterialService` |
| Methods | `Create`, `Update`, `Delete`, `GetDetail`, `GetList` |

**关键逻辑**:
- 生成单据编号：`SL` + 时间戳 + 序列号
- 保存时更新单据日期
- 删除仅限草稿状态

**保存流程（含附件）**:
```
事务开始
  ├── 1. 保存/更新主表 (StockLoss)
  ├── 2. 全量替换明细 (StockLossItem)
  │       - 删除旧明细（软删除）
  │       - 插入新明细
  └── 3. 全量替换附件关联 (StockLossFile)
          - 删除旧关联（硬删除，File 表不删）
          - 插入新关联
事务提交
```

- [ ] 完成

### 2.3 创建 Service 文件 - 流程操作

| 项目 | 内容 |
|------|------|
| File | `main/app/service/stock_loss.go` |
| Purpose | 提交/审核/驳回流程 |
| Requirements | Req 2, Req 3, Req 4 |
| Methods | `Submit`, `Approve`, `Reject`, `Resubmit` |

**关键逻辑**:

1. **Submit**:
   - 校验状态 = 已保存
   - 库存校验（按基本单位汇总）
   - 更新状态 = 已提交，记录 submit_time
   - 创建批注

2. **Approve**（先 ERP 后库存）:
   - 校验状态 = 已提交
   - 再次库存校验（仅校验）
   - 调用 ERP API 创建 Stock Entry
   - ERP 失败：返回错误提示
   - ERP 成功：开启事务
     - 更新 erp_code
     - 更新状态 = 已通过
     - 扣减 TTPOS 库存
     - 创建批注
   - 提交事务

3. **Reject**:
   - 校验状态 = 已提交
   - 校验驳回原因非空
   - 更新状态 = 已驳回，记录 reject_time, reject_reason
   - 创建批注

4. **Resubmit**:
   - 校验状态 = 已驳回
   - 库存校验
   - 更新状态 = 已提交，记录新 submit_time
   - 创建批注

- [ ] 完成

### 2.4 实现库存校验逻辑

| 项目 | 内容 |
|------|------|
| File | `main/app/service/stock_loss.go` |
| Purpose | 库存充足校验 |
| Requirements | Req 2 |
| Method | `checkStockSufficient(items []*model.StockLossItem) error` |

**校验逻辑**:
```go
// 1. 按 MaterialUuid 分组
// 2. 通过 MaterialUnitUuid 查询 ttpos_material_unit.conversion_rate
// 3. 计算基本单位数量：BaseQty = Quantity × conversion_rate
// 4. 汇总同一物料的基本单位数量
// 5. 查询当前库存
// 6. 比较，不足则返回错误列表
```

- [ ] 完成

### 2.5 扩展 ERP RPC 客户端

| 项目 | 内容 |
|------|------|
| File | `main/app/service/rpc/erp/stock.go` |
| Purpose | 新增 CreateStockEntry 方法 |
| Requirements | Req 3 |
| Dependency | 需确认 BMP 侧接口 |

**注意**: 需先确认 `ttpos-bmp` 是否已有 Stock Entry 创建接口，若无需先在 BMP 侧实现。

- [ ] 完成

### 2.6 实现批注功能

| 项目 | 内容 |
|------|------|
| File | `main/app/service/stock_loss.go` |
| Purpose | 批注记录与查询 |
| Requirements | Req 6 |
| Methods | `createAnnotation`, `GetAnnotations` |

- [ ] 完成

---

## Phase 3: API 层

### 3.1 创建 Controller 文件

| 项目 | 内容 |
|------|------|
| File | `main/app/api/v1/shop/stock_loss.go` |
| Purpose | HTTP Handler |
| Requirements | Req 1-6 |
| Endpoints | 10 个接口 |

**接口列表**:
- `GET /api/v1/shop/stock_loss/list`
- `GET /api/v1/shop/stock_loss/detail`
- `POST /api/v1/shop/stock_loss/create`
- `POST /api/v1/shop/stock_loss/update`
- `POST /api/v1/shop/stock_loss/delete`
- `POST /api/v1/shop/stock_loss/submit`
- `POST /api/v1/shop/stock_loss/approve`
- `POST /api/v1/shop/stock_loss/reject`
- `POST /api/v1/shop/stock_loss/resubmit`
- `GET /api/v1/shop/stock_loss/annotations`

- [ ] 完成

### 3.2 注册路由

| 项目 | 内容 |
|------|------|
| File | `main/router/shop.go` 或相应路由文件 |
| Purpose | 注册 stock_loss 路由组 |

- [ ] 完成

### 3.3 添加功能权限

| 项目 | 内容 |
|------|------|
| File | 权限配置文件 |
| Purpose | 添加「报损管理」功能权限 |
| Requirements | Req 7 |

**权限配置**:
- 权限标识: `stock_loss_manage`
- 默认勾选: 超管、收银员
- 现有商户: 默认勾选

- [ ] 完成

---

## Phase 4: 测试

### 4.1 编写 Service 单元测试

| 项目 | 内容 |
|------|------|
| File | `main/app/service/stock_loss_test.go` |
| Purpose | Service 层单元测试 |
| Coverage | ≥ 80% |

**测试用例**:
- CRUD 操作
- 状态流转：草稿→提交→通过
- 状态流转：草稿→提交→驳回→重新提交→通过
- 库存校验：充足/不足
- 库存校验：多单位汇总
- ERP 同步：成功/失败回滚

- [ ] 完成

### 4.2 编写 Repository 单元测试

| 项目 | 内容 |
|------|------|
| File | `main/app/repository/stock_loss_repo_test.go` |
| Purpose | Repository 层单元测试 |
| Coverage | ≥ 70% |

- [ ] 完成

---

## 提交清单

### 代码质量

- [ ] `go mod tidy` 执行
- [ ] `go fmt ./...` 执行
- [ ] `go vet ./...` 通过
- [ ] 测试通过: `go test ./...`

### 功能完整性

- [ ] 所有验收标准通过
- [ ] API 响应格式正确（data 为对象）
- [ ] 多语言字段使用 LocaleResponse
- [ ] 切片使用 make 初始化
- [ ] 日志包含 company_uuid

### 迁移同步

- [ ] 迁移文件已创建
- [ ] shop_01.sql 已更新

### 文档更新

- [ ] Swagger 文档已生成: `cd main && swag init`

---

## 依赖确认

| 依赖项 | 状态 | 说明 |
|--------|------|------|
| BMP Stock Entry 接口 | ⚠️ 待确认 | 需确认 ttpos-bmp 是否已有创建 Stock Entry 的 gRPC 接口 |
| 单位换算数据 | ✅ 已确认 | TTPOS 本地维护 |
| 出入库明细接口 | ✅ 可复用 | `WarehouseService.GetWarehouseInOutList` |

---

**版本**: v1.0.0
**创建日期**: 2026-02-02
