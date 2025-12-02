# Bug-251202-001 修复任务清单

> **当前状态**: 🟡 规划中
> **开始时间**: 2025-12-02
> **预计完成**: 2025-12-02
> **负责人**: 待指派

---

## 📋 任务列表

### 1. 代码修复

- [x] **扩展响应结构** `main/app/dto/resp/statistics_channel_resp.go`
  - 需求: 在 `ChannelSalesBlock` 中新增 `OrderAmountMealAvg` 字段（`order_meal_avg_amount`）
  - 预计时间: 0.5小时
  - 负责人: 
  - 参考: `main/app/dto/resp/business_data_resp/base.go:384` - `OrderAmountMealAvg` 字段定义
  - ✅ 已完成
  
- [x] **扩展 Repository 模型** `main/app/model/statistics.go`
  - 需求: 在 `ChannelSaleRepoResult` 中新增 `OrderAmountMealAvg` 字段（`sql.NullFloat64` 类型）
  - 预计时间: 0.5小时
  - 负责人: 
  - 参考: `main/app/model/statistics.go:378-386` - 现有字段定义
  - ✅ 已完成
  
- [x] **Repository 层计算** `main/app/repository/statistics.go:1470-1477`
  - 需求: 在 `CountChannelSale` 方法的桌台渠道查询中，添加人均订单金额计算 SQL
  - 计算公式: `ROUND(SUM(t.desk_order_amount) / NULLIF(SUM(IF(t.desk_uuid > 0 AND t.is_takeout = 0, t.meal_num, 0)), 0), 2) AS order_amount_meal_avg`
  - 预计时间: 1小时
  - 负责人: 
  - 注意: 使用 `NULLIF` 避免除零错误，保留两位小数
  - ✅ 已完成
  
- [x] **Service 层转换** `main/app/service/business.go:2824-2856`
  - 需求: 在 `convertToBlock` 函数中，添加人均订单金额的转换逻辑
  - 要求: 使用 `decimal.Decimal` 类型确保精度，保留两位小数（`.Round(2).InexactFloat64()`）
  - 预计时间: 1小时
  - 负责人: 
  - 参考: `main/app/service/statistics.go:2157-2163` - 综合运营统计的实现方式
  - ✅ 已完成
  
- [x] **导出功能扩展 - 多语言标签** `main/app/service/business.go:3048-3122`
  - 需求: 在 `labelMap` 中添加 `order_meal_avg` 的多语言标签
  - 语言: zh, en, th, zhtw, ja, ko, my, tr, sv
  - 预计时间: 1小时
  - 负责人: 
  - 参考: 现有标签定义格式
  - ✅ 已完成
  
- [x] **导出功能扩展 - Excel 列** `main/app/service/business.go:3164-3201`
  - 需求: 在桌台渠道导出部分，添加人均订单金额行
  - 位置: 在"人数"行之后，"最小订单金额"行之前
  - 预计时间: 1小时
  - 负责人: 
  - 注意: 需要更新 `tableEndRow` 的行号
  - ✅ 已完成

### 2. 测试验证

- [ ] **编写 Repository 单元测试** `main/app/repository/statistics_test.go`
  - 需求: 测试 `CountChannelSale` 方法返回的人均订单金额计算正确
  - 测试用例:
    - 正常情况：有订单和人数，验证计算正确
    - 边界情况：人数为 0，验证返回 NULL
    - 精度测试：验证保留两位小数
  - 预计时间: 2小时
  - 负责人: 
  
- [ ] **编写 Service 单元测试** `main/app/service/business_test.go`
  - 需求: 测试 `CountChannelSales` 方法转换人均订单金额正确
  - 测试用例:
    - 正常情况：验证转换正确
    - NULL 值处理：验证 NULL 值转换为 0
    - 精度测试：验证使用 decimal 保留两位小数
  - 预计时间: 2小时
  - 负责人: 
  
- [ ] **API 接口集成测试**
  - 需求: 调用渠道营业统计接口，验证桌台渠道包含人均订单金额字段
  - 测试步骤:
    1. 调用 `GET /api/v1/shop/statistics/channel_sales`
    2. 验证响应中 `table.order_meal_avg_amount` 字段存在
    3. 验证字段值计算正确（订单总金额 / 用餐人数）
    4. 验证精度：保留两位小数
  - 预计时间: 1小时
  - 负责人: 
  
- [ ] **导出功能集成测试**
  - 需求: 导出渠道营业统计，验证桌台渠道包含人均订单金额列
  - 测试步骤:
    1. 调用导出接口
    2. 下载 Excel 文件
    3. 验证桌台渠道包含"人均订单金额"列
    4. 验证数值正确且保留两位小数
    5. 验证多语言标签正确显示
  - 预计时间: 1小时
  - 负责人: 
  
- [ ] **手动功能测试**
  - 需求: 在测试环境验证功能正常
  - 测试场景:
    - 正常数据：验证人均订单金额计算正确
    - 边界情况：人数为 0、订单金额为 0
    - 大量数据：验证性能表现
  - 预计时间: 1小时
  - 负责人: 

### 3. 文档更新

- [ ] **更新 API 文档** `main/docs/swagger.yaml` 或 `main/docs/swagger.json`
  - 需求: 更新渠道营业统计接口文档，说明新增字段 `order_meal_avg_amount`
  - 字段说明: 人均订单金额，计算公式为订单总金额 / 用餐人数，保留两位小数
  - 预计时间: 0.5小时
  - 负责人: 
  
- [ ] **更新代码注释**
  - 需求: 在相关代码中添加注释，说明人均订单金额的计算逻辑
  - 文件:
    - `main/app/repository/statistics.go:1477` - SQL 计算注释
    - `main/app/service/business.go:2855` - Service 层转换注释
  - 预计时间: 0.5小时
  - 负责人: 

### 4. 部署上线

- [ ] **代码审查**
  - 需求: 通过 Code Review，确保代码质量和规范
  - 检查项:
    - 代码风格符合 `.cursor/rules/go-main.mdc` 规范
    - 计算逻辑正确，使用 decimal 保证精度
    - 边界情况处理完善
    - 注释清晰完整
  - 预计时间: 1小时
  - 负责人: 
  
- [ ] **发布到测试环境**
  - 需求: 部署到测试环境并验证
  - 验证项:
    - 接口功能正常
    - 导出功能正常
    - 性能表现正常
  - 预计时间: 1小时
  - 负责人: 
  
- [ ] **发布到生产环境**
  - 需求: 生产发布并监控
  - 监控项:
    - 接口响应时间
    - 错误率
    - 数据准确性
  - 预计时间: 1小时
  - 负责人: 

---

## 📊 任务统计

- **总任务数**: 15
- **已完成**: 6
- **进行中**: 0
- **未开始**: 9
- **完成率**: 40%

---

## 🔗 相关链接

- Bug 报告: `bug.md`
- 修复方案: `solution.md`
- 关联 Spec: `docs/shared/specs/active/story-shop-channel-sales/` - 渠道营业统计功能规格
- 参考实现: `main/app/service/statistics.go:2157-2163` - 综合运营统计人均订单金额计算

---

## 📝 注意事项

1. **精度要求**: 所有金额计算必须使用 `decimal.Decimal` 类型，保留两位小数
2. **边界处理**: 当用餐人数为 0 时，人均订单金额应为 0（SQL 中使用 `NULLIF` 处理）
3. **代码风格**: 遵循 `.cursor/rules/go-main.mdc` 规范，参考综合运营统计的实现方式
4. **测试覆盖**: 确保单元测试和集成测试覆盖所有边界情况
5. **向后兼容**: 新增字段为可选字段，不影响现有前端代码

