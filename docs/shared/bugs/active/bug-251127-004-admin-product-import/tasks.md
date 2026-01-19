# Bug-251127-004 修复任务清单

> **当前状态**: 🟡 规划中  
> **开始时间**: 2025-11-27  
> **预计完成**: 2025-11-29  
> **预计工作量**: 4-6 小时

---

## 📋 任务列表

### 1. 后端修复 (2.5小时)

- [x] **添加 DTO 验证标签** `main/app/dto/req/product.go`
  - 需求: 为 `ProductImportListItemReq.NumType` 添加 `binding:"required,oneof=1 2"`
  - 需求: 为 `ProductImportItemReq.NumType` 添加 `binding:"required,oneof=1 2"`
  - 预计时间: 0.5小时
  - 负责人: 待分配
  - 代码位置: 第 271 行和第 311 行
  
- [x] **修正默认值逻辑** `main/app/service/product.go`
  - 需求: 修改第 4671 行附近的 ImportProductList 方法
  - 需求: 修改第 4797 行附近的 ImportProduct 方法
  - 需求: 确保非法值时默认为整数计价（内部值 0）
  - 预计时间: 1小时
  - 负责人: 待分配
  - 关键修改:
    - 4671 行: `products.NumType = utils.IfInt(item.NumType == 2, 1, 0)`
    - 4797 行: `NumType: utils.IfInt(item.NumType == 2, 1, 0)`
  
- [ ] **优化错误处理** `main/app/api/v1/shop/shop_product.go`
  - 需求: 优化 ImportProductList 和 ImportProduct 的错误消息
  - 需求: 添加 formatValidationError 函数（可选）
  - 预计时间: 0.5小时
  - 负责人: 待分配
  - 代码位置: 第 902 行和第 928 行附近
  
- [ ] **添加 Service 层验证** `main/app/service/product.go`
  - 需求: 在 ImportProduct 方法中添加计价方式预验证
  - 需求: 提供清晰的中文错误提示
  - 预计时间: 0.5小时
  - 负责人: 待分配
  - 代码位置: 第 4744 行 ImportProduct 方法开始处

### 2. 前端修复 (2小时)

- [ ] **添加前端验证逻辑** `admin/views/shop/src/views/product/import.vue`
  - 需求: 实现 validateImportData 函数
  - 需求: 验证计价方式必填
  - 需求: 验证计价方式枚举值（1或2）
  - 预计时间: 1小时
  - 负责人: 待分配
  
- [ ] **优化预览显示** `admin/views/shop/src/views/product/import.vue`
  - 需求: 实现 formatNumType 函数
  - 需求: 预览页面正确显示"整数计价"或"小数计价"
  - 预计时间: 0.5小时
  - 负责人: 待分配
  
- [ ] **优化 Excel 解析** `admin/views/shop/src/views/product/import.vue`
  - 需求: 确保 num_type 字段正确解析为数字类型
  - 需求: 数据清洗和类型转换
  - 预计时间: 0.5小时
  - 负责人: 待分配

### 3. 测试验证 (1-1.5小时)

- [ ] **编写后端单元测试** `main/app/dto/req/product_test.go`
  - 需求: 测试 NumType 验证标签是否生效
  - 需求: 覆盖正常值（1,2）和非法值（0,3,-1）
  - 预计时间: 0.5小时
  - 负责人: 待分配
  
- [ ] **编写前端单元测试** `admin/views/shop/src/views/product/__tests__/import.spec.ts`
  - 需求: 测试 validateImportData 函数
  - 需求: 覆盖必填、枚举值、正常值场景
  - 预计时间: 0.5小时
  - 负责人: 待分配
  
- [ ] **集成测试**
  - 需求: 测试完整导入流程
  - 需求: 测试必填约束、枚举值校验、默认值逻辑
  - 测试场景: 见 solution.md 集成测试部分
  - 预计时间: 0.5小时
  - 负责人: 待分配

### 4. 文档更新 (0.5小时)

- [ ] **更新导入模板说明**
  - 需求: Excel 模板中明确标注计价方式说明
  - 需求: 提供示例值
  - 预计时间: 0.2小时
  - 负责人: 待分配
  
- [ ] **更新 API 文档** (如需要)
  - 需求: 更新 Swagger 文档
  - 需求: 更新 validation 规则说明
  - 预计时间: 0.3小时
  - 负责人: 待分配

### 5. 部署上线 (1小时)

- [ ] **代码审查**
  - 需求: 提交 PR 进行 Code Review
  - 需求: 检查验证逻辑完整性
  - 需求: 检查前后端一致性
  - 预计时间: 0.3小时
  - 负责人: 待分配
  
- [ ] **发布到测试环境**
  - 需求: 部署后端代码
  - 需求: 部署前端代码
  - 需求: 执行手动测试（见测试清单）
  - 预计时间: 0.3小时
  - 负责人: 待分配
  
- [ ] **发布到生产环境**
  - 需求: 选择低峰期发布
  - 需求: 监控错误日志
  - 需求: 验证功能正常
  - 预计时间: 0.4小时
  - 负责人: 待分配

---

## 📊 任务统计

- **总任务数**: 16
- **已完成**: 2
- **进行中**: 0
- **未开始**: 14
- **完成率**: 12.5%

---

## 🎯 关键里程碑

| 里程碑 | 截止日期 | 状态 |
|--------|---------|------|
| 后端修复完成 | 2025-11-28 上午 | ⏳ 待开始 |
| 前端修复完成 | 2025-11-28 下午 | ⏳ 待开始 |
| 测试验证完成 | 2025-11-28 晚上 | ⏳ 待开始 |
| 代码审查通过 | 2025-11-29 上午 | ⏳ 待开始 |
| 发布到生产 | 2025-11-29 下午 | ⏳ 待开始 |

---

## 🔧 技术要点

### 后端修改重点

**DTO 验证标签**:
```go
// 修改前
NumType int `json:"num_type"` // 数量计算方法, 1-整数 2-小数

// 修改后
NumType int `json:"num_type" binding:"required,oneof=1 2"` // 数量计算方法, 1-整数 2-小数
```

**默认值逻辑**:
```go
// 修改前（第 4797 行）
NumType: utils.IfInt(item.NumType == 1, 0, 1),
// 问题：当 NumType=0 时，结果为 1（小数）

// 修改后
NumType: utils.IfInt(item.NumType == 2, 1, 0),
// 正确：当 NumType=2 时结果为 1（小数），否则为 0（整数）
```

### 前端修改重点

**验证逻辑**:
```typescript
// 必填验证
if (!row.num_type && row.num_type !== 0) {
  errors.push({ row, message: '计价方式不能为空' });
}

// 枚举值验证
if (row.num_type && ![1, 2, '1', '2'].includes(row.num_type)) {
  errors.push({ row, message: '计价方式必须是1或2' });
}
```

---

## ⚠️ 注意事项

1. **值映射关系务必清楚**:
   - Excel/前端：1=整数，2=小数
   - 数据库：0=整数，1=小数

2. **验证标签顺序**:
   - `binding:"required,oneof=1 2"` - 先检查必填，再检查枚举值

3. **错误消息清晰度**:
   - 必须明确告知用户：1=整数计价，2=小数计价

4. **向后兼容**:
   - 修改不影响现有商品数据
   - 只影响新导入的商品

5. **测试覆盖**:
   - 必须测试空值、非法值、边界值
   - 必须测试前后端验证一致性

---

## 🔗 相关链接

- **Bug 报告**: `bug.md`
- **修复方案**: `solution.md`
- **关联任务**: [DooTask #37080](http://t.hitosea.com/#/task/37080)
- **相关文档**: 
  - [Gin 验证器文档](https://github.com/go-playground/validator)
  - [Go Main 开发规范](../../../../.cursor/rules/go-main.mdc)

---

## 📝 开发日志

### 2025-11-27

- ✅ 创建 Bug 报告（bug.md）
- ✅ 完成技术分析
- ✅ 创建修复方案（solution.md）
- ✅ 创建任务清单（tasks.md）
- ⏳ 待分配开发人员

---

**创建时间**: 2025-11-27 15:05  
**最后更新**: 2025-11-27 15:05  
**文档版本**: v1.0  
**创建人**: weifashi

