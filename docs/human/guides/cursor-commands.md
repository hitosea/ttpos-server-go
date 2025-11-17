# Cursor 自定义指令集

> 本目录存放团队自定义的 Cursor 指令，旨在通过简单的命令自动化常见任务。

## 🚀 快速开始

在 Cursor Chat 中输入 `/` 即可查看所有可用指令。

### 核心指令

| 优先级 | 指令                      | 描述                    | 状态   |
| ------ | ------------------------- | ----------------------- | ------ |
| ⭐⭐⭐    | `/propose`                | 创建需求提案（Scrum）   | ✅ MVP  |
| ⭐⭐⭐    | `/onboard`                | 项目快速入门引导        | 规划中  |
| ⭐⭐⭐    | `/create-spec`            | 创建 Spec 文档          | 规划中  |
| ⭐⭐⭐    | `/check-tasks`            | 检查任务完成进度        | 规划中  |
| ⭐⭐     | `/create-api-doc`         | 为已实现的 API 创建文档 | 规划中  |
| ⭐⭐     | `/create-component-test`  | 为组件创建测试          | 规划中  |
| ⭐⭐     | `/create-model-test`      | 为 Model 创建测试       | 规划中  |
| ⭐⭐     | `/create-api-test`        | 为 API 创建测试         | 规划中  |
| ⭐⭐     | `/create-controller-test` | 为 Controller 创建测试  | 规划中  |
| ⭐      | `/search-knowledge`       | 搜索知识图谱            | 规划中 |
| ⭐      | `/add-knowledge`          | 添加知识到 Graphiti     | 规划中 |
| ⭐      | `/check-doc-quality`      | 检查文档质量            | 规划中 |
| ⭐      | `/report-bug`             | 报告问题                | 规划中 |
| ⭐      | `/plan-refactor`          | 重构规划                | 规划中 |

---

## 📖 指令详解

### 1. `/propose` - 创建需求提案 ⭐ 新增

**使用场景**: 需求发起、Scrum 流程启动

**使用方式**:

```bash
/propose quick-payment       # 创建快速支付功能提案
/propose report-export       # 创建报表导出功能提案
/propose dark-mode           # 创建深色模式提案
```

**功能特点**:

- ✅ 自动使用 `docs/agent/templates/proposal-template.md` 创建提案
- ✅ 自动填充日期、提案人信息
- ✅ 创建目录 `docs/team/proposals/{YYYY-MM-DD}-{feature-name}.md`
- ✅ 提供 Scrum 评审清单

**详见**: `.cursor/commands/propose.md`

---