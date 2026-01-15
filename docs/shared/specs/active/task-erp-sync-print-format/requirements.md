# ERP 打印格式同步指令 需求文档

## 📋 基本信息

| 项目              | 内容                                                                 |
| ----------------- | -------------------------------------------------------------------- |
| **Spec ID**       | task-erp-sync-print-format                                           |
| **Level**         | task                                                                 |
| **来源 Proposal** | [erp-sync-print-format](../../../team/proposals/2026-01/erp-sync-print-format.md) |
| **创建日期**      | 2026-01-15                                                           |
| **负责人**        | rikugun                                                              |
| **目标 Sprint**   | 待定                                                                 |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 开发中     |
| **审核人**   | rikugun    |
| **审核日期** | 2026-01-15 |

---

## 📝 用户故事

**作为** 运维人员
**我想** 通过 CLI 指令从 ERP 系统拉取 Wallace 开头的打印格式文档并保存为本地 JSON 文件
**以便于** 提升运维效率，减少手动导出操作

---

## 功能需求

### Requirement 1: 执行指令获取打印格式文档

**用户故事**: 作为运维人员，我想执行 sync-print-format 指令获取 ERP 打印格式，以便于自动化同步配置

#### 验收标准

1. **WHEN** 执行 `sync-print-format --siteCode {code}` **THEN** 系统 **SHALL** 从 site 配置中获取对应 ERP 地址
2. **WHEN** 获取到 ERP 地址 **THEN** 系统 **SHALL** 调用 `/app/print-format` API 获取所有 "Wallace" 开头的文档列表
3. **IF** siteCode 参数缺失 **THEN** 系统 **SHALL** 输出参数错误提示并退出

### Requirement 2: 保存为本地 JSON 文件

**用户故事**: 作为运维人员，我想将获取的打印格式文档保存为 JSON 文件，以便于本地使用和版本管理

#### 验收标准

1. **WHEN** 成功获取文档列表 **THEN** 系统 **SHALL** 将每个文档保存为独立的 `.json` 文件
2. **WHEN** 保存文件 **THEN** 系统 **SHALL** 使用文档名称作为文件名，保存到 `manifest/printformat/html/` 目录
3. **IF** 目标文件已存在 **THEN** 系统 **SHALL** 覆盖现有文件

### Requirement 3: 输出同步结果摘要

**用户故事**: 作为运维人员，我想在同步完成后看到结果摘要，以便于确认操作是否成功

#### 验收标准

1. **WHEN** 同步完成 **THEN** 系统 **SHALL** 输出同步结果摘要，包含：成功数量、失败数量、总数量
2. **IF** 存在失败项 **THEN** 系统 **SHALL** 输出失败的文档名称和错误原因
3. **WHEN** 所有文档同步成功 **THEN** 系统 **SHALL** 以退出码 0 退出

---

## 非功能需求

### 测试要求

- [ ] 单元测试覆盖率 ≥ 80%
- [ ] 包含 API 调用 mock 测试
- [ ] 包含文件写入测试

### 平台兼容性

- [x] Linux (生产环境)
- [x] macOS (开发环境)
- [ ] Windows (可选)

### 日志要求

- [ ] 关键操作记录日志（开始同步、获取文档、保存文件）
- [ ] 错误信息包含详细上下文

---

## 约束条件

### 技术约束

- Go 版本: 1.23+
- 框架: GoFrame v2.x
- 必须遵循 CLAUDE.md 和 ttpos-bmp/.cursor/rules/go-rules.mdc 规范
- 使用 `gerror` 处理错误（不用标准库 errors）
- 参考现有 `ErpMigrate` 指令实现模式

### 资源约束

- Story Point: 3 (≤ 5)

### 依赖约束

- 依赖 site 配置表获取 ERP 地址
- 依赖 ERP `/app/print-format` API 可用

---

## 风险和缓解

### 风险 1: ERP API 接口变更

**影响**: 中
**缓解措施**: 增加 API 响应校验，接口变更时输出明确错误信息

### 风险 2: 网络不稳定导致同步失败

**影响**: 低
**缓解措施**: 实现单文档失败不影响其他文档同步，最终输出失败列表

---

## 接口设计（初稿）

### CLI 命令格式

```bash
# 基本用法
./ttpos-erp sync-print-format --siteCode 1

# 参数说明
--siteCode  站点代码，用于获取对应 ERP 地址（必填）
--dirBase   输出目录，默认 ./manifest/printformat/html/（可选）
```

### 输出示例

```
开始同步打印格式...
站点: 1, ERP 地址: http://192.168.100.206:15080
获取文档列表: 5 个 Wallace 开头的文档
  - Wallace Invoice: 保存成功
  - Wallace Receipt: 保存成功
  - Wallace Report: 保存成功

同步完成!
  成功: 3
  失败: 0
  总计: 3
```

---

**版本**: v1.0.0
**创建日期**: 2026-01-15
