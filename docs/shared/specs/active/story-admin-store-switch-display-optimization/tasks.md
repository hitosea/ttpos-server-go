# 任务清单：新管理端-切换门店-优化店名显示

**关联设计**: [design.md](./design.md)  
**创建日期**: 2026-01-05  
**负责人**: 曾振华

---

## 任务分解

### Task 1: 修改响应结构体
- **文件**: 
  - `main/app/dto/resp/saas_staff.go`
  - `main/app/dto/resp/base.go`
- **操作**: 在相关结构体中新增 `StoreCode` 字段
- **预计时间**: 10分钟
- **状态**: ✅ 已完成

**具体变更**:

1. **CompanyStaffResp**（门店切换列表）:
```go
type CompanyStaffResp struct {
	CompanyUuid uint64   `json:"company_uuid"` // 门店UUID
	CompanyName string   `json:"company_name"` // 门店名称
	StoreCode   string   `json:"store_code"`   // 店铺编号（新增）
	Roles       []string `json:"roles"`        // 角色列表
	IsSuper     int      `json:"is_super"`     // 是否超级管理员
}
```

2. **Company**（Shop基础信息）:
```go
type Company struct {
	// ... 其他字段 ...
	StoreCode string `json:"store_code"` // 店铺编码，shop端用于显示（新增）
}
```

---

### Task 2: 实现获取 StoreCode 逻辑
- **文件**: `main/app/service/auth.go`
- **操作**: 在相关方法中读取 `storeSetting` 并填充 `StoreCode`
- **预计时间**: 40分钟
- **状态**: ✅ 已完成

#### 2.1 切换门店列表（GetCompanyStaffList）

**实现位置**: 
- 在 `availableCompanyList` 构建逻辑中（约1794行附近）
- 在获取角色列表后，构建响应结构体之前

**实际实现**:
```go
// 1. 从 saas 数据库查询 company 信息
targetCompany, err := companyRepo.GetCompanyInfoByUuid(cs.CompanyUuid)
if err != nil || targetCompany == nil {
    continue // 查询失败则跳过该门店
}

// 2. 查询 companySetting
companySettingRepo := repository.NewCompanySettingRepo(shopDb)
companySetting, _ := companySettingRepo.GetOne(func(db *gorm.DB) *gorm.DB {
    return db.Where("company_uuid = ?", cs.CompanyUuid)
})
if err != nil {
    continue // 查询失败则跳过该门店
}

// 3. 创建独立的 context 并设置必要信息
var storeCode string
ctxCopy := ctx.Copy()
ctxCopy.SetCompanyUuid(cs.CompanyUuid)
ctxCopy.SetDB(shopDb)
ctxCopy.SetCompany(*targetCompany)
ctxCopy.SetCompanySetting(companySetting)

// 4. 通过 settingSrv 获取 storeSetting
if storeSetting, err := s.settingSrv.GetStoreSetting(ctxCopy); err == nil {
    storeCode = storeSetting.StoreCode
}

// 5. 填充到响应结构体
availableCompanyList = append(availableCompanyList, &resp.CompanyStaffResp{
    CompanyUuid: cs.CompanyUuid,
    CompanyName: company.Name,
    StoreCode:   storeCode, // 新增
    Roles:       roleNames,
    IsSuper:     cs.IsSuper,
})
```

#### 2.2 Shop 基础信息（ShopBase）

**实现位置**: 
- 在 `ShopBase` 方法中（约1460行附近）
- 在构建 `resp.ShopBase` 响应时

**实际实现**:
```go
// 1. 获取 storeSetting
storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
if err != nil {
    return shopBase, errors.WithMessage(err)
}

// 2. 在 Company 结构体中填充 StoreCode
Company: resp.Company{
    Uuid:       company.Uuid,
    Name:       company.Name,
    // ... 其他字段 ...
    StoreCode:  storeSetting.StoreCode, // 新增
}
```

**关键改进**:
1. 完整查询 company 和 companySetting 信息
2. 正确设置 context，确保 `GetStoreSetting` 能够访问必要数据
3. 使用 `ctx.Copy()` 创建独立上下文，避免污染
4. 查询失败时跳过门店，而非返回空 StoreCode
5. 同时支持切换门店和登录基础信息两个场景

---

### Task 3: 实现排序逻辑
- **文件**: `main/app/service/auth.go`
- **操作**: 在 `availableCompanyList` 构建完成后，添加自定义排序逻辑
- **预计时间**: 30分钟
- **状态**: ✅ 已完成

**实现位置**: 
- 在所有门店添加到 `availableCompanyList` 之后
- 在返回结果之前

**排序规则**:
1. **StoreCode 为空的排在最前面**
2. **空 StoreCode 之间**：`IsSuper > 0` 的排前面，都 > 0 则按 `CompanyName` 排序
3. **非空 StoreCode**：含数字的排前面，同组内按字符串排序（不区分大小写）

**实际实现**:
```go
// 对 availableCompanyList 按 StoreCode 排序
// 排序规则：
// 1. StoreCode 为空的排在最前面
// 2. 空 StoreCode 之间：IsSuper > 0 的排前面，都 > 0 则按 CompanyName 排序
// 3. 非空 StoreCode：含数字的排前面，同组内按字符串排序（不区分大小写）
sort.Slice(availableCompanyList, func(i, j int) bool {
    item1 := availableCompanyList[i]
    item2 := availableCompanyList[j]
    
    isEmpty1 := item1.StoreCode == ""
    isEmpty2 := item2.StoreCode == ""

    // 1. StoreCode 为空的排在最前面
    if isEmpty1 != isEmpty2 {
        return isEmpty1
    }

    // 2. 如果都是空字符串，按 IsSuper 和 CompanyName 排序
    if isEmpty1 && isEmpty2 {
        if (item1.IsSuper > 0) != (item2.IsSuper > 0) {
            return item1.IsSuper > 0
        }
        return strings.ToLower(item1.CompanyName) < strings.ToLower(item2.CompanyName)
    }

    // 3. 如果都有 StoreCode，按原规则排序
    code1Lower := strings.ToLower(item1.StoreCode)
    code2Lower := strings.ToLower(item2.StoreCode)
    
    hasDigit1 := containsDigit(code1Lower)
    hasDigit2 := containsDigit(code2Lower)
    
    if hasDigit1 != hasDigit2 {
        return hasDigit1
    }
    
    return code1Lower < code2Lower
})

// containsDigit 检查字符串是否包含数字
func containsDigit(s string) bool {
    for _, c := range s {
        if c >= '0' && c <= '9' {
            return true
        }
    }
    return false
}
```

**注意事项**:
- 辅助函数 `containsDigit` 定义为包级私有函数
- 排序保持原始大小写，只在比较时转小写
- 添加了必要的 import: `sort` 和 `strings`
- 空 StoreCode 优先显示，便于识别未设置店铺编号的门店

---

### Task 4: 添加单元测试
- **文件**: `main/app/service/auth_test.go` (如不存在则创建)
- **操作**: 编写排序逻辑的单元测试
- **预计时间**: 30分钟
- **状态**: ⏳ 待开始（可选）

**测试用例**:
```go
func TestSortCompanyListByStoreCode(t *testing.T) {
	testCases := []struct {
		name     string
		input    []*resp.CompanyStaffResp
		expected []string // 期望的 StoreCode 顺序
	}{
		{
			name: "混合含数字和不含数字",
			input: []*resp.CompanyStaffResp{
				{StoreCode: "No.05", CompanyName: "门店5"},
				{StoreCode: "Store-A", CompanyName: "门店A"},
				{StoreCode: "No.10", CompanyName: "门店10"},
				{StoreCode: "Store-B", CompanyName: "门店B"},
				{StoreCode: "No.01", CompanyName: "门店1"},
			},
			expected: []string{"No.01", "No.05", "No.10", "Store-A", "Store-B"},
		},
		{
			name: "含空字符串",
			input: []*resp.CompanyStaffResp{
				{StoreCode: "", CompanyName: "无编号门店"},
				{StoreCode: "No.01", CompanyName: "门店1"},
				{StoreCode: "Store-A", CompanyName: "门店A"},
			},
			expected: []string{"No.01", "", "Store-A"},
		},
		{
			name: "大小写混合",
			input: []*resp.CompanyStaffResp{
				{StoreCode: "STORE-A", CompanyName: "门店A"},
				{StoreCode: "store-b", CompanyName: "门店B"},
				{StoreCode: "Store-C", CompanyName: "门店C"},
			},
			expected: []string{"STORE-A", "store-b", "Store-C"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行排序
			sort.Slice(tc.input, func(i, j int) bool {
				// ... 排序逻辑 ...
			})

			// 验证结果
			for i, item := range tc.input {
				if item.StoreCode != tc.expected[i] {
					t.Errorf("排序错误: 位置%d 期望%s, 实际%s", i, tc.expected[i], item.StoreCode)
				}
			}
		})
	}
}
```

---

### Task 5: 集成测试
- **文件**: Postman/接口测试工具
- **操作**: 验证接口返回结果
- **预计时间**: 15分钟
- **状态**: ⏳ 待开始

**测试步骤**:
1. 准备测试数据：创建多个门店，设置不同的 `store_code`
2. 调用获取门店列表接口
3. 验证响应中包含 `store_code` 字段
4. 验证门店列表按 `store_code` 排序
5. 验证边界场景（空 `store_code`）

**测试数据示例**:
```sql
-- 门店1: store_code = "No.05"
-- 门店2: store_code = "Store-A"
-- 门店3: store_code = "No.01"
-- 门店4: store_code = ""
```

**期望结果**:
```json
[
  {"store_code": "No.01", "company_name": "门店3"},
  {"store_code": "No.05", "company_name": "门店1"},
  {"store_code": "", "company_name": "门店4"},
  {"store_code": "Store-A", "company_name": "门店2"}
]
```

---

### Task 6: 代码审查与优化
- **文件**: 所有变更文件
- **操作**: 代码自审和优化
- **预计时间**: 15分钟
- **状态**: ⏳ 待开始

**检查清单**:
- [ ] 代码符合 Go Main 开发规范
- [ ] 添加了必要的中文注释
- [ ] 错误处理完善
- [ ] 性能影响可接受
- [ ] 无 linter 错误
- [ ] 无安全风险

---

### Task 7: 前端对接（可选）
- **负责人**: 前端开发
- **操作**: 适配新的 `store_code` 字段
- **预计时间**: 30分钟/终端
- **状态**: ⏳ 待前端开始

**涉及终端**:
- [ ] 新管理端
- [ ] 收银端
- [ ] 点餐助手
- [ ] 平板端
- [ ] 厨显端

**前端变更**:
1. 更新接口类型定义，添加 `store_code` 字段
2. 修改展示逻辑，拼接 `store_code` 和 `company_name`
3. 处理 `store_code` 为空的场景

---

## 验收标准

### 后端验收
- ✅ `CompanyStaffResp` 包含 `StoreCode` 字段
- ✅ `Company` 结构体包含 `StoreCode` 字段
- ✅ 查询 company 和 companySetting 信息
- ✅ 正确设置 context 用于获取 storeSetting
- ✅ 接口返回的门店列表已排序（空 StoreCode 优先）
- ✅ 排序规则符合设计文档
- ✅ Shop 基础信息接口返回 StoreCode
- ✅ 添加了 `gorm.io/gorm` 导入
- ✅ 无 linter 错误
- ⏳ 单元测试通过（可选）
- ⏳ 集成测试通过（待测试）

### 前端验收（各终端）
- [ ] 切换门店页面展示"店铺编号+商家名称"
- [ ] 切换后店名显示"店铺编号+商家名称"
- [ ] 门店列表按店铺编号排序
- [ ] 店铺编号为空时只显示商家名称

---

## 风险与依赖

### 风险
- **低风险**: 新增字段，不影响现有功能
- **兼容性**: 前端未适配时，忽略 `store_code` 字段

### 依赖
- **上游**: 无
- **下游**: 前端各终端需适配

---

## 总预计时间
**后端**: 2小时  
**前端**: 2.5小时（5个终端 × 30分钟）  
**总计**: 4.5小时

---

**最后更新**: 2026-01-05  
**状态**: 待开始

