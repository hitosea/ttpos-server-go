# story-erp-init-documents-update

> ERP 文档初始化支持更新模式

---

## 📋 Spec 状态

| 项目         | 内容                                                                                       |
| ------------ | ------------------------------------------------------------------------------------------ |
| **Spec 名称** | story-erp-init-documents-update                                                            |
| **当前阶段** | 需求确认（产品审核）                                                                       |
| **创建日期** | 2026-01-04                                                                                 |
| **负责人**   | rikugun                                                                                    |
| **来源**     | [Proposal](../../../../team/proposals/2026-01/erp-init-documents-update-support.md)        |

---

## 📂 文档结构

```
story-erp-init-documents-update/
├── README.md           # 本文件，Spec 状态说明
├── requirements.md     # ✅ 已创建 - 需求规格文档（已通过审核）
├── design.md           # ✅ 已创建 - 技术设计文档
└── tasks.md            # ✅ 已创建 - 任务分解清单
```

---

## 🔄 工作流状态

### 当前阶段：开发实现

```
Proposal → /spec-create → 产品审核 → 【/spec-design】→ 开发实现
             ↑ 已完成       ↑ 已完成     ↑ 已完成        ↑ 当前阶段
```

---

## ✅ 已完成

- [x] Proposal 创建和评审
- [x] `/spec-create` 创建需求文档
- [x] 建立 Proposal ↔ Spec 双向链接
- [x] 填充基本信息和用户故事
- [x] 定义验收标准
- [x] 评估技术复杂度和 Story Point (SP=1)
- [x] 产品审核通过
- [x] `/spec-design` 创建技术设计和任务分解
- [x] 技术方案设计完成
- [x] 任务分解完成

---

## ⏳ 待处理

### 开发实现阶段

- [ ] Task 1.1: 修改 `initDocumentsFromDir` 方法
- [ ] Task 1.2: 更新方法注释
- [ ] Task 2.1: 编写单元测试
- [ ] Task 2.2: 手动集成测试
- [ ] Task 3.1: 运行代码检查
- [ ] Task 3.2: 更新相关文档
- [ ] Task 3.3: 清理测试数据

---

## 🎯 下一步操作

### 开发实现

1. **执行任务 1.1**: 修改 `initDocumentsFromDir` 方法
   - 文件: `ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup.go` (第 520-527 行)
   - 参考: `tasks.md` 中的详细 Prompt 和实现步骤

2. **执行任务 1.2**: 更新方法注释
   - 说明支持创建和更新两种模式

3. **执行任务 2.1**: 编写单元测试
   - 测试覆盖率目标: ≥ 80%
   - 6 个测试用例

4. **执行任务 2.2**: 手动集成测试
   - 在开发环境验证功能

5. **代码审查和提交**
   - 运行代码检查
   - 提交代码（参考 `.cursor/rules/version.mdc`）

---

## 📊 Story Point 评估

| 维度         | 评分 | 说明                     |
| ------------ | ---- | ------------------------ |
| 技术复杂度   | 1    | 简单条件判断，无架构变更 |
| 功能复杂度   | 1    | 单一功能点，逻辑清晰     |
| 不确定性     | 0    | Update 方法行为明确      |
| **总计 SP**  | **1** | ≤ 5，可进入 Sprint       |

---

## 🔗 相关链接

### 文档链接

- **Proposal**: [docs/team/proposals/2026-01/erp-init-documents-update-support.md](../../../../team/proposals/2026-01/erp-init-documents-update-support.md)
- **Requirements**: [requirements.md](./requirements.md)
- **源代码**: `ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup.go`

### 规范参考

- `.cursor/rules/go-bmp.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/go-rules.mdc` - GoFrame 开发规范
- `.cursor/rules/specs.mdc` - Spec 规范

---

## 📝 备注

### 核心改进

调整 `initDocumentsFromDir` 方法，使其能够：
- 检查 JSON 数据中的 `name` 字段
- 当 `name` 不为空时调用 `service.Document().Update()` 更新
- 当 `name` 为空时调用 `service.Document().Create()` 创建
- 实现配置管理的幂等性

### 业务价值

- ✅ 支持重复执行初始化流程
- ✅ 简化配置管理和版本升级
- ✅ 降低手动操作错误率
- ✅ 通过 JSON 文件统一管理配置

---

**最后更新**: 2026-01-04  
**更新人**: rikugun  
**状态**: 开发实现中

