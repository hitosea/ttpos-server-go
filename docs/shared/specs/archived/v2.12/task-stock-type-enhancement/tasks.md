# 盘点单类型字段扩展 任务分解

> 本文档定义盘点单类型字段扩展的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-2 小时（SP ≤ 0.5）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 6  
**已完成**: 4  
**进行中**: -  
**完成率**: 67%

---

## Phase 1: 数据库迁移和模型同步

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_type_options_to_stock_reconciliation_table.php`
  - Purpose: 修改 `type` 字段注释，新增日盘、周盘、月盘类型说明
  - Requirements: 1.1
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，修改 ttpos_stock_reconciliation 表的 type 字段注释，增加类型 3-日盘、4-周盘、5-月盘 | Context: 仅修改注释，不修改字段类型和默认值，保持现有数据不变 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查字段是否存在，确保幂等性 | Success: 迁移文件创建成功，字段注释更新正确

- [x] 1.2 同步更新种子数据文件

  - File: `admin/database/seeds/shop_01.sql`
  - Purpose: 同步迁移文件的变更到种子数据，确保新商户初始化时使用正确的表结构
  - Requirements: 1.2
  - Leverage: 迁移文件: Task 1.1，现有种子文件: `admin/database/seeds/shop_01.sql`
  - Prompt: Role: Database Engineer | Task: 在 shop_01.sql 中找到 ttpos_stock_reconciliation 表的定义，将 type 字段的注释更新为与迁移文件一致 | Context: 必须与迁移文件保持完全一致 | Restrictions: 遵循 .cursor/rules/database.mdc，这是强制要求 | Success: shop_01.sql 中 type 字段注释已更新，与迁移文件一致

- [x] 1.3 更新 Go Model 注释

  - File: `main/app/model/stock_reconciliation.go`
  - Purpose: 同步数据库字段注释到 Go Model，保持代码与数据库定义一致
  - Requirements: 1.3
  - Leverage: 迁移文件: Task 1.1，现有 Model: `main/app/model/stock_reconciliation.go`
  - Prompt: Role: Go Developer | Task: 在 stock_reconciliation.go 中找到 Type 字段，将 comment 标签更新为"盘点类型 1-指定物品盘点 2-全部物品盘点 3-日盘 4-周盘 5-月盘" | Context: 仅修改注释，不修改字段类型和标签 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 注释已更新，代码格式正确

- [x] 1.4 更新 DTO 请求参数验证

  - File: `main/app/dto/req/stock_reconciliation.go`
  - Purpose: 更新保存盘点单请求参数的 Type 字段，支持新类型并添加参数验证
  - Requirements: 1.1
  - Leverage: 现有 DTO: `main/app/dto/req/stock_reconciliation.go`
  - Prompt: Role: Go Developer | Task: 更新 StockReconciliationSaveReq 的 Type 字段，添加 binding:"oneof=1 2 3 4 5" 验证标签，更新注释 | Context: 使用 Gin 的参数验证标签，确保只能传入合法的类型值 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 验证规则已更新，注释已更新

- [ ] 1.5 执行数据库迁移

  - File: -
  - Purpose: 在数据库中应用字段注释变更
  - Requirements: 1.1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd /home/coder/workspaces/ttpos-server-go/admin && php think migrate:run`
  - Success: 迁移执行成功，字段注释已更新

---

## Phase 2: 测试验证

- [ ] 2.1 验证迁移和数据一致性

  - File: -
  - Purpose: 确保迁移成功，现有数据不受影响，新类型可正常使用
  - Requirements: 所有功能需求
  - Leverage: 所有前置任务
  - Test Cases:
    1. 查看 type 字段定义，验证注释已更新
    2. 查询现有盘点单（type=1 或 2），验证数据正常
    3. 重复执行迁移，验证幂等性
    4. 创建新类型盘点单（type=3、4、5），验证保存成功
    5. 测试参数验证：传入非法类型值（如 0、6），验证返回错误
  - Success: 所有测试通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt`
- [ ] 迁移文件语法正确
- [ ] 种子文件 SQL 语法正确

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] 种子文件 `shop_01.sql` 已更新
- [ ] Go Model 注释已更新
- [ ] 迁移文件包含清晰注释

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 强制要求：种子文件已同步

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/task-stock-type-enhancement/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/task-stock-type-enhancement/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/task-stock-type-enhancement/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/task-stock-type-enhancement/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/task-stock-type-enhancement/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **实现代码**: 按照规范实现功能
5. **运行检查**: 执行迁移，验证结果
6. **标记完成**: 将 `[ ]` 改为 `[x]`
7. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-29  
**维护者**: 曾振华

