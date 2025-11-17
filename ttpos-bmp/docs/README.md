# TTPOS 业务中台文档体系

> 基于 GoFrame v2.x 框架 + gRPC 微服务架构的业务中台模块文档

---

## 📋 项目概览

TTPOS 业务中台 (ttpos-bmp) 是 TTPOS 系统的核心业务处理模块，采用微服务架构设计，提供以下核心服务：

- **ttpos-manager**: 管理后台模块（用户、权限、系统配置等）
- **ttpos-shop**: 门店管理模块（商品、订单、会员等）
- **ttpos-erp**: ERP业务模块（进销存、财务、报表等）
- **ttpos-takeout**: 外卖配送模块（订单、配送、第三方对接等）
- **ttpos-message**: 消息中心模块（消息队列、通知服务等）

---

## 📖 目录导航

### 🤖 [Agent 专用](./agent/)
**用途：** 工作流执行清单和结构化模板
**受众：** AI Agent 自动化执行
**风格：** 步骤检查清单 + 决策树 + 快速命令

- [工作流程](./agent/workflows/) - 核心开发工作流
- [文档模板](./agent/templates/) - 标准开发模板

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
| **学习开发规范**   | [Go开发规范](./human/guides/golang-development-guide.md)                   |

---

## 🏗️ 技术栈

- **基础框架**: GoFrame v2.x (应用框架 + ORM + 工具链)
- **服务通讯**: gRPC (微服务通讯协议)
- **服务发现**: Nacos (服务注册与发现)
- **消息队列**: RocketMQ (支持延迟消息)
- **数据库迁移**: golang-migrate (版本管理)
- **监控追踪**: SkyWalking (分布式追踪)

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
- [GoFrame开发规范](./human/guides/goframe-development-guide.md)
- [Protobuf开发规范](./human/guides/protobuf-development-guide.md)
- [数据库迁移指南](../MIGRATION_QUICK_START.md)
- [部署指南](../README.MD#部署步骤)

### 项目文档
- [项目主文档](../README.MD) - 项目介绍和安装指南
- [Makefile](../Makefile) - 开发工具和命令参考
- [Docker配置](../docker-compose.yml) - 容器化部署配置

---

## 📝 文档维护

### 文档更新原则
1. Agent文档保持<300行
2. 所有模板保持最新
3. 定期审查文档时效性
4. 建立交叉引用避免孤岛

### 贡献指南
- 明确标注受众 (🤖/👤/📚)
- 在README中建立索引
- 参考主项目文档体系结构

---

## 🆘 需要帮助？

1. **查看快速开始** [QUICK_START.md](./QUICK_START.md)
2. **阅读项目文档** [../README.MD](../README.MD)
3. **查看开发工具** [../Makefile](../Makefile)
4. **了解技术栈** [技术栈](#技术栈)

---

**最后更新:** 2025-11-17