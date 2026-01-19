---
name: propose
description: 创建新功能或改进的需求提案
---

# /propose - 创建需求提案

## 使用场景

产品经理或团队成员提出新功能想法或改进建议。

## 使用方式

```bash
/propose quick-payment                    # 创建快速支付功能提案
/propose report-export                    # 创建报表导出功能提案
/propose dark-mode                        # 创建深色模式提案
/propose feature-name 编号:36917          # 基于 DooTask 任务创建提案
/propose feature-name DooTask #36917      # 支持 DooTask # 格式
```

## 参数

- `feature_name`: 必填，提案的功能名称（英文 kebab-case，如以其他语言传入时必须转换为英文）
- `task_id`: 可选，DooTask 任务编号（格式：`编号:36917` 或 `DooTask #36917`）
- `version`: 可选，版本号（格式：`v2.10.0` 或 `v 2.10.0`）

## 功能特点

- ✅ 自动使用 `proposal-template.md` 创建文件
- ✅ 自动填充日期和提案人（从 Git 配置读取）
- ✅ 创建路径: `docs/team/proposals/{YYYY-MM}/{version}-{feature_name}.md`
- ✅ 自动创建月份目录（如不存在）
- ✅ 提供 Scrum 评审清单
- ✅ 命名与 Spec 格式对齐，便于自动关联
- ✅ **自动读取 DooTask 任务**（当提供任务编号时）
  - 自动获取任务标题、描述、需求详情
  - 将任务内容填充到提案的"背景和动机"、"解决方案概述"等章节
  - 自动在"关联任务"字段中记录任务编号

## DooTask 任务集成

### 任务编号格式

支持以下格式：
- `编号:36917` - 标准格式
- `DooTask #36917` - 完整格式
- `#36917` - 简化格式

### 自动读取流程

当提供任务编号时，Agent 会：

1. **识别任务编号**
   - 从参数中提取任务 ID（如 `36917`）
   - 验证格式是否正确

2. **获取任务内容**
   - 提取任务标题、描述、需求、验收标准等信息

3. **填充提案内容**
   - **背景和动机**: 使用任务描述和需求背景
   - **解决方案概述**: 使用任务中的功能说明
   - **核心功能点**: 从任务中提取功能列表
   - **关联任务**: 自动记录 `DooTask #{task_id}`

4. **上下文信息**
   - 将任务内容作为上下文信息，供后续对话使用
   - 在创建提案时自动引用任务详情

📖 详见: `docs/human/guides/cursor-commands.md#1-propose`

---

**版本**: v1.1.0  
**维护者**: 产品组 + Scrum Master  
**状态**: ✅ MVP + DooTask 集成
