---
name: onboard
description: 项目快速入门引导（后端版）
---

# /onboard - 项目快速入门

## 使用场景

新成员入职 + 老成员快速回顾项目。

## 使用方式

```bash
/onboard                # 完整导航菜单
/onboard quick          # 5分钟快速入门
/onboard architecture   # 项目架构详解
/onboard workflow       # 工作流程
/onboard faq            # 常见问题
```

## 参数

- `mode` (可选): 入门模式
  - `quick`: 5分钟快速入门
  - `architecture`: 架构详解
  - `workflow`: 工作流程
  - `faq`: 常见问题
  - 默认: 完整导航菜单

## 功能特点

- ✅ 动态读取 `.cursor/rules/*.mdc` 和 `docs/` 内容
- ✅ 分模式展示，避免信息过载
- ✅ 提供清晰的下一步行动
- ✅ 适合新老成员（Go/PHP 开发者）

## 输出内容

### quick 模式
- 项目概览（TTPOS 后端三模块：Main, Admin, BMP）
- 技术栈（Go 1.23+ / PHP 8.0+ / Vue 3）
- 环境搭建（Docker Compose 快速启动）
- 下一步行动

### architecture 模式
- 三模块架构（Main, Admin, BMP）
- 依赖关系（MySQL 8.0+, Redis 6.0+, RabbitMQ）
- 微服务架构（Nacos, OpenTelemetry, SkyWalking）
- 3层依赖规范

### workflow 模式
- 常见工作流导航
- 需求管理（/propose）→ 功能开发（/spec-create）→ Bug 修复
- API 对接 → 数据库迁移

### faq 模式
- 常见问题解答（Go/PHP/Docker/Redis Cluster）
- 疑难排查

📖 详见: `docs/agent/workflows/onboarding.md`

## 6天快速上手路径

| 阶段 | 内容 | 时长 |
|---|---|---|
| D1-D2 | 环境搭建 + 项目理解 | 2天 |
| D3-D4 | 学习规范 + 简单任务 | 2天 |
| D5-D6 | 中等难度任务 | 2天 |

---

**版本**: v1.0.0  
**创建日期**: 2025-11-17  
**维护者**: 知识管理组  
**状态**: ✅ MVP

