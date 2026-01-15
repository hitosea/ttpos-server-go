# Proposal 详细规范

## 命名规则详解

### App 标识符

| App       | 代码名    | 平台   | 使用场景                     |
| --------- | --------- | ------ | ---------------------------- |
| pos       | pos       | 全平台 | 收银、点餐、结账相关功能     |
| shop      | shop      | 全平台 | 后台管理、配置、报表相关功能 |
| kds       | kds       | 全平台 | 厨房显示、制作状态相关功能   |
| qds       | qds       | 全平台 | 排队叫号、等待管理相关功能   |
| assistant | assistant | 全平台 | 收银辅助、简化操作相关功能   |
| tablet    | tablet    | 全平台 | 平板点餐、桌边服务相关功能   |
| mobile    | mobile    | Web    | 顾客扫码点餐相关功能         |
| menu      | menu      | Web    | 电子菜单展示相关功能         |
| member    | member    | Web    | 会员管理、积分储值相关功能   |
| kiosk     | kiosk     | 全平台 | 自助点餐机相关功能           |
| all       | -         | 多端   | 跨多个终端的通用功能         |

### Feature Name 规范

**格式要求:**
- 小写英文字母
- 单词间用连字符 `-` 分隔
- 长度建议 2-4 个单词

**命名建议:**

| 类型     | 示例                        | 说明               |
| -------- | --------------------------- | ------------------ |
| 功能添加 | `quick-payment`             | 动词+名词          |
| 功能优化 | `order-list-optimization`   | 名词+optimization  |
| 配置项   | `receipt-print-config`      | 名词+config        |
| 集成     | `grab-delivery-integration` | 服务名+integration |
| UI 改进  | `category-badge-display`    | 名词+display       |

**避免的命名:**
- ❌ `newFeature` (驼峰命名)
- ❌ `new_feature` (下划线)
- ❌ `f1` (无意义缩写)
- ❌ `pos-quick-payment-for-cash-and-card` (过长)

## 目录管理

### 月份目录

提案按创建时间归档到月份目录：

```
docs/team/proposals/
├── 2025-11/    # 2025年11月创建的提案
├── 2025-12/    # 2025年12月创建的提案
└── 2026-01/    # 2026年1月创建的提案
```

### 索引维护

`docs/team/proposals/README.md` 是提案索引，格式：

```markdown
## {YYYY-MM}

| Proposal                                  | 说明     | 状态 |
| ----------------------------------------- | -------- | ---- |
| [proposal-name](YYYY-MM/proposal-name.md) | 简短说明 | 状态 |
```

## 提案内容规范

### 必填章节

1. **📋 提案信息** - 基本元数据
2. **🎯 背景和动机** - 问题和价值
3. **💡 解决方案概述** - 方案描述

### 选填章节

4. **📊 初步评估** - 复杂度和风险
5. **🤝 需求评审** - 评审记录
6. **📝 附录** - User Story 初稿

### 影响范围标注

明确标注涉及的终端和模块：

```markdown
**涉及终端**：
- [x] POS 收银端
- [ ] Shop 商家管理端

**涉及模块**：
- [x] UI 组件
- [x] API 接口
```

## 状态管理

### 状态定义

| 状态     | Badge  | 下一步行动             |
| -------- | ------ | ---------------------- |
| 待评审   | 无     | 等待评审会议           |
| ✅ 已批准 | 已批准 | 创建 Spec，分配负责人  |
| ❌ 已拒绝 | 已拒绝 | 归档，记录拒绝原因     |
| 🔄 需修改 | 需修改 | 补充信息后重新提交评审 |

### 状态更新

在提案文件的 `提案信息` 表格中更新：

```markdown
| **状态** | ✅ 已批准 |
```

同时更新 `README.md` 索引中的状态列。

## 与 Spec 的关系

```
Proposal (提案)        Spec (规格)
docs/team/proposals/   docs/shared/specs/active/
--------------------   -------------------------
pos-quick-payment.md → story-pos-quick-payment/
                       ├── requirements.md
                       ├── design.md
                       └── tasks.md
```

### 关联记录

提案通过后，在提案中记录关联的 Spec：

```markdown
| **关联 Spec** | story-pos-quick-payment |
```

## 时间规范

所有时间使用东八区 (Asia/Shanghai)：

```bash
# 获取日期
TZ=Asia/Shanghai date +%Y-%m-%d  # 2025-01-04

# 获取月份目录名
TZ=Asia/Shanghai date +%Y-%m     # 2025-01
```

## 目标用户类型

### 标准用户类型

| 用户类型   | 说明                           |
| ---------- | ------------------------------ |
| 收银员     | POS 操作人员，负责点餐结账     |
| 店长       | 门店管理者，负责运营和报表     |
| 商户管理员 | 后台配置人员，管理商品和设置   |
| 厨房人员   | 制作人员，查看订单和出餐       |
| 服务员     | 前厅服务，点餐和桌台管理       |
| 顾客       | 终端消费者，扫码点餐和自助服务 |

### 终端→用户映射

| 终端      | 主要用户                   | 次要用户   |
| --------- | -------------------------- | ---------- |
| pos       | 收银员                     | 店长       |
| shop      | 商户管理员、店长           | -          |
| kds       | 厨房人员                   | -          |
| qds       | 服务员                     | 顾客       |
| assistant | 收银员                     | 服务员     |
| tablet    | 服务员                     | 顾客       |
| mobile    | 顾客                       | -          |
| menu      | 顾客                       | -          |
| member    | 顾客                       | 商户管理员 |
| kiosk     | 顾客                       | -          |
| all       | 根据功能确定               | -          |

> **Interview 规则**: Q4 目标用户选项必须从上表中选择，禁止自行推断用户类型。

## 相关资源

- 提案模板: [template.md](template.md)
- 提案索引: `docs/team/proposals/README.md`
- 创建命令: `/propose`

