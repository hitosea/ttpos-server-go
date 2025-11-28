# Bug-251128-001: 编写卖点内容过长，保存报错

## 基本信息

| 字段       | 值                    |
| ---------- | --------------------- |
| Bug ID     | bug-251128-001        |
| 模块       | product               |
| 严重程度   | medium                |
| 发现版本   | v2.10.9               |
| 发现日期   | 2025-11-28            |
| 发现者     | 王昱                  |
| 状态       | 🟡 规划中              |

## 问题描述

### 现象

在商品编辑页面，当卖点内容超过数据库字段长度限制时，保存商品会报错：

```
Error 1406 (22001): Data too long for column 'en_name' at row 1
```

### 复现步骤

1. 进入商品编辑页面（shop 终端）
2. 在卖点字段中输入超过 255 字符的内容（例如：500 字符）
3. 点击保存
4. 系统报错：`保存卖点多语言失败: Error 1406 (22001): Data too long for column 'en_name' at row 1`

### 预期行为

- 前端应该限制卖点输入长度，不允许超过数据库字段限制
- 或者后端在保存前进行截断处理
- 保存成功，不报错

### 实际行为

- 前端允许输入 500 字符（`CheckLenLocal` 检查长度为 500）
- 后端保存到数据库时，`en_name` 字段是 `VARCHAR(255)`，超出长度限制
- 保存失败，返回数据库错误

## 环境信息

- **部署环境**: 开发环境
- **数据库版本**: MySQL 8.0+
- **相关服务**: Main 模块（Go + Gin）
- **客户端类型**: shop（店铺后台）

## 影响范围

### 受影响的模块

- **Main 模块**（Go）：商品保存逻辑
- **Admin 模块**（PHP）：商品编辑页面（前端验证）

### 受影响的终端

- **shop**：店铺后台商品编辑功能

### 受影响的字段

- 卖点多语言字段（`selling_point_i18n`）
- 数据库表：
  - `ttpos_multi_language_name`：`en_name`, `zh_name`, `zh_tw_name`, `th_name` 等（均为 `VARCHAR(255)`）
  - `ttpos_product_package`：`describe` 字段（`VARCHAR(255)`）

## 初步分析

### 错误日志

```
[app/service/product.go:6023]保存卖点多语言失败: Error 1406 (22001): Data too long for column 'en_name' at row 1
```

### 相关代码位置

1. **前端验证**：`main/app/dto/common_resp.go:134` - `CheckLenLocal` 方法检查长度为 500
2. **后端保存**：`main/app/service/product.go:6022` - `saveSellingPointMultiLanguage` 方法
3. **数据库字段**：`ttpos_multi_language_name.en_name` - `VARCHAR(255)`
4. **PHP 验证**：`admin/app/shop/model/product/Product.php:371` - 验证长度为 500 字符

### 问题根源

1. **前端验证不一致**：
   - Go Main 模块前端验证允许 500 字符
   - PHP Admin 模块验证允许 500 字符
   - 但数据库字段 `en_name` 等只有 `VARCHAR(255)`

2. **数据库字段长度限制**：
   - `ttpos_multi_language_name` 表的所有语言字段都是 `VARCHAR(255)`
   - 无法存储超过 255 字符的内容

3. **验证逻辑问题**：
   - 前端验证长度（500）与数据库字段长度（255）不匹配
   - 后端保存前没有进行二次验证或截断处理

### 可能原因

1. 数据库字段设计时未考虑卖点内容可能较长
2. 前端验证长度设置错误（应该是 255 而不是 500）
3. 后端缺少数据库字段长度验证

## 相关链接

### 代码文件

- `main/app/service/product.go:6022` - 保存卖点多语言逻辑
- `main/app/dto/common_resp.go:134` - 前端长度验证
- `admin/app/shop/model/product/Product.php:371` - PHP 模块验证逻辑
- `main/app/model/multi_language_name.go:11` - 多语言名称模型定义

### 数据库表结构

- `ttpos_multi_language_name` 表
- 字段：`en_name VARCHAR(255)`, `zh_name VARCHAR(255)`, `th_name VARCHAR(255)` 等

### 日志文件

- `main/log/2025-11-28.log` - 错误日志记录

## 下一步

1. ✅ 确认数据库字段长度限制（调整为 1000 字符）
2. ✅ 统一前端和后端的验证逻辑（统一为 1000 字符）
3. ✅ 创建修复方案和任务清单

## 修复方案

- 修复方案: `solution.md`
- 任务清单: `tasks.md`

---

**创建时间**: 2025-11-28 09:41  
**最后更新**: 2025-11-28 09:44

