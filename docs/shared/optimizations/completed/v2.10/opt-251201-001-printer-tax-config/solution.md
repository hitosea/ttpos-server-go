# Opt-251201-001 优化方案（简化版）

## 需求概述

在 `/shop/printer/customize/config/info` 接口返回的测试数据中，`order.is_contain_tax` 字段需要随商家的实际消费税配置动态返回。

**当前问题**：测试数据中的 `order.is_contain_tax` 字段未动态计算，导致商家预览打印模板时看到的税费显示状态与实际配置不一致。

**影响范围**：
- 接口：`/shop/printer/customize/config/info`
- 方法：`GetTestData`（Line 427-516）
- 文件：`main/app/service/printer.go`

## 优化方案

### 实施方案（唯一方案）

在 `GetTestData` 方法中，动态计算 `order.is_contain_tax` 的值，基于商家的税费配置。

#### 修改位置

`main/app/service/printer.go` 的 `GetTestData` 方法（Line 427-516）

#### 实施代码

在 Line 514 之前（`return testData, nil` 之前）添加以下代码：

```go
// 动态设置 order.is_contain_tax 根据商家配置
if testData["order"] != nil {
    orderData, ok := testData["order"].(map[string]interface{})
    if ok {
        // 获取税率设置
        taxRateSetting, err := settingSrv.GetTaxRateSetting(ctx)
        if err == nil {
            // 根据商家配置设置 is_contain_tax
            // 0-关闭消费税, 1-商品未含税, 2-商品已含税
            taxFeeType := taxRateSetting.GetTaxFeeType()
            orderData["is_contain_tax"] = uint(taxFeeType)
        }
    }
}
```

## 实施步骤

1. **修改代码**（0.5 小时）
   - 在 `GetTestData` 方法中添加动态计算逻辑

2. **测试验证**（0.5 小时）
   - 测试不同税费配置下的返回值
   - 验证前端预览显示正确

3. **上线**（0.25 小时）
   - Code Review
   - 发布到生产环境

**总工时**: 1.25 小时

## 收益评估

- **一致性**: 预览与实际打印保持一致
- **用户体验**: 商家可以准确判断打印效果
- **维护成本**: 减少客服成本

## 风险评估

**风险等级**: 🟢 极低

- 只修改测试数据，不影响实际打印逻辑
- 向后兼容，不破坏现有功能

---

**方案版本**: v2.0（简化版）  
**预计实施时间**: 1.25 小时  
**风险等级**: 极低
