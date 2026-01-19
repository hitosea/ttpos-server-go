# Bug-260114-003 修复任务清单

> **当前状态**: 🟡 规划中
> **开始时间**: 2026-01-14
> **预计完成**: 2026-01-15

---

## 📋 任务列表

### 1. 代码修复

- [x] **修复 CountTakeoutProduct 函数** `main/app/repository/statistics_takeout.go:766-844`
  - 需求: 修改 SQL 查询，添加 `product_bom` 表关联，使用 `product_bom.price` 作为商品单价
  - 详细说明:
    - 普通商品：通过 `takeout_order_item_modifier` 关联 `product_bom`（`ttpos_modifier_type = 'flavor'`）
    - 套餐商品：通过 `takeout_order_item` 关联 `product_bom`（`product_flavor_uuid = 0 AND product_sauce_uuid = 0`）
    - 使用 `COALESCE` 处理 NULL 值
  - 预计时间: 2小时
  - 负责人: 已完成
  - 完成时间: 2026-01-14

### 2. 测试验证

- [ ] **编写单元测试** `main/app/repository/statistics_takeout_test.go`
  - 需求: 为 `CountTakeoutProduct` 函数编写单元测试
  - 测试场景:
    - 普通商品价格获取（通过 modifier 关联 product_bom）
    - 套餐商品价格获取（直接关联 product_bom）
    - NULL 值处理（没有关联 product_bom 的情况）
    - 多语言支持（商品名称和规格名称）
  - 预计时间: 2小时
  - 负责人:

- [ ] **集成测试**
  - 需求: 端到端测试修复效果
  - 测试场景:
    - 创建多个订单（包含普通商品和套餐商品）
    - 调用统计接口，验证返回结果
    - 对比修复前后的统计数据
  - 预计时间: 1小时
  - 负责人:

- [ ] **手动验证**
  - 需求: 在测试环境复现并验证
  - 测试步骤:
    1. 部署修复代码到测试环境
    2. 进入收银端-营业数据页面
    3. 选择"按商品统计"
    4. 查看外卖订单商品的单价
    5. 验证单价是否正确（对比 product_bom 表中的价格）
  - 预计时间: 1小时
  - 负责人:

### 3. 文档更新

- [ ] **更新代码注释**
  - 需求: 在 `CountTakeoutProduct` 函数中添加详细注释
  - 说明:
    - 价格获取逻辑（普通商品和套餐商品的不同处理）
    - product_bom 表的关联条件
    - 商品名称和规格名称的获取方式
  - 预计时间: 0.5小时
  - 负责人:

- [ ] **更新故障排查指南**（如果需要）
  - 需求: 如果相关文档存在，更新统计功能的价格来源说明
  - 预计时间: 0.5小时
  - 负责人:

### 4. 部署上线

- [ ] **代码审查**
  - 需求: 通过 Code Review
  - 审查重点:
    - SQL 查询逻辑正确性
    - 性能影响评估
    - 错误处理完整性
  - 预计时间: 1小时
  - 负责人:

- [ ] **发布到测试环境**
  - 需求: 部署并验证
  - 验证内容:
    - 功能正常
    - 性能正常
    - 数据准确性
  - 预计时间: 0.5小时
  - 负责人:

- [ ] **发布到生产环境**
  - 需求: 生产发布并监控
  - 监控指标:
    - 统计查询性能
    - 数据准确性
    - 错误日志
  - 预计时间: 1小时
  - 负责人:

---

## 📊 任务统计

- **总任务数**: 8
- **已完成**: 1
- **进行中**: 0
- **未开始**: 7
- **完成率**: 12.5%

---

## 🔗 相关链接

- Bug 报告: `bug.md`
- 修复方案: `solution.md`
- 代码位置: `main/app/repository/statistics_takeout.go:766-844`
- 相关模型:
  - `main/app/model/product.go:ProductBom`
  - `main/app/modules/takeout/domain/model/takeout_order_item.go`
  - `main/app/modules/takeout/domain/model/takeout_order_item_modifier.go`
