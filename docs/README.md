# TTPOS 后端文档体系

> 本文档体系采用三层架构设计，支持 Agent 自动化和人类学习的不同需求

---

## 📋 文档迁移和待办

- **[📝 快速参考](./QUICK_START.md)** - 5分钟了解文档体系现状和待办
- [文档缺失清单](./MISSING_DOCS_REPORT.md) - 对比前端仓库的详细缺失盘点（51+个文档）
- [明天的工作计划](./TODO_NEXT.md) - 3天完成 P0 文档的具体行动计划

---

## 📖 目录导航

### 🤖 [Agent 专用](./agent/)
**用途：** 工作流执行清单和结构化模板  
**受众：** AI Agent 自动化执行  
**风格：** 步骤检查清单 + 决策树 + 快速命令

- [工作流程](./agent/workflows/) - 6个核心工作流（需求管理、功能开发、Bug修复、API对接、数据库迁移、微服务集成）
- [文档模板](./agent/templates/) - 11个标准模板（支持Go/PHP/Vue多语言）

### 👤 [人类专用](./human/)
**用途：** 学习资料和系统设计文档  
**受众：** 开发者学习和参考  
**风格：** 详细解释 + WHY + HOW + 完整示例

- [学习指南](./human/guides/) - 开发指南和最佳实践
- [架构设计](./human/architecture/) - 系统架构和技术决策
- [业务知识](./human/business/) - 业务规则和工作流程
- [技术决策](./human/decisions/) - ADR 技术决策记录

### 📚 [共用资源](./shared/)
**用途：** Agent 和人类都需要的资源  
**受众：** 通用  
**风格：** 结构化 + 简洁

- [功能规格](./shared/specs/) - 需求和设计文档
- [API文档](./shared/api/) - 接口文档
- [问题排查](./shared/troubleshooting/) - 故障处理指南
- [第三方集成](./shared/integrations/) - 外部服务集成文档

### 👥 [团队协作](./team/)
**用途：** 团队沟通和项目管理  
**受众：** 团队成员  
**风格：** 正式文档

- [需求提案](./team/proposals/) - 需求提案和评审
- [活动日志](./team/activities/) - 团队活动记录

---

## 🚀 快速开始

### 我想...

| 场景               | 查看                                                                        |
| ------------------ | --------------------------------------------------------------------------- |
| **开发新功能**     | [功能开发工作流](./agent/workflows/feature-development.md)                  |
| **修复Bug**        | [Bug修复工作流](./agent/workflows/bug-fixing.md)                            |
| **创建数据库迁移** | [数据库迁移工作流](./agent/workflows/database-migration.md)                 |
| **开发gRPC服务**   | [微服务集成工作流](./agent/workflows/microservice-integration.md)           |
| **对接第三方API**  | [API对接工作流](./agent/workflows/api-integration.md)                       |
| **了解项目架构**   | [架构设计文档](./human/architecture/)                                       |
| **学习开发规范**   | [Go规范](../.cursor/rules/golang.mdc) / [PHP规范](../.cursor/rules/php.mdc) |

---

## 🎯 文档体系原则

### Agent 视角 vs 人类视角

**创建文档前，必须明确受众：**

```yaml
IF 文档主要给 Agent 阅读 THEN
  受众: 🤖 Agent
  位置: docs/agent/
  长度: <300行
  风格: 步骤检查清单 + IF-THEN决策树
  
ELSE IF 文档主要给人类学习 THEN
  受众: 👤 人类
  位置: docs/human/
  长度: 不限
  风格: 详细解释 + WHY + HOW + 示例
  
ELSE
  受众: 📚 共用
  位置: docs/shared/ 或 docs/team/
  风格: 结构化 + 简洁
```

---

## 🔗 相关资源

### 核心规范 (必读)
- [Agent速查表](../.cursor/rules/AGENT_QUICK_REF.mdc) - 所有规则的压缩映射表
- [工作流导航](../.cursor/rules/workflows.mdc) - 场景识别和流程导航
- [Go开发规范](../.cursor/rules/golang.mdc)
- [PHP开发规范](../.cursor/rules/php.mdc)
- [Vue开发规范](../.cursor/rules/vue.mdc)

### 项目概览
- [项目介绍](../.cursor/rules/intro.mdc) - 项目概览和技术栈
- [项目结构](../.cursor/rules/structs.mdc) - 目录结构和模块划分
- [README](../README.md) - 项目主文档

### ttpos-bmp 专用规范
- [GoFrame规范](../ttpos-bmp/.cursor/rules/go-rules.mdc)
- [Protobuf规范](../ttpos-bmp/.cursor/rules/proto-rules.mdc)

---

## 📝 文档维护

### 文档更新原则
1. Agent文档保持<300行
2. 所有模板保持最新
3. 定期审查文档时效性
4. 建立交叉引用避免孤岛

### 贡献指南
- 参考 [文档创建规范](../.cursor/rules/documentation.mdc)
- 使用标准模板 ([模板目录](./agent/templates/))
- 明确标注受众 (🤖/👤/📚)
- 在README中建立索引

---

## 🆘 需要帮助？

1. **优先查阅** [Agent速查表](../.cursor/rules/AGENT_QUICK_REF.mdc)
2. **搜索 Graphiti** 查询历史经验
3. **查看工作流** 找到对应的执行步骤
4. **参考模板** 使用标准文档模板
5. **阅读规范** 查看 `.cursor/rules/` 中的规范文件

---

**最后更新:** 2025-11-16

