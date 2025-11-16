# 后端文档缺失清单

> 对比前端仓库 (ttpos-flutter) 与后端仓库 (ttpos-server-go) 的文档差异

**生成日期**: 2025-11-16  
**对比基准**: ttpos-flutter/docs/  
**目标仓库**: ttpos-server-go/docs/

---

## 📊 缺失统计概览

| 类别             | 前端已有 | 后端已有 | 缺失数量 | 缺失率 |
| ---------------- | -------- | -------- | -------- | ------ |
| **Agent 模板**   | 18个     | 0个      | 18个     | 100%   |
| **Agent 工作流** | 11个     | 6个      | 5个      | 45%    |
| **人类架构文档** | 6个      | 6+个     | 5个      | ~50%   |
| **人类业务文档** | 3+个     | 0个      | 3+个     | 100%   |
| **人类指南**     | 4个      | 1个      | 3个      | 75%    |
| **人类测试文档** | 6个      | 0个      | 6个      | 100%   |
| **共享API文档**  | 3+个     | 1个      | 2+个     | ~67%   |
| **共享问题排查** | 3个      | 0个      | 3个      | 100%   |
| **团队运营文档** | 1+个     | 0个      | 1+个     | 100%   |
| **团队报告**     | 5+个     | 0个      | 5+个     | 100%   |

**总计**: 约 **51+ 个文档** 缺失

---

## 🤖 Agent 层缺失文档

### 一、模板文件 (18个全部缺失)

#### 1.1 Spec 文档模板 ⭐⭐⭐ (P0)
```
缺失文件:
✗ docs/agent/templates/requirements-template.md
✗ docs/agent/templates/design-template.md
✗ docs/agent/templates/tasks-template.md
```

**影响**: 
- Agent 无法自动创建标准化的 Spec 文档
- `/create-spec` 命令无法正常工作
- 需求文档格式不统一

**适配需求**:
- 前端模板需适配后端技术栈 (Go/PHP/Vue)
- 设计模板需包含数据库设计、API设计、微服务设计
- 任务模板需区分 Go/PHP 任务

#### 1.2 需求管理模板 ⭐⭐⭐ (P0)
```
缺失文件:
✗ docs/agent/templates/proposal-template.md
```

**影响**:
- Agent 无法自动创建需求提案
- `/propose` 命令无法正常工作
- 提案格式不标准

**适配需求**:
- 需适配后端业务场景
- 需包含数据库影响评估
- 需包含 API 接口设计概要

#### 1.3 缺陷管理模板 ⭐⭐ (P1)
```
缺失文件:
✗ docs/agent/templates/bug-report-template.md
✗ docs/agent/templates/defect-severity-guide.md
```

**影响**:
- Bug 报告格式不统一
- 缺陷等级定义不明确
- 无法自动生成标准化 Bug 报告

**适配需求**:
- 需适配后端日志格式 (SkyWalking、Go log、PHP log)
- 需包含数据库影响评估
- 需区分前后端问题

#### 1.4 测试相关模板 ⭐⭐ (P1)
```
缺失文件:
✗ docs/agent/templates/test-report-template.md
✗ docs/agent/templates/verification-checklist.md
```

**影响**:
- 测试报告格式不统一
- 验收清单不标准
- 测试流程不清晰

**适配需求**:
- 需包含后端测试覆盖率要求
- 需包含 API 测试、单元测试、集成测试
- 需区分 Go、PHP、Vue 的测试方式

#### 1.5 文档生成模板 ⭐⭐ (P1)
```
缺失文件:
✗ docs/agent/templates/api-doc.md
✗ docs/agent/templates/architecture-doc.md
✗ docs/agent/templates/business-rule.md
✗ docs/agent/templates/business-workflow.md
✗ docs/agent/templates/troubleshooting-guide.md
✗ docs/agent/templates/migration-guide.md
```

**影响**:
- Agent 无法自动生成标准化文档
- 文档格式各异
- 文档质量不可控

**适配需求**:
- api-doc: 需支持 REST API 和 gRPC API
- architecture-doc: 需包含微服务架构说明
- business-rule: 需适配后端业务逻辑
- troubleshooting-guide: 需适配后端技术栈

#### 1.6 特殊模板 ⭐ (P2)
```
缺失文件:
✗ docs/agent/templates/context-packet.md
✗ docs/agent/templates/decision-record.md
✗ docs/agent/templates/graphiti-episode.md
```

**影响**:
- 上下文传递不标准
- 技术决策记录格式不统一
- Graphiti 记录不规范

**适配需求**:
- context-packet: 通用，可直接复用
- decision-record: 需适配后端技术决策场景
- graphiti-episode: 需添加后端 Group ID

---

### 二、工作流文档 (5个缺失)

#### 2.1 测试流程 ⭐⭐⭐ (P0)
```
缺失文件:
✗ docs/agent/workflows/test-submission.md
✗ docs/agent/workflows/test-verification.md
✗ docs/agent/workflows/defect-management.md
```

**影响**:
- 提测流程不清晰
- 验收标准不明确
- 缺陷管理流程缺失

**适配需求**:
- 需包含后端测试环境配置
- 需包含 API 测试、压力测试
- 需包含数据库回滚机制

#### 2.2 团队协作 ⭐⭐ (P1)
```
缺失文件:
✗ docs/agent/workflows/onboarding.md
✗ docs/agent/workflows/proposal-spec-linking.md
```

**影响**:
- 新人入职流程不清晰
- Proposal 和 Spec 关联机制缺失

**适配需求**:
- onboarding: 需适配后端技术栈学习路径
- proposal-spec-linking: 通用，可直接复用

#### 2.3 性能优化 ⭐⭐ (P1)
```
前端已有: docs/agent/workflows/performance-optimization.md
后端缺失: ✗
```

**影响**:
- 性能问题排查流程不清晰
- 优化方法不系统

**适配需求**:
- 需适配后端性能分析工具 (pprof、Xdebug)
- 需包含数据库优化、API 优化
- 需包含微服务性能优化

---

## 👤 人类层缺失文档

### 三、架构设计文档 (5个缺失)

#### 3.1 核心架构文档 ⭐⭐⭐ (P0)
```
缺失文件:
✗ docs/human/architecture/overview.md
✗ docs/human/architecture/modules.md
✗ docs/human/architecture/code-style-guide.md
```

**影响**:
- 缺少系统架构总览
- 模块依赖关系不清晰
- 代码风格指南不完整

**现状**:
- 后端有部分架构文档 (entities/, features/, refactor/)
- 但缺少总览性、结构化的核心文档

**适配需求**:
- overview.md: 需包含三语言架构 (Go/PHP/Vue)
- modules.md: 需说明 main、admin、ttpos-bmp 的关系
- code-style-guide.md: 前端有632行详细指南，后端需创建类似文档

#### 3.2 组件指南 ⭐⭐ (P1)
```
缺失文件:
✗ docs/human/architecture/component-guidelines.md
✗ docs/human/architecture/monorepo.md
```

**影响**:
- 缺少组件设计指南
- Monorepo 管理策略不明确

**适配需求**:
- component-guidelines: 前端是 Flutter 组件，后端可改为"模块设计指南"
- monorepo.md: 需说明 main、admin、ttpos-bmp 三个仓库的协作

---

### 四、业务知识文档 (3+个缺失)

#### 4.1 核心业务文档 ⭐⭐⭐ (P0)
```
缺失文件:
✗ docs/human/business/glossary.md
```

**影响**:
- 业务术语不统一
- 新人学习业务困难

**适配需求**:
- 需包含后端特有术语 (Repository、Service、DTO、DAO)
- 需包含餐饮行业术语 (点餐、结账、桌台、厨显等)

#### 4.2 业务规则和流程 ⭐⭐ (P1)
```
前端已有:
✓ docs/human/business/rules/README.md
✓ docs/human/business/workflows/README.md

后端缺失:
✗ docs/human/business/rules/ 下无具体文档
✗ docs/human/business/workflows/ 下仅有5个文档
```

**影响**:
- 业务规则文档不完整
- 业务流程说明不足

---

### 五、开发指南 (3个缺失)

#### 5.1 快速开始指南 ⭐⭐⭐ (P0)
```
缺失文件:
✗ docs/human/guides/installation.md
✗ docs/human/guides/cursor-commands.md
✗ docs/human/guides/documentation-guide.md
```

**影响**:
- 新人环境配置困难
- Cursor 命令使用不明确
- 文档创建规范不清晰

**适配需求**:
- installation.md: 需包含 Go、PHP、MySQL、Redis、Docker 环境配置
- cursor-commands.md: 需适配后端命令
- documentation-guide.md: 通用，可直接复用

---

### 六、测试文档 (6个全部缺失)

#### 6.1 测试标准 ⭐⭐⭐ (P0)
```
缺失目录和文件:
✗ docs/human/testing/README.md
✗ docs/human/testing/standards/README.md
✗ docs/human/testing/standards/api-testing.md
✗ docs/human/testing/standards/component-testing.md
✗ docs/human/testing/standards/controller-testing.md
✗ docs/human/testing/standards/model-testing.md
```

**影响**:
- 测试覆盖率要求不明确
- 测试编写规范缺失
- 测试用例质量不可控

**适配需求**:
- api-testing.md: 需包含 REST API 和 gRPC 测试
- component-testing.md: 可改为 "service-testing.md"
- controller-testing.md: 需适配 Go Gin 和 PHP ThinkPHP
- model-testing.md: 需适配 Go struct 和 PHP Model

#### 6.2 测试示例 ⭐⭐ (P1)
```
缺失目录:
✗ docs/human/testing/examples/README.md
```

**影响**:
- 缺少测试示例参考
- 新人编写测试困难

---

## 📚 共享层缺失文档

### 七、API 文档 (2+个缺失)

#### 7.1 API 规范 ⭐⭐⭐ (P0)
```
前端已有: docs/shared/api/conventions.md
后端缺失: ✗
```

**影响**:
- API 设计规范不统一
- 响应格式不标准

**适配需求**:
- 需包含 REST API 规范
- 需包含 gRPC API 规范
- 需明确响应格式 (data 不能为 null 或数组)

#### 7.2 API 文档目录 ⭐⭐ (P1)
```
前端已有:
✓ docs/shared/api/backend/README.md
✓ docs/shared/api/backend/pos/desk_api.md
✓ docs/shared/api/backend/assistant/desk_api.md

后端缺失:
✗ 缺少按模块组织的 API 文档
```

**影响**:
- API 文档分散
- 难以查找特定模块的 API

---

### 八、问题排查文档 (3个缺失)

#### 8.1 通用问题 ⭐⭐⭐ (P0)
```
缺失文件:
✗ docs/shared/troubleshooting/common-issues.md
```

**影响**:
- 常见问题无文档记录
- 重复问题重复排查

**适配需求**:
- 需包含后端常见问题 (数据库连接、Redis、Nacos、gRPC)
- 需包含多语言问题 (Go panic、PHP error、Vue build error)

#### 8.2 平台特定问题 ⭐⭐ (P1)
```
前端已有:
✓ docs/shared/troubleshooting/platform/README.md
✓ docs/shared/troubleshooting/platform/web.md

后端缺失:
✗ 缺少按平台分类的问题文档
```

**适配需求**:
- 可改为按技术栈分类: golang/, php/, database/, microservice/

---

### 九、版本迁移文档 ⭐⭐ (P1)

```
前端已有: docs/shared/migration/README.md
后端缺失: ✗
```

**影响**:
- 版本升级无指导文档
- API 变更无迁移指南

---

## 👥 团队层缺失文档

### 十、运营文档 (1+个缺失)

#### 10.1 知识管理推广 ⭐⭐ (P1)
```
前端已有: docs/team/operations/knowledge_management_rollout.md
后端缺失: ✗
```

**影响**:
- 知识管理制度无文档
- 团队协作规范不明确

---

### 十一、团队报告 (5+个缺失)

#### 11.1 季度报告 ⭐ (P2)
```
前端已有:
✓ docs/team/reports/2025-Q3/2025-Q3-开发部门突出贡献报告.md
✓ docs/team/reports/2025-Q3/2025-Q3-项目组季度投资人报告.md
✓ docs/team/reports/2025-Q3/cursor_team_pay_rules.md
✓ docs/team/reports/2025-Q3/cursor_team_payment.202509.md
✓ docs/team/reports/2025-Q3/cursor_team_usage.md

后端缺失: ✗ (整个 reports/ 目录)
```

**影响**:
- 缺少季度总结
- 团队贡献无记录

---

### 十二、其他目录 ⭐ (P2)

```
前端已有: docs/others/ (第三方集成参考文档)
  - erpnext/
  - lineman/
  - tr/ (土耳其第三方)

后端缺失: ✗
```

**影响**:
- 第三方集成参考文档缺失
- 已有 docs/shared/integrations/lineman/ 但内容较少

---

## 📋 优先级分类汇总

### P0 - 必须补充 (1周内)

**Agent 层**:
1. ✗ 3个 Spec 模板 (requirements, design, tasks)
2. ✗ 1个提案模板 (proposal)
3. ✗ 3个测试流程工作流 (test-submission, test-verification, defect-management)

**人类层**:
4. ✗ 3个核心架构文档 (overview, modules, code-style-guide)
5. ✗ 1个业务术语表 (glossary)
6. ✗ 1个环境配置指南 (installation)
7. ✗ 6个测试文档 (testing/)

**共享层**:
8. ✗ 1个 API 规范 (conventions)
9. ✗ 1个常见问题文档 (common-issues)

**合计**: 约 **20个文档** (P0)

---

### P1 - 建议补充 (1月内)

**Agent 层**:
1. ✗ 2个缺陷管理模板 (bug-report, defect-severity-guide)
2. ✗ 2个测试模板 (test-report, verification-checklist)
3. ✗ 6个文档生成模板 (api-doc, architecture-doc, business-rule, business-workflow, troubleshooting-guide, migration-guide)
4. ✗ 2个团队协作工作流 (onboarding, proposal-spec-linking)
5. ✗ 1个性能优化工作流 (performance-optimization)

**人类层**:
6. ✗ 2个架构指南 (component-guidelines, monorepo)
7. ✗ 业务规则和流程文档
8. ✗ 2个开发指南 (cursor-commands, documentation-guide)

**共享层**:
9. ✗ 按模块组织的 API 文档
10. ✗ 按技术栈分类的问题排查文档
11. ✗ 版本迁移文档

**团队层**:
12. ✗ 知识管理运营文档

**合计**: 约 **20+个文档** (P1)

---

### P2 - 可选补充 (按需)

**Agent 层**:
1. ✗ 3个特殊模板 (context-packet, decision-record, graphiti-episode)

**团队层**:
2. ✗ 季度报告体系
3. ✗ 其他第三方参考文档

**合计**: 约 **10+个文档** (P2)

---

## 🎯 补充建议

### 立即行动 (明天开始)

#### Phase 1: Agent 核心模板 (预计1天)
```
优先级: P0
数量: 4个
文件:
- docs/agent/templates/requirements-template.md
- docs/agent/templates/design-template.md
- docs/agent/templates/tasks-template.md
- docs/agent/templates/proposal-template.md
```

#### Phase 2: Agent 测试流程 (预计0.5天)
```
优先级: P0
数量: 3个
文件:
- docs/agent/workflows/test-submission.md
- docs/agent/workflows/test-verification.md
- docs/agent/workflows/defect-management.md
```

#### Phase 3: 人类核心架构 (预计1天)
```
优先级: P0
数量: 3个
文件:
- docs/human/architecture/overview.md
- docs/human/architecture/modules.md
- docs/human/architecture/code-style-guide.md
```

#### Phase 4: 测试标准体系 (预计1天)
```
优先级: P0
数量: 6个
文件:
- docs/human/testing/README.md
- docs/human/testing/standards/README.md
- docs/human/testing/standards/api-testing.md
- docs/human/testing/standards/service-testing.md
- docs/human/testing/standards/controller-testing.md
- docs/human/testing/standards/model-testing.md
```

#### Phase 5: 共享文档核心 (预计0.5天)
```
优先级: P0
数量: 3个
文件:
- docs/shared/api/conventions.md
- docs/shared/troubleshooting/common-issues.md
- docs/human/business/glossary.md
- docs/human/guides/installation.md
```

---

### 短期计划 (1周内)

**目标**: 完成所有 P0 文档 (20个)

**执行方式**:
1. 从前端仓库复制模板
2. 适配后端技术栈
3. 填充后端特有内容
4. 建立交叉引用

**质量标准**:
- Agent 文档 <300 行
- 人类文档详细说明 WHY
- 代码示例适配 Go/PHP/Vue
- 所有占位符都已填充

---

### 中期计划 (1月内)

**目标**: 完成所有 P1 文档 (20+个)

**执行方式**:
1. 按优先级逐个补充
2. 结合实际使用反馈优化
3. 补充实际案例
4. 完善交叉引用

---

### 长期计划 (持续)

**目标**: 完成 P2 文档 + 持续优化

**执行方式**:
1. 根据实际需求补充
2. 记录最佳实践
3. 持续更新和维护
4. 定期审查和优化

---

## 🔄 适配注意事项

### 技术栈差异

| 前端 (Flutter)  | 后端 (TTPOS Server)   | 适配方式               |
| --------------- | --------------------- | ---------------------- |
| Flutter/Dart    | Go + PHP + Vue        | 所有代码示例需改写     |
| GetX Controller | Go Service/Repository | 架构模式需重新说明     |
| Freezed Model   | Go Struct + PHP Model | 数据模型文档需重写     |
| Dio API         | Go Gin + PHP ThinkPHP | API 封装规范需重写     |
| Web/Mobile 平台 | 服务端架构            | 平台问题改为技术栈问题 |

### 业务场景差异

| 前端         | 后端        | 适配方式                 |
| ------------ | ----------- | ------------------------ |
| UI 组件      | 业务逻辑    | 组件指南改为模块设计指南 |
| 客户端性能   | 服务端性能  | 性能指标和优化方法不同   |
| Flutter 测试 | Go/PHP 测试 | 测试框架和工具不同       |

### 文档风格一致性

保持与前端仓库相同的风格：
- ✅ Agent 文档 <300 行，结构化
- ✅ 人类文档详细，包含 WHY
- ✅ 明确受众标识 (🤖/👤/📚)
- ✅ 完整的交叉引用
- ✅ 清晰的示例代码

---

## 📊 工作量评估

### 文档创建工作量

| 优先级 | 文档数量 | 预估时间 | 说明                   |
| ------ | -------- | -------- | ---------------------- |
| P0     | 20个     | 5-7天    | 需从前端复制并大量适配 |
| P1     | 20+个    | 10-14天  | 部分可复用，部分需重写 |
| P2     | 10+个    | 按需补充 | 低优先级，可延后       |

### 人力建议

- **理想**: 2人全职 1周完成 P0
- **最低**: 1人全职 2周完成 P0
- **实际**: 可结合 Agent 辅助加速

---

## 🎯 成功标准

### 完成标准

- ✅ 所有 P0 文档补充完成
- ✅ 所有代码示例适配后端技术栈
- ✅ 所有交叉引用正确
- ✅ Agent 可以正常使用所有工作流和模板
- ✅ 新人可以通过文档快速上手

### 质量标准

- ✅ 文档格式符合规范
- ✅ 受众明确 (🤖/👤/📚)
- ✅ 内容准确无误
- ✅ 示例代码可运行
- ✅ 无死链接

---

## 📝 备注

### 已有但需优化的部分

虽然后端已有一些文档，但质量和结构需要提升：

1. **docs/human/architecture/entities/** - 18个实体文档
   - ✅ 已有文档
   - ⚠️ 缺少实体关系图
   - ⚠️ 缺少总览说明

2. **docs/human/architecture/features/** - 26个功能特性文档
   - ✅ 已有文档
   - ⚠️ 格式不统一
   - ⚠️ 缺少分类索引

3. **docs/human/architecture/refactor/** - 10个重构文档
   - ✅ 已有文档
   - ⚠️ 与新架构关系不清晰

4. **docs/human/business/workflows/** - 5个业务流程文档
   - ✅ 已有文档
   - ⚠️ 数量偏少
   - ⚠️ 需补充更多业务流程

### 可直接复用的文档

部分前端文档可以直接或稍作修改后复用：

- ✅ context-packet.md (通用)
- ✅ graphiti-episode.md (加后端 Group ID)
- ✅ proposal-spec-linking.md (通用)
- ✅ documentation-guide.md (通用)
- ✅ verification-checklist.md (稍作适配)

---

## 🔗 相关资源

- **前端仓库**: `/Users/ben/projects/ttpos-flutter/docs/`
- **本次迁移总结**: `docs/MIGRATION_SUMMARY.md` (已删除)
- **Agent 速查表**: `.cursor/rules/AGENT_QUICK_REF.mdc`
- **工作流导航**: `.cursor/rules/workflows.mdc`

---

**报告生成**: 2025-11-16  
**下次更新**: 明天继续补充文档时更新此报告  
**维护者**: @开发者

---

## ✅ 行动计划

明天开始执行：

1. **上午**: Phase 1 - 补充 4 个 Agent 核心模板
2. **下午**: Phase 2 - 补充 3 个测试流程工作流
3. **次日上午**: Phase 3 - 补充 3 个核心架构文档
4. **次日下午**: Phase 4 - 补充测试标准体系
5. **第三天**: Phase 5 - 补充共享文档核心

预计 **3天** 完成所有 P0 文档！🎯

