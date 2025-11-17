# 业务中台文档体系快速参考

> 5分钟了解 TTPOS 业务中台文档体系

---

## 🎯 项目概览

TTPOS 业务中台 (ttpos-bmp) 是基于 GoFrame v2.x + gRPC 微服务架构的核心业务处理系统。

### 核心模块
- **ttpos-manager**: 管理后台（用户、权限、配置）
- **ttpos-shop**: 门店管理（商品、订单、会员）
- **ttpos-erp**: ERP系统（进销存、财务、报表）
- **ttpos-takeout**: 外卖配送（订单、配送、第三方对接）
- **ttpos-message**: 消息中心（通知、队列服务）

---

## 📂 文档结构速览

```
docs/
├── README.md                 # 📖 文档体系总览
├── QUICK_START.md           # 🚀 本文件 - 快速开始
│
├── human/                   # 👤 人类专用（学习资料）
│   ├── guides/              # 开发指南和最佳实践
│   ├── architecture/        # 系统架构和技术决策
│   │   ├── entities/        # 数据实体定义
│   │   └── features/        # 功能模块设计
│   ├── business/            # 业务规则和工作流程
│   └── decisions/           # ADR 技术决策记录
│
├── shared/                  # 📚 共用资源
│   ├── specs/               # 功能规格说明
│   ├── integrations/        # 第三方服务集成
│   └── troubleshooting/     # 问题排查指南
│
├── team/                    # 👥 团队协作
│   ├── proposals/           # 需求提案和评审
│   └── activities/          # 团队活动记录
│
└── agent/                   # 🤖 Agent 专用（工作流+模板）
    ├── workflows/           # 开发工作流
    └── templates/           # 标准文档模板
```

---

## 🚀 快速查找

### 我要开发新功能
1. **查规范**: [GoFrame开发规范](./human/guides/goframe-development-guide.md)
2. **看流程**: [功能开发工作流](./agent/workflows/feature-development.md)
3. **用模板**: [功能开发模板](./agent/templates/feature-template.md)

### 我要修复Bug
1. **看流程**: [Bug修复工作流](./agent/workflows/bug-fixing.md)
2. **查问题**: [故障排查指南](./shared/troubleshooting/)
3. **搜经验**: 项目Git历史或团队文档

### 我要创建数据库迁移
1. **看流程**: [数据库迁移工作流](./agent/workflows/database-migration.md)
2. **查规范**: [../MIGRATION_QUICK_START.md](../MIGRATION_QUICK_START.md)
3. **用工具**: `make migrate-create`

### 我要开发gRPC服务
1. **看流程**: [微服务集成工作流](./agent/workflows/microservice-integration.md)
2. **查规范**: [Protobuf开发规范](./human/guides/protobuf-development-guide.md)
3. **用工具**: `gf gen pb`

### 我是新人入职
1. **项目概览**: [../README.MD](../README.MD)
2. **技术栈**: [技术栈](#技术栈)
3. **环境配置**: [快速开始](#快速开始)
4. **开发指南**: [./human/guides/](./human/guides/)

---

## 🏗️ 技术栈快速参考

| 组件 | 用途 | 文档入口 |
|------|------|----------|
| **GoFrame v2.x** | 应用框架 + ORM | [开发指南](./human/guides/goframe-development-guide.md) |
| **gRPC** | 微服务通讯 | [Protobuf规范](./human/guides/protobuf-development-guide.md) |
| **Nacos** | 服务发现 | [服务发现文档](./shared/integrations/nacos-integration.md) |
| **RocketMQ** | 消息队列 | [消息队列文档](./shared/integrations/rocketmq-integration.md) |
| **golang-migrate** | 数据库迁移 | [../MIGRATION_QUICK_START.md](../MIGRATION_QUICK_START.md) |

---

## 💡 使用技巧

### 对于开发者
```yaml
学习路径:
  1. 阅读项目主文档了解概览
  2. 查看技术栈和架构设计
  3. 按模块查阅相关文档
  4. 参考工作流了解开发流程
  5. 使用指南学习最佳实践
```

### 开发环境设置
```bash
# 1. 环境配置
cp .env.example .env
make conf

# 2. 启动中间件
make mid

# 3. 本地开发
make run.manager  # 或其他模块
```

---

## 📊 文档完善状态

### ✅ 已完成
- 项目文档体系结构
- 基础README和快速开始
- 部分架构和实体文档
- 核心开发指南框架

### ⏳ 待完善
- 完整的Agent工作流和模板
- 详细的业务流程文档
- 第三方集成完整文档
- 故障排查指南

---

## 🔗 核心入口

| 文件 | 用途 |
|------|------|
| [项目主文档](../README.MD) | 项目介绍和安装 ⭐⭐⭐ |
| [Makefile](../Makefile) | 开发工具和命令 ⭐⭐⭐ |
| [数据库迁移指南](../MIGRATION_QUICK_START.md) | 数据库管理 ⭐⭐ |
| [Docker配置](../docker-compose.yml) | 部署配置 ⭐⭐ |

---

## 🆘 需要帮助？

1. **查看项目文档** [../README.MD](../README.MD)
2. **阅读快速开始** 本文件
3. **查看开发工具** [../Makefile](../Makefile)
4. **了解技术栈** [技术栈](#技术栈)

---

**最后更新:** 2025-11-17