---
name: propose
description: 创建新功能或改进的需求提案
---

# /propose - 创建需求提案

## 使用场景

产品经理或团队成员提出新功能想法或改进建议。

## 使用方式

```bash
/propose quick-payment       # 创建快速支付功能提案
/propose report-export       # 创建报表导出功能提案
/propose dark-mode           # 创建深色模式提案
```

## 参数

- `feature_name`: 必填，提案的功能名称（kebab-case）

## 功能特点

- ✅ 自动使用 `proposal-template.md` 创建文件
- ✅ 自动填充日期和提案人（从 Git 配置读取）
- ✅ 创建路径: `docs/team/proposals/{YYYY-MM-DD}-{feature_name}.md`
- ✅ 提供 Scrum 评审清单

📖 详见: `docs/human/guides/cursor-commands.md#1-propose`

---

**版本**: v1.0.0  
**维护者**: 产品组 + Scrum Master  
**状态**: ✅ MVP
