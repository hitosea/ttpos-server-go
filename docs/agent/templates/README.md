# 文档模板

> Agent 使用的标准化文档模板

---

## 📋 模板列表

### 核心模板（待创建）
- [ ] `proposal-template.md` - 需求提案模板
- [ ] `requirements-template.md` - 需求规格模板
- [ ] `design-template.md` - 技术设计模板
- [ ] `tasks-template.md` - 任务分解模板
- [ ] `api-doc.md` - API 文档模板
- [ ] `troubleshooting-guide.md` - 问题排查模板
- [ ] `decision-record.md` - ADR 决策记录模板

### 后端专属模板（待创建）
- [ ] `database-migration-template.md` - 数据库迁移模板
- [ ] `grpc-service-template.md` - gRPC 服务模板
- [ ] `migration-guide.md` - 版本迁移指南模板
- [ ] `test-report-template.md` - 测试报告模板

---

## 📝 与 .spec-workflow 的关系

### 目录对比

| 目录                        | 用途                 | 推荐使用 |
| --------------------------- | -------------------- | -------- |
| `docs/agent/templates/`     | Agent 执行的标准模板 | ✅ 推荐   |
| `.spec-workflow/templates/` | 工作区历史模板       | 可选参考 |

### 使用建议

1. **优先使用** `docs/agent/templates/` 中的模板
   - 与工作流完全集成
   - Agent 友好格式
   - 支持多语言（Go/PHP/Vue）

2. **可选参考** `.spec-workflow/templates/` 
   - 作为历史参考
   - 部分内容可能过时

---

## 🔗 相关资源

### 工作流
- [需求管理](../workflows/requirement-management.md)
- [功能开发](../workflows/feature-development.md)

### 规范
- [文档创建规范](../../../.cursor/rules/documentation.mdc)
- [Spec 规范](../../../.cursor/rules/specs.mdc)

---

**说明**: 模板文件正在迁移中，优先级从 P0 模板开始创建。

**最后更新**: 2025-11-16

