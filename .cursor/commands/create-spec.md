---
name: create-spec
description: 创建功能规格文档
---

# /create-spec - 创建 Spec 文档

## 使用场景

快速创建功能规格文档，开始新功能开发。

## 使用方式

```bash
/create-spec story-order-quick-payment   # 订单模块功能
/create-spec task-bmp-member-sync        # BMP 微服务任务
/create-spec story-admin-shop-audit      # 管理后台功能
```

## 参数

- `spec_name`: 必填，Spec 名称
  - 格式: `{level}-{module}-{feature}`
  - level: `story` (用户故事) 或 `task` (技术任务)
  - module: 业务模块（order, member, product, shop, admin, bmp...）
  - feature: 功能名称（kebab-case）

## 功能特点

- ✅ 自动使用标准模板 (requirements/design/tasks)
- ✅ 自动填充基本信息 (功能名称、终端、日期)
- ✅ **智能关联 Proposal** (搜索 docs/team/proposals/)
- ✅ 创建完整目录结构
- ✅ 提供下一步指引

## 智能关联 Proposal

- 在 `docs/team/proposals/` 中搜索匹配的 Proposal
- 自动填充 requirements.md 中的"来源 Proposal"链接
- 回写 Proposal，更新状态和 Spec 链接
- 建立双向可追溯关系

📖 详见:

- `docs/human/guides/cursor-commands.md#3-create-spec`
- `docs/agent/workflows/proposal-spec-linking.md`

## 输出产物

```
docs/shared/specs/active/{level}-{module}-{feature}/
├── requirements.md  # 需求规格（自动关联 Proposal）
├── design.md        # 技术设计（Go/PHP/Vue）
└── tasks.md         # 任务分解（含 AI Prompt）
```

## 后端特定适配

- ✅ 支持三模块（Main: Go + Gin, Admin: PHP + ThinkPHP, BMP: Go + GoFrame）
- ✅ 自动生成数据库迁移任务（SQL + Go Model）
- ✅ 自动生成 API 设计（RESTful / gRPC）
- ✅ 自动填充 Service/Repository/Model 文件路径

## 错误处理

| 错误类型     | 处理方式     |
| ------------ | ------------ |
| 参数格式错误 | 显示正确格式 |
| 目录已存在   | 询问是否覆盖 |
| 模板缺失     | 提示恢复模板 |

---

**版本**: v2.1.0 (后端适配版)  
**创建日期**: 2025-11-16  
**更新日期**: 2025-11-17  
**维护者**: 知识管理组  
**状态**: ✅ MVP
