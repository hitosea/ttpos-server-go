# 批量重新生成订单材料消耗和POS发票 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容         |
| ---------- | ------------ |
| **提案人** | xiezhihuan   |
| **日期**   | 2025-12-17   |
| **目标版本** | v2.11.0    |
| **状态**   | 已创建 Spec  |
| **关联任务** | -            |
| **关联 Spec** | [story-main-batch-regenerate-order-material-and-invoice](../../shared/specs/archived/v2.12/story-main-batch-regenerate-order-material-and-invoice/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

当前系统已有四个独立的命令用于重新生成订单相关的数据：

1. **`regenerate-order-material`** - 重新生成订单材料用料记录
2. **`regenerate-sales-outbound`** - 重新生成销售出库汇总记录
3. **`regenerate-sale-order-material-outbound`** - 重新生成销售订单材料出库记录
4. **`regenerate-order-pos-invoice`** - 重新生成订单POS发票

但在实际业务场景中，当需要批量修复某个公司从某个日期开始的所有订单数据时，存在以下问题：

1. **操作繁琐**：需要手动按日期执行日期级别的命令，然后逐个订单执行三个命令，操作步骤多、耗时长
2. **容易遗漏**：手动操作容易遗漏某些日期、订单或某个步骤
3. **无法断点续传**：如果执行过程中中断，无法从断点继续，需要重新开始
4. **缺乏进度追踪**：无法了解整体进度和已完成/失败的日期、订单情况
5. **错误处理困难**：某个日期或订单失败后，难以定位和修复问题
6. **执行顺序复杂**：需要先按日期执行日期级别的汇总命令，再执行该日期下所有订单的命令，顺序容易出错

**示例场景**：
> 某公司在 2025-12-01 修改了成本卡配置，需要重新计算从该日期开始的所有订单的材料消耗和POS发票。当前需要手动逐个订单执行三个命令（regenerate-order-material、regenerate-sale-order-material-outbound、regenerate-order-pos-invoice），等该日期下所有订单都处理完成后，再执行日期级别的汇总命令（regenerate-sales-outbound）重新生成当天的销售出库汇总记录，操作繁琐且容易出错。

### 业务价值

解决这个问题能带来以下业务价值：

- **提升运维效率**：批量处理订单，减少手动操作时间
- **降低操作风险**：自动化流程减少人为错误
- **支持断点续传**：程序中断后可从断点继续，避免重复操作
- **进度可视化**：实时查看处理进度和结果统计
- **错误追踪**：记录详细的错误日志，便于问题排查和修复

### 目标用户

- [ ] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] 其他: **运维人员、技术支持人员、开发人员**

---

## 💡 解决方案概述

### 方案描述

创建一个新的命令行工具 `batch-regenerate-order-material-and-invoice`，用于批量重新生成指定公司和日期范围内的所有订单的材料消耗和POS发票。

**核心流程**：
1. **生成任务清单**：根据公司列表和起始日期，查询所有符合条件的订单，按日期分组，生成JSON格式的任务清单
2. **任务清单结构**：四层结构（公司 → 日期 → 订单 → 步骤）
3. **逐任务执行**：按照任务清单按日期分组执行，每个日期先执行该日期下所有订单的步骤，等所有订单执行完成后，再执行日期级别的步骤（regenerate-sales-outbound）
4. **状态管理**：每个步骤执行完成后更新状态，支持断点续传
5. **日志记录**：记录详细的执行日志和错误信息

**命令格式**：
```bash
./main batch-regenerate-order-material-and-invoice \
  --company-uuids <公司UUID列表，逗号分隔> \
  --start-date <起始日期，格式：YYYY-MM-DD> \
  [--task-file <任务清单文件路径>] \
  [--resume] \
  [--dry-run] \
  [--show-progress] \
  [--progress-interval <秒数>]
```

**参数说明**：
- `--company-uuids`：公司UUID列表，逗号分隔（必填）
- `--start-date`：起始日期，格式：YYYY-MM-DD（必填）
- `--task-file`：任务清单文件路径（可选，默认：./batch-regenerate-task-{timestamp}.json）
- `--resume`：从现有任务清单继续执行（可选）
- `--dry-run`：预览模式，仅生成任务清单（可选）
- `--show-progress`：显示详细进度信息（可选，默认：false）
- `--progress-interval`：进度信息刷新间隔（秒数，可选，默认：5秒）

### 核心功能点

1. **任务清单生成**
   - 根据公司UUID列表和起始日期查询所有符合条件的订单
   - 按日期分组订单，为每个日期生成日期级别的步骤（regenerate-sales-outbound）
   - 为每个订单生成3个步骤的任务（regenerate-order-material、regenerate-sale-order-material-outbound、regenerate-order-pos-invoice）
   - 将任务清单保存为JSON文件
   - 支持指定任务清单文件路径

2. **任务清单结构（四层）**
   ```json
   {
     "companies": [
       {
         "company_uuid": 123456,
         "company_name": "测试门店",
         "dates": [
           {
             "date": "2025-12-01",
             "date_step": {
               "step": 1,
               "name": "regenerate-sales-outbound",
               "status": "pending",  // pending, running, completed, failed
               "start_time": null,
               "end_time": null,
               "error": null
             },
             "orders": [
               {
                 "sale_order_uuid": 789012,
                 "order_no": "SO20251201001",
                 "order_date": "2025-12-01",
                 "steps": [
                   {
                     "step": 1,
                     "name": "regenerate-order-material",
                     "status": "pending",
                     "start_time": null,
                     "end_time": null,
                     "error": null
                   },
                   {
                     "step": 2,
                     "name": "regenerate-sale-order-material-outbound",
                     "status": "pending",
                     "start_time": null,
                     "end_time": null,
                     "error": null
                   },
                   {
                     "step": 3,
                     "name": "regenerate-order-pos-invoice",
                     "status": "pending",
                     "start_time": null,
                     "end_time": null,
                     "error": null
                   }
                 ]
               }
             ]
           }
         ]
       }
     ],
     "summary": {
       "total_companies": 1,
       "total_dates": 1,
       "total_orders": 1,
       "total_date_steps": 1,
       "total_order_steps": 3,
       "completed_steps": 0,
       "failed_steps": 0,
       "pending_steps": 4
     },
     "created_at": "2025-12-17T10:00:00Z",
     "updated_at": "2025-12-17T10:00:00Z"
   }
   ```

3. **任务执行流程**
   - 按公司顺序遍历
   - 按日期顺序遍历
   - 对于每个日期：
     - 先按订单顺序遍历该日期下的所有订单
     - 对于每个订单，按步骤顺序执行（regenerate-order-material → regenerate-sale-order-material-outbound → regenerate-order-pos-invoice）
     - 等该日期下所有订单都执行完成后，再执行日期级别的步骤（regenerate-sales-outbound）重新生成当天的销售出库汇总记录
   - 每个步骤执行前检查状态，跳过已完成的步骤
   - 每个步骤执行完成后更新状态和时间戳
   - 如果步骤失败，记录错误信息并继续下一个订单或日期
   - 日期级别的步骤只有在该日期下所有订单的所有步骤都完成后才会执行

4. **断点续传**
   - 支持 `--resume` 参数，从现有任务清单文件继续执行
   - 自动跳过状态为 `completed` 的步骤
   - 重新执行状态为 `failed` 的步骤
   - 保持任务清单文件的状态同步

5. **进度显示（可选功能）**
   - **公司级别进度**：
     - 显示总共有多少个公司
     - 显示还有多少个公司未开始
     - 显示哪些公司已经完成
     - 显示当前正在处理哪个公司
   - **日期级别进度**：
     - 显示当前公司下总共有多少个日期
     - 显示当前公司下还有多少个日期未开始
     - 显示当前公司下哪些日期已经完成
     - 显示当前正在处理哪个日期
   - **订单级别进度**：
     - 显示当前日期下总共有多少个订单
     - 显示当前日期下还有多少个订单未开始
     - 显示当前日期下哪些订单已经完成
     - 显示当前正在处理哪个订单
   - **步骤级别进度**：
     - 显示当前订单的步骤执行情况
     - 显示当前步骤的执行状态（pending/running/completed/failed）
   - **总体统计**：
     - 显示总体完成百分比
     - 显示已完成的步骤数、失败的步骤数、待执行的步骤数
     - 显示预计剩余时间（基于已完成步骤的平均耗时）

6. **日志记录**
   - 控制台输出：实时显示当前处理的公司、日期、订单和步骤，以及详细的进度信息
   - 文件日志：记录详细的执行日志到文件
   - 错误日志：记录每个失败步骤的错误信息
   - 统计信息：显示总体进度和完成情况

7. **错误处理**
   - 单个步骤失败不影响其他步骤和订单
   - 记录详细的错误信息（错误类型、错误消息、堆栈信息）
   - 支持手动修复后重新执行失败的任务

### 影响范围

**涉及终端**：
- [ ] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [ ] API 接口
- [x] 数据模型（订单、材料、出库、发票相关表）
- [x] 业务逻辑（复用现有的四个命令的业务逻辑）
- [ ] 第三方集成
- [x] 其他: **命令行工具、任务清单管理**

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：
- 复用现有的四个命令的业务逻辑，无需重新实现
- 需要实现任务清单的生成、状态管理和断点续传逻辑
- 需要处理JSON文件的读写和状态同步
- 需要实现详细的日志记录和错误处理

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 4-6 天（包含进度显示功能）
- **预估 SP**: 6-9 SP（待技术评审确认）
- **说明**：进度显示功能为可选功能，如果不需要此功能，可以减少1天工作量

**任务分解**：
1. 实现任务清单生成逻辑（查询订单、按日期分组、生成四层JSON结构）（1天）
2. 实现任务执行引擎（遍历公司→日期→订单→步骤、调用现有命令逻辑、状态更新）（1.5天）
3. 实现日期级别步骤执行逻辑（先执行所有订单步骤，等所有订单步骤完成后才执行日期级别步骤）（0.5天）
4. 实现订单步骤完成状态检查逻辑（检查该日期下所有订单的所有步骤是否都已完成）（0.5天）
5. 实现断点续传功能（读取任务清单、状态检查、跳过已完成步骤）（0.5天）
6. 实现日志记录系统（控制台输出、文件日志、错误记录）（0.5天）
7. 实现错误处理和统计信息（错误收集、进度统计、结果汇总）（0.5天）
8. 实现进度显示功能（可选，公司/日期/订单/步骤级别进度、总体统计、预计剩余时间）（1天）
9. 编写单元测试和集成测试（1天）

### 风险识别

**潜在风险**：
1. **数据一致性风险**：批量操作可能影响大量订单数据，需要确保每个步骤的原子性
2. **性能风险**：大批量订单处理可能耗时很长，需要优化查询和执行逻辑
3. **任务清单文件风险**：JSON文件可能被意外修改或删除，导致状态不一致
4. **并发风险**：如果同时运行多个实例处理同一批订单，可能产生数据冲突
5. **错误恢复风险**：部分订单失败后，需要手动修复并重新执行，流程可能复杂

**缓解措施**：
1. **事务处理**：每个步骤的执行使用数据库事务，确保原子性
2. **分批处理**：支持分批处理订单，避免一次性加载过多数据
3. **文件锁定**：使用文件锁机制防止多个实例同时操作同一任务清单文件
4. **状态验证**：执行前验证任务清单文件的完整性和一致性
5. **详细日志**：记录详细的执行日志，便于问题排查和手动修复
6. **进度保存**：定期保存任务清单状态，避免程序崩溃导致进度丢失

---

## 🔗 相关资源

### 参考需求

- **关联命令**：
  - `regenerate-order-material` - [提案文档](./regenerate-order-material.md) | [Spec文档](../../shared/specs/archived/v2.12/story-main-regenerate-order-material/requirements.md)
  - `regenerate-sales-outbound` - [提案文档](./regenerate-daily-sales-outbound-summary.md) | [Spec文档](../../shared/specs/archived/v2.12/story-main-regenerate-sales-outbound-summary/requirements.md)
  - `regenerate-sale-order-material-outbound` - [提案文档](./regenerate-sale-bill-material-outbound.md) | [Spec文档](../../shared/specs/archived/v2.12/story-main-regenerate-sale-bill-material-outbound/requirements.md)
  - `regenerate-order-pos-invoice` - [提案文档](./regenerate-order-pos-invoice.md) | [Spec文档](../../shared/specs/archived/v2.12/story-main-regenerate-order-pos-invoice/requirements.md)

### 相关文档

- 命令行工具规范: `.cursor/rules/go-main.mdc`
- 数据库规范: `.cursor/rules/database.mdc`
- 订单模型: `main/app/model/sale_order.go`
- 订单服务: `main/app/service/order.go`
- 销售出库汇总服务: `main/app/service/sales_outbound_summary_service.go`

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

- [ ] 创建 Spec：`story-main-batch-regenerate-order-material-and-invoice`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 运维人员/技术支持人员  
**我想** 通过命令行工具批量重新生成指定公司和日期范围内的所有订单的材料消耗和POS发票  
**以便于** 快速修复批量订单数据，无需手动逐个订单执行多个命令

### AC 验收标准（初稿）

1. **WHEN** 执行 `batch-regenerate-order-material-and-invoice --company-uuids <UUID列表> --start-date <日期>` **THEN** 系统 **SHALL** 查询所有符合条件的订单并生成JSON任务清单文件

2. **WHEN** 任务清单生成完成 **THEN** 系统 **SHALL** 自动开始执行任务，按公司→日期→订单→步骤的顺序处理

3. **WHEN** 执行每个日期时 **THEN** 系统 **SHALL** 先执行该日期下所有订单的步骤，等所有订单执行完成后，再执行日期级别的步骤（regenerate-sales-outbound）重新生成当天的销售出库汇总记录

4. **WHEN** 执行每个订单步骤时 **THEN** 系统 **SHALL** 调用对应的命令逻辑（regenerate-order-material、regenerate-sale-order-material-outbound、regenerate-order-pos-invoice）

5. **WHEN** 该日期下所有订单的所有步骤都完成后 **THEN** 系统 **SHALL** 自动执行日期级别的步骤（regenerate-sales-outbound）

6. **WHEN** 步骤执行完成 **THEN** 系统 **SHALL** 更新任务清单中对应步骤的状态为 `completed`，并记录开始时间和结束时间

7. **WHEN** 步骤执行失败 **THEN** 系统 **SHALL** 更新任务清单中对应步骤的状态为 `failed`，并记录错误信息，然后继续处理下一个订单或日期

8. **WHEN** 该日期下所有订单的所有步骤都完成后 **THEN** 系统 **SHALL** 自动执行日期级别的步骤（regenerate-sales-outbound），即使某些订单步骤失败，只要所有订单步骤都执行过（completed或failed），就会执行日期级别步骤

9. **IF** 使用 `--resume` 参数 **THEN** 系统 **SHALL** 从现有任务清单文件继续执行，跳过状态为 `completed` 的步骤

10. **WHEN** 程序重新执行时 **THEN** 系统 **SHALL** 自动跳过已完成的步骤，只执行未完成或失败的步骤

11. **WHEN** 执行过程中 **THEN** 系统 **SHALL** 在控制台实时显示当前处理的公司、日期、订单和步骤，并记录详细的日志到文件

12. **IF** 使用 `--show-progress` 参数 **THEN** 系统 **SHALL** 显示详细的进度信息，包括：
    - 公司级别：总共有多少个公司、还有多少个公司未开始、哪些公司已经完成、当前正在处理哪个公司
    - 日期级别：当前公司下总共有多少个日期、还有多少个日期未开始、哪些日期已经完成、当前正在处理哪个日期
    - 订单级别：当前日期下总共有多少个订单、还有多少个订单未开始、哪些订单已经完成、当前正在处理哪个订单
    - 步骤级别：当前订单的步骤执行情况、当前步骤的执行状态
    - 总体统计：总体完成百分比、已完成的步骤数、失败的步骤数、待执行的步骤数、预计剩余时间

13. **IF** 使用 `--progress-interval` 参数 **THEN** 系统 **SHALL** 按照指定的间隔（秒数）刷新进度信息显示

12. **WHEN** 所有任务执行完成 **THEN** 系统 **SHALL** 输出统计信息（总公司数、总日期数、总订单数、完成数、失败数、耗时等）

13. **IF** 使用 `--dry-run` 参数 **THEN** 系统 **SHALL** 仅生成任务清单，不实际执行任务

### 技术实现要点

1. **任务清单生成**：
   - 查询符合条件的订单：`SELECT * FROM ttpos_sale_order WHERE company_uuid IN (?) AND created_at >= ? AND status = ?`
   - 按日期分组订单
   - 为每个日期生成日期级别的步骤（regenerate-sales-outbound）
   - 为每个订单生成3个步骤的任务（regenerate-order-material、regenerate-sale-order-material-outbound、regenerate-order-pos-invoice）
   - 将任务清单保存为JSON文件

2. **任务执行引擎**：
   - 读取任务清单JSON文件
   - 遍历公司→日期→订单→步骤
   - 对于每个日期：
     - 先按订单顺序执行该日期下所有订单的步骤
     - 等该日期下所有订单的所有步骤都完成后（completed或failed），再执行日期级别的步骤（regenerate-sales-outbound）
   - 调用对应的Service方法执行步骤
   - 更新任务清单状态
   - 检查日期下所有订单步骤的完成状态，决定是否执行日期级别步骤

3. **步骤执行映射**：
   - **日期级别步骤**：
     - `regenerate-sales-outbound` → `salesOutboundSummarySrv.RegenerateSalesOutboundSummary()`（需要公司UUID和日期）
   - **订单级别步骤**：
     - Step 1: `regenerate-order-material` → `salesOutboundSummarySrv.RegenerateOrderMaterial()`
     - Step 2: `regenerate-sale-order-material-outbound` → `salesOutboundSummarySrv.RegenerateSaleBillMaterialOutbound()`
     - Step 3: `regenerate-order-pos-invoice` → `salesOutboundSummarySrv.RegenerateOrderPosInvoice()`

4. **状态管理**：
   - 使用文件锁防止并发操作
   - 定期保存任务清单状态（每完成一个步骤）
   - 支持断点续传（读取现有任务清单，跳过已完成步骤）

5. **日志记录**：
   - 控制台输出：使用颜色区分不同状态（成功/失败/进行中）
   - 文件日志：记录到 `logs/batch-regenerate-{timestamp}.log`
   - 错误日志：记录到任务清单JSON文件的 `error` 字段

6. **命令行参数**：
   ```go
   --company-uuids string       公司UUID列表，逗号分隔（必填）
   --start-date string          起始日期，格式：YYYY-MM-DD（必填）
   --task-file string           任务清单文件路径（可选，默认：./batch-regenerate-task-{timestamp}.json）
   --resume                      从现有任务清单继续执行（可选）
   --dry-run                     预览模式，仅生成任务清单（可选）
   --show-progress               显示详细进度信息（可选，默认：false）
   --progress-interval int       进度信息刷新间隔（秒数，可选，默认：5秒）
   ```

7. **进度显示实现**：
   - 使用定时器定期更新进度信息（默认每5秒刷新一次）
   - 计算各级别的完成百分比和剩余数量
   - 使用颜色和格式化输出，清晰展示进度信息
   - 支持实时更新，不阻塞主执行流程
   - 进度信息格式示例：
     ```
     ========================================
     批量重新生成订单材料消耗和POS发票 - 进度信息
     ========================================
     总体进度: 45.2% (123/272 步骤完成)
     预计剩余时间: 约 15 分钟
     
     公司进度:
       ✓ 已完成: 测试门店1 (2/2 日期完成)
       → 进行中: 测试门店2 (1/3 日期完成, 当前: 2025-12-02)
       ○ 未开始: 测试门店3 (0/1 日期完成)
     
     当前日期进度 (2025-12-02):
       订单进度: 60.0% (12/20 订单完成)
       当前订单: SO20251202013
       当前步骤: regenerate-sale-order-material-outbound (2/3)
     
     步骤统计:
       已完成: 123
       失败: 2
       待执行: 147
     ========================================
     ```

### 线框图/原型（可选）

命令行使用示例：

```bash
# 生成任务清单并执行
./main batch-regenerate-order-material-and-invoice \
  --company-uuids "123456,789012" \
  --start-date "2025-12-01"

# 从现有任务清单继续执行
./main batch-regenerate-order-material-and-invoice \
  --task-file "./batch-regenerate-task-20251217100000.json" \
  --resume

# 预览模式，仅生成任务清单
./main batch-regenerate-order-material-and-invoice \
  --company-uuids "123456" \
  --start-date "2025-12-01" \
  --dry-run

# 显示详细进度信息（每3秒刷新一次）
./main batch-regenerate-order-material-and-invoice \
  --company-uuids "123456,789012" \
  --start-date "2025-12-01" \
  --show-progress \
  --progress-interval 3
```

任务清单文件示例：

```json
{
  "companies": [
    {
      "company_uuid": 123456,
      "company_name": "测试门店",
      "dates": [
        {
          "date": "2025-12-01",
          "date_step": {
            "step": 1,
            "name": "regenerate-sales-outbound",
            "status": "pending",
            "start_time": null,
            "end_time": null,
            "error": null,
            "note": "该日期下所有订单步骤完成后才执行"
          },
          "orders": [
            {
              "sale_order_uuid": 789012,
              "order_no": "SO20251201001",
              "order_date": "2025-12-01",
              "steps": [
                {
                  "step": 1,
                  "name": "regenerate-order-material",
                  "status": "completed",
                  "start_time": "2025-12-17T10:00:10Z",
                  "end_time": "2025-12-17T10:00:15Z",
                  "error": null
                },
                {
                  "step": 2,
                  "name": "regenerate-sale-order-material-outbound",
                  "status": "completed",
                  "start_time": "2025-12-17T10:00:16Z",
                  "end_time": "2025-12-17T10:00:20Z",
                  "error": null
                },
                {
                  "step": 3,
                  "name": "regenerate-order-pos-invoice",
                  "status": "completed",
                  "start_time": "2025-12-17T10:00:21Z",
                  "end_time": "2025-12-17T10:00:25Z",
                  "error": null
                }
              ]
            }
          ]
        }
      ]
    }
  ],
  "summary": {
    "total_companies": 1,
    "total_dates": 1,
    "total_orders": 1,
    "total_date_steps": 1,
    "total_order_steps": 3,
    "completed_steps": 3,
    "failed_steps": 0,
    "pending_steps": 1,
    "note": "日期级别步骤（regenerate-sales-outbound）将在该日期下所有订单步骤完成后执行"
  },
  "created_at": "2025-12-17T10:00:00Z",
  "updated_at": "2025-12-17T10:00:25Z"
}
```

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
**创建日期**: 2025-12-17  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`, `.cursor/rules/go-main.mdc`

