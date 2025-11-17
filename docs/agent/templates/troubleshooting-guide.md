# {问题分类} 排查指南

> 受众：🤖 Agent + 👤 开发者。用于记录可复用的调查步骤、解决方案和预防措施。

## 文档信息

- **状态**：Draft / In Review / Published
- **版本**：v1.0.0
- **创建日期**：YYYY-MM-DD
- **最后更新**：YYYY-MM-DD
- **维护者**：[@责任人]
- **适用技术栈**：Go (main/) / PHP (admin/) / GoFrame (ttpos-bmp/) / Database / Infra
- **Related Episode**：`type-topic-YYYY-MM`（Graphiti 名称）

---

## 1. 问题概述

### 1.1 现象

- 用户/系统看到的报错、超时、日志关键字
- API/命令示例（如 `/api/v1/...`、`php think ...`）

### 1.2 影响范围

- 终端：POS / Shop / KDS / API / Admin / 其他
- 环境：dev / staging / production
- 版本：`v2.x` / `Sprint xx`
- 严重等级：P0 / P1 / P2 / P3

---

## 2. 快速诊断

### 2.1 症状检查清单

- [ ] 症状 1
- [ ] 症状 2
- [ ] 症状 3

### 2.2 判断矩阵

| 症状组合        | 可能原因   | 跳转                       |
| --------------- | ---------- | -------------------------- |
| 症状 1 + 症状 2 | {原因描述} | [解决方案 A](#4-解决方案a) |
| 症状 2 + 症状 3 | {原因描述} | [解决方案 B](#4-解决方案b) |

---

## 3. 排查步骤

### 3.1 环境/配置

```bash
# 示例：检查数据库连接
cd /Users/benbige/Projects/ttpos-server-go/main
go run cmd/check_db/main.go
```

- 预期结果：
- 若异常 → 跳转到具体解决方案。

### 3.2 日志/监控

```bash
# Go 主服务日志
tail -f main/storage/logs/app.log | grep payment

# PHP 管理后台日志
tail -f admin/runtime/log/*.log
```

- 关键字段/错误示例：
- 对应解决方案：

### 3.3 数据状态

```sql
-- 示例：查看锁等待
SELECT * FROM information_schema.innodb_lock_waits;
```

---

## 4. 解决方案

### 4.1 解决方案 A

- **适用场景**：{症状/原因描述}
- **步骤**：
  1.
  2.
  3.
- **代码/配置修改**：

```go
// main/app/service/xxx/xxx_srv.go
```

- **验证**：

### 4.2 解决方案 B

（同上结构）

---

## 5. 预防措施

- 代码规范/PR 检查项
- 监控/告警（Prometheus、SkyWalking 等）
- 文档/Spec 更新要求

---

## 6. 相关资源

- `.cursor/rules/xxx.mdc`
- `docs/shared/specs/...`
- `docs/shared/api/...`
- Graphiti Episode：`{name}`（入库后替换）
- 其他：

---

## 7. 变更记录

| 日期       | 版本   | 修改人 | 说明     |
| ---------- | ------ | ------ | -------- |
| YYYY-MM-DD | v1.0.0 | @姓名  | 初始版本 |

---

> **完成后务必：** 在本文件及 Graphiti Episode 之间建立双向链接，并在 `docs/shared/troubleshooting/README.md` 的问题清单中登记。\*\*\*
