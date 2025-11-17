# Agent 专用文档

> 为 AI Agent 自动化执行设计的工作流和模板

---

## 📋 文档特征

- **目标：** 快速执行，精准定位
- **长度：** <300 行
- **风格：** 步骤检查清单 + IF-THEN 决策树 + 快速命令
- **格式：** YAML/表格/代码块优先

---

## 📂 目录结构

### [工作流程](./workflows/)

**用途：** 具体场景的执行清单

| 工作流     | 文件                                                                   | 适用场景              |
| ---------- | ---------------------------------------------------------------------- | --------------------- |
| 需求管理   | [requirement-management.md](./workflows/requirement-management.md)     | 需求发起、评审、确认  |
| 功能开发   | [feature-development.md](./workflows/feature-development.md)           | 新增功能、优化功能    |
| Bug 修复   | [bug-fixing.md](./workflows/bug-fixing.md)                             | 修复缺陷、解决问题    |
| API 对接   | [api-integration.md](./workflows/api-integration.md)                   | 第三方集成、接口开发  |
| 数据库迁移 | [database-migration.md](./workflows/database-migration.md)             | 新增表、修改字段      |
| 微服务集成 | [microservice-integration.md](./workflows/microservice-integration.md) | gRPC 服务、Nacos 注册 |

### [文档模板](./templates/)

**用途：** 标准化文档结构

| 模板类型   | 文件                                                                         | 适用场景       |
| ---------- | ---------------------------------------------------------------------------- | -------------- |
| API 文档   | [api-doc-template.md](./templates/api-doc-template.md)                       | API 接口文档   |
| 需求规格   | [requirements-template.md](./templates/requirements-template.md)             | 需求定义       |
| 技术设计   | [design-template.md](./templates/design-template.md)                         | 技术方案       |
| 任务分解   | [tasks-template.md](./templates/tasks-template.md)                           | 任务清单       |
| 问题排查   | [troubleshooting-guide.md](./templates/troubleshooting-guide.md)             | 故障处理       |
| 技术决策   | 待补充（模板未提供）                                                         | ADR 记录       |
| 需求提案   | [proposal-template.md](./templates/proposal-template.md)                     | 需求提案       |
| 数据库迁移 | [database-migration-template.md](./templates/database-migration-template.md) | PHP Phinx 迁移 |
| gRPC 服务  | [grpc-service-template.md](./templates/grpc-service-template.md)             | Protobuf 定义  |
| 迁移指南   | 待补充（模板未提供）                                                         | 版本迁移       |
| 测试报告   | 待补充（模板未提供）                                                         | 测试报告       |

### [Graphiti 草稿](./graphiti/)

**用途：** 存放基于 `docs/agent/templates/graphiti-episode.md` 生成的 Episode 草稿，准备通过 MCP 入库。  
**使用场景：**
- 耗时 >30 分钟的排障/设计
- 需要跨团队复用的经验
- 重大决策沉淀

**互链要求：**
- 在相关文档末尾添加 `Related Episode: {name}`
- 在 Episode `Related Docs` 中列出对应 Spec / Troubleshooting / Report

---

## 🚀 使用方式

### Agent 自动执行

1. **场景识别：** Agent 根据用户描述识别场景
2. **查找工作流：** 在 `workflows/` 中找到对应文档
3. **按步执行：** 逐步检查清单执行
4. **使用模板：** 需要创建文档时，使用 `templates/` 中的模板

### 工作流调用示例

```yaml
用户: "我要开发一个快捷支付功能"
Agent识别: 功能开发场景
执行步骤: 1. 读取 workflows/feature-development.md
  2. 按照8步流程执行
  3. 使用 templates/requirements-template.md 创建需求文档
  4. 使用 templates/tasks-template.md 分解任务
  5. 逐任务实现并验证
```

---

## ⚠️ 注意事项

### 文档编写规范

```yaml
DO (应该):
  - 使用步骤检查清单 ✓
  - 使用 IF-THEN 决策树 ✓
  - 提供快速命令 ✓
  - 控制在300行以内 ✓

DON'T (不应该):
  - 写冗长解释 ✗
  - 讨论"为什么" ✗
  - 超过300行 ✗
  - 混合Agent和人类内容 ✗
```

### 多语言支持

所有模板都支持 Go/PHP/Vue 三种技术栈，在使用时需要选择：

```markdown
## 适用语言

- [x] Go (main/)
- [x] Go (ttpos-bmp GoFrame)
- [ ] PHP (admin/)
- [ ] Vue (admin/views/)
```

---

## 🔗 相关资源

### 核心规范

- [Agent 速查表](../../AGENT.md) - 所有规则的压缩映射表
- [工作流导航](../../.cursor/rules/workflows.mdc) - 场景识别和流程导航

### 人类参考资料

- [学习指南](../human/guides/) - 详细教程（人类专用）
- [架构设计](../human/architecture/) - 系统设计（人类专用）

---

**最后更新:** 2025-11-16
