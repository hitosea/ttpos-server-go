# Cursor 自定义指令集

> 本目录存放团队自定义的 Cursor 指令，旨在通过简单的命令自动化常见任务。

## 🚀 快速开始

在 Cursor Chat 中输入 `/` 即可查看所有可用指令。

### 核心指令

| 优先级 | 指令                      | 描述                    | 状态   |
| ------ | ------------------------- | ----------------------- | ------ |
| ⭐⭐⭐ | `/propose`                | 创建需求提案（Scrum）   | ✅ MVP |
| ⭐⭐⭐ | `/onboard`                | 项目快速入门引导        | ✅ MVP |
| ⭐⭐⭐ | `/create-spec`            | 创建 Spec 文档          | ✅ MVP |
| ⭐⭐⭐ | `/archive-spec`           | 归档已完成的 Spec       | ✅ MVP |
| ⭐⭐⭐ | `/deprecate-spec`         | 废弃不再需要的 Spec     | ✅ MVP |
| ⭐⭐⭐ | `/restore-spec`           | 恢复已归档/废弃的 Spec  | ✅ MVP |
| ⭐⭐⭐ | `/check-tasks`            | 检查任务完成进度        | 规划中 |
| ⭐⭐   | `/create-api-doc`         | 为已实现的 API 创建文档 | 规划中 |
| ⭐⭐   | `/create-component-test`  | 为组件创建测试          | 规划中 |
| ⭐⭐   | `/create-model-test`      | 为 Model 创建测试       | 规划中 |
| ⭐⭐   | `/create-api-test`        | 为 API 创建测试         | 规划中 |
| ⭐⭐   | `/create-controller-test` | 为 Controller 创建测试  | 规划中 |
| ⭐     | `/search-knowledge`       | 搜索知识图谱            | 规划中 |
| ⭐     | `/add-knowledge`          | 添加知识到 Graphiti     | 规划中 |
| ⭐     | `/check-doc-quality`      | 检查文档质量            | 规划中 |
| ⭐     | `/report-bug`             | 报告问题                | 规划中 |
| ⭐     | `/plan-refactor`          | 重构规划                | 规划中 |

---

## 📖 指令详解

### `/propose` - 创建需求提案 ⭐ 新增

**使用场景**: 需求发起、Scrum 流程启动

**使用方式**:

```bash
/propose quick-payment                    # 创建快速支付功能提案
/propose report-export                    # 创建报表导出功能提案
/propose dark-mode                        # 创建深色模式提案
/propose feature-name 编号:36917          # 基于 DooTask 任务创建提案
/propose feature-name DooTask #36917      # 支持 DooTask # 格式
```

**功能特点**:

- ✅ 自动使用 `docs/agent/templates/proposal-template.md` 创建提案
- ✅ 自动填充日期、提案人信息
- ✅ 创建目录 `docs/team/proposals/{YYYY-MM}/{feature-name}.md`
- ✅ 提供 Scrum 评审清单
- ✅ **自动读取 DooTask 任务**（当提供任务编号时）
  - 自动获取任务标题、描述、需求详情
  - 将任务内容填充到提案的"背景和动机"、"解决方案概述"等章节
  - 自动在"关联任务"字段中记录任务编号
  - 将任务内容作为上下文信息，供后续对话使用
  - 将任务内容作为上下文信息，供后续对话使用

**详见**: `.cursor/commands/propose.md`

---

### `/archive-spec` - 归档 Spec

**使用场景**: 功能开发完成并发布后，将 Spec 归档到版本目录

**使用方式**:

```bash
/archive-spec @story-order-quick-payment                    # 自动检测版本号
/archive-spec @story-order-quick-payment --version v2.10    # 指定版本号
```

**功能特点**:

- ✅ **任务必须全部完成才能归档**
- ✅ 自动检测版本号（优先级：命令参数 > 关联 Proposal 中的目标版本 > 交互询问）
- ✅ 移动 Spec 目录到 `archived/{version}/`
- ✅ 在 `requirements.md` 头部添加归档标记
- ✅ 同步更新关联 Proposal 的链接和状态
- ✅ 记录活动日志

**详见**: `.cursor/commands/archive-spec.md`

---

### `/deprecate-spec` - 废弃 Spec

**使用场景**: 将不再需要、被替代或取消的 Spec 标记为废弃

**使用方式**:

```bash
/deprecate-spec @story-old-payment --reason "被 story-new-payment 替代"
/deprecate-spec @story-abandoned-feature --reason "需求取消"
```

**功能特点**:

- ✅ 必须提供废弃原因
- ✅ 移动 Spec 目录到 `deprecated/`
- ✅ 创建 `DEPRECATED.md` 记录废弃原因、时间、操作人
- ✅ 在 `requirements.md` 头部添加废弃标记
- ✅ 同步更新关联 Proposal 的链接和状态
- ✅ 记录活动日志

**详见**: `.cursor/commands/deprecate-spec.md`

---

### `/restore-spec` - 恢复 Spec

**使用场景**: 将已归档或已废弃的 Spec 恢复到 active 目录

**使用方式**:

```bash
/restore-spec @story-order-quick-payment                    # 自动检测来源
/restore-spec @story-order-quick-payment --from archived    # 指定从 archived 恢复
/restore-spec @story-old-payment --from deprecated          # 指定从 deprecated 恢复
```

**功能特点**:

- ✅ 自动检测 Spec 来源（`archived/` 或 `deprecated/`）
- ✅ 移动 Spec 目录到 `active/`
- ✅ 移除归档/废弃标记和 `DEPRECATED.md` 文件
- ✅ 同步更新关联 Proposal 的链接和状态
- ✅ 记录活动日志

**详见**: `.cursor/commands/restore-spec.md`

---
