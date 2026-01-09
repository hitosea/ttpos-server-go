# 单品备注原因管理 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | 王昱     |
| **日期**   | 2025-12-05 |
| **目标版本** | - |
| **状态**   | 已创建 Spec   |
| **关联任务** | - |
| **关联 Spec** | [story-main-order-item-remark-reason-management](../../../shared/specs/archived/v2.12/story-main-order-item-remark-reason-management/requirements.md)      |

---

## 🎯 背景和动机

### 问题描述

目前系统中"整单备注"已有完整的原因管理功能（增删改查），但"单品备注"仅支持自由文本输入，缺少预设原因管理功能。商户无法统一管理常用的单品备注原因，导致：

1. 收银员需要手动输入备注，效率较低
2. 不同收银员输入的备注格式不统一，影响后厨理解
3. 无法统计和分析常用的单品备注原因

### 业务价值

- 提升收银效率：收银员可通过选择预设原因快速添加单品备注
- 统一备注格式：通过预设原因管理，确保备注信息格式统一
- 改善后厨体验：统一的备注格式便于后厨理解订单需求
- 数据统计支持：为后续分析常用备注原因提供数据基础

### 目标用户

- [x] 商户管理员（配置单品备注原因）
- [x] 收银员（使用预设原因添加单品备注）
- [ ] 厨房人员
- [ ] 顾客
- [ ] 其他: ________

---

## 💡 解决方案概述

### 方案描述

参考"整单备注"的实现逻辑，为"单品备注"添加原因管理功能。在旧商户后台（PHP）和新管理端（Go）的业务设置中，新增"单品备注"原因管理模块，支持多语言、增删改查等操作，与"整单备注"逻辑保持一致。

**本次仅实现后端 API，前端界面后续实现。**

### 核心功能点

1. **新管理端（Go Main 模块）**
   - 获取单品备注原因列表 API：`GET /shop/setting/order_item_remark`
   - 新增单品备注原因 API：`POST /shop/setting/order_item_remark/add`
   - 编辑单品备注原因 API：`POST /shop/setting/order_item_remark/edit`
   - 删除单品备注原因 API：`DELETE /shop/setting/order_item_remark`

2. **旧商户后台（PHP Admin 模块）**
   - 在 `admin/app/shop/controller/setting/Business.php` 中新增 `orderItemRemark()` 方法
   - 支持批量增删改操作（与 `orderRemark()` 方法逻辑一致）

3. **数据模型**
   - 创建 `OrderItemRemark` 模型和 `order_item_remark` 数据表（参考 `OrderRemark` 表结构）
   - 支持多语言名称（MultiLanguageName）
   - 限制数量不超过 100 个
   - 字数限制（非字符）：100 

### 影响范围

**涉及终端**：
- [ ] POS 收银端（后续使用）
- [x] Shop 商家管理端（配置管理）
- [ ] KDS 厨显端
- [ ] QDS 排号叫号端
- [ ] Assistant 助手端（后续使用）
- [ ] Tablet 平板端
- [ ] Mobile 扫码端
- [ ] Menu 电子菜单端
- [ ] Member 会员端

**涉及模块**：
- [ ] UI 组件（本次不实现）
- [x] API 接口（本次实现）
- [x] 数据模型（本次实现）
- [x] 业务逻辑（本次实现）
- [ ] 第三方集成
- [ ] 其他: ________

---

## 📊 初步评估

### 技术复杂度

- [ ] **低**：纯 UI 调整，无业务逻辑变更
- [x] **中**：需要前后端联调，基础业务逻辑
- [ ] **高**：涉及架构调整、第三方集成、复杂算法

**说明**：参考现有"整单备注"实现，技术方案成熟，复杂度中等。

### 工作量预估

[粗略估算，具体 SP 在技术方案评审时确定]

- **预计天数**: 2-3 天
- **预估 SP**: 3-5 SP（待技术评审确认）

**工作项分解**：
1. 数据模型设计（0.5 天）
2. Go Main 模块 API 实现（1 天）
3. PHP Admin 模块 API 实现（0.5 天）
4. 单元测试和联调（0.5-1 天）

### 风险识别

**潜在风险**：
1. 数据迁移：如果已有商户使用单品备注，需要考虑数据兼容性
2. API 设计：需要确保与"整单备注"API 设计保持一致
3. 多语言支持：需要确保多语言名称的验证逻辑正确

**缓解措施**：
1. 数据迁移：本次仅新增原因管理功能，不影响现有单品备注文本字段，无需数据迁移
2. API 设计：严格参考 `OrderRemark` 相关 API 的实现逻辑
3. 多语言支持：复用现有的多语言验证逻辑

---

## 🔗 相关资源

### 参考需求

- 类似功能: `整单备注` 原因管理功能
  - Go API: `main/app/api/v1/shop/shop_setting.go` (GetOrderRemark, AddOrderRemark, EditOrderRemark, DeleteOrderRemark)
  - Go Service: `main/app/service/other.go` (AddOrderRemark, EditOrderRemark, DeleteOrderRemark)
  - Go Repository: `main/app/repository/base/order_remark.go`
  - PHP API: `admin/app/shop/controller/setting/Business.php` (orderRemark 方法)

### 相关文档

- 数据库迁移示例: `admin/database/migrations/20251020134645_create_order_remark_table.php`
- 订单备注功能: `main/app/service/order_manage.go` (OrderRemark 方法)
- 订单商品备注功能: `main/app/service/order_product.go` (OrderProductRemark 方法)

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

- [x] 创建 Spec：`story-main-order-item-remark-reason-management`
- [ ] 分配负责人：{姓名}
- [ ] 目标 Sprint：Sprint {N}

---

## 📝 附录

### User Story（初稿）

**作为** 商户管理员  
**我想** 在业务设置中管理单品备注原因列表  
**以便于** 收银员可以快速选择预设原因添加单品备注，提升效率和统一性

**作为** 收银员  
**我想** 在添加单品备注时选择预设原因  
**以便于** 快速完成备注操作，避免手动输入错误

### AC 验收标准（初稿）

1. **WHEN** 商户管理员访问业务设置-原因管理 **THEN** 系统 **SHALL** 显示"整单备注"和"单品备注"两个选项
2. **WHEN** 商户管理员新增单品备注原因 **THEN** 系统 **SHALL** 支持多语言名称输入，并限制数量不超过 100 个
3. **WHEN** 商户管理员编辑单品备注原因 **THEN** 系统 **SHALL** 更新多语言名称信息
4. **WHEN** 商户管理员删除单品备注原因 **THEN** 系统 **SHALL** 软删除记录，不影响历史订单数据
5. **WHEN** 调用获取单品备注原因列表 API **THEN** 系统 **SHALL** 返回当前商户的所有有效原因列表（支持多语言）
6. **IF** 单品备注原因数量已达到 100 个 **THEN** 系统 **SHALL** 拒绝新增操作并提示错误信息

### 技术实现要点

1. **数据表设计**（参考 `order_remark` 表）：
   - 表名：`order_item_remark`
   - 模型名：`OrderItemRemark`
   - 字段：id（主键）, uuid, name, multi_language_name_uuid, create_time, update_time, delete_time
   - 索引：id（主键），uuid（唯一索引）

2. **API 路径设计**（参考整单备注）：
   - `GET /shop/setting/order_item_remark` - 获取列表
   - `POST /shop/setting/order_item_remark/add` - 新增
   - `POST /shop/setting/order_item_remark/edit` - 编辑
   - `DELETE /shop/setting/order_item_remark` - 删除

3. **PHP 接口设计**（参考 orderRemark 方法）：
   - 路径：`/index.php/shop/setting.Business/orderItemRemark`
   - 支持 GET（查询）和 POST（批量增删改）
   - 参数格式：`order_item_remark` 数组，包含 id, remark（JSON 多语言）, action（add/edit/delete）

### 线框图/原型（可选）

[本次仅实现后端 API，前端界面后续实现]

---

## 📄 模板使用说明

### 何时使用此模板

- ✅ 产品经理提出新功能想法
- ✅ 用户反馈需求建议
- ✅ 技术团队提出改进方案
- ✅ 需要团队讨论和评审的需求

### 与 Spec 的区别

| 阶段        | 文档类型      | 详细程度 | 用途               |
| ----------- | ------------- | -------- | ------------------ |
| **需求发起** | Proposal      | 粗略     | 团队评审、决策是否做 |
| **需求确认** | Requirements  | 详细     | User Story + AC，开发依据 |
| **技术设计** | Design        | 详细     | 技术方案，实现指导 |
| **任务分解** | Tasks         | 详细     | 开发执行，进度追踪 |

### 流转路径

```
提案 (Proposal) 
  ↓ 评审批准
需求文档 (Requirements) 
  ↓ 技术评审
设计文档 (Design) 
  ↓ SP 评估 ≤ 5
任务分解 (Tasks)
  ↓
开发实现
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-05  
**维护者**: 产品组 + Scrum Master  
**相关规范**: `.cursor/rules/scrum_story_point.mdc`, `.cursor/rules/specs.mdc`

