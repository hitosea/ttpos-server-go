# BMP MQ Topic 文档梳理 需求提案

> 本提案用于梳理 `ttpos-bmp` 中实际使用的 MQ topic，并补齐按模块的对接文档，降低联调与运维成本。

---

## 📋 提案信息

| 项目 | 内容 |
| ---------- | -------- |
| **提案人** | rikugun |
| **日期** | 2025-12-12 |
| **目标版本** | - |
| **状态** | 已落地（文档已补齐） |
| **关联任务** | - |
| **关联 Spec** | - |

---

## 🎯 背景和动机

### 问题描述

- `ttpos-bmp` 内 RocketMQ topic 分散在常量/业务逻辑/运维清单/架构文档中，缺少一个“按模块可落地”的统一说明。
- 新模块接入或跨服务订阅时，常见问题包括：topic 名称不一致、消息体字段不清晰、生产/消费位置不好找、运维 topic 初始化遗漏。

### 业务价值

- 降低联调成本：一眼能定位 topic、消息体 schema、生产/消费入口。
- 降低运维风险：topic 初始化清单与代码使用对齐。
- 提升可维护性：topic 变更能同步更新到单一目录。

### 目标用户

- [x] 后端开发
- [x] 运维/部署
- [ ] 前端
- [ ] 测试

---

## 💡 解决方案概述

### 方案描述

在 `ttpos-bmp/docs/shared/` 下新增 `mq-topics/` 目录，按模块拆分维护 topic 文档（包含用途、消息体 schema、生产/消费源码位置、注意事项），并补齐 `ttpos-bmp/manifest/topics.txt` 中的漏项，形成“代码-文档-运维清单”的闭环。

### 核心功能点

1. 按模块整理 topic：`ttpos-erp` / `ttpos-message` / `ttpos-takeout` / `ttpos-websocket` / `common`
2. 为每个 topic 补齐：用途、消息体结构、生产/消费位置、备注
3. 同步修正运维 topic 清单：`ttpos-bmp/manifest/topics.txt`

### 影响范围

**涉及模块**：
- [x] 文档（`ttpos-bmp/docs/shared/mq-topics/`）
- [x] 运维清单（`ttpos-bmp/manifest/topics.txt`）
- [ ] 业务逻辑

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：文档与清单补齐

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1

### 风险识别

- **潜在风险**：topic/消息体变更后文档不更新
- **缓解措施**：将 `mq-topics/` 作为变更检查项（topic 新增/重命名必须同步更新）

---

## 🔗 相关资源

- 产出文档：`ttpos-bmp/docs/shared/mq-topics/README.md`
