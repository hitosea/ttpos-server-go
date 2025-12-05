# 并台、转台、转菜操作使用多个订单锁 任务分解

> 本文档定义并台、转台、转菜操作使用多个订单锁的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 16  
**已完成**: 10  
**进行中**: -  
**完成率**: 63%

---

## Phase 1: 锁管理工具函数实现

- [ ] 1.1 创建锁管理工具函数文件

  - File: `main/pkg/lock/lock_util.go`
  - Purpose: 创建多订单锁管理工具函数
  - Requirements: 4.1, 4.2, 4.3, 4.4
  - Leverage: 现有锁接口: `main/pkg/lock/system_lock.go`，锁实现: `main/pkg/lock/lock_redsync.go`
  - Prompt: Role: Go Developer specializing in concurrency control | Task: 创建 lock_util.go，实现 sortAndDeduplicateUuids、LockMultipleUuids 和 UnlockMultipleUuids 函数 | Context: 抽取排序和去重逻辑为公用方法 sortAndDeduplicateUuids，LockMultipleUuids 和 UnlockMultipleUuids 都使用这个公用方法，按 UUID 升序获取锁，按相反顺序释放锁 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: 工具函数创建成功，公用方法抽取正确，去重和排序逻辑正确

- [ ] 1.2 实现 sortAndDeduplicateUuids 公用方法

  - File: `main/pkg/lock/lock_util.go`
  - Purpose: 实现 UUID 列表的去重和排序公用方法
  - Requirements: 4.3, 4.4
  - Leverage: Task 1.1 的文件
  - Prompt: Role: Go Developer | Task: 实现 sortAndDeduplicateUuids 函数，对 UUID 列表进行去重和排序 | Context: 自动去重（使用 map），自动过滤无效 UUID（0 值），按 UUID 升序排序，返回排序后的唯一 UUID 列表 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 函数实现正确，去重和排序逻辑正确

- [ ] 1.3 实现 LockMultipleUuids 函数

  - File: `main/pkg/lock/lock_util.go`
  - Purpose: 实现锁定多个 UUID 的函数（按 UUID 排序）
  - Requirements: 4.1, 4.3, 4.4
  - Leverage: Task 1.2 的 sortAndDeduplicateUuids 方法
  - Prompt: Role: Go Developer | Task: 实现 LockMultipleUuids 函数，支持锁定多个 UUID | Context: 调用 sortAndDeduplicateUuids 进行去重和排序，依次调用 lock.LockUuid()，返回排序后的 UUID 列表 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 函数实现正确，使用公用方法，返回排序后的列表

- [ ] 1.4 实现 UnlockMultipleUuids 函数

  - File: `main/pkg/lock/lock_util.go`
  - Purpose: 实现释放多个 UUID 锁的函数（按相反顺序）
  - Requirements: 4.2
  - Leverage: Task 1.2 的 sortAndDeduplicateUuids 方法
  - Prompt: Role: Go Developer | Task: 实现 UnlockMultipleUuids 函数，支持释放多个 UUID 锁 | Context: 调用 sortAndDeduplicateUuids 进行去重和排序（确保与 LockMultipleUuids 使用相同的策略），按相反顺序释放锁（从后往前），依次调用 lock.UnlockUuid() | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 函数实现正确，使用公用方法，释放顺序正确

- [ ] 1.5 编写工具函数单元测试

  - File: `main/pkg/lock/lock_util_test.go`
  - Purpose: 确保工具函数的正确性
  - Requirements: 4.1, 4.2, 4.3, 4.4
  - Leverage: 现有测试: `main/pkg/lock/system_lock_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 lock_util.go 编写单元测试，覆盖率 100% | Context: 测试 sortAndDeduplicateUuids 的去重和排序逻辑，测试 LockMultipleUuids 的锁获取顺序，测试 UnlockMultipleUuids 的锁释放顺序（相反顺序），测试过滤无效 UUID，测试公用方法的一致性 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 100%，所有测试通过

---

## Phase 2: 并台操作锁机制修改

- [x] 2.1 修改并台操作的锁机制

  - File: `main/app/service/desk.go:799-1010`
  - Purpose: 移除 companyUuid 锁，改为锁定所有涉及的订单
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 4.5
  - Leverage: 现有代码: `main/app/service/desk.go:799-1010`，Task 1.1-1.3 的工具函数
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 修改 MergeDesk 方法的锁机制，移除 companyUuid 锁，改为锁定所有涉及的订单 | Context: 收集主订单 UUID 和被合并桌台的订单 UUID，使用 LockMultipleUuids 锁定所有订单，使用 defer 和 UnlockMultipleUuids 释放锁 | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持现有业务逻辑不变 | Success: 锁机制修改成功，业务逻辑保持不变

- [ ] 2.2 编写并台操作单元测试

  - File: `main/app/service/desk_test.go`
  - Purpose: 确保并台操作的锁机制正确
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 现有测试: `main/app/service/desk_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为修改后的 MergeDesk 方法编写单元测试 | Context: 测试多订单锁的获取和释放，测试锁获取顺序，测试业务逻辑正确性 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

---

## Phase 3: 转台操作锁机制修改

- [x] 3.1 修改转台操作的锁机制

  - File: `main/app/service/desk.go:694-797`
  - Purpose: 增加目标桌台的锁，避免与开台操作并发冲突
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 4.5
  - Leverage: 现有代码: `main/app/service/desk.go:694-797`，开台操作: `main/app/service/order_base.go:96-102`，Task 1.1-1.3 的工具函数
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 修改 ChangeDesk 方法的锁机制，增加目标桌台的锁 | Context: 收集源订单 UUID 和目标桌台 UUID，使用 LockMultipleUuids 锁定（按 UUID 排序），使用 defer 和 UnlockMultipleUuids 释放锁，确保与开台操作使用相同的桌台锁 | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持现有业务逻辑不变 | Success: 锁机制修改成功，业务逻辑保持不变

- [ ] 3.2 编写转台操作单元测试

  - File: `main/app/service/desk_test.go`
  - Purpose: 确保转台操作的锁机制正确
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: 现有测试: `main/app/service/desk_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为修改后的 ChangeDesk 方法编写单元测试 | Context: 测试源订单和目标桌台的锁获取和释放，测试锁获取顺序，测试业务逻辑正确性 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [x] 3.3 编写转台操作与开台操作的并发测试

  - File: `test/integration/desk_concurrency_test.go`
  - Purpose: 验证转台操作与开台操作同时操作同一桌台的并发场景
  - Requirements: 2.2
  - Leverage: 现有测试文件
  - Prompt: Role: QA Engineer specializing in concurrency testing | Task: 编写转台操作与开台操作的并发测试 | Context: 模拟转台操作和开台操作同时操作同一桌台，验证转台操作锁定目标桌台后，开台操作会被正确阻塞，验证数据一致性 | Restrictions: 使用 goroutine 模拟并发，使用 sync.WaitGroup 等待所有 goroutine 完成 | Success: 并发测试通过，验证了锁的正确性和数据一致性

---

## Phase 4: 转菜操作锁机制修改

- [x] 4.1 修改转菜操作的锁机制

  - File: `main/app/service/order_product.go:1060-1220`
  - Purpose: 增加目标订单的锁
  - Requirements: 3.1, 3.2, 3.3, 3.4, 4.5
  - Leverage: 现有代码: `main/app/service/order_product.go:1060-1220`，DeskRepository: `main/app/repository/desk.go`，Task 1.1-1.3 的工具函数
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 修改 InstantOrderCartProductChangeDesk 方法的锁机制，增加目标订单的锁 | Context: 通过目标桌台 UUID 查询目标订单 UUID，收集源订单 UUID 和目标订单 UUID，使用 LockMultipleUuids 锁定（按 UUID 排序），使用 defer 和 UnlockMultipleUuids 释放锁，处理目标桌台没有关联订单的错误场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc，保持现有业务逻辑不变 | Success: 锁机制修改成功，业务逻辑保持不变，错误处理正确

- [ ] 4.2 编写转菜操作单元测试

  - File: `main/app/service/order_product_test.go`
  - Purpose: 确保转菜操作的锁机制正确
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: 现有测试: `main/app/service/order_product_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为修改后的 InstantOrderCartProductChangeDesk 方法编写单元测试 | Context: 测试源订单和目标订单的锁获取和释放，测试锁获取顺序，测试目标桌台没有关联订单的错误场景，测试业务逻辑正确性 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

---

## Phase 5: 测试和优化

- [x] 5.1 死锁测试

  - File: `test/integration/deadlock_test.go`
  - Purpose: 验证所有操作都按订单 UUID 排序获取锁，避免死锁
  - Requirements: 所有功能需求
  - Leverage: 现有测试文件
  - Prompt: Role: QA Engineer specializing in deadlock testing | Task: 编写死锁测试，验证所有操作都按订单 UUID 排序获取锁 | Context: 模拟多个操作同时涉及相同订单，验证不会发生死锁，验证锁获取顺序一致 | Restrictions: 使用 goroutine 模拟并发，监控死锁情况 | Success: 死锁测试通过，验证了锁获取顺序的一致性

- [ ] 5.2 集成测试

  - File: `test/integration/table_operations_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试并台、转台、转菜操作的完整流程，测试多订单锁的获取和释放逻辑，测试数据一致性 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 5.3 并发性能测试

  - File: `test/performance/concurrency_test.go`
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具
  - Prompt: Role: Performance Engineer | Task: 编写并发性能测试 | Context: 测试不同订单的操作可以并发执行，测试并台操作锁定多个订单的性能影响，测试本地响应时间 < 200ms（不含锁等待时间） | Restrictions: 使用压力测试工具 | Success: 性能测试通过，响应时间达标

- [ ] 5.4 代码审查

  - File: 所有修改的文件
  - Purpose: 确保代码质量和规范遵循
  - Requirements: 所有功能需求
  - Leverage: 代码审查清单
  - Prompt: Role: Code Reviewer | Task: 审查所有修改的代码 | Context: 检查锁获取顺序一致性，检查错误处理，检查代码规范遵循，检查业务逻辑正确性 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 代码审查通过，所有问题已修复

- [ ] 5.5 文档更新

  - File: `docs/shared/api/desk_api.md`, `docs/shared/api/order_api.md`（如有需要）
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: 如有 API 变更，更新 API 文档；更新设计文档中的实现细节 | Restrictions: 文档准确完整 | Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
  - 工具函数: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有变更）
- [ ] 设计文档已更新
- [ ] CHANGELOG.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-main-table-multi-order-lock/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-main-table-multi-order-lock/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-main-table-multi-order-lock/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-main-table-multi-order-lock/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-main-table-multi-order-lock/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go 后端开发

```
Role: Go Developer specializing in {具体领域}

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/api.mdc, .cursor/rules/database.mdc

Restrictions:
- 接口以 I 开头，实现以 Impl 结尾
- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- URL 使用 snake_case
- data 字段必须是对象
- 不使用 panic，返回 error
- 使用 errors.WithMessage 包装错误
- 锁的获取和释放必须在同一个函数中，使用 defer 确保释放

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70% (Service) 或 ≥ 80% (Repository) 或 100% (工具函数)
```

### 测试工程师

```
Role: QA Engineer with {Go} testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Service) 或 ≥ 80% (Repository) 或 100% (工具函数)

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- 并发场景测试（如适用）
- 死锁测试（如适用）

Restrictions:
- 遵循 .cursor/rules/go-main.mdc
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-11/2025-11-26.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-26  
**维护者**: xiezhihuan

