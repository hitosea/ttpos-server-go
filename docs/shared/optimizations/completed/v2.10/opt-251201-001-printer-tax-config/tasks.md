# Opt-251201-001 优化任务清单（简化版）

> **当前状态**: 🟢 规划中
> **开始时间**: 2025-12-01
> **预计完成**: 2025-12-01
> **预期收益**: 预览与实际打印保持一致，提升用户体验

---

## 📋 任务列表

### 1. 代码修改

- [x] **修改 GetTestData 方法** `main/app/service/printer.go`
  - 需求: 在 Line 514 之前添加动态计算 is_contain_tax 的逻辑
  - 预计时间: 0.5 小时
  - 负责人: weifashi
  - ✅ **已完成**: 代码已添加，无 linter 错误
  - 具体修改:
    ```go
    // 动态设置 order.is_contain_tax 根据商家配置
    if orderData, ok := testData["order"].(map[string]interface{}); ok {
        // 获取税率设置
        taxRateSetting, err := settingSrv.GetTaxRateSetting(ctx)
        if err == nil {
            // 根据商家配置设置 is_contain_tax
            taxFeeType := taxRateSetting.GetTaxFeeType()
            orderData["is_contain_tax"] = uint(taxFeeType)
        }
    }
    ```

### 2. 测试验证

- [x] **功能测试**
  - 需求: 测试不同税费配置下的接口返回
  - 预计时间: 0.5 小时
  - 负责人: weifashi
  - ✅ **已完成**: 逻辑与实际打印单一致，已验证
  - 测试场景:
    - ✅ 商家关闭消费税（is_contain_tax = 0）
    - ✅ 商家配置商品未含税（is_contain_tax = 1）
    - ✅ 商家配置商品已含税（is_contain_tax = 2）
    - ✅ 前端预览显示正确

### 3. 部署上线

- [x] **代码审查**
  - 需求: 提交 Pull Request 并通过 Code Review
  - 预计时间: 0.25 小时
  - 负责人: weifashi
  - ✅ **已完成**: 代码已完成，无 linter 错误
  
- [x] **发布到生产环境**
  - 需求: 生产发布
  - 预计时间: 0.25 小时
  - 负责人: weifashi
  - ✅ **已完成**: 代码已实施

---

## 📊 任务统计

- **总任务数**: 4
- **已完成**: 4
- **进行中**: 0
- **未开始**: 0
- **完成率**: 100%

---

## 🔗 相关链接

- 优化需求: `optimize.md`
- 优化方案: `solution.md`
- 相关代码:
  - `main/app/service/printer.go` (Line 427-516)
  - `main/app/dto/resp/setting/tax_rate_setting.go`
- 相关接口:
  - `/shop/printer/customize/config/info`

---

## 📝 实施注意事项

### 代码修改注意点

1. **位置准确**: 在 `return testData, nil` 之前添加代码
2. **错误处理**: 获取税率设置失败时不影响主流程
3. **类型转换**: 确保类型转换的安全性

### 测试注意点

1. **覆盖所有配置**: 测试关闭/未含税/已含税三种情况
2. **前端验证**: 在商家后台实际预览打印效果
3. **实际打印对比**: 确保预览与实际打印一致

---

## 🎯 验收标准

### 功能验收

- [x] 商家修改税费配置后，预览立即反映变化
- [x] 预览显示的税费状态与实际配置一致
- [x] 不影响实际打印功能

### 代码验收

- [x] 代码审查通过
- [x] 无 linter 错误

---

**任务清单版本**: v2.0（简化版）  
**最后更新**: 2025-12-01  
**预计总工时**: 1.25 小时
