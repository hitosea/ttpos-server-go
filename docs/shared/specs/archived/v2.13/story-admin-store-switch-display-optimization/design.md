# 设计文档：新管理端-切换门店-优化店名显示

**关联需求**: [requirements.md](./requirements.md)  
**创建日期**: 2026-01-05  
**设计者**: 曾振华

---

## 1. 设计概览

### 1.1 核心变更
1. 在 `CompanyStaffResp` 结构体中增加 `StoreCode` 字段
2. 在获取门店列表时读取 `storeSetting` 并填充 `StoreCode`
3. 实现基于 `StoreCode` 的自定义排序逻辑

### 1.2 技术方案
- **后端模块**: Go Main (main/app/)
- **影响文件**: 
  - `main/app/dto/resp/saas_staff.go`
  - `main/app/service/auth.go`
- **数据来源**: `ttpos_setting` 表 (key='storeSetting')

---

## 2. 数据结构设计

### 2.1 响应结构体变更

#### 变更1：CompanyStaffResp（门店列表）

**文件**: `main/app/dto/resp/saas_staff.go`

```go
type CompanyStaffResp struct {
	CompanyUuid uint64   `json:"company_uuid"` // 门店UUID
	CompanyName string   `json:"company_name"` // 门店名称
	StoreCode   string   `json:"store_code"`   // 店铺编号（新增）
	Roles       []string `json:"roles"`        // 角色列表
	IsSuper     int      `json:"is_super"`     // 是否超级管理员
}
```

**变更说明**:
- 新增 `StoreCode` 字段，类型为 `string`
- JSON 标签为 `store_code`，遵循 snake_case 命名规范

#### 变更2：Company（Shop基础信息）

**文件**: `main/app/dto/resp/base.go`

```go
type Company struct {
	Uuid                 uint64 `json:"uuid"`                    // 商家UUID
	Name                 string `json:"name"`                    // 商家名称
	Logo                 string `json:"logo"`                    // 商家logo
	TimeZone             string `json:"time_zone"`               // 时区，形如 Asia/Shanghai
	ExpireTime           int64  `json:"expire_time"`             // 店铺到期时间，0表示没有过期时间
	IsOpenMember         int    `json:"is_open_member"`          // 是否开启会员功能: 0不开启, 1开启
	IsOpenBuffet         int    `json:"is_open_buffet"`          // 是否开启自助餐功能: 0不开启, 1开启
	IsOpenH5Order        int    `json:"is_open_h5_order"`        // 是否开启扫码接单功能: 0不开启, 1开启
	IsOpenOldOrder       int    `json:"is_open_old_order"`       // 是否开启旧订单功能: 0不开启, 1开启
	IsOpenRider          bool   `json:"is_open_rider"`           // 是否开启外送
	IsEnableErp          bool   `json:"is_show_inventory"`       // 是否显示移动管理端进销存功能
	IsOpenMap            bool   `json:"is_open_map"`             // 是否开启地图
	IsOpenDataManagement bool   `json:"is_open_data_management"` // 是否开启数据管理功能
	IsOpenKiosk          bool   `json:"is_open_kiosk"`           // 是否开启自助点餐机功能
	IsOpenGrabDelivery   bool   `json:"is_open_grab_delivery"`   // 是否开启Grab外卖功能
	StoreCode            string `json:"store_code"`              // 店铺编码，shop端用于显示（新增）
}
```

**变更说明**:
- 新增 `StoreCode` 字段，类型为 `string`
- JSON 标签为 `store_code`
- 用于 Shop 端（新管理端）显示店铺编号

### 2.2 数据来源

**表**: `ttpos_setting`  
**查询条件**: `key = 'storeSetting'`  
**字段路径**: `values` (JSON) -> `store_code`

**StoreSetting JSON 结构**:
```json
{
  "name": "门店名称",
  "store_code": "No.05",
  "company": "公司名称",
  "address": "地址",
  "phone": "电话",
  ...
}
```

---

## 3. 业务逻辑设计

### 3.1 获取门店列表流程（切换门店场景）

**文件**: `main/app/service/auth.go`  
**方法**: `GetCompanyStaffList`

**流程图**:
```
开始
  ↓
获取员工关联的门店列表 (companyList)
  ↓
遍历每个门店 (cs)
  ↓
├─ 过滤已禁用门店 (IsDisable == 1)
├─ 过滤已过期/异常门店
├─ 获取员工角色列表
├─ 【新增】从 saas 数据库查询 company 信息
├─ 【新增】查询 companySetting
├─ 【新增】设置 context (company, companySetting)
├─ 【新增】获取门店的 storeSetting (含 store_code)
├─ 构建 CompanyStaffResp (含 StoreCode)
  ↓
【新增】对 availableCompanyList 进行排序
  ↓
返回排序后的门店列表
  ↓
结束
```

### 3.2 获取 Shop 基础信息流程（登录后场景）

**文件**: `main/app/service/auth.go`  
**方法**: `ShopBase`

**流程图**:
```
开始
  ↓
获取员工、公司、设置等信息
  ↓
调用 settingSrv.GetStoreSetting(ctx)
  ↓
构建 ShopBase 响应
  ↓
├─ Company 结构体中填充 StoreCode
│   StoreCode: storeSetting.StoreCode
├─ Profile 结构体中也包含 StoreCode
│   StoreCode: storeSetting.StoreCode
  ↓
返回 ShopBase
  ↓
结束
```

**关键代码**:
```go
storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
if err != nil {
    return shopBase, errors.WithMessage(err)
}

return resp.ShopBase{
    // ... 其他字段 ...
    Company: resp.Company{
        // ... 其他字段 ...
        StoreCode: storeSetting.StoreCode, // 新增
    },
    Profile: resp.ShopProfile{
        // ... 其他字段 ...
        StoreCode: storeSetting.StoreCode, // 已有
    },
    // ... 其他字段 ...
}
```

### 3.3 获取 StoreCode 逻辑

#### 场景1：切换门店列表（GetCompanyStaffList）

**实际实现代码**:
```go
// 1. 获取 shopDB
shopDb := s.dbm.GetDB(cs.CompanyUuid)

// 2. 从 saas 数据库查询 company 信息
targetCompany, err := companyRepo.GetCompanyInfoByUuid(cs.CompanyUuid)
if err != nil || targetCompany == nil {
    continue // 查询失败则跳过该门店
}

// 3. 查询 companySetting
companySettingRepo := repository.NewCompanySettingRepo(shopDb)
companySetting, _ := companySettingRepo.GetOne(func(db *gorm.DB) *gorm.DB {
    return db.Where("company_uuid = ?", cs.CompanyUuid)
})
if err != nil {
    continue // 查询失败则跳过该门店
}

// 4. 创建独立的 context 并设置必要信息
var storeCode string
ctxCopy := ctx.Copy()
ctxCopy.SetCompanyUuid(cs.CompanyUuid)
ctxCopy.SetDB(shopDb)
ctxCopy.SetCompany(*targetCompany)
ctxCopy.SetCompanySetting(companySetting)

// 5. 通过 settingSrv 获取 storeSetting
if storeSetting, err := s.settingSrv.GetStoreSetting(ctxCopy); err == nil {
    storeCode = storeSetting.StoreCode
}

// 6. 填充到响应结构体
availableCompanyList = append(availableCompanyList, &resp.CompanyStaffResp{
    CompanyUuid: cs.CompanyUuid,
    CompanyName: company.Name,
    StoreCode:   storeCode, // 新增字段
    Roles:       roleNames,
    IsSuper:     cs.IsSuper,
})
```

#### 场景2：Shop 基础信息（ShopBase）

**实际实现代码**:
```go
// 1. 获取 storeSetting（context 已正确设置）
storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
if err != nil {
    return shopBase, errors.WithMessage(err)
}

// 2. 填充到 Company 结构体
Company: resp.Company{
    Uuid:       company.Uuid,
    Name:       company.Name,
    // ... 其他字段 ...
    StoreCode:  storeSetting.StoreCode, // 新增字段
}

// 3. Profile 中也有 StoreCode（已有字段）
Profile: resp.ShopProfile{
    Address:   storeSetting.Address,
    // ... 其他字段 ...
    StoreCode: storeSetting.StoreCode,
}
```

**关键改进点**:
1. **完整的 context 设置**: 确保 `GetStoreSetting` 能够正确访问 company 和 companySetting 信息
2. **独立 context**: 使用 `ctx.Copy()` 创建独立上下文，避免污染原始 context
3. **错误处理**: 如果 company 或 companySetting 查询失败，跳过该门店
4. **分离关注点**: 通过 settingSrv 统一处理 storeSetting 的获取逻辑
5. **多场景支持**: 同时支持切换门店列表和登录后的基础信息展示

### 3.4 排序逻辑设计

**排序规则**:
1. **StoreCode 为空的排在最前面**
2. **空 StoreCode 之间的排序**：
   - `IsSuper > 0` 的排在前面
   - 如果 `IsSuper` 都大于 0 或都不大于 0，则按 `CompanyName` 字符串排序（不区分大小写）
3. **非空 StoreCode 的排序**：
   - 将 `StoreCode` 转为小写（用于比较）
   - 分为两组：含数字的 / 不含数字的
   - 含数字的组排在不含数字的组前面
   - 每组内部按字符串排序（字典序）

**实现方案**:
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
        return isEmpty1 // true 排在前面
    }

    // 2. 如果都是空字符串，按 IsSuper 和 CompanyName 排序
    if isEmpty1 && isEmpty2 {
        // IsSuper > 0 的排在前面
        if (item1.IsSuper > 0) != (item2.IsSuper > 0) {
            return item1.IsSuper > 0
        }
        // 如果 IsSuper 都大于 0 或都不大于 0，按 CompanyName 排序
        return strings.ToLower(item1.CompanyName) < strings.ToLower(item2.CompanyName)
    }

    // 3. 如果都有 StoreCode，按原规则排序
    code1Lower := strings.ToLower(item1.StoreCode)
    code2Lower := strings.ToLower(item2.StoreCode)

    // 判断是否含数字
    hasDigit1 := containsDigit(code1Lower)
    hasDigit2 := containsDigit(code2Lower)

    // 含数字的排在不含数字的前面
    if hasDigit1 != hasDigit2 {
        return hasDigit1
    }

    // 同组内按字符串排序
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

**排序示例**:
```
输入:
- {StoreCode: "", CompanyName: "总部", IsSuper: 1}
- {StoreCode: "", CompanyName: "管理店", IsSuper: 0}
- {StoreCode: "No.05", CompanyName: "门店5", IsSuper: 0}
- {StoreCode: "Store-A", CompanyName: "门店A", IsSuper: 0}
- {StoreCode: "No.10", CompanyName: "门店10", IsSuper: 0}
- {StoreCode: "", CompanyName: "测试店", IsSuper: 1}
- {StoreCode: "No.01", CompanyName: "门店1", IsSuper: 0}

处理:
1. 分组：
   - 空 StoreCode: ["总部"(IsSuper=1), "管理店"(IsSuper=0), "测试店"(IsSuper=1)]
   - 含数字 StoreCode: ["No.05", "No.10", "No.01"]
   - 不含数字 StoreCode: ["Store-A"]

2. 空 StoreCode 组内排序：
   - 先按 IsSuper: ["总部"(1), "测试店"(1), "管理店"(0)]
   - IsSuper 相同的按 CompanyName: ["测试店", "总部", "管理店"]

3. 含数字 StoreCode 组内排序：
   - 转小写: ["no.05", "no.10", "no.01"]
   - 字符串排序: ["no.01", "no.05", "no.10"]

4. 合并: ["测试店", "总部", "管理店", "No.01", "No.05", "No.10", "Store-A"]

输出:
1. {StoreCode: "", CompanyName: "测试店", IsSuper: 1}
2. {StoreCode: "", CompanyName: "总部", IsSuper: 1}
3. {StoreCode: "", CompanyName: "管理店", IsSuper: 0}
4. {StoreCode: "No.01", CompanyName: "门店1", IsSuper: 0}
5. {StoreCode: "No.05", CompanyName: "门店5", IsSuper: 0}
6. {StoreCode: "No.10", CompanyName: "门店10", IsSuper: 0}
7. {StoreCode: "Store-A", CompanyName: "门店A", IsSuper: 0}
```

---

## 4. 接口设计

### 4.1 响应结构变更

**接口**: `GET /api/v1/saas/company_staff_list` (推测)

**响应示例**:
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "company_uuid": 123456,
      "company_name": "华莱士xxx店",
      "store_code": "No.01",
      "roles": ["收银员", "店长"],
      "is_super": 0
    },
    {
      "company_uuid": 123457,
      "company_name": "华莱士yyy店",
      "store_code": "No.05",
      "roles": ["店长"],
      "is_super": 1
    }
  ]
}
```

**变更点**:
- 新增 `store_code` 字段
- 列表已按 `store_code` 排序

---

## 5. 错误处理

### 5.1 异常场景

| 场景 | 处理方式 |
|------|----------|
| company 信息查询失败 | 跳过该门店，不添加到列表 |
| companySetting 查询失败 | 跳过该门店，不添加到列表 |
| storeSetting 获取失败 | `StoreCode` 设为空字符串 `""`，排在最前面 |
| storeSetting JSON 解析失败 | `StoreCode` 设为空字符串 `""`，排在最前面 |
| store_code 字段不存在 | `StoreCode` 设为空字符串 `""`，排在最前面 |
| StoreCode 为空 | 正常处理，排在最前面，按 IsSuper 和 CompanyName 排序 |

### 5.2 日志记录

```go
// 查询 company 失败
if err != nil || targetCompany == nil {
    log.Warnf("获取门店信息失败: companyUuid=%d, err=%v", cs.CompanyUuid, err)
    continue
}

// 查询 companySetting 失败
if err != nil {
    log.Warnf("获取门店设置失败: companyUuid=%d, err=%v", cs.CompanyUuid, err)
    continue
}

// 获取 storeSetting 失败（静默处理，StoreCode 为空）
if storeSetting, err := s.settingSrv.GetStoreSetting(ctxCopy); err == nil {
    storeCode = storeSetting.StoreCode
} // err 时 storeCode 保持空字符串
```

---

## 6. 性能优化

### 6.1 批量查询优化

**问题**: 在循环中逐个查询 `storeSetting` 可能导致性能问题（N+1查询）

**优化方案（可选）**:
- 方案1: 保持现状，单次查询通常在可接受范围内
- 方案2: 后续优化时考虑批量查询或缓存

### 6.2 排序性能

- 时间复杂度: O(n log n)
- 空间复杂度: O(1)
- 门店数量通常 < 100，性能影响可忽略

---

## 7. 测试设计

### 7.1 单元测试

**测试文件**: `main/app/service/auth_test.go`

**测试用例**:
1. `TestSortByStoreCode_WithDigits`: 测试含数字的排序
2. `TestSortByStoreCode_WithoutDigits`: 测试不含数字的排序
3. `TestSortByStoreCode_Mixed`: 测试混合场景
4. `TestSortByStoreCode_Empty`: 测试空字符串处理
5. `TestSortByStoreCode_CaseInsensitive`: 测试大小写不敏感

### 7.2 集成测试

**测试场景**:
1. 获取门店列表，验证 `StoreCode` 字段存在
2. 验证门店列表按 `StoreCode` 排序
3. 验证 `StoreCode` 为空时的处理

### 7.3 边界测试

**测试数据**:
```go
testCases := []struct {
    input    []string
    expected []string
}{
    {
        input:    []string{"No.05", "Store-A", "No.10", "Store-B", "No.01"},
        expected: []string{"No.01", "No.05", "No.10", "Store-A", "Store-B"},
    },
    {
        input:    []string{"", "No.01", "Store-A"},
        expected: []string{"No.01", "", "Store-A"},
    },
    {
        input:    []string{"ABC", "123", "A1B"},
        expected: []string{"123", "A1B", "ABC"},
    },
}
```

---

## 8. 前端对接说明

### 8.1 字段使用

**展示格式**: `{store_code}{company_name}`

**示例代码** (Vue 3):
```vue
<template>
  <div v-for="company in companyList" :key="company.company_uuid">
    {{ company.store_code }}{{ company.company_name }}
  </div>
</template>

<script setup lang="ts">
interface CompanyStaff {
  company_uuid: number
  company_name: string
  store_code: string  // 新增字段
  roles: string[]
  is_super: number
}
</script>
```

### 8.2 兼容性处理

**旧版本兼容**:
```typescript
// 如果 store_code 不存在或为空，只显示 company_name
const displayName = computed(() => {
  return company.store_code 
    ? `${company.store_code}${company.company_name}` 
    : company.company_name
})
```

---

## 9. 上线计划

### 9.1 上线步骤

1. 后端部署（本次变更）
2. 前端适配（各终端）
3. 灰度验证
4. 全量发布

### 9.2 回滚方案

**影响评估**: 低风险
- 新增字段，不影响现有功能
- 前端未适配时，忽略 `store_code` 字段即可

**回滚步骤**:
1. 回滚后端代码
2. 清理相关日志

---

## 10. 相关文档

- [Go Main 开发规范](/.cursor/rules/go-main.mdc)
- [API 设计规范](/.cursor/rules/api.mdc)
- [数据库规范](/.cursor/rules/database.mdc)

---

**最后更新**: 2026-01-05  
**审核状态**: 待审核

