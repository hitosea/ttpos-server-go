# SaveMenuSnapshot 菜单快照保存 任务分解

> 本文档定义 SaveMenuSnapshot 功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 9  
**已完成**: 6  
**进行中**: -  
**完成率**: 67% (核心功能已完成，测试待补充)

---

## Phase 1: Proto 定义和代码生成

- [x] 1.1 修改 takeout.proto 添加 SaveMenuSnapshot 定义

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/takeout/takeout.proto`
  - Purpose: 定义 SaveMenuSnapshot gRPC 接口
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 现有 `GetMenuSnapshot` 定义
  - Changes:
    ```protobuf
    // 新增消息定义
    message SaveMenuSnapshotReq {
      string provider_name = 1; // 渠道名称: grab,lineman
      string shop_uuid = 2;     // 店铺 UUID
      string menu_data = 3;     // 菜单数据 JSON 字符串
      string request_id = 4;    // 请求 ID
    }

    message SaveMenuSnapshotResp {
      ResponseInfo responseInfo = 1;
    }

    // 在 TakeoutService 中新增:
    rpc SaveMenuSnapshot (SaveMenuSnapshotReq) returns (SaveMenuSnapshotResp) {}
    ```
  - Success: Proto 文件语法正确，字段定义完整

- [x] 1.2 执行代码生成

  - File: -
  - Purpose: 生成 Go gRPC 代码
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 现有 Makefile
  - Command: `cd ttpos-bmp/app/ttpos-takeout && make dao`
  - Success: `api/takeout/takeout.pb.go` 和 `takeout_grpc.pb.go` 更新成功

---

## Phase 2: Service 接口定义

- [x] 2.1 更新 Takeout Service 接口

  - File: `ttpos-bmp/app/ttpos-takeout/internal/service/takeout.go`
  - Purpose: 在 ITakeout 接口中添加 SaveMenuSnapshot 方法
  - Requirements: 1.3
  - Leverage: 现有 `GetMenuSnapshot` 接口定义
  - Changes:
    ```go
    type ITakeout interface {
        // 现有方法...
        SaveMenuSnapshot(ctx context.Context, req *takeout.SaveMenuSnapshotReq) (*takeout.SaveMenuSnapshotResp, error)
    }
    ```
  - Success: 接口定义完整，编译通过

---

## Phase 3: Logic 层实现

- [x] 3.1 实现 SaveMenuSnapshot 业务逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/takeout/takeout.go`
  - Purpose: 实现菜单快照保存核心逻辑
  - Requirements: 1.4, 1.5, 1.6
  - Leverage: 
    - 现有 `GetMenuSnapshot` 实现
    - `internal/dao/channel_menu_snapshot.go`
  - Key Logic:
    1. 参数校验（provider_name, shop_uuid, menu_data 必填）
    2. 保存/更新菜单快照到 channel_menu_snapshot 表
    3. 如果 provider_name == "grab"，异步调用 Grab 通知
  - Success: 菜单快照保存成功，异步触发 Grab 通知

- [x] 3.2 实现 Grab NotifyMenuUpdate 方法 (已存在于 grab.go)

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab.go`
  - Purpose: 调用 Grab Update Menu Notification API
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage:
    - 现有 OAuth 认证逻辑 `getAccessToken()`
    - 现有 `sdk_wrapper.go` HTTP 调用封装
    - `shop_provider_cfg` 表获取 merchantID
  - Key Logic:
    1. 根据 shop_uuid 获取 Grab merchantID
    2. 获取 OAuth access token
    3. 调用 `POST /partner/v1/merchant/menu/notification`
    4. 处理响应：成功、失败、409 锁冲突
  - Success: Grab API 调用成功，409 响应正常处理

---

## Phase 4: Controller 层实现

- [x] 4.1 实现 gRPC Controller 处理

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/takeout/takeout.go`
  - Purpose: 接收 gRPC 请求，调用 Logic 层
  - Requirements: 1.4
  - Leverage: 现有 `GetMenuSnapshot` Controller 实现
  - Changes:
    ```go
    func (c *cTakeout) SaveMenuSnapshot(ctx context.Context, req *takeout.SaveMenuSnapshotReq) (*takeout.SaveMenuSnapshotResp, error) {
        return service.Takeout().SaveMenuSnapshot(ctx, req)
    }
    ```
  - Success: gRPC 请求正确路由到 Logic 层

---

## Phase 5: 测试

- [ ] 5.1 编写 SaveMenuSnapshot 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/takeout/takeout_test.go`
  - Purpose: 测试 SaveMenuSnapshot 业务逻辑
  - Requirements: 所有功能需求
  - Leverage: 现有测试文件结构
  - Test Cases:
    1. 参数校验：provider_name 为空
    2. 参数校验：shop_uuid 为空
    3. 参数校验：menu_data 为空
    4. 正常保存：新建快照
    5. 正常保存：更新已存在快照
    6. Grab 渠道：触发通知
    7. 非 Grab 渠道：不触发通知
  - Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 5.2 编写 Grab NotifyMenuUpdate 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab/grab_test.go`
  - Purpose: 测试 Grab 菜单更新通知逻辑
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: Mock HTTP Client
  - Test Cases:
    1. 正常响应：200 成功
    2. 锁冲突响应：409
    3. 错误响应：500
    4. 超时处理
    5. merchantID 不存在
  - Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 5.3 集成测试验证

  - File: -
  - Purpose: 端到端验证 SaveMenuSnapshot 功能
  - Requirements: 所有需求
  - Test Steps:
    1. 启动 ttpos-takeout 服务
    2. 调用 SaveMenuSnapshot 保存菜单（provider_name=test）
    3. 调用 GetMenuSnapshot 验证数据已保存
    4. 调用 SaveMenuSnapshot 保存菜单（provider_name=grab）
    5. 验证 Grab API 被调用（查看日志）
  - Success: 所有测试场景通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标（Logic 层 ≥ 70%）
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成：
  - [x] SaveMenuSnapshot 能正确保存菜单快照
  - [x] provider_name/shop_uuid 为空时返回错误
  - [x] provider_name == "grab" 时触发 Grab API
  - [x] Grab API 失败不影响主流程

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 禁止修改 dao/entity/do/ 目录

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-takeout-save-menu-snapshot/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-save-menu-snapshot/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-save-menu-snapshot/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-takeout-save-menu-snapshot/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务（按 Phase 顺序）
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 参考 design.md 中的实现方案
4. **实现代码**: 按照规范实现功能
5. **运行检查**: `go fmt`, `go vet`, `go test`
6. **标记完成**: 将 `[ ]` 改为 `[x]`
7. **提交代码**: Git commit

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-11  
**负责人**: rikugun
