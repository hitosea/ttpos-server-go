# 共用资源文档

> Agent 和人类都需要的资源文档

---

## 📂 目录结构

### [功能规格](./specs/)
**用途：** 需求和设计文档  
**格式：** Spec 目录，包含 requirements.md, design.md, tasks.md

```
specs/
└── story-{module}-{feature}/
    ├── requirements.md    # 需求定义（User Story + AC）
    ├── design.md          # 技术设计方案
    └── tasks.md           # 任务分解清单（Agent友好）
```

**命名规范：**
```yaml
格式: {level}-{module}-{feature}
示例: story-order-quick-payment
说明:
  - level: story (用户故事) / task (技术任务)
  - module: order, member, product, payment 等业务模块
  - feature: kebab-case 格式的功能名称
```

### [API 文档](./api/)
**用途：** 接口文档  
**格式：** Markdown，包含接口定义、参数、响应示例

目前包含：
- [后端API文档](./api/backend/) - Go 服务API（待完善）

推荐结构：
```markdown
# {模块名称} API

## 概述
模块功能描述

## 接口列表
### {接口名称}
- **URL:** POST /api/v1/order/create
- **请求参数:**
- **响应示例:**
- **错误码:**

## 代码示例
Go/PHP/JavaScript 调用示例

## 注意事项
重要提示和最佳实践
```

### [问题排查](./troubleshooting/)
**用途：** 常见问题和解决方案  
**格式：** 问题分类文档

目前包含：
- [常见问题](./troubleshooting/common-issues.md)
- [平台兼容性](./troubleshooting/platform/) - Web等平台问题

推荐结构：
```markdown
# {问题类别}

## 问题现象
具体错误信息

## 问题原因
根本原因分析

## 解决方案
步骤化解决方法

## 预防措施
如何避免再次发生

## 相关资源
相关文档链接
```

### [第三方集成](./integrations/)
**用途：** 外部服务集成文档  
**格式：** 服务目录，包含接入指南和API参考

目前包含：
- [LINE MAN 集成](./integrations/lineman/) - LINE MAN 外卖平台对接

推荐结构：
```
integrations/{service}/
├── README.md           # 概述
├── setup.md           # 环境配置
├── api-reference.md   # API参考
└── troubleshooting.md # 常见问题
```

---

## 🎯 使用场景

### Agent 使用

```yaml
场景: 实现新功能
步骤:
  1. 读取 specs/{spec-name}/requirements.md 了解需求
  2. 查看 specs/{spec-name}/design.md 了解技术方案
  3. 按照 specs/{spec-name}/tasks.md 逐任务执行
  4. 参考 api/ 文档了解接口规范
```

### 人类使用

```yaml
场景: 了解功能实现
步骤:
  1. 阅读 specs/ 了解需求和设计
  2. 查看 api/ 了解接口定义
  3. 参考 troubleshooting/ 解决问题
  4. 查阅 integrations/ 对接第三方服务
```

---

## 📋 文档规范

### Spec 文档规范

参考 [Spec规范](../../.cursor/rules/specs.mdc)

**requirements.md 结构：**
- User Story (作为...我想...以便...)
- Acceptance Criteria (验收条件)
- 非功能需求

**design.md 结构：**
- 技术方案
- 数据模型
- API 设计
- 依赖关系

**tasks.md 结构：**
- [ ] 任务1 (带文件路径、需求引用、AI提示)
- [ ] 任务2
- ...

### API 文档规范

参考 [API规范](../../.cursor/rules/go-main.mdc) 中的 API 响应规范

**关键规范：**
- URL 使用 snake_case (不用 kebab-case)
- 响应格式: `{code, message, data}`
- data 不能是 null 或数组
- 分页信息放在 meta 中

---

## 🔗 相关资源

### Agent 资源
- [功能开发工作流](../agent/workflows/feature-development.md)
- [需求管理工作流](../agent/workflows/requirement-management.md)
- [文档模板](../agent/templates/)

### 规范文档
- [Spec规范](../../.cursor/rules/specs.mdc)
- [Go开发规范](../../.cursor/rules/go-main.mdc)
- [PHP开发规范](../../.cursor/rules/php.mdc)

### 项目指导
- [产品概述](../../.spec-workflow/steering/product.md)
- [技术栈说明](../../.spec-workflow/steering/tech.md)
- [项目结构](../../.spec-workflow/steering/structure.md)

---

**最后更新:** 2025-11-16

