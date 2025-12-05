# 物品管理-原产地功能需求

> 本文档定义物品管理中原产地字段的功能需求。

## 📋 概述

在物品详情中增加"原产地"字段，支持从国家字典库的197个国家/地区中选择。该功能用于记录物品的原产地信息，便于商家管理和追溯。

**版本**: v2.11.0  
**创建日期**: 2025-12-05  
**来源**: DooTask #37483

---

## 🎯 业务目标

- 为物品增加原产地信息记录能力
- 支持从197个国家/地区中选择原产地
- 提供国家列表接口，支持多语言显示
- 前端可自行实现国家搜索功能

---

## 📝 功能需求

### R1: 数据库设计

**R1.1** `ttpos_material` 表增加原产地国家字段

- 字段名: `origin_country_code`
- 类型: `varchar(10)` 
- 默认值: `''`
- 注释: `原产地国家编码（ISO 3166-1 alpha-2，如：CN, US, TH）`
- 约束: 可为空，存储国家编码（英文国家名）

**R1.2** 数据库迁移

- 创建迁移文件: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_origin_country_code_to_material_table.php`
- 更新 Seeds 文件: `admin/database/seeds/shop_01.sql`

---

### R2: 国家枚举常量

**R2.1** 创建国家枚举文件

- 文件路径: `main/app/constant/country.go`
- 包含197个国家/地区的枚举定义
- 每个国家包含：
  - Code: 国家编码（ISO 3166-1 alpha-2，如：CN, US, TH）
  - Name: 英文国家名
  - LocaleNames: 多语言国家名称（zh, zhtw, en, ja, ko, my, th, tr, de, sv）

**R2.2** 国家数据来源

- 使用 AI 查询197个国家并输出枚举文件
- 参考 ISO 3166-1 标准
- 支持多语言名称

---

### R3: API 接口设计

**R3.1** 物品详情接口 - 增加原产地字段

- 接口: `GET /api/v1/shop/material/detail`
- 响应字段: `origin_country_code` (string, 可选)
- 响应字段: `origin_country` (object, 可选)
  - `code`: 国家编码
  - `locale_name`: 多语言国家名称

**R3.2** 物品创建接口 - 增加原产地字段

- 接口: `POST /api/v1/shop/material/add`
- 请求字段: `origin_country_code` (string, 可选)

**R3.3** 物品编辑接口 - 增加原产地字段

- 接口: `POST /api/v1/shop/material/edit`
- 请求字段: `origin_country_code` (string, 可选)

**R3.4** 国家列表接口

- 接口: `GET /api/v1/shop/country/list`
- 响应格式:
  ```json
  {
    "code": 1,
    "message": "success",
    "data": {
      "list": [
        {
          "code": "CN",
          "locale_name": {
            "zh": "中国",
            "zhtw": "中國",
            "en": "China",
            "ja": "中国",
            "ko": "중국",
            "my": "တရုတ်",
            "th": "จีน",
            "tr": "Çin",
            "de": "China",
            "sv": "Kina"
          }
        }
      ]
    }
  }
  ```
- 要求: 
  - 返回197个国家列表
  - 国家名称多语言结构返回
  - `code` 表示国家编码，用英文国家名替代（实际使用 ISO 3166-1 alpha-2 编码）

---

### R4: 前端功能

**R4.1** 物品详情页面

- 显示原产地字段
- 支持选择国家（下拉选择器）
- 显示多语言国家名称

**R4.2** 物品创建/编辑页面

- 增加原产地选择器
- 支持从197个国家中选择
- 国家搜索由前端自己实现

---

## 🔒 约束条件

### 技术约束

- 国家编码使用 ISO 3166-1 alpha-2 标准（2位字母，如：CN, US, TH）
- 多语言支持：zh, zhtw, en, ja, ko, my, th, tr, de, sv
- 原产地字段为可选字段，可为空

### 业务约束

- 原产地信息仅用于记录，不影响业务逻辑
- 国家列表接口无需分页（197个国家一次性返回）
- 国家搜索功能由前端实现，后端不提供搜索接口

---

## ✅ 验收标准

1. ✅ `ttpos_material` 表成功增加 `origin_country_code` 字段
2. ✅ 国家枚举文件包含197个国家/地区
3. ✅ 物品详情接口返回原产地信息
4. ✅ 物品创建/编辑接口支持设置原产地
5. ✅ 国家列表接口返回197个国家，支持多语言
6. ✅ 前端可正常选择原产地国家
7. ✅ 数据库迁移和 Seeds 文件已更新

---

## 📚 相关文档

- DooTask: #37483
- 数据库规范: `.cursor/rules/database.mdc`
- API 设计规范: `.cursor/rules/api.mdc`
- Go Main 规范: `.cursor/rules/go-main.mdc`

---

**审核状态**: ✅ 已通过  
**审核日期**: 2025-12-05  
**审核者**: 产品组

