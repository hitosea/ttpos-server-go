# 自动外卖显示功能 - 手动测试指南

## 测试前准备

1. 确保数据库已执行迁移,`ttpos_product_category` 表包含 `is_display_in_takeout` 字段
2. 准备一个测试分类,初始 `is_display_in_takeout = 0`
3. 准备一个测试商品

## 测试场景1: 创建外卖商品时自动设置分类外卖显示

### 前置条件
- 分类UUID: 假设为 1000
- 分类 `is_display_in_takeout = 0`

### 测试步骤
1. 登录新管理端
2. 进入商品管理页面
3. 选择一个商品,点击"添加外卖配置"
4. 选择外卖类型: Grab
5. 选择分类: UUID=1000的分类
6. 设置其他必填字段
7. 点击"保存"

### 预期结果
1. 外卖商品创建成功
2. 查询数据库: `SELECT is_display_in_takeout FROM ttpos_product_category WHERE uuid = 1000`
3. 结果应该为 `1`

### SQL验证
```sql
-- 创建前
SELECT uuid, name, is_display_in_takeout FROM ttpos_product_category WHERE uuid = 1000;
-- 结果: is_display_in_takeout = 0

-- 创建外卖商品后
SELECT uuid, name, is_display_in_takeout FROM ttpos_product_category WHERE uuid = 1000;
-- 结果: is_display_in_takeout = 1
```

## 测试场景2: 编辑外卖商品时更换分类

### 前置条件
- 分类A UUID: 1000 (已有外卖商品使用)
- 分类B UUID: 2000 (`is_display_in_takeout = 0`)

### 测试步骤
1. 登录新管理端
2. 进入外卖商品列表
3. 选择一个外卖商品(当前使用分类A)
4. 点击"编辑"
5. 将分类从A改为B
6. 点击"保存"

### 预期结果
1. 外卖商品编辑成功
2. 查询数据库: 分类B的 `is_display_in_takeout = 1`
3. 分类A的状态保持不变

### SQL验证
```sql
-- 编辑前
SELECT uuid, name, is_display_in_takeout FROM ttpos_product_category WHERE uuid = 2000;
-- 结果: is_display_in_takeout = 0

-- 编辑外卖商品后
SELECT uuid, name, is_display_in_takeout FROM ttpos_product_category WHERE uuid = 2000;
-- 结果: is_display_in_takeout = 1
```

## 测试场景3: 幂等性测试

### 前置条件
- 分类UUID: 1000 (`is_display_in_takeout = 1`)

### 测试步骤
1. 创建第二个外卖商品,使用相同的分类(UUID=1000)
2. 保存

### 预期结果
1. 外卖商品创建成功
2. 分类的 `is_display_in_takeout` 仍然为 1
3. 不会报错,不会有异常日志

### 日志检查
查看日志文件,应该看到类似信息:
```
自动设置分类外卖显示成功 categoryUuid=1000
```

## 测试场景4: 特色分类支持

### 前置条件
- 特色分类UUID: 3000 (`is_special = 1`, `is_display_in_takeout = 0`)

### 测试步骤
1. 创建外卖商品
2. 在"特色分类"字段选择UUID=3000的特色分类
3. 保存

### 预期结果
1. 外卖商品创建成功
2. 特色分类的 `is_display_in_takeout = 1`

## 测试场景5: 总部商品编辑

### 前置条件
- 当前不是总店
- 外卖商品是总部下发的 (`headquarter_uuid != 0`)

### 测试步骤
1. 尝试编辑总部外卖商品的分类

### 预期结果
1. 总部商品编辑时,分类不可修改
2. 不会触发自动设置分类外卖显示的逻辑

## 测试场景6: 容错性测试

### 前置条件
- 使用一个不存在的分类UUID: 9999

### 测试步骤
1. 直接在代码中模拟调用 `SetCategoryDisplayInTakeout(ctx, 9999)`

### 预期结果
1. 不会抛出异常
2. 返回 nil (不阻塞主流程)
3. 日志中记录警告信息

## API测试

可以使用 Postman 或 curl 测试API:

### 创建外卖商品
```bash
curl -X POST http://localhost:8080/api/v1/shop/product_takeout/add \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "product_package_uuid": 12345,
    "takeout_type": 1,
    "category_uuid": 1000,
    "status": 1,
    "flavors": []
  }'
```

### 验证分类
```bash
curl -X GET http://localhost:8080/api/v1/shop/product/category?uuid=1000 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

检查响应中的 `is_display_in_takeout` 字段是否为 `1`

## 回归测试

确保本功能不影响现有功能:

1. ✅ 创建普通商品 - 不应该触发外卖显示设置
2. ✅ 编辑普通商品分类 - 不应该触发外卖显示设置
3. ✅ 编辑分类信息 - 原有逻辑正常工作
4. ✅ 删除外卖商品 - 分类状态保持不变(符合需求)

## 测试检查清单

- [ ] 场景1: 创建外卖商品时自动设置
- [ ] 场景2: 编辑外卖商品时自动设置
- [ ] 场景3: 幂等性验证
- [ ] 场景4: 特色分类支持
- [ ] 场景5: 总部商品不触发
- [ ] 场景6: 容错性测试
- [ ] API测试
- [ ] 回归测试
- [ ] 日志记录正常
- [ ] 性能无明显下降

## 问题排查

### 如果自动设置失败

1. 检查日志文件,搜索 `SetCategoryDisplayInTakeout`
2. 确认分类UUID是否正确
3. 确认数据库连接正常
4. 检查 `productSrv` 是否正确初始化

### 日志位置
- 应用日志: `main/storage/logs/`
- 搜索关键字: `自动设置分类外卖显示`

---

**测试完成标准**: 所有场景通过,无异常日志,回归测试通过

**测试日期**: ________  
**测试人员**: ________  
**测试结果**: ________

