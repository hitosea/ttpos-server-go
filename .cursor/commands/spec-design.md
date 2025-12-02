---
name: spec-design
description: 创建技术设计和任务分解（设计阶段）
---

# /spec-design - 创建 Spec 设计文档

## 使用场景

需求审核通过后，创建技术设计文档和任务分解，进入开发阶段。

> **前置条件**: 必须先使用 `/spec-create` 创建需求文档，且审核状态为「已通过」。

## 使用方式

```bash
/spec-design story-order-quick-payment
/spec-design task-shop-export-report
```

## 参数

- `spec_name`: 必填，已存在的 Spec 名称
  - 格式: `{level}-{module}-{feature}`
  - 必须是已通过审核的 Spec

## 前置检查

执行前会自动检查：

1. **Spec 目录存在**: `docs/shared/specs/active/{spec_name}/`
2. **requirements.md 存在**: 需求文档已创建
3. **审核状态为「已通过」**: 产品审核已通过

### 检查失败处理

| 检查项 | 失败处理 |
|--------|----------|
| Spec 目录不存在 | 提示先使用 `/spec-create` |
| requirements.md 不存在 | 提示先使用 `/spec-create` |
| 审核状态不是「已通过」 | 提示等待产品审核通过 |

## 功能特点

- ✅ 检查 requirements.md 存在且已通过审核
- ✅ 创建 design.md（技术设计）
- ✅ 创建 tasks.md（任务分解）
- ✅ 自动填充基本信息
- ✅ 提供开发指引

## 后端特定适配

- ✅ 支持三模块（Main: Go + Gin, Admin: PHP + ThinkPHP, BMP: Go + GoFrame）
- ✅ 自动生成数据库迁移任务（SQL + Go Model）
- ✅ 自动生成 API 设计（RESTful / gRPC）
- ✅ 自动填充 Service/Repository/Model 文件路径

## 输出产物

```
docs/shared/specs/active/{level}-{module}-{feature}/
├── requirements.md  # 已存在（审核状态: 已通过）
├── design.md        # 新创建：技术设计
└── tasks.md         # 新创建：任务分解
```

## 工作流位置

```
Proposal → 评审 → /spec-create → 产品审核 → /spec-design → 开发
                                                 ↑
                                              当前命令
```

## 执行流程

```mermaid
graph TD
    A[/spec-design spec-name] --> B{检查 Spec 目录}
    B -->|不存在| C[❌ 提示先 /spec-create]
    B -->|存在| D{检查 requirements.md}
    D -->|不存在| C
    D -->|存在| E{检查审核状态}
    E -->|待审核| F[❌ 提示等待审核]
    E -->|需修改| G[❌ 提示先修改需求]
    E -->|已通过| H[✅ 创建 design.md]
    H --> I[✅ 创建 tasks.md]
    I --> J[✅ 输出开发指引]
```

## 错误处理

| 错误类型 | 处理方式 |
|---|---|
| Spec 不存在 | 提示使用 `/spec-create` 创建 |
| requirements.md 不存在 | 提示使用 `/spec-create` 创建 |
| 审核状态非「已通过」 | 提示等待产品审核 |
| design.md 已存在 | 询问是否覆盖 |
| tasks.md 已存在 | 询问是否覆盖 |
| 模板缺失 | 提示恢复模板 |

## 相关命令

| 命令 | 用途 |
|------|------|
| `/propose` | 创建需求提案 |
| `/spec-create` | 创建需求规格 |
| `/spec-design` | 创建技术设计 + 任务分解（当前命令） |
| `/check-tasks` | 检查任务进度 |

## 模板引用

- `docs/agent/templates/design-template.md` - 技术设计模板
- `docs/agent/templates/tasks-template.md` - 任务分解模板

---

**版本**: v1.0.0  
**创建日期**: 2025-11-25  
**维护者**: 知识管理组  
**状态**: ✅ MVP

