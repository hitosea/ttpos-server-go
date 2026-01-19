# {功能名称} 任务分解

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 0
**已完成**: 0
**完成率**: 0%

---

## Phase 1: {阶段名称}（如：核心实现）

- [ ] 1.1 {任务标题}

  - File: `{文件路径}`
  - Purpose: {任务目的}
  - Requirements: {需求编号}
  - Leverage: `{可复用代码路径}`

- [ ] 1.2 {任务标题}

  - File: `{文件路径}`
  - Purpose: {任务目的}
  - Requirements: {需求编号}
  - Leverage: Task 1.1 的实现

- [ ] 1.3 编写单元测试
  - File: `test/{对应路径}/{name}_test.dart`
  - Purpose: 确保代码质量
  - Requirements: {需求编号}
  - Leverage: 现有测试模板

---

## Phase 2: {阶段名称}（如：应用集成）

- [ ] 2.1 {任务标题} - POS 端集成

  - File: `apps/pos/lib/{路径}`
  - Purpose: 在 POS 端集成功能
  - Requirements: {需求编号}
  - Leverage: Phase 1 实现

- [ ] 2.2 {任务标题} - Web 端集成

  - File: `apps/{mobile|menu|member}/lib/{路径}`
  - Purpose: 在 Web 端集成功能
  - Requirements: {需求编号}
  - Leverage: Phase 1 实现 (确保 Web 兼容)

---

## Phase 3: {阶段名称}（如：测试优化）

- [ ] 3.1 集成测试

  - File: `integration_test/{name}_test.dart`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求

- [ ] 3.2 文档更新
  - File: `docs/{相关文档}`, `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求

---

## 提交清单

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] 代码通过 `melos run lint`
- [ ] 测试覆盖率达标
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现

### 平台兼容

- [ ] Android 测试通过
- [ ] iOS 测试通过
- [ ] Web 测试通过（HTML 渲染器）

---

**模板版本**: v1.0.0
