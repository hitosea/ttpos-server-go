# Grab 菜单转换 - 默认语言自动设置

> 根据店铺时区自动设置 Grab 菜单的默认语言

---

## 📋 需求背景

在 Grab 菜单转换过程中，默认语言之前是硬编码为 `"en"`（英语）。但不同国家/地区的店铺应该根据其时区自动设置对应的默认语言，以提供更好的本地化体验。

例如：
- 泰国店铺（`Asia/Bangkok`）→ 泰语（`th`）
- 中国店铺（`Asia/Shanghai`）→ 中文（`zh`）
- 马来西亚店铺（`Asia/Kuala_Lumpur`）→ 马来语（`ms`）

---

## ✅ 实现方案

### 1. 时区到语言的映射

新增 `pkg/language/language.go` 中的 `GetLanguageByTimezone()` 函数，根据时区返回对应的语言代码。

**支持的时区映射：**

| 国家/地区 | 时区示例 | 语言代码 |
|----------|---------|---------|
| 中国 | `Asia/Shanghai`、`Asia/Hong_Kong` | `zh` |
| 泰国 | `Asia/Bangkok` | `th` |
| 马来西亚 | `Asia/Kuala_Lumpur` | `ms` |
| 新加坡 | `Asia/Singapore` | `en` |
| 越南 | `Asia/Ho_Chi_Minh` | `vi` |
| 印度尼西亚 | `Asia/Jakarta` | `id` |
| 柬埔寨 | `Asia/Phnom_Penh` | `km` |
| 缅甸 | `Asia/Yangon` | `my` |
| 其他 | 未知时区 | `en`（默认） |

### 2. 修改 GrabConverter 构造函数

```go
// NewGrabConverter 创建 Grab 转换器
// timezone: 店铺时区，用于自动设置默认语言
func NewGrabConverter(dbm *database.DBManager, timezone ...string) *GrabConverter {
	defaultLang := "en"
	if len(timezone) > 0 && timezone[0] != "" {
		defaultLang = language.GetLanguageByTimezone(timezone[0])
	}
	return &GrabConverter{
		dbm:                    dbm,
		menuRepo:               persistence.NewMenuDataRepository(dbm),
		amountConversionFactor: 100,
		defaultLanguage:        defaultLang,
	}
}
```

**特点：**
- `timezone` 参数为可选参数（变长参数）
- 向后兼容：如果不传入 timezone，默认使用英语（`en`）
- 自动查找：根据传入的时区自动查找对应的语言代码

### 3. 调用处修改

#### 3.1 菜单导入（`takeout_menu.go`）

```go
func (s *takeoutSrv) importMenuWithLog(ctx context.Context, reqs request.ImportMenuRequest, reimportLogUuid *uint64) (*resp.GrabMenuImportResp, error) {
	// ...
	
	// 获取店铺时区
	companySetting := ctx.GetCompanySetting()
	timezone := companySetting.GetTimezone()

	// 创建带时区的转换器
	grabConverter := grab.NewGrabConverter(s.dbm, timezone)
	
	// ...
}
```

#### 3.2 菜单导出（`takeout_app_service.go`）

```go
func (s *takeoutAppService) ExportMenu(ctx context.Context, req request.ExportMenuRequest) (interface{}, error) {
	// ...
	
	// 获取店铺时区
	companySetting := ctx.GetCompanySetting()
	timezone := companySetting.GetTimezone()

	// 根据平台类型处理
	switch strings.ToLower(req.Platform) {
	case "grab":
		// 创建带时区的转换器
		grabConverter := grab.NewGrabConverter(s.dbm, timezone)
		platformData, err = grabConverter.LoadMenuFromDatabase(ctx, companyUuid, req.CurrencyUnit, []uint64{})
		// ...
	}
	
	// ...
}
```

---

## 🧪 测试

新增单元测试 `pkg/language/language_test.go`，测试各种时区到语言的映射：

```bash
cd /home/coder/workspaces/ttpos-server-go/main
go test -v ./pkg/language/language_test.go ./pkg/language/language.go
```

**测试覆盖：**
- ✅ 中国各个时区（北京、重庆、香港）
- ✅ 东南亚各国时区（泰国、马来西亚、新加坡、越南、印尼、菲律宾、柬埔寨、缅甸）
- ✅ 印度时区
- ✅ 默认时区（UTC、空字符串）
- ✅ 未知时区

所有测试均通过 ✅

---

## 📝 涉及文件

### 新增文件
- `main/pkg/language/language_test.go` - 时区映射单元测试

### 修改文件
1. `main/pkg/language/language.go`
   - 新增 `GetLanguageByTimezone()` 函数
   - 添加 `strings` 包导入

2. `main/app/modules/takeout/infrastructure/adapter/grab/grab_menu_converter.go`
   - 修改 `NewGrabConverter()` 构造函数，支持 timezone 参数
   - 添加 `language` 包导入

3. `main/app/service/takeout/takeout_menu.go`
   - 在 `importMenuWithLog()` 中获取时区并传入转换器

4. `main/app/modules/takeout/application/takeout_app_service.go`
   - 在 `ExportMenu()` 中获取时区并传入转换器

5. `main/app/modules/takeout/application/takeout_order_service.go`
   - 更新注释说明（构造函数不传入 timezone）

---

## 🔄 向后兼容性

**完全向后兼容！**

- `NewGrabConverter(dbm)` - 不传 timezone，默认使用英语
- `NewGrabConverter(dbm, timezone)` - 传入 timezone，自动设置对应语言

所有现有代码无需修改即可正常运行。

---

## 📊 效果示例

### 之前（硬编码）
所有店铺的菜单默认语言都是英语（`en`）

### 之后（自动设置）
- 泰国曼谷店铺 → 默认语言：泰语（`th`）
- 中国上海店铺 → 默认语言：中文（`zh`）
- 马来西亚吉隆坡店铺 → 默认语言：马来语（`ms`）
- 越南胡志明市店铺 → 默认语言：越南语（`vi`）

---

## 🎯 实现原理

1. **时区数据来源**：`company_setting` 表的 `timezone` 字段
2. **语言映射逻辑**：
   - 优先精确匹配时区字符串（如 `Asia/Bangkok` → `th`）
   - 其次模糊匹配（如 `UTC+08:00` → `en`）
   - 最后回退到默认英语（`en`）

3. **应用场景**：
   - 菜单导入：从 Grab 导入菜单时设置默认语言
   - 菜单导出：从 TTPOS 导出菜单到 Grab 时设置默认语言
   - 多语言翻译：过滤掉默认语言，只保留其他语言的翻译

---

## 🚀 后续优化建议

1. **支持更多时区**：可以根据实际业务需要扩展时区映射表
2. **配置化管理**：将时区映射表放到配置文件或数据库中，便于动态调整
3. **语言回退机制**：如果某个语言不存在翻译，可以回退到默认语言

---

**完成时间**: 2026-01-14  
**开发者**: AI Assistant  
**Story Point**: 3
