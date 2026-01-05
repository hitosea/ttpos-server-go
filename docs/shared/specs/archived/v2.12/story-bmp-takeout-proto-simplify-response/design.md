# 简化菜单更新接口响应结构 设计文档

> 本文档定义 简化菜单更新接口响应结构 的技术设计和实现方案。

## 📋 概述

本次需求是对 `menu.proto` 中的响应结构进行简化重构，移除 `UpdateMenuItemResp` 和 `UpdateMenuModifierResp` 中的冗余错误字段（`error_code` 和 `error_message`），统一使用 `takeout.ApiResponse` 的 `code` 和 `message` 进行错误处理。这是一个低风险的技术重构，不涉及业务逻辑变更。

---

## 🎯 规范对齐

### Go BMP 规范 (go-bmp.mdc)

- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务注册到 Nacos
- 遵循 GoFrame 项目结构
- Proto 代码生成使用 `gf gen pb`

### API 设计规范 (api.mdc)

- 统一使用 `takeout.ApiResponse` 进行错误处理
- 响应格式：`{code, message, data{}}`
- 避免数据冗余

---

## 🔄 代码复用分析

### 可复用的现有组件

- **takeout.ApiResponse**: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/takeout_api.proto` - 统一的 API 响应结构，已包含 `code` 和 `message` 字段
- **Proto 生成工具**: GoFrame 的 `gf gen pb` 命令

### 集成点

- **MenuService RPC**: 已使用 `takeout.ApiResponse` 作为返回类型
- **DTO 转换**: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu_update.go` 中的 `UpdateMenuResult` 结构（如需要调整）

---

## 🏗️ 架构设计

### 当前架构

```
Proto 定义 (menu.proto)
  ↓
生成 Go 代码 (gf gen pb)
  ↓
DTO 转换 (menu_update.go)
  ↓
Logic 层 (grab_menu.go)
  ↓
gRPC Controller
```

### 修改点

1. **Proto 文件**: 移除冗余字段
2. **生成的 Go 代码**: 自动更新（通过 `gf gen pb`）
3. **DTO 结构**: 检查并更新（如需要）

---

## 📊 数据模型

### Proto 定义修改

#### 修改前

```protobuf
message UpdateMenuItemResp {
  bool success = 1;
  string merchant_id = 2;
  string record_id = 3;
  string record_type = 4;
  string error_code = 5;     // ← 移除
  string error_message = 6;  // ← 移除
}

message UpdateMenuModifierResp {
  bool success = 1;
  string merchant_id = 2;
  string record_id = 3;
  string record_type = 4;
  string error_code = 5;     // ← 移除
  string error_message = 6;  // ← 移除
}
```

#### 修改后

```protobuf
message UpdateMenuItemResp {
  bool success = 1;
  string merchant_id = 2;
  string record_id = 3;
  string record_type = 4;
  // error_code 和 error_message 已移除，由 ApiResponse 统一处理
}

message UpdateMenuModifierResp {
  bool success = 1;
  string merchant_id = 2;
  string record_id = 3;
  string record_type = 4;
  // error_code 和 error_message 已移除，由 ApiResponse 统一处理
}
```

### DTO 结构（如需要调整）

当前 `UpdateMenuResult` 结构：

```go
type UpdateMenuResult struct {
    Success      bool   `json:"success"`
    MerchantID   string `json:"merchant_id"`
    RecordID     string `json:"record_id"`
    RecordType   string `json:"record_type"`
    ErrorCode    string `json:"error_code,omitempty"`    // ← 检查是否需要移除
    ErrorMessage string `json:"error_message,omitempty"` // ← 检查是否需要移除
}
```

**注意**: DTO 中的 `ErrorCode` 和 `ErrorMessage` 字段用于内部逻辑处理，与 proto 响应结构不同。需要确认是否有代码依赖这些字段。

---

## 🔌 API 设计

### gRPC API（不变）

RPC 方法定义保持不变：

```protobuf
service MenuService {
    rpc UpdateMenuItem (UpdateMenuItemReq) returns (takeout.ApiResponse) {}
    rpc UpdateMenuModifier (UpdateMenuModifierReq) returns (takeout.ApiResponse) {}
}
```

### 响应结构

**成功响应**:
```json
{
  "code": 1,
  "message": "success",
  "data": {
    "success": true,
    "merchant_id": "M-12345",
    "record_id": "ITEM-001",
    "record_type": "ITEM"
  }
}
```

**错误响应**:
```json
{
  "code": 0,
  "message": "调用 Grab UpdateMenuItem API 失败: 具体错误信息",
  "data": {}
}
```

**关键点**: 错误信息现在统一通过 `ApiResponse.message` 传递，不再在 `data` 中的 `error_code` 和 `error_message` 字段重复。

---

## 🧩 组件和接口

### Logic 层（无需修改）

当前 `UpdateMenuItem` 和 `UpdateMenuModifier` 方法已经返回 `UpdateMenuResult`，其中包含 `ErrorCode` 和 `ErrorMessage`。这些字段用于内部逻辑处理，与 proto 响应结构是分离的。

**检查点**: 确认是否有代码依赖 `UpdateMenuResult.ErrorCode` 或 `UpdateMenuResult.ErrorMessage`。

---

## 🚨 错误处理

### 错误场景

#### 场景 1: API 调用失败

- **处理方式**: 错误信息通过 `ApiResponse.message` 返回
- **用户影响**: 客户端从 `ApiResponse.message` 获取错误信息
- **代码示例**:
  ```go
  // Logic 层返回错误
  return &grabDto.UpdateMenuResult{
      Success:      false,
      ErrorCode:    "API_ERROR",  // 内部使用
      ErrorMessage: err.Error(),  // 内部使用
  }, gerror.Wrap(err, "调用 Grab UpdateMenuItem API 失败")
  
  // Controller 层转换为 ApiResponse
  // ApiResponse.message = err.Error()
  ```

#### 场景 2: 参数验证失败

- **处理方式**: 验证错误通过 `ApiResponse.message` 返回
- **用户影响**: 客户端从 `ApiResponse.message` 获取验证错误信息

---

## 🔒 安全设计

- **不涉及安全变更**: 本次重构不涉及安全相关修改

---

## 🧪 测试策略

### 单元测试

**测试内容**:
- Proto 文件语法正确性
- 代码生成成功
- 编译通过

### 集成测试

**测试内容**:
- gRPC 接口调用正常
- 错误信息正确传递到 `ApiResponse.message`
- 响应格式符合预期

---

## 📈 性能优化

- **不涉及性能优化**: 本次重构不涉及性能相关修改

---

## 📚 实现清单

### Phase 1: Proto 文件修改

- [ ] 修改 `menu.proto`，移除 `UpdateMenuItemResp.error_code` 和 `UpdateMenuItemResp.error_message`
- [ ] 修改 `menu.proto`，移除 `UpdateMenuModifierResp.error_code` 和 `UpdateMenuModifierResp.error_message`

### Phase 2: 代码生成和验证

- [ ] 执行 `gf gen pb` 重新生成 proto 代码
- [ ] 验证生成的 Go 代码不包含已移除的字段
- [ ] 编译检查，确保无编译错误

### Phase 3: 代码检查和更新

- [ ] 检查是否有代码依赖已移除的字段
- [ ] 如有依赖，更新为使用 `ApiResponse.code` 和 `ApiResponse.message`
- [ ] 更新相关测试（如有）

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-16  
**作者**: rikugun  
**审核者**: {审核者}

