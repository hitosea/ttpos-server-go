# gRPC 菜单更新服务 任务分解

> 本文档定义 gRPC 菜单项/修饰符更新服务的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 9  
**已完成**: 7  
**进行中**: -  
**完成率**: 77.8%

---

## Phase 0: 常量定义（前置准备）

- [x] 0.1 创建响应码常量文件

  - File: `ttpos-bmp/app/ttpos-takeout/internal/consts/response.go`
  - Purpose: 定义统一的响应码和消息常量，避免硬编码
  - Requirements: 3.3
  - Leverage: 
    - 现有常量: `ttpos-bmp/app/ttpos-takeout/internal/consts/consts.go`
  - 新增内容:
    ```go
    package consts

    // ResponseCode gRPC 响应码常量
    type ResponseCode string

    const (
        CodeSuccess        ResponseCode = "0"     // 成功
        CodeInvalidParam   ResponseCode = "4001"  // 参数校验失败
        CodeServiceError   ResponseCode = "5001"  // 服务内部错误
        CodeSerializeError ResponseCode = "5002"  // 序列化错误
        CodeExternalAPIError ResponseCode = "5003" // 外部 API 错误
    )

    // ResponseMessage 响应消息常量
    const (
        MsgSuccess           = "success"
        MsgSerializeFailed   = "序列化响应数据失败"
        MsgMerchantIDEmpty   = "merchant_id 不能为空"
        MsgItemIDEmpty       = "item_id 不能为空"
        MsgModifierIDEmpty   = "modifier_id 不能为空"
        MsgModifierNameEmpty = "modifier_name 不能为空"
    )
    ```
  - Success: 常量文件创建成功，可被其他包引用

---

## Phase 1: Proto 定义（0.25 天）

- [x] 1.1 更新 menu.proto 添加新消息类型

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
  - Purpose: 定义 UpdateMenuItem 和 UpdateMenuModifier 的请求/响应消息
  - Requirements: 1.1, 1.2, 2.1, 2.2
  - Leverage: 
    - 现有 Proto: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
    - DTO 参考: `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu_update.go`
    - Grab Proto 风格: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto`
  - 新增内容:
    - `AdvancedPricing` - 高级定价配置
    - `Purchasability` - 购买能力配置
    - `UpdateMenuItemReq` - 更新菜单项请求
    - `UpdateMenuItemResp` - 更新菜单项响应
    - `UpdateMenuModifierReq` - 更新修饰符请求
    - `UpdateMenuModifierResp` - 更新修饰符响应
  - Success: Proto 消息定义完成，字段与 DTO 对应

- [x] 1.2 更新 menu.proto 添加新 RPC 方法

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/menu/menu.proto`
  - Purpose: 在 MenuService 中添加 UpdateMenuItem 和 UpdateMenuModifier RPC
  - Requirements: 1.3, 2.3
  - Leverage: 现有 MenuService 定义
  - 新增内容:
    - `rpc UpdateMenuItem (UpdateMenuItemReq) returns (takeout.ApiResponse) {}`
    - `rpc UpdateMenuModifier (UpdateMenuModifierReq) returns (takeout.ApiResponse) {}`
  - Success: RPC 方法定义完成，返回 takeout.ApiResponse

- [x] 1.3 生成 Protobuf Go 代码

  - File: `ttpos-bmp/app/ttpos-takeout/api/menu/*.pb.go`
  - Purpose: 根据 Proto 定义生成 Go 代码
  - Requirements: 1.1-1.3, 2.1-2.3
  - Leverage: 现有 Makefile
  - Command: 
    ```bash
    cd ttpos-bmp/app/ttpos-takeout
    make proto
    ```
  - Success: 代码生成成功，无编译错误

---

## Phase 2: Controller 实现（0.25 天）

- [x] 2.1 实现 UpdateMenuItem Controller 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Purpose: 实现 UpdateMenuItem gRPC 方法
  - Requirements: 1.4, 1.5, 3.2, 3.3
  - Leverage:
    - 现有方法: `GetMenuSnapshot`, `SaveMenuSnapshot` 风格
    - Service: `service.GrabMenu().UpdateMenuItem()`
    - DTO: `internal/model/dto/grab/menu_update.go`
    - 常量: `internal/consts/response.go` (Task 0.1)
  - 实现要点:
    - 参数校验（merchant_id, item_id 必填）
    - 使用 `consts.CodeInvalidParam`、`consts.MsgMerchantIDEmpty` 等常量
    - Proto → DTO 转换（处理 optional 字段）
    - 调用 Service 层
    - DTO → Proto 响应转换
    - 使用 anypb.Any 包装响应
  - Prompt: 
    ```
    Role: Go Developer with GoFrame gRPC expertise
    Task: 实现 UpdateMenuItem Controller 方法
    Context: 
    - 参考现有 GetMenuSnapshot/SaveMenuSnapshot 实现风格
    - 调用 service.GrabMenu().UpdateMenuItem()
    - 使用 takeout.ApiResponse 统一响应
    - 使用 consts.CodeXxx 和 consts.MsgXxx 常量
    Restrictions: 
    - 遵循 .cursor/rules/go-bmp.mdc
    - Proto optional 字段需判断 nil
    - 禁止硬编码错误码字符串
    Success: 方法实现完成，使用常量，参数校验正确
    ```
  - Success: UpdateMenuItem 方法实现完成

- [x] 2.2 实现 UpdateMenuModifier Controller 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Purpose: 实现 UpdateMenuModifier gRPC 方法
  - Requirements: 2.4, 2.5, 3.2, 3.3
  - Leverage:
    - Task 2.1 的实现
    - Service: `service.GrabMenu().UpdateMenuModifier()`
  - 实现要点:
    - 参数校验（merchant_id, modifier_id, modifier_name 必填）
    - Proto → DTO 转换
    - 调用 Service 层
    - DTO → Proto 响应转换
  - Success: UpdateMenuModifier 方法实现完成

- [x] 2.3 添加必要的 import 声明

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu.go`
  - Purpose: 添加新方法所需的 import
  - Requirements: -
  - 新增 import:
    ```go
    "ttpos-bmp/app/ttpos-takeout/internal/consts"
    grabDto "ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab"
    ```
  - Success: 编译通过，无 import 错误

---

## Phase 3: 测试验证（可选，视时间）

- [ ] 3.1 编写 Controller 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/menu/menu_test.go`
  - Purpose: 测试 Controller 参数校验逻辑
  - Requirements: 所有验收标准
  - Leverage: 现有测试文件
  - 测试用例:
    - 测试 merchant_id 为空返回 4001
    - 测试 item_id 为空返回 4001
    - 测试 modifier_name 为空返回 4001
    - 测试 optional 字段处理
  - Success: 测试覆盖率 ≥ 70%

- [ ] 3.2 集成测试（grpcurl）

  - File: -
  - Purpose: 验证 gRPC 接口端到端功能
  - Requirements: 所有功能需求
  - 测试命令:
    ```bash
    # 测试 UpdateMenuItem
    grpcurl -plaintext -d '{
      "merchant_id": "test-merchant",
      "item_id": "test-item",
      "available_status": "UNAVAILABLE"
    }' localhost:9001 menu.MenuService/UpdateMenuItem
    
    # 测试 UpdateMenuModifier  
    grpcurl -plaintext -d '{
      "merchant_id": "test-merchant",
      "modifier_id": "test-modifier",
      "modifier_name": "Test Modifier",
      "is_free": true
    }' localhost:9001 menu.MenuService/UpdateMenuModifier
    ```
  - Success: 接口调用成功，响应格式正确

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] Proto 语法正确，生成代码无错误

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] UpdateMenuItem RPC 可正常调用
- [ ] UpdateMenuModifier RPC 可正常调用
- [ ] 参数校验正确返回 4001
- [ ] 服务错误正确返回 5001
- [ ] 成功响应正确返回 0

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-bmp.mdc`
- [ ] 使用 `takeout.ApiResponse` 统一响应
- [ ] 使用 `anypb.Any` 包装响应数据

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/task-bmp-grpc-menu-update-service/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/task-bmp-grpc-menu-update-service/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/task-bmp-grpc-menu-update-service/tasks.md
```

### 执行流程

1. **选择任务**: 从 Phase 1 开始，按顺序执行
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 参考 design.md 中的代码示例
4. **实现代码**: 按照规范实现功能
5. **运行检查**: `go fmt`, `go vet`, `make proto`
6. **标记完成**: 将 `[ ]` 改为 `[x]`
7. **提交代码**: Git commit

---

## 快速开始

### 执行顺序

```
Phase 0: 常量定义
  0.1 创建响应码常量文件
        ↓
Phase 1: Proto 定义
  1.1 添加消息类型 → 1.2 添加 RPC 方法 → 1.3 生成代码
        ↓
Phase 2: Controller 实现
  2.3 添加 import → 2.1 UpdateMenuItem → 2.2 UpdateMenuModifier
        ↓
Phase 3: 测试验证 (可选)
  3.1 单元测试 → 3.2 集成测试
```

### 一键执行脚本

```bash
# 1. 进入项目目录
cd ttpos-bmp/app/ttpos-takeout

# 2. 创建响应码常量文件（Task 0.1）
vim internal/consts/response.go

# 3. 编辑 Proto 文件（Task 1.1, 1.2）
vim manifest/protobuf/menu/menu.proto

# 4. 生成代码（Task 1.3）
make proto

# 5. 编辑 Controller（Task 2.1, 2.2, 2.3）
vim internal/controller/rpc/menu/menu.go

# 6. 格式化和检查
go fmt ./...
go vet ./...

# 7. 运行测试（可选）
go test ./internal/controller/rpc/menu/...

# 8. 启动服务测试
go run main.go
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-16  
**维护者**: AI Agent

