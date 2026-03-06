# story-erp-stock-entry-deferred 任务清单

## 进度总览

| 项目 | 数值 |
|------|------|
| 总 SP | 5 |
| 总任务数 | 10 |
| 已完成 | 0 |
| 完成率 | 0% |

---

## Phase 1: 数据模型

### 1.1 创建盘点快照表

| 项目 | 内容 |
|------|------|
| File | `admin/database/migrations/`, `admin/database/seeds/shop_01.sql` |
| Purpose | 存储盘点时的未出库订单快照 |
| Requirements | ttpos_stocktake_snapshot 表 |

- [ ] 完成

### 1.2 新增 Stock Entry Type

| 项目 | 内容 |
|------|------|
| File | ERPNext 配置 / BMP 初始化脚本 |
| Purpose | 新增 Material Inventory Deduction 类型 |
| Requirements | Purpose: Material Consumption for Manufacture |

- [ ] 完成

---

## Phase 2: 核心逻辑

### 2.1 Stock Entry 合并服务

| 项目 | 内容 |
|------|------|
| File | `main/app/service/erp_stock_entry.go` |
| Purpose | 查询未出库订单，按 item+warehouse 合并，调用 ERP |
| Requirements | 合并规则、分批处理、状态更新 |

- [ ] 完成

### 2.2 BMP Stock Entry 合并逻辑

| 项目 | 内容 |
|------|------|
| File | `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_entry_merge.go` |
| Purpose | 接收合并请求，调用 ERPNext API 创建 Stock Entry |
| Requirements | Purpose=Material Consumption for Manufacture, Type=Material Inventory Deduction |

- [ ] 完成

### 2.3 盘点快照生成

| 项目 | 内容 |
|------|------|
| File | `main/app/service/erp_stock_entry.go` |
| Purpose | 添加盘点时生成未出库订单快照 |
| Requirements | 按 item+warehouse 合并，与盘点单关联 |

- [ ] 完成

### 2.4 盘点差异计算集成

| 项目 | 内容 |
|------|------|
| File | 盘点相关服务文件 |
| Purpose | 预期库存 = ERP 账面 - 快照扣减量 |
| Requirements | 盘点提交无需等待 Stock Entry |

- [ ] 完成

---

## Phase 3: 触发机制

### 3.1 盘点事件触发

| 项目 | 内容 |
|------|------|
| File | 盘点相关服务文件 |
| Purpose | 添加盘点单据时触发 Stock Entry 合并 |
| Requirements | 异步触发，不阻塞盘点流程 |

- [ ] 完成

### 3.2 0 点定时任务

| 项目 | 内容 |
|------|------|
| File | `main/app/tasks/erp_stock_entry_task.go` |
| Purpose | 门店时区 0 点触发 Stock Entry 合并 |
| Requirements | 按时区分组，批量处理 |
| Leverage | 现有 tasks 框架 |

- [ ] 完成

### 3.3 失败重试和告警

| 项目 | 内容 |
|------|------|
| File | 相关 consumer / task 文件 |
| Purpose | Stock Entry 失败时重试和告警 |
| Requirements | 5 分钟间隔，最多 3 次，超过告警 |

- [ ] 完成

---

## Phase 4: 测试

### 4.1 编写单元测试

| 项目 | 内容 |
|------|------|
| File | `main/app/service/erp_stock_entry_test.go` |
| Purpose | 合并逻辑、快照生成、定时任务测试 |
| Requirements | 覆盖率 >= 80% |

- [ ] 完成

---

## 提交清单

### 代码质量
- [ ] `go mod tidy` 执行
- [ ] `go fmt ./...` 执行
- [ ] `go vet ./...` 通过
- [ ] 测试通过

### 功能完整性
- [ ] 盘点触发 Stock Entry 合并
- [ ] 0 点定时触发 Stock Entry 合并
- [ ] 相同 item+warehouse 合并为一条
- [ ] 盘点快照正确生成
- [ ] 盘点提交无需等待 Stock Entry

### 迁移同步
- [ ] ttpos_stocktake_snapshot 表迁移
- [ ] shop_01.sql 已更新
- [ ] ERPNext Stock Entry Type 配置
