# 人类专用文档

> 为开发者学习和参考设计的详细文档

---

## 📖 文档特征

- **目标：** 深度理解，系统学习
- **长度：** 不限
- **风格：** 详细解释 + WHY + HOW + 完整示例
- **格式：** 叙述性文字 + 设计权衡 + 最佳实践

---

## 📂 目录结构

### [学习指南](./guides/)
**用途：** 开发者入门和最佳实践

[待补充 by @开发者]

推荐内容：
- 快速开始指南
- 开发环境配置
- 常见问题解答
- Cursor 指令使用

### [架构设计](./architecture/)
**用途：** 系统架构和技术文档

包含内容：
- [实体模型](./architecture/entities/) - 18个数据模型文档
- [架构设计](./architecture/) - 系统架构文档
- [重构计划](./architecture/refactor/) - 10个重构文档
- [功能特性](./architecture/features/) - 功能特性参考文档

### [业务知识](./business/)
**用途：** 业务规则和工作流程

包含内容：
- [业务术语](./business/) - 餐饮行业术语
- [业务流程](./business/workflows/) - 业务工作流
- [业务规则](./business/rules/) - 业务逻辑规则

### [技术决策](./decisions/)
**用途：** ADR (Architecture Decision Records)

[待补充 by @开发者]

推荐格式：
```
{YYYY-MM-DD}-{决策标题}.md
例如：2025-11-16-choose-gin-framework.md
```

---

## 🎯 使用场景

### 我想学习...

| 场景 | 查看 |
|---|---|
| **项目架构** | [架构设计](./architecture/) |
| **数据模型** | [实体模型](./architecture/entities/) |
| **业务流程** | [业务知识](./business/) |
| **技术决策** | [技术决策](./decisions/) |
| **重构计划** | [重构文档](./architecture/refactor/) |

### 我想理解...

| 问题 | 查看 |
|---|---|
| **为什么选择 Gin 框架？** | [技术决策](./decisions/) |
| **数据表之间的关系？** | [实体模型](./architecture/entities/) |
| **订单处理的完整流程？** | [业务流程](./business/workflows/) |
| **系统如何扩展？** | [架构设计](./architecture/) |

---

## 📝 文档编写规范

### 人类视角文档特征

```yaml
目标: 深度理解，系统学习
长度: 不限
结构:
  - WHY: 设计原因 ✓
  - HOW: 详细步骤 ✓
  - EXAMPLE: 完整示例 ✓
  - TROUBLESHOOTING: 常见问题 ✓
风格:
  - 叙述性文字
  - 详细解释
  - 设计权衡
  - 最佳实践
```

### 文档模板

```markdown
# {文档标题}

## 背景
为什么需要这个功能/设计？

## 设计原则
遵循哪些原则？

## 详细设计
如何实现的？（包含代码示例）

## 设计权衡
考虑了哪些方案？为什么选择当前方案？

## 最佳实践
使用时的注意事项和建议

## 常见问题
Q&A

## 相关资源
相关文档链接
```

---

## ⚠️ 注意事项

### DO (应该)
- ✓ 详细解释设计原因
- ✓ 提供完整代码示例
- ✓ 讨论设计权衡
- ✓ 包含常见问题解答
- ✓ 引用相关技术文章

### DON'T (不应该)
- ✗ 只有命令没有解释
- ✗ 缺少示例代码
- ✗ 忽略"为什么"
- ✗ 过于简洁难以理解

---

## 🔗 相关资源

### Agent 执行清单
- [Agent工作流](../agent/workflows/) - 具体执行步骤（Agent专用）
- [Agent模板](../agent/templates/) - 结构化模板（Agent专用）

### 核心规范
- [Go开发规范](../../.cursor/rules/golang.mdc)
- [PHP开发规范](../../.cursor/rules/php.mdc)
- [Vue开发规范](../../.cursor/rules/vue.mdc)

### 项目文档
- [项目README](../../README.md)
- [项目介绍](../../.cursor/rules/intro.mdc)
- [项目结构](../../.cursor/rules/structs.mdc)

---

**最后更新:** 2025-11-16

