# Grab 支持的语言代码过滤

## 问题
TTPOS 系统支持 `sv`（瑞典语），但 Grab 平台不支持这个语言代码。如果包含不支持的语言代码，可能导致菜单提交失败或验证错误。

## 解决方案

### 新增方法
在 `GrabConverter` 中添加了 `filterSupportedLanguages` 方法：

```go
// filterSupportedLanguages 过滤出 Grab 支持的语言代码
func (c *GrabConverter) filterSupportedLanguages(translations map[string]string) map[string]string {
	supportedLangs := map[string]bool{
		"zh":   true, // 中文
		"en":   true, // 英文
		"th":   true, // 泰语
		"zhtw": true, // 繁体中文
		"ja":   true, // 日语
		"ko":   true, // 韩语
		"my":   true, // 缅甸语
		"tr":   true, // 土耳其语
		"id":   true, // 印尼语
		"vi":   true, // 越南语
		"km":   true, // 高棉语
		"ms":   true, // 马来语
	}

	filtered := make(map[string]string)
	for lang, value := range translations {
		if supportedLangs[lang] && value != "" {
			filtered[lang] = value
		}
	}

	return filtered
}
```

### 应用位置
1. **分类名称转换** (`convertTTPOSCategory`)
2. **商品名称转换** (`convertTTPOSProduct`)

## Grab 支持的语言（官方文档确认）

根据 Grab 官方文档，**仅支持以下 8 种语言代码**：

| 代码 | 语言 | 应用国家 |
|------|------|----------|
| `en` | English | 所有国家（通用） |
| `zh` | 中文 | Thailand, Singapore, Indonesia |
| `th` | ภาษาไทย | Thailand |
| `ms` | Bahasa Melayu | Malaysia |
| `vi` | Tiếng Việt | Vietnam |
| `id` | Bahasa Indonesia | Indonesia |
| `km` | ភាសាខ្មែរ | Cambodia |
| `my` | မြန်မာ | Myanmar |

## TTPOS 支持但 Grab 不支持的语言

以下语言会被自动过滤，不会出现在导出给 Grab 的菜单中：

- ❌ `zhtw` - 繁体中文（Traditional Chinese）
- ❌ `ja` - 日语（Japanese）
- ❌ `ko` - 韩语（Korean）
- ❌ `tr` - 土耳其语（Turkish）
- ❌ `sv` - 瑞典语（Swedish）

## 示例

### 转换前（TTPOS 数据）
```json
{
  "nameTranslation": {
    "zh": "老板推荐",
    "en": "Manager's Recommendation",
    "th": "แนะนำโดยผู้จัดการ",
    "my": "အလုပ်ရှင်အကြံပြုချက်",
    "tr": "Patronun Tavsiyesi",
    "sv": "Chefens Rekommendation"  // ← 不支持的语言
  }
}
```

### 转换后（Grab 格式）
```json
{
  "nameTranslation": {
    "zh": "老板推荐",
    "en": "Manager's Recommendation",
    "th": "แนะนำโดยผู้จัดการ",
    "my": "အလုပ်ရှင်အကြံပြုချက်",
    "tr": "Patronun Tavsiyesi"
    // sv 已被过滤掉
  }
}
```

## 其他平台扩展

如果将来添加其他外卖平台（如 FoodPanda、LINE MAN），每个平台都应该实现自己的 `filterSupportedLanguages` 方法，因为不同平台支持的语言可能不同。

## 测试建议

1. **验证语言过滤**：确保 `sv` 不出现在 Grab 菜单中
2. **空值过滤**：确保空字符串的翻译也被过滤掉
3. **支持的语言**：确保所有 Grab 支持的语言都能正确通过

---

**修复日期**: 2025-12-10
**修复文件**: `main/app/modules/takeout/infrastructure/adapter/grab/grab_converter.go`
**相关方法**: `filterSupportedLanguages`, `convertTTPOSCategory`, `convertTTPOSProduct`

