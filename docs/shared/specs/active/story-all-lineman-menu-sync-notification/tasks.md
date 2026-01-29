# LINE MAN 菜单同步通知入库任务分解

> 本文档定义 LINE MAN 菜单同步通知入库功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 6  
**已完成**: 4  
**进行中**: -  
**完成率**: 67%

**预估 SP**: 2  
**技术栈**: Go BMP (ttpos-takeout)

---

## Phase 1: 核心实现（已完成）

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 更新 API 定义

  - File: `ttpos-bmp/app/ttpos-takeout/api/lineman/v1/menu.go`
  - Purpose: 定义 MenuSyncNotification 接口的请求和响应结构体
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5
  - Leverage: 
    - 现有定义: `TriggerSyncMenuReq/Res` - 路径参数和响应格式
    - 通用结构: `LinemanCommonResData` - 响应数据字段
  - Status: ✅ 已完成
  - Implementation:
    ```go
    // MenuSyncNotificationReq 菜单同步通知请求
    type MenuSyncNotificationReq struct {
        g.Meta            `path:"/partners/:partnerId/stores/:storeId/menus/notification" method:"post"`
        PartnerId         string `json:"partnerId" v:"required"`
        StoreId           string `json:"storeId" v:"required"`
        MenuSyncRequestId string `json:"menuSyncRequestId" v:"required"`
        UpdatedAt         string `json:"updatedAt" v:"required"`
        Status            string `json:"status" v:"required|in:SUCCESS,FAILED"`
        Error             string `json:"error"`
    }
    
    // MenuSyncNotificationRes 菜单同步通知响应
    type MenuSyncNotificationRes struct {
        g.Meta `mime:"application/json"`
        LinemanCommonResData
    }
    ```

---

- [x] 1.2 新增 MenuSyncTypeNotify 常量

  - File: `ttpos-bmp/app/ttpos-takeout/internal/consts/consts.go`
  - Purpose: 定义菜单同步通知类型常量，用于 menu_log 记录
  - Requirements: 2.2, 2.6
  - Leverage:
    - 现有常量: `MenuSyncTypeFull`, `MenuSyncTypeBatchUpdateItem` - 常量定义模式
  - Status: ✅ 已完成
  - Implementation:
    ```go
    const (
        // 已存在
        MenuSyncTypeFull                MenuSyncType = "FULL"
        MenuSyncTypeBatchUpdateItem     MenuSyncType = "BATCH_UPDATE_ITEM"
        MenuSyncTypeBatchUpdateModifier MenuSyncType = "BATCH_UPDATE_MODIFIER"
        
        // 新增
        MenuSyncTypeNotify              MenuSyncType = "NOTIFY"
    )
    ```

---

- [x] 1.3 实现 Logic 层

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_sync_notification.go`（新建）
  - Purpose: 实现业务编排逻辑，记录 menu_log
  - Requirements: 1.1, 2.1, 2.2, 2.3, 2.4, 2.5, 3.1, 3.2, 3.3, 3.4
  - Leverage:
    - 现有 Logic: `trigger_sync_menu.go` - Logic 结构和参数校验模式
    - Service 接口: `service.ChannelMenu().LogMenuSync()` - 记录日志
    - 常量: `consts.ProviderLineman`, `consts.MenuSyncTypeNotify`
  - Status: ✅ 已完成
  - Implementation:
    ```go
    // HandleMenuSyncNotification 处理菜单同步通知
    func (s *sLineman) HandleMenuSyncNotification(ctx context.Context, req *v1.MenuSyncNotificationReq) error {
        // 1. 参数校验
        if req.MenuSyncRequestId == "" {
            return gerror.NewCode(gcode.CodeInvalidParameter, "menuSyncRequestId 不能为空")
        }
        if req.Status == "" {
            return gerror.NewCode(gcode.CodeInvalidParameter, "status 不能为空")
        }
        if req.Status != "SUCCESS" && req.Status != "FAILED" {
            return gerror.NewCode(gcode.CodeInvalidParameter, "status 必须为 SUCCESS 或 FAILED")
        }
        
        // 2. 状态映射
        var menuLogStatus string
        if req.Status == "SUCCESS" {
            menuLogStatus = "SUCCESS"
        } else {
            menuLogStatus = "FAIL"
        }
        
        // 3. 构建错误信息
        var errorMsg string
        if req.Status == "FAILED" {
            errorMsg = req.Error
        }
        
        // 4. 调用 ChannelMenu Service 写入日志
        err := service.ChannelMenu().LogMenuSync(
            ctx,
            req.StoreId,
            string(consts.ProviderLineman),
            string(consts.MenuSyncTypeNotify),
            req.MenuSyncRequestId,
            menuLogStatus == "SUCCESS",
            "",
            errorMsg,
        )
        
        if err != nil {
            g.Log().Errorf(ctx, "[Lineman] 记录菜单同步通知失败: storeId=%s, requestId=%s, status=%s, error=%v",
                req.StoreId, req.MenuSyncRequestId, req.Status, err)
            return gerror.WrapCode(gcode.CodeInternalError, err, "记录菜单同步通知失败")
        }
        
        g.Log().Infof(ctx, "[Lineman] 菜单同步通知已记录: storeId=%s, requestId=%s, status=%s",
            req.StoreId, req.MenuSyncRequestId, req.Status)
        
        return nil
    }
    ```

---

- [x] 1.4 实现 Controller 层

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/lineman/lineman_v1_menu_sync_notification.go`
  - Purpose: 实现 MenuSyncNotification HTTP 接口，接收 Lineman 通知请求
  - Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7
  - Leverage:
    - 现有实现: `lineman_v1_trigger_sync_menu.go` - Controller 结构和错误处理模式
    - API 定义: `api/lineman/v1/menu.go` - Request/Response 结构体
  - Status: ✅ 已完成
  - Implementation:
    ```go
    func (c *ControllerV1) MenuSyncNotification(ctx context.Context, req *v1.MenuSyncNotificationReq) (res *v1.MenuSyncNotificationRes, err error) {
        // 1. 类型断言调用 Logic 层
        lineman, ok := service.Lineman().(interface {
            HandleMenuSyncNotification(context.Context, *v1.MenuSyncNotificationReq) error
        })
        if !ok {
            g.Log().Errorf(ctx, "[Lineman] HandleMenuSyncNotification 方法未实现")
            return &v1.MenuSyncNotificationRes{
                LinemanCommonResData: v1.LinemanCommonResData{
                    Status:  "fail",
                    Code:    "500",
                    Message: "系统内部错误",
                },
            }, nil
        }
        
        // 2. 调用 Logic 层
        err = lineman.HandleMenuSyncNotification(ctx, req)
        if err != nil {
            // 3. 错误码映射
            errCode := gerror.Code(err)
            switch errCode {
            case gcode.CodeInvalidParameter:
                return &v1.MenuSyncNotificationRes{
                    LinemanCommonResData: v1.LinemanCommonResData{
                        Status:  "fail",
                        Code:    "400",
                        Message: err.Error(),
                    },
                }, nil
            case gcode.CodeNotFound:
                return &v1.MenuSyncNotificationRes{
                    LinemanCommonResData: v1.LinemanCommonResData{
                        Status:  "fail",
                        Code:    "404",
                        Message: err.Error(),
                    },
                }, nil
            default:
                return &v1.MenuSyncNotificationRes{
                    LinemanCommonResData: v1.LinemanCommonResData{
                        Status:  "fail",
                        Code:    "500",
                        Message: "系统内部错误",
                    },
                }, nil
            }
        }
        
        // 4. 返回成功响应
        return &v1.MenuSyncNotificationRes{
            LinemanCommonResData: v1.LinemanCommonResData{
                Status:  "ok",
                Code:    "200",
                Message: "Menu sync notification received successfully",
            },
        }, nil
    }
    ```

---

## Phase 2: 测试验证（待执行）

- [ ] 2.1 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_sync_notification_test.go`（新建）
  - Purpose: 验证 Logic 层业务逻辑正确性
  - Requirements: 所有 requirements
  - Leverage:
    - 现有测试: `menu_sync_test.go` - 测试结构和 mock 模式
  - Prompt:
    ```
    Role: Go Developer specializing in GoFrame Testing
    
    Task: 编写 HandleMenuSyncNotification 单元测试
    
    Context:
    - File: ttpos-bmp/app/ttpos-takeout/internal/logic/lineman/menu_sync_notification_test.go
    - Leverage: menu_sync_test.go (同目录)
    - Requirements: requirements.md 所有需求
    - Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    
    Test Cases:
    1. TestHandleMenuSyncNotification_Success
       - menuSyncRequestId 必填校验
       - status 必填校验
       - status 枚举校验（SUCCESS/FAILED）
    
    2. TestHandleMenuSyncNotification_StatusMapping
       - SUCCESS → SUCCESS
       - FAILED → FAIL
    
    3. TestHandleMenuSyncNotification_ErrorMsg
       - status=SUCCESS 时，error_msg 为空
       - status=FAILED 时，error_msg 记录 error 字段
    
    4. TestHandleMenuSyncNotification_ServiceCallFailed
       - LogMenuSync 调用失败，返回 CodeInternalError
    
    Success Criteria:
    - 所有测试用例通过
    - 测试覆盖率 > 80%
    - 代码通过 go test ./...
    ```

---

- [ ] 2.2 集成测试

  - Tool: Postman / curl
  - Purpose: 验证端到端流程和 HTTP 响应
  - Requirements: 所有 requirements
  - Leverage:
    - 现有集成测试: TriggerSyncMenu 测试用例
  - Test Cases:
    
    **Test Case 1: 成功通知**
    ```bash
    curl -X POST "http://localhost:14031/v1/partners/partner123/stores/store123/menus/notification" \
      -H "Authorization: Bearer {token}" \
      -H "Content-Type: application/json" \
      -d '{
        "menuSyncRequestId": "req_123456",
        "updatedAt": "2022-11-01T13:08:06+07:00",
        "status": "SUCCESS",
        "error": ""
      }'
    
    Expected:
    HTTP 200
    {
      "status": "ok",
      "code": "200",
      "message": "Menu sync notification received successfully"
    }
    ```
    
    **Test Case 2: 失败通知**
    ```bash
    curl -X POST "http://localhost:14031/v1/partners/partner123/stores/store123/menus/notification" \
      -H "Authorization: Bearer {token}" \
      -H "Content-Type: application/json" \
      -d '{
        "menuSyncRequestId": "req_123456",
        "updatedAt": "2022-11-01T13:08:06+07:00",
        "status": "FAILED",
        "error": "Invalid menu format"
      }'
    
    Expected:
    HTTP 200
    {
      "status": "ok",
      "code": "200",
      "message": "Menu sync notification received successfully"
    }
    
    Database:
    menu_log 记录 status=FAIL, error_msg="Invalid menu format"
    ```
    
    **Test Case 3: 参数缺失**
    ```bash
    curl -X POST "http://localhost:14031/v1/partners/partner123/stores/store123/menus/notification" \
      -H "Authorization: Bearer {token}" \
      -H "Content-Type: application/json" \
      -d '{
        "updatedAt": "2022-11-01T13:08:06+07:00",
        "status": "SUCCESS"
      }'
    
    Expected:
    HTTP 200
    {
      "status": "fail",
      "code": "400",
      "message": "menuSyncRequestId 不能为空"
    }
    ```
    
    **Test Case 4: 状态非法**
    ```bash
    curl -X POST "http://localhost:14031/v1/partners/partner123/stores/store123/menus/notification" \
      -H "Authorization: Bearer {token}" \
      -H "Content-Type: application/json" \
      -d '{
        "menuSyncRequestId": "req_123456",
        "updatedAt": "2022-11-01T13:08:06+07:00",
        "status": "INVALID"
      }'
    
    Expected:
    HTTP 200
    {
      "status": "fail",
      "code": "400",
      "message": "status 必须为 SUCCESS 或 FAILED"
    }
    ```

---

## Phase 3: 部署上线（待执行）

- [ ] 3.1 代码审查

  - Purpose: 确保代码符合规范
  - Checklist:
    - [ ] 代码遵循 Go BMP 规范
    - [ ] 错误处理使用 gerror
    - [ ] 日志使用 g.Log()
    - [ ] 未修改自动生成文件（dao/, model/entity/, model/do/）
    - [ ] 中文注释完整
    - [ ] 通过 go fmt, go vet

---

- [ ] 3.2 部署验证

  - Purpose: 在测试环境验证部署
  - Steps:
    1. 部署到测试环境
    2. 运行集成测试
    3. 检查日志输出
    4. 检查 menu_log 表记录
  - Checklist:
    - [ ] 服务正常启动
    - [ ] 接口响应正常
    - [ ] menu_log 记录正确
    - [ ] 日志格式正确

---

## 📊 任务依赖关系

```
1.1 (API 定义) ──┐
                 ├──> 1.3 (Logic 层) ──> 1.4 (Controller 层) ──> 2.2 (集成测试)
1.2 (常量定义) ──┘                      ↓
                                     2.1 (单元测试)
                                        ↓
                                     3.1 (代码审查)
                                        ↓
                                     3.2 (部署验证)
```

---

## 📝 完成标准

### Phase 1 完成标准
- [x] 所有 Controller、Logic 实现完成
- [x] 常量定义完成
- [x] 代码通过 go fmt, go vet
- [x] 无 linter 错误

### Phase 2 完成标准
- [ ] 单元测试覆盖率 > 80%
- [ ] 所有集成测试通过
- [ ] menu_log 表记录正确

### Phase 3 完成标准
- [ ] 代码审查通过
- [ ] 测试环境验证通过
- [ ] 生产环境部署成功

---

## 🔄 实施进度

| Phase | 任务数 | 已完成 | 进度 |
|-------|--------|--------|------|
| Phase 1: 核心实现 | 4 | 4 | 100% ✅ |
| Phase 2: 测试验证 | 2 | 0 | 0% |
| Phase 3: 部署上线 | 2 | 0 | 0% |
| **总计** | **8** | **4** | **50%** |

---

## 📚 相关文档

- [requirements.md](./requirements.md) - 需求规格
- [design.md](./design.md) - 技术设计
- [Go BMP 开发规范](../../../../.cursor/rules/go-bmp.mdc)
- [Lineman API 协议](https://docs.google.com/spreadsheets/d/1CKRl7tRLtp6dCAcXQqWhPvS_0M378-vdKpucR6ZtNbg/edit?gid=571121603#gid=571121603)

---

**版本**: v1.0.0  
**创建日期**: 2026-01-15  
**最后更新**: 2026-01-15  
**当前状态**: Phase 1 已完成，Phase 2 待执行
