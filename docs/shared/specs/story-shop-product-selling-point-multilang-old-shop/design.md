# 旧商家后台 商品卖点多语言支持 设计文档

> 本文档定义旧商家后台商品卖点多语言支持的技术方案与实现细节。

## 📋 概述

功能聚焦在旧商家后台（PHP + Vue）与主业务服务（Go main/）之间的商品卖点链路：在数据库中为 `ttpos_product_package` 引入 `describe_multi_language_name_uuid` 字段，复用 `ttpos_multi_language_name` 表管理多语言卖点，并让后台编辑器/终端展示/缓存刷新全链路感知该字段。范围包含：

- 数据库迁移与历史数据回填
- PHP Shop 后台 Model + Controller + 验证器更新
- Vue 旧后台多语言输入 UI
- Go main 商品查询/缓存返回多语言卖点
- 图片更新后的缓存刷新保障

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- Service 通过接口注入 Repository，禁止直接依赖实现。
- API 响应 `data` 必须为对象，URL 使用 snake_case。
- 错误使用 `errors.WithMessage`，不得 panic。
- DTO/Model 需要使用 `json` 与 `gorm` 标签并运行 `go fmt`.

### PHP 规范 (php.mdc)

- Controller 保持轻量，参数校验放在 Validate/Request 层。
- Service/Model 负责多语言写入逻辑，使用事务包裹保存。
- 输入过滤防止 XSS，严格遵循数组字段白名单。

### API 设计规范 (api.mdc)

- 对外字段命名：`selling_point`（当前语言）、`selling_point_i18n`（map）、`describe_multi_language_name_uuid`（内部）。
- data 字段永远是对象。

### 数据库规范 (database.mdc)

- 新增列遵循 `snake_case`，类型 `BIGINT UNSIGNED NOT NULL DEFAULT 0`。
- 所有迁移需包含 `up()`/`down()` 操作，允许回滚。
- 时间字段使用 `int`，迁移要设置索引。

---

## 🔄 代码复用分析

### 可复用组件

- **多语言名称模型**:
  - PHP: `admin/app/common/model/locale/MultiLanguageName.php`（现有保存/更新能力）
  - Go: `main/app/model/multi_language_name.go`（提供 JSON/语言映射方法）
- **商品缓存**: `main/pkg/cache/product_cache.go`（统一的商品详情缓存）
- **前端多语言输入组件**: `admin/views/shop/src/components/LocaleInput.vue`（已有 name/i18n 录入模式，可扩展为卖点输入）

### 集成点

- Shop 后台商品保存接口 `/shop/product/store.Product/save`。
- Go 商品查询接口 `GET /api/v1/cashier/product/detail`、H5/Member/Tablet 使用的列表接口。
- Redis 缓存键 `product_detail:{company_uuid}:{product_uuid}`。

---

## 🏗️ 架构设计

### 分层设计原则

```
Shop Vue (多语言输入)
   ↓ 调用
Shop PHP Controller (参数校验)
   ↓ 调用
Shop Product Model (保存逻辑) ----> MultiLanguageName Service (写入表)
   ↓ 发布事件
Redis 缓存刷新队列
   ↓
Go main API/Service (读取 describe_multi_language_name_uuid → join ttpos_multi_language_name)
   ↓
前台终端（会员/平板/H5）
```

### 模块职责

- **Vue**: 表单 + 500 字符限制 + 主语言复制。
- **PHP Controller**: JSON 结构 `selling_point_i18n` 解析、长度校验、至少一个语言必填。
- **PHP Model**: 调用 `MultiLanguageName->saveNames()` 保存/更新，回填 `describe_multi_language_name_uuid`。
- **Go Service**: `GetProductDetail()` JOIN 多语言表；`ProductCache` 内部存 `SellingPointI18n`。
- **缓存刷新**: 后台保存成功后调用 `ProductUpdatedEvent`，Go 监听后清理 `product_detail` 缓存。

---

## 🗄️ 数据库设计

### `ttpos_product_package` 字段扩展

```sql
ALTER TABLE `ttpos_product_package`
ADD COLUMN `describe_multi_language_name_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '商品卖点多语言UUID' AFTER `describe`,
ADD INDEX `idx_describe_multi_lang_uuid` (`describe_multi_language_name_uuid`);
```

**迁移策略**

1. `up()`：添加列 + 索引；遍历每批 500 条商品。
2. 对 `describe` 非空记录：
   - 新建或复用 `ttpos_multi_language_name` 记录（仅填充 `zh_name` 初始值）。
   - 回填 `describe_multi_language_name_uuid`。
3. `down()`：删除索引与列。

### 数据模型

```go
// main/app/model/product_package.go
type ProductPackage struct {
    Describe                  string `gorm:"column:describe"`
    DescribeMultiLanguageUuid uint64 `gorm:"column:describe_multi_language_name_uuid"`
}
```

```php
// admin/app/shop/model/product/Product.php
protected $autoWriteTimestamp = false;
protected $schema = [
    'describe' => 'string',
    'describe_multi_language_name_uuid' => 'integer',
];
```

---

## 📊 数据模型 & DTO

### Go DTO

```go
// main/app/dto/resp/product_resp/product.go
type ProductDetailResp struct {
    Uuid                 uint64             `json:"uuid"`
    SellingPoint         string             `json:"selling_point"`
    SellingPointI18n     *dto.LocaleResponse `json:"selling_point_i18n"`
    DescribeMultiLangUuid uint64            `json:"describe_multi_language_name_uuid"`
    ImageUrl             string             `json:"image_url"`
}
```

`SellingPointI18n` 直接重用 `dto.LocaleResponse`，方便终端一次性缓存。

### PHP Request Payload（示例）

```json
{
  "selling_point": "默认卖点",
  "selling_point_i18n": {
    "zh": "清香泰式冬阴功",
    "en": "Fresh Tom Yum Soup",
    "th": "ต้มยำรสสด"
  },
  "images": [...]
}
```

---

## 🔌 API 设计

### Shop 后台

- **URL**: `POST /shop/product/store.Product/save`
- **Body**: 增加 `selling_point_i18n`，结构为语言代码 → 字符串。
- **验证**:
  - `selling_point_i18n.zh` 必填
  - 每个语言长度 ≤ 500
  - 空字符串自动转 `""`，不创建多语言记录
- **响应**: 增加 `describe_multi_language_name_uuid` 字段，便于前端比对。

### Go main

- **URL**: `GET /api/v1/cashier/products/detail`
- **Query**: `lang=zh|en|...`（沿用现有 `lang` 参数）
- **处理**:
  - 通过 `describe_multi_language_name_uuid` JOIN `ttpos_multi_language_name`
  - 优先返回请求语言，fallback 主语言
- **响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "uuid": 900011,
    "selling_point": "Fresh Tom Yum Soup",
    "selling_point_i18n": {
      "zh": "清香冬阴功",
      "en": "Fresh Tom Yum Soup",
      "th": "ต้มยำรสสด"
    },
    "image_url": "https://cdn/xxx.jpg"
  }
}
```

---

## 🧩 组件与接口设计

### PHP 侧

- **验证器**: `admin/app/shop/validate/ProductValidate.php`
  - 新增规则 `selling_point_i18n`（array），限制 500 字符。
- **Model**: `Product::sanitizeProductData()` 中增加 `selling_point_i18n` 处理，调用 `MultiLanguageName->saveLocaleTexts()`。
- **事件**: 保存成功后触发 `ProductUpdated` 事件（若已有），附带 `describe_multi_language_name_uuid` 与图片 UUID。

### Go 侧

- **Repository** (`main/app/repository/product_repo.go`)
  - 在 `select` 语句中 JOIN `ttpos_multi_language_name`（alias `mln`）。
  - 提供 `WithDescribeMultiLang()` 选项返回 `dto.LocaleResponse`。
- **Service** (`main/app/service/product.go`)
  - 解析 `mln` 字段并赋值 `SellingPoint/SellingPointI18n`。
  - 当 `mln` 为空时 fallback `product.Describe`。
- **Cache** (`main/pkg/cache/product_cache.go`)
  - 缓存结构加上 `selling_point` 与 `selling_point_i18n`。
  - 新增 `InvalidateDescribe(companyUuid, productUuid uint64)`。

### Vue 侧

- 在 `admin/views/shop/src/views/product/store/product/part/Buyset.vue`、`.../add.vue`、`.../edit.vue` 中：
  - 引入多语言输入组件
  - 提供“复制主语言”按钮
  - 显示剩余字数（`maxlength=500` + `show-word-limit`）

---

## ⚡ 缓存设计

- **Key**: `ttpos:product:detail:{company_uuid}:{product_uuid}`
- **字段**: 包含图片 URL、`selling_point`, `selling_point_i18n`, `describe_multi_language_name_uuid`.
- **策略**:
  - 保存商品后直接删除缓存键，终端重新加载即获得新数据。
  - 兜底：设置 5 分钟 TTI，避免 stale 数据。
- **监控**:
  - 在清理失败时写入 `warn` 日志 + Prometheus 计数。

---

## 🚨 错误处理

- 如果 `selling_point_i18n` 缺少主语言，PHP 验证器返回 `400`。
- 若多语言保存异常，Shop Model 回滚整个事务并返回错误信息。
- Go Service JOIN 失败时记录日志并返回空字符串，防止接口崩溃。

---

## 🔒 安全设计

- Shop API 继续使用商户 token 权限校验。
- 输入内容进行 `htmlspecialchars` 防止脚本注入。
- Go 层输出前再次进行 `template.HTMLEscapeString`（如终端需原始文本，则仅对 `<` `>` 进行 encode）。

---

## 🧪 测试策略

1. **PHP 单测**：`ProductTest` 新增多语言保存/更新/迁移 UT。
2. **Go 单测**：`product_service_test.go` 验证 `lang` 优先级、fallback、缓存。
3. **前端联调**：使用 Cypress 或 Vitest 检查 500 字符限制与语言切换。
4. **集成测试**：模拟更新卖点后在会员端接口验证 5 分钟内刷新。

---

## 📈 性能与扩展

- JOIN 多语言表需对 `describe_multi_language_name_uuid` 建索引，保证单次查询 < 30ms。
- 缓存对象包含 `selling_point_i18n`（最多 9 个字段），体积 < 2KB，可安全缓存。
- 支持未来新增语言：只需为 `ttpos_multi_language_name` 添加列，业务逻辑无需调整。

---

## 实现清单（高层）

1. 数据库迁移 + 数据回填
2. PHP Model/Controller/Validator 更新
3. Vue 表单改造 + 文案
4. Go DTO/Service/Cache 修改
5. 缓存刷新链路联调
6. 自动化测试 & 文档更新

详细任务见 `tasks.md`。

---

## Graphiti & 活动日志

- 若迁移/多语言最佳实践形成经验，请更新 `docs/agent/templates/graphiti-episode.md` 并互链。
- 活动日志：`docs/team/activities/王昱/2025-11/2025-11-25.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-11-25  
**作者**: 技术团队  
**审核者**: 待定

