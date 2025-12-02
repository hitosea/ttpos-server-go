# 多语言字段规范示例

> Go Main 模块中多语言字段使用的正确和错误示例

---

## 规范说明

**所有 API 接口的多语言字段，无论是请求参数还是响应参数，都必须使用 `dto.LocaleResponse` 结构**

**规则：**

- ✅ **必须使用** `dto.LocaleResponse` 结构体
- ✅ **字段名称必须使用 `LocaleName` 或 `LocaleXXXName` 格式**（如 `LocaleName`、`LocaleAttributeName`、`LocaleUnitName` 等）
- ❌ **禁止使用** 字符串、map 或其他自定义结构
- ❌ **禁止使用** `Name`、`AttributeName` 等不带 `Locale` 前缀的字段名
- ✅ **包含所有支持的语言字段**：ZH, TH, EN, ZHTW, JA, KO, MY, TR, SV
- ✅ **请求参数和响应参数都必须使用** `dto.LocaleResponse`

---

## ✅ 正确示例

### 响应参数示例

```go
// ✅ 正确 - 响应参数使用 dto.LocaleResponse 和正确的字段命名
type ProductResp struct {
    Uuid          uint64            `json:"uuid"`
    LocaleName    dto.LocaleResponse `json:"locale_name"`    // 多语言名称
    LocaleRemark  dto.LocaleResponse `json:"locale_remark"`  // 多语言备注
    LocaleAttributeName dto.LocaleResponse `json:"locale_attribute_name"` // 属性名称（多语言）
}

// 在 Service 中构建响应
func (s *productSrv) GetProduct(ctx context.Context, uuid uint64) (*ProductResp, error) {
    // 从 model.MultiLanguageName 转换
    localeSrv := service.NewLocaleSrv()
    localeResp := localeSrv.GetLocaleNames(multiLanguageName)
    
    return &ProductResp{
        Uuid:          uuid,
        LocaleName:    localeResp,  // ✅ 使用 LocaleResponse 和正确的字段名
        LocaleRemark:  localeResp,
    }, nil
}
```

### 请求参数示例

```go
// ✅ 正确 - 请求参数使用 dto.LocaleResponse 和正确的字段命名
type ProductCreateReq struct {
    LocaleName    dto.LocaleResponse `json:"locale_name" binding:"required"`    // 多语言名称
    LocaleRemark  dto.LocaleResponse `json:"locale_remark"`                      // 多语言备注
}
```

### 从 JSON 字符串转换

```go
// 从数据库的 name 字段（JSON 格式）转换为 LocaleResponse
import "ttpos-server-go/pkg/language"

jsonStr := `{"zh":"苹果","en":"Apple","th":"แอปเปิล"}`
localeResp := language.JsonToLocaleResponse(jsonStr)  // 返回 *dto.LocaleResponse
```

### 从 model.MultiLanguageName 转换

```go
// 使用 LocaleSrv 转换
localeSrv := service.NewLocaleSrv()
localeResp := localeSrv.GetLocaleNames(multiLanguageName)  // 返回 dto.LocaleResponse
```

### 前端请求示例

```json
{
  "name": {
    "zh": "苹果",
    "en": "Apple",
    "th": "แอปเปิล",
    "zhtw": "蘋果",
    "ja": "りんご",
    "ko": "사과",
    "my": "ပန်းသီး",
    "tr": "Elma",
    "sv": "Äpple"
  }
}
```

---

## ❌ 错误示例

### 请求参数错误示例

```go
// ❌ 错误 - 字段名缺少 Locale 前缀
type ProductCreateReq struct {
    Name dto.LocaleResponse `json:"name" binding:"required"`  // ❌ 错误，应使用 LocaleName
}

// ❌ 错误 - 请求参数使用字符串
type ProductCreateReq struct {
    Name string `json:"name" binding:"required"`  // ❌ 错误，应使用 dto.LocaleResponse
}

// ❌ 错误 - 请求参数使用 map
type ProductCreateReq struct {
    Name map[string]string `json:"name" binding:"required"`  // ❌ 错误
}

// ❌ 错误 - 请求参数使用自定义结构
type ProductCreateReq struct {
    Name struct {
        ZH string `json:"zh"`
        EN string `json:"en"`
    } `json:"name" binding:"required"`  // ❌ 错误，应使用 dto.LocaleResponse
}
```

### 响应参数错误示例

```go
// ❌ 错误 - 响应参数使用字符串
type ProductResp struct {
    Name string `json:"name"`  // ❌ 错误
}

// ❌ 错误 - 响应参数使用 map
type ProductResp struct {
    Name map[string]string `json:"name"`  // ❌ 错误
}

// ❌ 错误 - 响应参数使用自定义结构
type ProductResp struct {
    Name struct {
        ZH string `json:"zh"`
        EN string `json:"en"`
    } `json:"name"`  // ❌ 错误，应使用 dto.LocaleResponse
}
```

**问题：**
- 使用字符串、map 或自定义结构不符合项目规范
- 字段名缺少 `Locale` 前缀会导致命名不一致
- 无法统一处理多语言字段

---

## 相关文档

- [Go Main 核心约束](../../../../.cursor/rules/go-main.mdc) - 多语言字段规范

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

