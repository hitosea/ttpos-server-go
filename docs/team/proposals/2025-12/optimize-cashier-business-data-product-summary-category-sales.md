# （优化）收银机-营业数据按商品【合计】、按商品分类【销售额】取值调整 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | 王昱     |
| **日期**   | 2025-12-08 |
| **目标版本** | - |
| **状态**   | 待评审   |
| **关联任务** | - |
| **关联 Spec** | [story-cashier-business-data-product-summary-category-sales](../../../shared/specs/archived/v2.12/story-cashier-business-data-product-summary-category-sales/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

当前收银机营业数据统计中，按商品统计的【合计】字段和按商品分类统计的【销售额】字段的取值逻辑需要调整，以确保数据统计的准确性和一致性。

**当前情况**：
- 按商品统计（`/cashier/statistics/product`）返回的 `Subtotal`（合计）字段取值逻辑可能不符合业务预期
- 按商品分类统计（`/cashier/statistics/product_category`）返回的 `Prices`（销售额）字段取值逻辑需要优化

**影响**：
> 营业数据统计不准确可能导致商户对经营情况的误判，影响经营决策。

### 业务价值

- 提升营业数据统计的准确性
- 确保按商品和按商品分类统计的数据一致性
- 提高商户对经营数据的信任度
- 支持更准确的经营分析和决策

### 目标用户

- [x] 收银员
- [x] 商户管理员
- [ ] 厨房人员
- [ ] 顾客
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

> 调整收银机营业数据统计接口中，按商品统计的【合计】字段和按商品分类统计的【销售额】字段的计算逻辑，确保取值符合业务预期，并与整体营业数据保持一致。

### 核心功能点

1. 调整按商品统计（`CountProduct`）中 `Subtotal`（合计）字段的计算逻辑
2. 调整按商品分类统计（`CountProductCategory`）中 `Prices`（销售额）字段的计算逻辑
3. 确保两个统计口径的数据一致性和准确性
4. 验证调整后的数据与总营业数据的一致性

### 影响范围

**涉及终端**：
- [x] POS 收银端
- [ ] Shop 商家管理端
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件
- [x] API 接口
- [ ] 数据模型
- [x] 业务逻辑
- [ ] 第三方集成
- [ ] 其他: ________

**涉及文件**：
- `main/app/api/v1/cashier/cashier_statistics.go` - API 接口层
- `main/app/service/business.go` - 业务逻辑层（`CountProduct`、`CountProductCategory`）
- `main/app/service/statistics.go` - 统计服务层
- `main/app/repository/statistics.go` - 数据访问层（`CountProduct`、`CountCategory`）
- `main/app/dto/resp/business_data_resp/base.go` - 响应数据结构

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 2-3 天
- **预估 SP**: 3-5 SP（待技术评审确认）

### 风险识别

**潜在风险**：
1. 数据计算逻辑调整可能影响现有数据的展示
2. 需要确保调整后的数据与历史数据的一致性
3. 可能影响打印小票的数据展示

**缓解措施**：
1. 详细分析现有计算逻辑，明确调整方向
2. 进行充分的数据验证和测试
3. 检查打印模板中的数据展示逻辑

---

## 🔗 相关资源

### 参考需求

- 类似功能: 营业数据统计相关功能
- 竞品分析: -

### 相关文档

- 产品需求文档 (PRD): -
- 用户调研报告: -
- 技术预研文档: -

**相关代码文件**：
- `main/app/api/v1/cashier/cashier_statistics.go` - 收银端统计接口
- `main/app/service/business.go` - 营业数据业务逻辑
- `main/app/repository/statistics.go` - 统计数据访问层
- `main/app/dto/resp/business_data_resp/base.go` - 响应数据结构

---

## 🤝 需求评审

### 评审参与人

| 角色         | 姓名   | 签名/日期 |
| ------------ | ------ | --------- |
| 产品经理     | {姓名} |           |
| 技术负责人   | {姓名} |           |
| 开发代表     | {姓名} |           |
| 测试代表     | {姓名} |           |
| UI/UX 设计师 | {姓名} |           |

### 评审结论

- [ ] ✅ **批准**：进入技术方案设计阶段
- [ ] 🔄 **修改后批准**：需补充以下内容
- [ ] ❌ **拒绝**：不符合产品规划或优先级

**评审意见**：

```
[记录评审会议的关键讨论和决策]
```

**下一步行动**：

- [ ] 创建 Spec：`story-cashier-business-data-product-summary-category-sales`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 查看准确的按商品和按商品分类的营业数据统计  
**以便于** 准确了解商品销售情况，做出正确的经营决策

### AC 验收标准（初稿）

1. **WHEN** 查询按商品统计接口 **THEN** 系统 **SHALL** 返回正确的商品合计金额
2. **WHEN** 查询按商品分类统计接口 **THEN** 系统 **SHALL** 返回正确的分类销售额
3. **IF** 按商品合计金额求和 **THEN** 系统 **SHALL** 与总营业数据保持一致
4. **IF** 按商品分类销售额求和 **THEN** 系统 **SHALL** 与总营业数据保持一致

### 线框图/原型（可选）

[附加 UI 线框图或原型链接]

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

