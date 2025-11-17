# 人类专用文档

> 👤 详细的系统设计文档和学习资料

本目录包含面向人类开发者的详细文档，包括架构设计、业务流程、开发指南等。

## 📖 目录结构

### [guides/](./guides/) - 开发指南
**用途：** 开发规范、最佳实践、工具使用
- [GoFrame开发指南](./guides/goframe-development-guide.md)
- [Protobuf开发规范](./guides/protobuf-development-guide.md)
- [数据库设计规范](./guides/database-design-guide.md)

### [architecture/](./architecture/) - 系统架构
**用途：** 系统设计、技术决策、模块说明
- [modules/](./architecture/modules/) - 各模块详细文档
  - [manager/](./architecture/modules/manager/) - 管理模块
  - [shop/](./architecture/modules/shop/) - 门店模块
  - [erp/](./architecture/modules/erp/) - ERP模块
  - [takeout/](./architecture/modules/takeout/) - 外送模块
  - [message/](./architecture/modules/message/) - 消息模块
- [entities/](./architecture/entities/) - 数据实体定义
- [features/](./architecture/features/) - 功能模块设计

### [business/](./business/) - 业务知识
**用途：** 业务规则、流程、工作场景
- 业务流程文档
- 用户故事和用例
- 业务规则说明

### [decisions/](./decisions/) - 技术决策
**用途：** ADR（Architecture Decision Records）
- 技术选型决策
- 架构设计决策
- 重要技术决策记录

## 🎯 使用指南

### 新人入职
1. 先阅读 [项目主文档](../../README.MD)
2. 查看 [快速开始](../../QUICK_START.md)
3. 按需查阅相关指南和架构文档

### 功能开发
1. 查看对应模块的架构文档
2. 了解相关业务流程
3. 参考开发指南编写代码

### 问题排查
1. 查看架构文档了解系统设计
2. 查阅业务规则理解逻辑
3. 参考技术决策了解选型原因

## 📝 文档规范

### 编写原则
- 详细解释 WHY（为什么这样设计）
- 完整说明 HOW（如何实现）
- 提供示例代码和配置
- 建立文档间的交叉引用

### 更新维护
- 代码变更时同步更新相关文档
- 定期审查文档时效性
- 保持文档结构的一致性

---

**最后更新:** 2025-11-17