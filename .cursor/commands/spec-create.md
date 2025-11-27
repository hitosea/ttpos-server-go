---
name: spec-create
description: 创建功能规格文档（需求阶段）
---

# /spec-create - 创建 Spec 需求文档

## 使用场景

快速创建功能规格的需求文档，进入产品审核阶段。

> **注意**: 此命令只创建 `requirements.md`。产品审核通过后，使用 `/spec-design` 创建技术设计和任务分解。

## 使用方式

```bash
/spec-create story-order-quick-payment
/spec-create task-shop-export-report
```

## 参数

- `spec_name`: 必填，Spec 名称
  - 格式: `{level}-{module}-{feature}`
  - level: `story` (用户故事) 或 `task` (技术任务)
  - module: 业务模块（order, member, product, shop, admin, bmp...）
  - feature: 功能名称（kebab-case）

## 功能特点

- ✅ 创建 requirements.md（需求规格）
- ✅ 自动填充基本信息（功能名称、模块、日期）
- ✅ **智能关联 Proposal**（搜索 docs/team/proposals/）
- ✅ 初始化审核状态为「待审核」
- ✅ 提供下一步指引

## 智能关联 Proposal

- 在 `docs/team/proposals/` 中搜索匹配的 Proposal
- 自动填充 requirements.md 中的"来源 Proposal"链接
- 回写 Proposal，更新状态和 Spec 链接
- 建立双向可追溯关系

📖 详见: `docs/agent/workflows/requirement/linking.md`

## 输出产物

```
docs/shared/specs/active/{level}-{module}-{feature}/
└── requirements.md  # 需求规格（审核状态: 待审核）
```

## 工作流位置

```
Proposal → 评审 → /spec-create → 产品审核 → /spec-design → 开发
                       ↑                          ↑
                    当前命令                   下一步
```

## 审核状态说明

| 状态 | 说明 | 下一步 |
|------|------|--------|
| **待审核** | 需求刚创建，等待产品审核 | 产品审核 |
| **已通过** | 产品审核通过，可进入设计阶段 | `/spec-design` |
| **需修改** | 产品审核有意见，需要修改 | 修改后重新审核 |

## 后端特定适配

- ✅ 支持三模块（Main: Go + Gin, Admin: PHP + ThinkPHP, BMP: Go + GoFrame）
- ✅ 自动识别技术栈并在 requirements.md 中标注
- ✅ 自动填充模块影响范围（Go/PHP/Vue）

## 错误处理

| 错误类型 | 处理方式 |
|---|---|
| 参数格式错误 | 显示正确格式 |
| 目录已存在 | 询问是否覆盖 |
| 模板缺失 | 提示恢复模板 |

## 相关命令

| 命令 | 用途 |
|------|------|
| `/spec-propose` | 创建需求提案 |
| `/spec-create` | 创建需求规格（当前命令） |
| `/spec-design` | 创建技术设计 + 任务分解 |

---

**版本**: v3.0.0 (分离需求与设计阶段)  
**创建日期**: 2025-11-16  
**更新日期**: 2025-11-25  
**维护者**: 知识管理组  
**状态**: ✅ MVP
