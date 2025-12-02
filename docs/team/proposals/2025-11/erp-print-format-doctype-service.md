# ERP 增加 Print Format 的 Doctype 通用服务支持 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目          | 内容                                                                                                                                     |
| ------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| **提案人**    | rikugun                                                                                                                                  |
| **日期**      | 2025-11-24                                                                                                                               |
| **目标版本**  | v1.x.x                                                                                                                                   |
| **状态**      | ✅ 已创建 Spec                                                                                                                           |
| **关联任务**  | -                                                                                                                                        |
| **关联 Spec** | [docs/shared/specs/active/story-erp-print-format-doctype-support/](../../../shared/specs/active/story-erp-print-format-doctype-support/) |

---

## 🎯 背景和动机

### 问题描述

当前 ERP 模块中已有多个 Print Format 相关的 HTML 模板文件（如销售订单、采购收货单、送货单等），但缺少统一的 Print Format Doctype 通用服务支持。各个业务模块需要打印功能时，都需要单独实现，导致代码重复和维护成本高。

**现状**：

- `ttpos-bmp/app/ttpos-erp/manifest/printformat/html/` 目录下已有多个打印模板
- 缺少统一的 Print Format 服务接口
- 各业务模块无法复用打印格式管理功能

### 业务价值

- **统一管理**：提供统一的 Print Format 服务接口，便于各业务模块复用
- **降低维护成本**：减少代码重复，统一管理打印格式
- **提升开发效率**：新业务模块可以快速接入打印功能
- **标准化**：遵循 ERPNext 的 Print Format Doctype 规范

### 目标用户

- [ ] 收银员
- [x] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [x] 开发人员

---

## 💡 解决方案概述

### 方案描述

参考 `logic` 目录下现有服务的实现模式（如 `erpnext/doctype.go`、`selling/sale_order.go`），在 ERP 模块中新增 Print Format 的 Doctype 通用服务支持。

**实现思路**：

1. 在 `logic/erpnext/` 目录下新增 `print_format.go` 服务文件
2. 实现 Print Format 的 CRUD 操作（创建、查询、更新、删除）
3. 支持获取 Print Format 的元数据（Meta）
4. 支持根据 DocType 查询对应的 Print Format 列表
5. 在 `service/erpnext.go` 中注册 Print Format 服务接口

### 核心功能点

1. **Print Format 元数据查询**：获取 Print Format 的字段定义和结构信息
2. **Print Format 列表查询**：根据 DocType 查询对应的 Print Format 列表
3. **Print Format 详情查询**：根据名称查询 Print Format 的详细信息
4. **Print Format 创建/更新**：支持创建和更新 Print Format
5. **Print Format 删除**：支持删除 Print Format（软删除）

### 影响范围

**涉及终端**：

- [ ] POS 收银端
- [x] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：

- [ ] UI 组件
- [x] API 接口
- [ ] 数据模型
- [x] 业务逻辑
- [x] 第三方集成（ERPNext）
- [ ] 其他: **\_\_\_\_**

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 3-5 天
- **预估 SP**: 5 SP（待技术评审确认）

### 风险识别

**潜在风险**：

1. ERPNext Print Format API 的兼容性问题
2. 打印模板格式的标准化处理
3. 多语言支持的实现复杂度

**缓解措施**：

1. 参考现有 `doctype.go` 的实现模式，确保 API 调用方式一致
2. 先实现基础功能，再逐步完善高级特性
3. 充分测试各种 DocType 的 Print Format 场景

---

## 🔗 相关资源

### 参考需求

- 类似功能: `logic/erpnext/doctype.go` - Doctype 服务实现
- 竞品分析: ERPNext Print Format 官方文档

### 相关文档

- ERPNext Print Format API: `/api/v2/doctype/Print Format`
- 现有打印模板: `ttpos-bmp/app/ttpos-erp/manifest/printformat/html/`
- 服务注册机制: `logic/erpnext/erpnext.go`

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

- [x] 创建 Spec：`story-erp-print-format-doctype-support`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 开发人员  
**我想** 使用统一的 Print Format 服务接口  
**以便于** 快速为各业务模块接入打印功能，减少重复代码

### AC 验收标准（初稿）

1. **WHEN** 调用 Print Format Meta API **THEN** 系统 **SHALL** 返回 Print Format 的元数据信息
2. **WHEN** 根据 DocType 查询 Print Format 列表 **THEN** 系统 **SHALL** 返回该 DocType 对应的所有 Print Format
3. **WHEN** 根据名称查询 Print Format 详情 **THEN** 系统 **SHALL** 返回完整的 Print Format 信息（包括 HTML 模板内容）
4. **IF** 创建/更新 Print Format **THEN** 系统 **SHALL** 成功保存并返回结果
5. **IF** 删除 Print Format **THEN** 系统 **SHALL** 执行软删除操作

### 技术实现参考

**参考现有服务实现**：

- `logic/erpnext/doctype.go` - Doctype 服务实现模式
- `logic/selling/sale_order.go` - 业务服务实现模式
- `service/erpnext.go` - 服务注册机制

**API 端点**：

- Meta: `/api/v2/doctype/Print Format/meta`
- List: `/api/v2/document/Print Format`
- Get: `/api/v2/document/Print Format/{name}`
- Create/Update: `/api/v2/document/Print Format` (POST/PUT)
- Delete: `/api/v2/document/Print Format/{name}` (DELETE)

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段         | 文档类型     | 详细程度 | 用途                      |
| ------------ | ------------ | -------- | ------------------------- |
| **需求发起** | Proposal     | 粗略     | 团队评审、决策是否做      |
| **需求确认** | Requirements | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design       | 详细     | 技术方案，实现指导        |
| **任务分解** | Tasks        | 详细     | 开发执行，进度追踪        |

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
**创建日期**: 2025-11-24  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`, `.cursor/rules/go-bmp.mdc`
