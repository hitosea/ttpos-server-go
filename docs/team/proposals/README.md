# 需求提案

> 需求提案和评审记录

---

## 📋 说明

本目录存放需求提案文档，记录从想法到需求确认的过程。

---

## 📝 提案命名规范

### 格式
```
{YYYY-MM-DD}-{feature-name}.md
```

### 示例
```
2025-11-16-quick-payment.md
2025-11-16-member-integration.md
```

---

## 🔄 提案流程

```
想法 → 创建提案 → 需求评审 → 
  ├─ 批准 → 创建 Spec → 进入开发
  └─ 拒绝 → 归档（标注原因）
```

---

## 🎯 创建提案

### 使用 Agent 指令
```bash
/propose quick-payment
```

### 手动创建
```bash
cd docs/team/proposals
touch 2025-11-16-quick-payment.md
```

---

## 📋 提案应包含

1. **背景** - 为什么需要这个功能？
2. **目标** - 要解决什么问题？
3. **方案** - 打算怎么实现？
4. **价值** - 能带来什么收益？
5. **风险** - 有哪些潜在风险？
6. **评审** - 评审结论和决策

---

## 🔗 相关资源

### 工作流
- [需求管理工作流](../../agent/workflows/requirement-management.md)

### 模板
- [提案模板](../../agent/templates/proposal-template.md)

### 规范
- [Spec 规范](../../../.cursor/rules/specs.mdc)

---

**最后更新**: 2025-11-16

