# 功能规格文档

> 需求定义和技术设计

---

## 📂 目录结构

```
docs/shared/specs/
├── active/      # 开发中的需求（默认创建位置）
├── archived/    # 已完成并发布的需求（按版本归档）
│   └── v2.10/
└── deprecated/  # 已废弃的需求
```

### 状态说明

| 状态 | 目录 | 说明 |
|------|------|------|
| **Active** | `active/` | 正在开发中的需求，新创建的 Spec 默认放这里 |
| **Archived** | `archived/{version}/` | 已完成并随版本发布的需求，按 minor 版本号归档 |
| **Deprecated** | `deprecated/` | 已废弃、被替代或取消的需求 |

### 状态流转

```
创建 → active/
         ↓
      开发完成 & 发布
         ↓
    ┌────┴────┐
    ↓         ↓
archived/   deprecated/
 (上线)      (废弃)
```

### 状态管理命令

```bash
# 归档到指定版本（任务必须全部完成）
/archive-spec @story-order-quick-payment --version v2.10

# 废弃 Spec（必须提供原因）
/deprecate-spec @story-old-feature --reason "需求取消"

# 恢复 Spec 到 active
/restore-spec @story-order-quick-payment
```

---

## 📝 目录说明

本目录存放功能规格文档（Spec），每个 Spec 应包含：

- `requirements.md` - 需求定义（User Story + AC）
- `design.md` - 技术设计方案
- `tasks.md` - 任务分解清单（Agent 友好）

---

## 📝 已有文档

### 系统级规格

- [软件需求规范](./software_requirements_specification.md)
- [标准软件开发文档](./standard_software_development_documents.md)

### 产品规格（PRD）

- [会员客户端 PRD](./member_client_prd.md)
- [LINE MAN PRD](./ttpos_lineman_prd.md)
- [LINE MAN SRS](./ttpos_lineman_srs.md)

---

## 🎯 Spec 命名规范

### 格式

```
{level}-{module}-{feature}/
├── requirements.md
├── design.md
└── tasks.md
```

### 示例

```
story-order-quick-payment/
├── requirements.md    # 需求：快速支付功能
├── design.md          # 设计：技术方案
└── tasks.md           # 任务：分解清单
```

### 命名规则

- **level**: `story` (用户故事) / `task` (技术任务)
- **module**: `order`, `member`, `product`, `payment` 等业务模块
- **feature**: kebab-case 格式的功能名称

---

## 📋 创建新 Spec

### 使用 Agent 指令

```bash
/create-spec story-order-quick-payment
```

### 手动创建

```bash
mkdir -p docs/shared/specs/active/story-order-quick-payment
cd docs/shared/specs/active/story-order-quick-payment
touch requirements.md design.md tasks.md
```

### Graphiti & 活动日志

- 每个 Spec 的 `requirements.md`、`design.md`、`tasks.md` 底部必须包含 `Graphiti & 活动日志` 段落。
- Related Episode 使用 `docs/agent/templates/graphiti-episode.md` 模板生成草稿后填入名称。
- 同步在 `docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md` 中记录创建/更新动作。

---

## 🔗 相关资源

### 工作流

- [需求管理工作流](../../agent/workflows/requirement-management.md)
- [功能开发工作流](../../agent/workflows/feature-development.md)

### 模板

- [需求模板](../../agent/templates/requirements-template.md)
- [设计模板](../../agent/templates/design-template.md)
- [任务模板](../../agent/templates/tasks-template.md)

### 规范

- [Spec 规范](../../../.cursor/rules/specs.mdc)
- [Agent 速查表](../../../AGENTS.md)

---

**最后更新**: 2025-11-16
