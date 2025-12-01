# Bug-251127-005 修复方案

## 问题概述

物料详情接口 `/api/v1/shop/material/detail` 返回的数据中缺失 `allow_substore_visible` 字段，导致前端无法正确判断物料的门店可见性配置。

**影响范围**：
- **终端**：shop（店铺后台）
- **模块**：material（物料管理）
- **严重程度**：medium

## 根本原因

经过代码分析，确认了问题的根本原因：

### 1. 数据层完整 ✅
```go
// main/app/model/material.go:31
AllowSubstoreVisible int `gorm:"type:tinyint(1);default:1;column:allow_substore_visible;comment:'允许子店可见：1-允许，0-不允许'"`
```
数据库 model 中已定义该字段。

### 2. 请求 DTO 完整 ✅
```go
// main/app/dto/req/material.go:115
AllowSubstoreVisible int `json:"allow_substore_visible"` // 允许子店可见：1-允许，0-不允许
```
请求 DTO 中已包含该字段（用于创建/编辑物料）。

### 3. 响应 DTO 缺失 ❌
```go
// main/app/dto/resp/material_resp/material.go:54-79
type MaterialDetailResp struct {
    Uuid         uint64             `json:"uuid"`
    LocaleName   dto.LocaleResponse `json:"locale_name"`
    // ... 其他字段
    // 缺少 AllowSubstoreVisible 字段！
}
```
**响应结构体中缺少字段定义**。

### 4. Service 层未填充 ❌
```go
// main/app/service/material.go:428-455
return material_resp.MaterialDetailResp{
    Uuid:         material.Uuid,
    LocaleName:   material.MultiLanguageName.GetNames(),
    // ... 其他字段
    // 未填充 AllowSubstoreVisible 字段！
}
```
**Service 层构建响应时未包含该字段**。

### 结论

这是一个典型的**字段遗漏问题**：
- 数据库层和 Model 层都正常
- 但在构建响应 DTO 时遗漏了该字段

## 修复方案

### 方案选择

**方案 1: 在响应 DTO 中添加字段（推荐）**
- **优点**：
  - 简单直接，修改范围小
  - 符合现有代码结构
  - 风险低，不影响其他逻辑
- **缺点**：无
- **风险**：极低

**方案 2: 重构 DTO 映射逻辑**
- **优点**：可以一次性检查所有字段
- **缺点**：工作量大，可能引入新问题
- **风险**：中等

**✅ 最终选择：方案 1**

**理由**：
1. 问题明确，只需要添加单个字段
2. 修改范围小，测试成本低
3. 不影响现有功能
4. 符合敏捷开发原则

### 实施步骤

#### Step 1: 修改响应 DTO
在 `MaterialDetailResp` 结构体中添加 `AllowSubstoreVisible` 字段。

#### Step 2: 修改 Service 层
在 `GetMaterialDetail` 方法中填充该字段。

#### Step 3: 测试验证
- 编写单元测试
- 手动验证接口返回

#### Step 4: 更新文档
- Swagger 文档会自动更新（基于结构体 tag）

### 技术方案

#### 1. 修改响应 DTO

**文件**：`main/app/dto/resp/material_resp/material.go`

**位置**：`MaterialDetailResp` 结构体（第 54-79 行）

**修改**：在合适位置添加字段

```go
type MaterialDetailResp struct {
	Uuid                   uint64               `json:"uuid"`
	LocaleName             dto.LocaleResponse   `json:"locale_name"`
	Code                   string               `json:"code"`
	CategoryUuid           uint64               `json:"category_uuid"`
	CategoryName           string               `json:"category_name"`
	Status                 int                  `json:"status"`
	AllowSubstoreVisible   int                  `json:"allow_substore_visible"`  // 🆕 添加此字段
	Valuation              float64              `json:"valuation"`
	BarcodeValue           string               `json:"barcode_value"`
	// ... 其他字段保持不变
}
```

**字段说明**：
- **字段名**：`AllowSubstoreVisible`
- **JSON 标签**：`allow_substore_visible`
- **类型**：`int`
- **值域**：`1`=允许，`0`=不允许
- **注释**：`允许子店可见：1-允许，0-不允许（仅总店可用）`

#### 2. 修改 Service 层

**文件**：`main/app/service/material.go`

**位置**：`GetMaterialDetail` 方法（第 428-455 行）

**修改**：在返回的结构体中添加字段赋值

```go
return material_resp.MaterialDetailResp{
	Uuid:         material.Uuid,
	LocaleName:   material.MultiLanguageName.GetNames(),
	Code:         material.Code,
	CategoryUuid: material.CategoryUuid,
	CategoryName: material.Category.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
	Status:       int(utils.BoolToUint(material.Status)),
	AllowSubstoreVisible: material.AllowSubstoreVisible,  // 🆕 添加此行
	Valuation:    material.Valuation,
	BarcodeValue: material.BarcodeValue,
	// ... 其他字段保持不变
}
```

**数据来源**：
- 从 `material.AllowSubstoreVisible` 直接获取
- Model 中已有该字段，数据库查询时会自动填充

#### 3. 代码位置参考

```yaml
涉及文件:
  - main/app/dto/resp/material_resp/material.go:54-79
    作用: 添加响应字段定义
    修改行: ~60 (在 Status 之后添加)
    
  - main/app/service/material.go:428-455
    作用: 填充响应字段值
    修改行: ~436 (在 Status 之后添加)
```

## 影响分析

### 兼容性

✅ **完全向下兼容**

- **前端**：只是新增字段，不影响现有字段
- **API 签名**：接口路径、参数都不变
- **数据库**：不需要迁移（字段已存在）

### 性能影响

✅ **无性能影响**

- 不增加数据库查询
- 不增加计算逻辑
- 只是多返回一个字段

### 安全风险

⚠️ **低风险**

- `allow_substore_visible` 是业务配置字段
- 不涉及敏感信息
- 已在其他接口中使用（列表、更新等）

## 测试计划

### 单元测试

**文件**：`main/app/service/material_test.go`（需创建或更新）

**测试用例**：

```go
func TestGetMaterialDetail_AllowSubstoreVisible(t *testing.T) {
    // 测试场景 1：允许门店可见的物料
    material := &model.Material{
        Uuid: 123456,
        AllowSubstoreVisible: 1,
        // ... 其他字段
    }
    // 验证返回值中包含 AllowSubstoreVisible = 1
    
    // 测试场景 2：不允许门店可见的物料
    material := &model.Material{
        Uuid: 789012,
        AllowSubstoreVisible: 0,
        // ... 其他字段
    }
    // 验证返回值中包含 AllowSubstoreVisible = 0
}
```

### 集成测试

**场景 1：正常查询**
```bash
GET /api/v1/shop/material/detail?uuid=3699861597323265
Authorization: Bearer {valid_token}

预期返回:
{
  "code": 0,
  "message": "success",
  "data": {
    "uuid": 3699861597323265,
    "name": "...",
    "allow_substore_visible": 1,  // ✅ 应包含此字段
    "..."
  }
}
```

**场景 2：验证不同值**
- 测试 `allow_substore_visible = 1` 的物料
- 测试 `allow_substore_visible = 0` 的物料

### 手动测试

**测试环境**：ttpos-test1.ttpos.com

**测试步骤**：

1. **准备数据**
   - 确认数据库中存在测试物料
   - 验证物料的 `allow_substore_visible` 字段值

2. **接口测试**
   ```bash
   curl -X GET \
     'https://ttpos-test1.ttpos.com/api/v1/shop/material/detail?uuid=3699861597323265' \
     -H 'Authorization: Bearer {token}'
   ```

3. **验证结果**
   - ✅ HTTP 状态码 200
   - ✅ `code` 字段为 0
   - ✅ `data.allow_substore_visible` 字段存在
   - ✅ 字段值与数据库一致

4. **前端联调**（如果涉及）
   - 确认前端能正确读取该字段
   - 验证门店可见性逻辑是否正常

## 上线计划

### 发布时间

**建议**：下一个常规版本发布

**理由**：
- 非紧急 Bug（严重程度 medium）
- 不影响核心功能
- 可与其他功能一起发布

### 回滚方案

✅ **低风险，无需特殊回滚**

- 只是新增字段，不影响现有逻辑
- 如果有问题，可以直接发布修复版本

### 监控指标

发布后关注：

1. **接口响应时间**
   - 监控：`/api/v1/shop/material/detail` 的响应时间
   - 预期：无明显变化

2. **错误率**
   - 监控：该接口的错误率
   - 预期：无增加

3. **业务指标**
   - 监控：物料管理相关操作
   - 预期：正常

## 预防措施

### 如何避免类似问题

#### 1. 代码规范强化

**问题**：响应 DTO 与 Model 不一致

**方案**：
- 在 Code Review 时检查 DTO 完整性
- 创建检查清单：新增字段时需同步更新 DTO

#### 2. 自动化检测

**建议**：
- 编写 lint 规则：检测 Model 与 DTO 字段一致性
- 在 CI/CD 中集成检查

#### 3. 测试覆盖

**问题**：缺少字段未被测试发现

**方案**：
- 增加 API 响应字段完整性测试
- 在集成测试中验证所有字段

#### 4. 文档同步

**建议**：
- 更新 API 文档模板，明确列出所有字段
- 在需求评审时检查字段完整性

### 相关规范参考

- **API 设计规范**：`.cursor/rules/api.mdc`
- **Go 开发规范**：`.cursor/rules/go-main.mdc`
- **数据库规范**：`.cursor/rules/database.mdc`

---

**创建时间**：2025-11-27  
**创建者**：weifashi  
**审核状态**：待审核

