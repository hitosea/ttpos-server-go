# 简化菜单更新方法返回值 需求提案

> 本文档用于需求发起阶段，经团队评审后创建正式 Spec。

---

## 📋 提案信息

| 项目       | 内容     |
| ---------- | -------- |
| **提案人** | AI Agent |
| **日期**   | 2025-12-18 |
| **目标版本** | - |
| **状态**   | 已批准   |
| **关联任务** | - |
| **关联 Spec** | [task-bmp-simplify-menu-update-response](../../../shared/specs/archived/v2.12/task-bmp-simplify-menu-update-response/requirements.md) |

---

## 🎯 背景和动机

### 问题描述

当前 `UpdateMenuItem` 和 `UpdateMenuModifier` 方法返回 `*grabDto.UpdateMenuResult` 结构体，包含冗余信息：

```go
type UpdateMenuResult struct {
    Success      bool   `json:"success"`
    MerchantID   string `json:"merchant_id"`
    RecordID     string `json:"record_id"`
    RecordType   string `json:"record_type"`
    ErrorCode    string `json:"error_code,omitempty"`
    ErrorMessage string `json:"error_message,omitempty"`
}
```

问题：
1. **信息冗余**：`MerchantID`、`RecordID`、`RecordType` 调用方已知
2. **错误处理不一致**：同时通过 `Success` 字段和 `error` 返回值表示失败
3. **不符合 Go 惯例**：Go 标准是通过 `error` 返回值处理错误

### 业务价值

- 简化 API 设计，减少维护成本
- 统一错误处理模式
- 符合 Go 语言惯例

### 目标用户

- [x] 开发人员

---

## 💡 解决方案概述

### 方案描述

移除 `grabDto.UpdateMenuResult` 结构体，将两个方法的返回值简化为 `error`：

```go
// Before
func (s *sGrabMenu) UpdateMenuItem(ctx context.Context, req *grabDto.UpdateMenuItemReq) (*grabDto.UpdateMenuResult, error)
func (s *sGrabMenu) UpdateMenuModifier(ctx context.Context, req *grabDto.UpdateMenuModifierReq) (*grabDto.UpdateMenuResult, error)

// After
func (s *sGrabMenu) UpdateMenuItem(ctx context.Context, req *grabDto.UpdateMenuItemReq) error
func (s *sGrabMenu) UpdateMenuModifier(ctx context.Context, req *grabDto.UpdateMenuModifierReq) error
```

### 核心功能点

1. 移除 `grabDto.UpdateMenuResult` 结构体定义
2. 修改 `UpdateMenuItem` 返回值为 `error`
3. 修改 `UpdateMenuModifier` 返回值为 `error`
4. 更新 service interface 定义
5. 更新调用方代码

### 影响范围

**涉及模块**：
- [x] API 接口
- [x] 业务逻辑

**涉及文件**：
- `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
- `ttpos-bmp/app/ttpos-takeout/internal/model/dto/grab/menu_update.go`
- `ttpos-bmp/app/ttpos-takeout/internal/service/grab_menu.go` (interface)

---

## 📊 初步评估

### 技术复杂度

- [x] **低**：纯接口重构，无业务逻辑变更

### 工作量预估

- **预计天数**: 0.5 天
- **预估 SP**: 1

### 风险识别

**潜在风险**：
1. 需要同步更新调用方代码

**缓解措施**：
1. 全局搜索确认所有调用点

---

## 📝 附录

### 代码变更示例

**UpdateMenuItem 简化后**：

```go
func (s *sGrabMenu) UpdateMenuItem(ctx context.Context, req *grabDto.UpdateMenuItemReq) error {
    g.Log().Infof(ctx, "[Grab] UpdateMenuItem: merchantID=%s, itemID=%s", req.MerchantID, req.ItemID)

    // 1. 参数验证
    if err := g.Validator().Data(req).Run(ctx); err != nil {
        g.Log().Errorf(ctx, "[Grab] UpdateMenuItem validation failed: %v", err)
        return gerror.NewCode(gcode.CodeValidationFailed, err.Error())
    }

    // 2. 构建 SDK 请求
    updateItem := req.ToSDKUpdateMenuItem()
    updateReq := grabfood.UpdateMenuItemAsUpdateMenuRequest(updateItem)

    // 3. 调用 Grab API
    if err := service.Grab().UpdateMenuRecord(ctx, req.MerchantID, updateReq); err != nil {
        s.logMenuRecordUpdate(ctx, req.MerchantID, req.ItemID, grabDto.MenuItemUpdateFieldItem, false, err.Error())
        g.Log().Errorf(ctx, "[Grab] UpdateMenuItem API failed: merchantID=%s, itemID=%s, error=%v",
            req.MerchantID, req.ItemID, err)
        return gerror.Wrap(err, "调用 Grab UpdateMenuItem API 失败")
    }

    // 4. 记录成功日志
    s.logMenuRecordUpdate(ctx, req.MerchantID, req.ItemID, grabDto.MenuItemUpdateFieldItem, true, "")
    g.Log().Infof(ctx, "[Grab] UpdateMenuItem success: merchantID=%s, itemID=%s", req.MerchantID, req.ItemID)
    return nil
}
```

---

**版本**: v1.0.0  
**创建日期**: 2025-12-18

