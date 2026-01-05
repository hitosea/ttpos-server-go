# ERP 文档初始化支持更新模式 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | rikugun   |
| **日期**   | 2026-01-04   |
| **目标版本** | v2.14.0 |
| **状态**   | 已批准 - 进入需求阶段   |
| **关联任务** | - |
| **关联 Spec** | [story-erp-init-documents-update](../../../shared/specs/active/story-erp-init-documents-update/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

当前 `ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup.go` 中的 `initDocumentsFromDir` 方法在初始化 ERPNext 文档时，只支持创建（Create）操作。当 JSON 文件中的文档已经存在于 ERPNext 系统中时，会导致创建失败，无法更新已有文档的配置。

这在以下场景中造成问题：
1. **重复初始化**：当需要重新运行初始化流程时，已存在的文档会导致失败
2. **配置更新**：当需要更新已有文档的字段或配置时，无法通过初始化流程完成
3. **版本升级**：在 ERP 系统版本升级时，需要更新自定义字段等配置，但当前只能手动操作

### 业务价值

解决这个问题能带来以下业务价值：

- **提升运维效率**：支持幂等性操作，可以安全地重复执行初始化流程
- **简化配置管理**：通过 JSON 文件统一管理文档配置，支持版本控制和批量更新
- **降低错误率**：减少手动操作 ERPNext 后台的需求，避免人为配置错误
- **加速版本升级**：在系统升级时可以自动更新文档配置，无需手动逐个修改

### 目标用户

- [x] 运维人员
- [x] 开发人员（ERP 集成开发）
- [ ] 收银员
- [ ] 商户管理员
- [ ] 厨房人员
- [ ] 顾客

---

## 💡 解决方案概述

### 方案描述

调整 `initDocumentsFromDir` 方法的逻辑，在处理 JSON 文件时：

1. **检查文档名称**：读取 JSON 数据中的 `name` 字段
2. **判断操作类型**：
   - 如果 `name` 字段不为空，说明是更新已有文档，调用 `service.Document().Update()`
   - 如果 `name` 字段为空，说明是创建新文档，调用 `service.Document().Create()`
3. **保持向后兼容**：不影响现有的创建流程，只是增加更新支持

这样既支持新文档的创建，也支持已有文档的更新，使初始化流程具有幂等性。

### 核心功能点

1. **智能判断**：根据 JSON 数据中的 `name` 字段自动判断是创建还是更新
2. **更新支持**：当文档已存在时，使用 Update 方法更新配置
3. **向后兼容**：不影响现有的创建流程和 JSON 文件格式
4. **错误处理**：保持现有的错误日志记录机制

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
- [x] API 接口（ERPNext 文档操作）
- [ ] 数据模型
- [x] 业务逻辑（初始化流程）
- [ ] 第三方集成
- [ ] 其他: ________

**备注**：这是一个纯后端的内部优化，不涉及任何前端界面或用户交互。

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯业务逻辑调整，无架构变更
- [ ] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**理由**：只需要在现有方法中增加一个条件判断，调用不同的 service 方法，不涉及复杂的业务逻辑或架构调整。

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1（待技术评审确认）

**工作内容**：
1. 修改 `initDocumentsFromDir` 方法逻辑（30 分钟）
2. 测试创建和更新两种场景（1 小时）
3. 更新相关文档和注释（30 分钟）
4. 代码审查和提交（30 分钟）

### 风险识别

**潜在风险**：
1. **Update 方法行为不确定**：需要确认 `service.Document().Update()` 的具体行为和参数要求
2. **JSON 数据格式兼容性**：需要确保现有的 JSON 文件格式符合 Update 方法的要求

**缓解措施**：
1. 在开发前先调研 `service.Document().Update()` 的 API 文档和使用示例
2. 编写测试用例覆盖创建和更新两种场景
3. 保持错误处理机制，记录详细的日志信息

---

## 🔗 相关资源

### 参考需求

- 类似功能: 无
- 竞品分析: 无

### 相关文档

- 源代码文件: `ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup.go`
- ERPNext API 文档: https://frappeframework.com/docs/user/en/api
- GoFrame 开发规范: `.cursor/rules/go-bmp.mdc`

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | -      |           |
| 技术负责人   | rikugun |           |
| 开发代表     | -      |           |
| 测试代表     | -      |           |
| UI/UX 设计师 | -      |           |

### 评审结论

- [x] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
2026-01-04 自动批准：
- 技术复杂度低，工作量评估合理（0.5 天，SP=1）
- 方案清晰，风险可控
- 向后兼容，不影响现有功能
- 批准进入需求确认阶段
```

**下一步行动**：

- [x] 创建 Spec：`story-erp-init-documents-update`
- [x] 分配负责人：rikugun
- [ ] 目标 Sprint：Sprint N
- [ ] 产品审核：待审核 requirements.md
- [ ] 技术设计：产品审核通过后执行 `/spec-design`

---

## 📝 附录

### User Story（初稿）

**作为** 运维人员/开发人员  
**我想** 能够通过 JSON 文件更新已有的 ERPNext 文档配置  
**以便于** 实现配置的版本控制和批量更新，提升运维效率

### AC 验收标准（初稿）

1. **WHEN** JSON 文件中的 `name` 字段不为空 **THEN** 系统 **SHALL** 调用 `service.Document().Update()` 更新文档
2. **WHEN** JSON 文件中的 `name` 字段为空 **THEN** 系统 **SHALL** 调用 `service.Document().Create()` 创建文档
3. **WHEN** 更新操作失败 **THEN** 系统 **SHALL** 记录详细的错误日志
4. **WHEN** 重复执行初始化流程 **THEN** 系统 **SHALL** 能够成功更新已有文档而不报错

### 技术实现要点

#### 修改前的代码逻辑

```go
// 调用service.Document.Create创建文档
if _, err := service.Document().Create(ctx, config.DocType, docData); err != nil {
    g.Log().Error(ctx, fmt.Sprintf("创建%s失败", config.ItemName), err, g.Map{"file": path, "data": docData})
    //return gerror.Wrapf(err, "创建%s失败: %s", config.ItemName, path)
}
```

#### 修改后的代码逻辑（伪代码）

```go
// 检查 docData 中的 name 字段
docName, hasName := docData["name"].(string)

if hasName && docName != "" {
    // name 不为空，调用 Update 方法
    if _, err := service.Document().Update(ctx, config.DocType, docName, docData); err != nil {
        g.Log().Error(ctx, fmt.Sprintf("更新%s失败", config.ItemName), err, g.Map{"file": path, "data": docData})
    } else {
        g.Log().Infof(ctx, "%s更新成功: %s", config.ItemName, path)
    }
} else {
    // name 为空，调用 Create 方法
    if _, err := service.Document().Create(ctx, config.DocType, docData); err != nil {
        g.Log().Error(ctx, fmt.Sprintf("创建%s失败", config.ItemName), err, g.Map{"file": path, "data": docData})
    } else {
        g.Log().Infof(ctx, "%s创建成功: %s", config.ItemName, path)
    }
}
```

### 测试场景

1. **场景 1：创建新文档**
   - JSON 文件中 `name` 字段为空
   - 预期：调用 Create 方法，成功创建文档

2. **场景 2：更新已有文档**
   - JSON 文件中 `name` 字段不为空
   - 文档已存在于 ERPNext 系统中
   - 预期：调用 Update 方法，成功更新文档配置

3. **场景 3：重复执行初始化**
   - 第一次执行：创建文档
   - 第二次执行：更新文档
   - 预期：两次执行都成功，无错误

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 技术团队提出改进方案
- ✅ 优化现有功能
- ✅ 修复设计缺陷

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
**创建日期**: 2026-01-04  
**维护者**: rikugun  
**相关规范**: `.cursor/rules/go-bmp.mdc`, `.cursor/rules/specs.mdc`

