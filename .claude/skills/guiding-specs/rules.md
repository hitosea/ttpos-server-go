# Spec 规范

## Proposal → Spec 拆分策略

### 核心原则

**Proposal 是 1:N 关系到 Spec**，需要多维度叠加拆分，直到每个 Spec 满足 SP ≤ 5。

### 拆分维度（按优先级）

| 优先级 | 维度         | 触发条件           | 示例                          |
| ------ | ------------ | ------------------ | ----------------------------- |
| 1      | **终端**     | 涉及多终端         | all → pos, shop, kds          |
| 2      | **复杂度**   | SP > 5             | 拆到 SP ≤ 5                   |
| 3      | **功能模块** | 功能可独立交付     | 列表、详情、操作              |
| 4      | **用户角色** | 不同角色不同功能   | admin-settings, staff-actions |
| 5      | **Phase**    | 需要分批上线       | phase1-core, phase2-advanced  |
| 6      | **依赖层级** | 有前后依赖         | 基础组件 → 业务逻辑           |

### 拆分算法（递归）

```
Proposal: all-order-management
    │
    ├─ Step 1: 按终端拆
    │   ├─ story-pos-order-management (SP8) ← 还需拆
    │   ├─ story-shop-order-management (SP13) ← 还需拆
    │   └─ story-kds-order-management (SP3) ✅
    │
    └─ Step 2: 按复杂度拆（对 SP>5 的继续拆）
        ├─ story-pos-order-list (SP3) ✅
        ├─ story-pos-order-detail (SP3) ✅
        ├─ story-pos-order-action (SP2) ✅
        ├─ story-shop-order-list (SP5) ✅
        ├─ story-shop-order-export (SP3) ✅
        └─ story-shop-order-analytics (SP5) ✅
```

### 拆分决策流程

```
function 拆分(需求):
    if SP ≤ 5:
        return [创建Spec]
    
    # 按优先级顺序尝试拆分
    for 维度 in [终端, 功能模块, 用户角色, Phase, 依赖层级]:
        if 可按该维度拆分:
            子需求列表 = 按维度拆分(需求)
            return flatten([拆分(子) for 子 in 子需求列表])
    
    # 无法拆分，标记为复杂需求，需要 Spike 调研
    return [创建Spike]
```

### Spec 关联

同一 Proposal 拆出的 Spec 应：
- 在 `requirements.md` 中标注 `来源 Proposal`
- 使用 `关联 Spec` 字段互相引用
- 标注依赖顺序（如有）

---

## 命名格式

```
{level}-{app}-{feature}
```

**参数**
- `{level}`: `story` (用户故事) | `task` (技术任务)
- `{app}`: pos | shop | kds | qds | assistant | tablet | mobile | menu | member | `all`
- `{feature}`: kebab-case

**示例**
```
story-pos-quick-payment               # POS端快捷支付
task-shop-report-export               # Shop端报表导出
story-all-connectivity-indicator      # 全应用网络指示器
```

---

## 最小细粒度原则

### 核心原则

**按应用垂直拆分 Spec**

```yaml
❌ 错误:
story-pos-shop-kds-order-sync  # 跨3个应用,难以控制SP

✅ 正确:
story-pos-order-sync    # POS端,SP3
story-shop-order-sync   # Shop端,SP3
story-kds-order-sync    # KDS端,SP3
```

### 为什么

1. **独立交付** 每个应用可独立开发、测试、部署
2. **并行开发** 不同团队可同时开发
3. **SP 可控** 单应用 Spec 的 SP 更容易 ≤5
4. **职责清晰** 每个 Spec 有明确的负责团队

---

## 跨应用场景处理

### 场景 A: 相似功能,不同实现

**拆分为多个 Spec**
```
story-pos-category-display
story-assistant-category-display
story-tablet-category-display
```

### 场景 B: 完全相同功能和实现

**使用 all**
```
story-all-connectivity-indicator      # 共享代码
```

### 场景 C: 后端功能

**使用 task-all**
```
task-all-api-refactor                 # API 重构
task-all-database-migration          # 数据库迁移
```

---

## Spec 目录结构

```
docs/shared/specs/{level}-{app}-{feature}/
├── requirements.md    # 需求规格说明
├── design.md          # 技术设计文档
└── tasks.md           # 实施任务分解
```

---

## 与 Scrum Story Point 的关系

### 黄金规则

**只有 SP ≤ 5 的需求才能进入 Sprint**

### 如果 SP > 5

**必须拆分** → 按最小细粒度拆分

---

## Story Point 评估

### SP 等级表

| SP       | 工作量   | 技术复杂度 | 功能复杂度 | 使用条件 |
| -------- | -------- | ---------- | ---------- | -------- |
| **SP1**  | 0.5-1 天 | 极简单     | 极简单     | 随时可用 |
| **SP3**  | 1-2 天   | 简单       | 简单       | 随时可用 |
| **SP5**  | 2-3 天   | 中等       | 中等       | 随时可用 |
| **SP8**  | 4-6 天   | 复杂       | 复杂       | 必须拆分 |
| **SP13** | 1-2 周   | 很复杂     | 很复杂     | 必须拆分 |

### 因素加成

| 因素       | 加分    | 触发条件                  |
| ---------- | ------- | ------------------------- |
| 高风险     | +1      | payment, order, cart 相关 |
| 新技术     | +1      | 首次使用技术栈            |
| 多端适配   | +0.5/端 | 需要支持多个终端          |
| 测试要求高 | +0.5    | 覆盖率 100%、复杂测试场景 |
| 文档复杂   | +0.5    | 需要详细文档和迁移指南    |
| 团队协作   | +1      | 需要跨团队协作            |

---

## 命名规范清单

### ✅ 允许

- 单个单词: `payment`, `category`, `report`
- 多个单词: `quick-payment`, `category-display`
- 使用 `all`: `story-all-dark-mode`

### ❌ 禁止

- 使用数字代替: ~~`feature1`~~ → `quick-payment`
- 使用缩写: ~~`cat-disp`~~ → `category-display`
- 过长名称: ~~`product-category-display-with-search-and-filter`~~ → 拆分

---

## 状态流转

```
草稿 → 待审核 → 已通过 → 开发中 → 待测试 → 已验证 → 已完成 → 已归档
```
