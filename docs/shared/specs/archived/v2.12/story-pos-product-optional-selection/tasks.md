# 商品选择-可选属性/加料/套餐分组 任务清单

## 任务概览

本任务清单描述实现"商品选择-可选属性/加料/套餐分组"功能所需的所有任务。

**关联文档**:
- 需求文档: `requirements.md`
- 设计文档: `design.md`
- DooTask 任务: [#37618](https://task.example.com/37618)

---

## 后端任务（Go Main 模块）

### Task 1: 分析和确认现有代码实现情况
- **负责人**: 待分配
- **预计工时**: 1h
- **优先级**: 最高
- **状态**: ✅ 已完成

**描述**: 分析现有代码，确认哪些功能已经实现，哪些需要补充。

**已完成的检查**:
- ✅ 数据库表结构：所有需要的字段已存在
- ✅ Model 层：字段定义正确
- ✅ Repository 层：查询逻辑包含相关字段
- ✅ Service 层 - 查询：`GetProductList` 已正确返回所有字段
- ✅ Service 层 - 兼容性：`product_check.go` 包含 `is_must` 转换逻辑

**待补充的部分**:
- ❌ 订单验证逻辑：需要补充完整的验证逻辑
- ❌ 错误提示信息：需要统一错误提示格式

---

### Task 2: 补充订单提交时的验证逻辑
- **负责人**: 待分配
- **预计工时**: 3h
- **优先级**: 高
- **状态**: 🚧 进行中（部分已实现）
- **依赖**: Task 1

**描述**: 在 `main/app/service/order_product.go` 中补充完整的验证逻辑。

**子任务**:

#### 2.1 补充属性组最小选择数量验证
- **文件**: `main/app/service/order_product.go`
- **位置**: 订单提交验证函数中
- **实现内容**:
  ```go
  // 验证属性组选择数量
  for _, attrGroup := range product.AttributeGroups {
      selectedCount := countSelectedAttributes(orderItem, attrGroup.Uuid)
      if selectedCount < int(attrGroup.MinSelect) {
          return errors.New(fmt.Sprintf("【%s】最少选择%d份", attrGroup.LocaleName, attrGroup.MinSelect))
      }
      if attrGroup.MaxSelect > 0 && selectedCount > int(attrGroup.MaxSelect) {
          return errors.New(fmt.Sprintf("【%s】最多选择%d份", attrGroup.LocaleName, attrGroup.MaxSelect))
      }
  }
  ```

#### 2.2 补充加料最小选择数量验证
- **文件**: `main/app/service/order_product.go`
- **位置**: 订单提交验证函数中
- **实现内容**:
  ```go
  // 验证加料选择数量
  selectedSauceCount := len(orderItem.Sauces)
  if selectedSauceCount < int(product.SauceMinSelection) {
      return errors.New(fmt.Sprintf("加料最少选择%d份", product.SauceMinSelection))
  }
  if product.SauceMaxSelection > 0 && selectedSauceCount > int(product.SauceMaxSelection) {
      return errors.New(fmt.Sprintf("加料最多选择%d份", product.SauceMaxSelection))
  }
  ```

#### 2.3 确认套餐分组验证逻辑的完整性
- **文件**: `main/app/service/order_product.go`
- **位置**: 第 2197 行附近
- **实现内容**: 确认现有的套餐分组验证逻辑是否完整，是否包含：
  - 可选分组的最小数量验证
  - 可选分组的最大数量验证
  - 固定分组的必选验证

#### 2.4 补充套餐内商品属性验证
- **文件**: `main/app/service/order_product.go`
- **位置**: 套餐商品验证函数中
- **实现内容**: 验证套餐内每个商品的属性选择是否满足最小/最大要求

---

### Task 3: 统一错误提示信息
- **负责人**: 待分配
- **预计工时**: 1h
- **优先级**: 中
- **状态**: ⏳ 待开始
- **依赖**: Task 2

**描述**: 统一所有验证错误的提示信息格式，使用国际化支持。

**实现内容**:
1. 在 `i18n` 包中添加错误提示的翻译键：
   - `product.attribute_group_min_select`: `【{name}】最少选择{count}份`
   - `product.sauce_min_select`: `加料最少选择{count}份`
   - `product.package_group_min_select`: `【{name}】最少选择{count}份`

2. 修改验证逻辑，使用国际化翻译：
   ```go
   return errors.New(i18n.T("product.attribute_group_min_select", map[string]interface{}{
       "name": attrGroup.LocaleName,
       "count": attrGroup.MinSelect,
   }))
   ```

---

### Task 4: 编写单元测试
- **负责人**: 待分配
- **预计工时**: 4h
- **优先级**: 中
- **状态**: ⏳ 待开始
- **依赖**: Task 2, Task 3

**描述**: 为新增的验证逻辑编写单元测试。

**测试文件**: `main/app/service/order_product_test.go`

**测试用例**:
1. 套餐分组验证：
   - `TestValidatePackageGroup_MinCount_Pass`
   - `TestValidatePackageGroup_MinCount_Fail`
   - `TestValidatePackageGroup_MaxCount_Pass`
   - `TestValidatePackageGroup_MaxCount_Fail`

2. 属性组验证：
   - `TestValidateAttributeGroup_MinSelect_Pass`
   - `TestValidateAttributeGroup_MinSelect_Fail`
   - `TestValidateAttributeGroup_MaxSelect_Pass`
   - `TestValidateAttributeGroup_MaxSelect_Fail`

3. 加料验证：
   - `TestValidateSauce_MinSelect_Pass`
   - `TestValidateSauce_MinSelect_Fail`
   - `TestValidateSauce_MaxSelect_Pass`
   - `TestValidateSauce_MaxSelect_Fail`

4. 兼容性测试：
   - `TestCompatibility_IsMust_To_MinSelection`

---

### Task 5: API 文档更新
- **负责人**: 待分配
- **预计工时**: 1h
- **优先级**: 低
- **状态**: ⏳ 待开始
- **依赖**: Task 2

**描述**: 更新 Swagger API 文档，补充字段说明。

**实现内容**:
1. 确认 `GetProductList` 接口的响应字段注释是否完整
2. 补充字段说明：
   - `optional_min_count`: 套餐分组最小可选数量
   - `optional_count`: 套餐分组最大可选数量
   - `min_select`: 属性组/加料最小可选数量
   - `max_select`: 属性组/加料最大可选数量

3. 运行 `swag init` 生成最新的文档

---

## 前端任务（需要与前端团队协调）

### Task 6: 前端代码分析
- **负责人**: 前端负责人
- **预计工时**: 2h
- **优先级**: 高
- **状态**: ⏳ 待开始
- **依赖**: Task 1

**描述**: 分析前端代码（可能在 `ttpos-flutter` 仓库或 Vue 前端代码中），确认前端实现情况。

**检查内容**:
- 前端是否正确处理 `optional_min_count` 和 `optional_count` 字段
- 前端是否正确处理 `min_select` 和 `max_select` 字段
- 前端是否有客户端验证逻辑
- 前端是否有提示信息显示

---

### Task 7: 前端商品选择界面调整
- **负责人**: 前端负责人
- **预计工时**: 6h
- **优先级**: 高
- **状态**: ⏳ 待开始
- **依赖**: Task 6

**描述**: 调整前端商品选择界面，支持可选属性/加料/套餐分组。

**实现内容**:
1. 套餐分组界面：
   - 显示"可选（最少X份，最多Y份）"标签
   - 当选择数量 = `optional_count` 时，未选择的项置灰
   - 当选择数量 < `optional_min_count` 时，显示提示信息

2. 属性组界面：
   - 显示"可选（最少X份，最多Y份）"标签
   - 当选择数量 = `max_select` 时，未选择的项置灰
   - 当选择数量 < `min_select` 时，显示提示信息

3. 加料界面：
   - 显示"可选（最少X份，最多Y份）"标签
   - 当选择数量 = `sauce_max_selection` 时，未选择的项置灰
   - 当选择数量 < `sauce_min_selection` 时，显示提示信息

---

### Task 8: 前端提交前验证
- **负责人**: 前端负责人
- **预计工时**: 3h
- **优先级**: 高
- **状态**: ⏳ 待开始
- **依赖**: Task 7

**描述**: 实现前端提交前的验证逻辑，避免无效请求。

**实现内容**:
1. 验证套餐分组选择数量
2. 验证属性组选择数量
3. 验证加料选择数量
4. 显示统一的错误提示信息

---

### Task 9: 前端错误提示优化
- **负责人**: 前端负责人
- **预计工时**: 2h
- **优先级**: 中
- **状态**: ⏳ 待开始
- **依赖**: Task 8

**描述**: 优化前端的错误提示信息显示方式。

**实现内容**:
1. 当后端返回验证错误时，高亮显示未满足的分组/属性/加料
2. 提供"快速选择"按钮，帮助用户快速满足最小选择数量
3. 提供"重置选择"按钮，清空所有选择

---

## 测试任务

### Task 10: 集成测试
- **负责人**: 测试负责人
- **预计工时**: 4h
- **优先级**: 高
- **状态**: ⏳ 待开始
- **依赖**: Task 2, Task 7, Task 8

**描述**: 执行端到端的集成测试。

**测试场景**:
1. POS 端点餐：
   - 选择套餐，不选择某些可选分组，提交订单
   - 选择商品，不选择某些可选属性，提交订单
   - 选择商品，不选择加料，提交订单

2. 自助点餐机：
   - 选择套餐，超过最大数量，系统提示
   - 选择商品，少于最小数量，系统提示

3. 扫码点餐：
   - 所有场景同上

4. 其他终端（assistant、tablet、member）：
   - 抽样测试关键场景

---

### Task 11: 兼容性测试
- **负责人**: 测试负责人
- **预计工时**: 2h
- **优先级**: 中
- **状态**: ⏳ 待开始
- **依赖**: Task 10

**描述**: 测试新旧版本的兼容性。

**测试场景**:
1. 旧版本配置的商品（`is_must = 1`），新版本前端正确处理
2. 新版本配置的商品（`min_selection = 0`），旧版本前端正确处理（如果需要支持）
3. 旧数据迁移后，新版本正确处理

---

### Task 12: 性能测试
- **负责人**: 测试负责人
- **预计工时**: 2h
- **优先级**: 低
- **状态**: ⏳ 待开始
- **依赖**: Task 10

**描述**: 测试系统性能，确保响应时间符合要求。

**测试内容**:
1. 商品数据查询响应时间 < 500ms
2. 订单提交验证响应时间 < 300ms
3. 并发测试：100 用户同时点餐

---

## 文档和部署任务

### Task 13: 用户文档编写
- **负责人**: 产品经理
- **预计工时**: 2h
- **优先级**: 中
- **状态**: ⏳ 待开始
- **依赖**: Task 7

**描述**: 编写用户操作文档，帮助用户理解新功能。

**实现内容**:
1. 商家后台配置指南：如何配置可选属性/加料/套餐分组
2. 收银员操作指南：如何在点餐时选择商品
3. 顾客使用指南：如何在自助点餐时选择商品

---

### Task 14: 部署和灰度发布
- **负责人**: 运维负责人
- **预计工时**: 2h
- **优先级**: 高
- **状态**: ⏳ 待开始
- **依赖**: Task 10, Task 11

**描述**: 部署新版本，执行灰度发布。

**实现内容**:
1. 在测试环境部署，执行完整测试
2. 在生产环境部署，先在 1-2 家门店上线
3. 观察 1-2 天，确认无问题后全量发布
4. 准备回滚方案，如果出现问题可以快速回滚

---

### Task 15: 监控和日志配置
- **负责人**: 运维负责人
- **预计工时**: 1h
- **优先级**: 中
- **状态**: ⏳ 待开始
- **依赖**: Task 14

**描述**: 配置监控和日志，及时发现问题。

**实现内容**:
1. 配置错误监控：记录所有验证失败的情况
2. 配置性能监控：监控商品查询和订单提交的响应时间
3. 配置业务日志：记录商品选择和订单提交的关键操作
4. 配置告警：当错误率或响应时间超过阈值时，发送告警

---

## 任务依赖关系图

```
Task 1 (分析现有代码)
  ├─> Task 2 (补充验证逻辑)
  │    ├─> Task 3 (统一错误提示)
  │    ├─> Task 4 (单元测试)
  │    └─> Task 5 (API 文档更新)
  │
  └─> Task 6 (前端代码分析)
       └─> Task 7 (前端界面调整)
            └─> Task 8 (前端验证)
                 └─> Task 9 (前端错误提示优化)

Task 2 + Task 7 + Task 8
  └─> Task 10 (集成测试)
       ├─> Task 11 (兼容性测试)
       ├─> Task 12 (性能测试)
       └─> Task 14 (部署和灰度发布)
            └─> Task 15 (监控和日志配置)

Task 7
  └─> Task 13 (用户文档编写)
```

---

## 任务优先级排序

| 优先级 | 任务编号 | 任务名称 | 预计工时 |
|--------|---------|---------|---------|
| 最高 | Task 1 | 分析现有代码 | 1h |
| 高 | Task 2 | 补充验证逻辑 | 3h |
| 高 | Task 6 | 前端代码分析 | 2h |
| 高 | Task 7 | 前端界面调整 | 6h |
| 高 | Task 8 | 前端验证 | 3h |
| 高 | Task 10 | 集成测试 | 4h |
| 高 | Task 14 | 部署和灰度发布 | 2h |
| 中 | Task 3 | 统一错误提示 | 1h |
| 中 | Task 4 | 单元测试 | 4h |
| 中 | Task 9 | 前端错误提示优化 | 2h |
| 中 | Task 11 | 兼容性测试 | 2h |
| 中 | Task 13 | 用户文档编写 | 2h |
| 中 | Task 15 | 监控和日志配置 | 1h |
| 低 | Task 5 | API 文档更新 | 1h |
| 低 | Task 12 | 性能测试 | 2h |

**总预计工时**: 36h

---

## 当前进度

- ✅ 已完成: Task 1（分析现有代码）
- 🚧 进行中: Task 2（补充验证逻辑 - 部分已实现）
- ⏳ 待开始: Task 3-15

**下一步行动**:
1. 完成 Task 2.1: 补充属性组验证逻辑
2. 完成 Task 2.2: 补充加料验证逻辑
3. 完成 Task 2.3: 确认套餐分组验证逻辑
4. 完成 Task 2.4: 补充套餐内商品属性验证

---

**文档版本**: v1.0  
**最后更新**: 2025-12-23  
**维护者**: TTPOS Development Team

